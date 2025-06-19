package service

import (
	"auth-api/internal/config"
	"auth-api/internal/model"
	repo "auth-api/internal/repository/mongo"
	"auth-api/internal/repository/mongo/cache"
	"auth-api/pkg/auth"
	"auth-api/pkg/hash"
	"encoding/json"
	"net/http"

	"context"
	"crypto/rand"
	"database/sql"
	"errors"
	"math/big"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"golang.org/x/oauth2"
)

type AuthService struct {
	userRepo     repo.Users
	hasher       hash.PasswordHasher
	tokenManager auth.TokenManager
	cacher       cache.Cache
	oAuthConfig  config.OAuthConfig

	accessTokenTTL  time.Duration
	refreshTokenTTL time.Duration
}

type userSession struct {
	id   string
	role string
}

func (s *AuthService) GetAccessTokenTTL() time.Duration {
	return s.accessTokenTTL
}

func (s *AuthService) GetRefreshTokenTTL() time.Duration {
	return s.refreshTokenTTL
}

func NewAuthService(userRepo repo.Users,
	hasher hash.PasswordHasher, tokenManager auth.TokenManager, cacher cache.Cache, oAuthConfig config.OAuthConfig,
	accessTokenTTL time.Duration, refreshTokenTTL time.Duration) *AuthService {
	return &AuthService{
		userRepo:        userRepo,
		hasher:          hasher,
		tokenManager:    tokenManager,
		cacher:          cacher,
		oAuthConfig:     oAuthConfig,
		accessTokenTTL:  accessTokenTTL,
		refreshTokenTTL: refreshTokenTTL,
	}
}

func (s *AuthService) SignUp(ctx context.Context, userSignUp model.UserSignUp) (model.Tokens, model.User, error) {
	codeRedis, err := s.cacher.VerifyCodeCache.GetVerifyCode(ctx, userSignUp.VerifyCode.Recipient)
	if err != nil {
		return model.Tokens{}, model.User{}, err
	}

	if codeRedis != userSignUp.VerifyCode.Code {
		return model.Tokens{}, model.User{}, model.ErrVerifyCodeInvalid
	}

	passwordHash, err := s.hasher.Hash(userSignUp.User.Password)
	if err != nil {
		return model.Tokens{}, model.User{}, err
	}

	userSignUp.User.Blocked = model.UNBLOCKED
	userSignUp.User.RegisteredAt = time.Now()
	userSignUp.User.Password = passwordHash
	userSignUp.User.UserID = bson.NilObjectID

	userID, err := s.userRepo.Create(ctx, userSignUp.User)
	if err != nil {
		return model.Tokens{}, model.User{}, err
	}

	userSignUp.User.UserID = userID

	tokens, err := s.createSession(ctx, userSession{
		id:   userID.Hex(),
		role: model.USER_ROLE,
	})

	if err != nil {
		return model.Tokens{}, model.User{}, err
	}

	if err := s.cacher.VerifyCodeCache.DeleteVerifyCode(ctx, userSignUp.VerifyCode.Recipient); err != nil {
		return model.Tokens{}, model.User{}, err
	}

	userSignUp.User.Password = ""
	return tokens, userSignUp.User, nil
}

func (s *AuthService) SignIn(ctx context.Context, requestSignIn model.UserSignIn) (model.Tokens, model.User, error) {
	user, err := s.userRepo.GetByLogin(ctx, requestSignIn.Login)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.Tokens{}, model.User{}, model.ErrInvalidLogin
		}
		return model.Tokens{}, model.User{}, err
	}
	var errLogin error

	if user.Email == nil || *user.Email != requestSignIn.VerifyCode.Recipient {
		errLogin = model.ErrInvalidLogin
	}

	if errLogin != nil {
		if user.Phone == nil || *user.Phone != requestSignIn.VerifyCode.Recipient {
			return model.Tokens{}, model.User{}, errLogin
		}
		errLogin = nil // valid phone
	}

	codeRedis, err := s.cacher.VerifyCodeCache.GetVerifyCode(ctx, requestSignIn.VerifyCode.Recipient)
	if err != nil {
		return model.Tokens{}, model.User{}, err
	}

	if codeRedis != requestSignIn.VerifyCode.Code {
		return model.Tokens{}, model.User{}, model.ErrVerifyCodeInvalid
	}

	if user.IsBlocked() {
		return model.Tokens{}, model.User{}, model.ErrUserBlocked
	}

	if err := s.hasher.Compare(user.Password, requestSignIn.Password); err != nil {
		return model.Tokens{}, model.User{}, err
	}

	tokens, err := s.createSession(ctx, userSession{
		id:   user.UserID.Hex(),
		role: model.USER_ROLE,
	})

	if err != nil {
		return model.Tokens{}, model.User{}, err
	}

	if err := s.cacher.VerifyCodeCache.DeleteVerifyCode(ctx, requestSignIn.VerifyCode.Recipient); err != nil {
		return model.Tokens{}, model.User{}, err
	}

	user.Password = ""
	return tokens, user, nil
}

func (s *AuthService) Refresh(ctx context.Context, refreshToken string) (model.Tokens, error) {
	var (
		res model.Tokens
		err error
	)
	if res.RefreshToken, err = s.tokenManager.RefreshToken(refreshToken, s.refreshTokenTTL); err != nil {
		return model.Tokens{}, err
	}
	claims, err := s.tokenManager.ParseToken(res.RefreshToken, auth.RefreshToken)
	if err != nil {
		return model.Tokens{}, err
	}
	if res.AccessToken, err = s.tokenManager.NewJWT(claims.UserID, claims.Role, s.accessTokenTTL, auth.AccessToken); err != nil {
		return model.Tokens{}, err
	}
	return res, nil
}

func (s *AuthService) VerifyToken(ctx context.Context, accessToken string) error {
	return s.tokenManager.VerifyToken(accessToken, auth.AccessToken)
}

func (s *AuthService) VerifyCode(ctx context.Context, login string) (model.VerifyCodeInput, error) {
	var vc model.VerifyCodeInput
	var err error
	vc.Code, err = generateRandomString(model.LENGTH_CODE)
	if err != nil {
		return model.VerifyCodeInput{}, err
	}
	vc.Recipient = login

	if err = s.cacher.VerifyCodeCache.SetVerifyCode(ctx, vc); err != nil {
		return model.VerifyCodeInput{}, err
	}

	return vc, nil
}

func (s *AuthService) EntranceViaSocialMedia(ctx context.Context, request model.SocialMediaRequest) (model.Tokens, model.User, error) {
	var token *oauth2.Token
	var err error
	var oauthData *model.OAuthUserData

	switch request.Provider {
	case model.PROVIDER_GOOGLE:
		token, err = s.exchangeGoogleCode(ctx, request.Code)
		if err != nil {
			return model.Tokens{}, model.User{}, model.ErrFailedTokenExchange
		}
		oauthData, err = s.getGoogleUserData(ctx, token)

	case model.PROVIDER_GITHUB:
		token, err = s.exchangeGithubCode(ctx, request.Code)
		if err != nil {
			return model.Tokens{}, model.User{}, model.ErrFailedTokenExchange
		}
		oauthData, err = s.getGithubUserData(ctx, token)

	case model.PROVIDER_FACEBOOK:
		token, err = s.exchangeFacebookCode(ctx, request.Code)
		if err != nil {
			return model.Tokens{}, model.User{}, model.ErrFailedTokenExchange
		}
		oauthData, err = s.getFacebookUserData(ctx, token)

	default:
		return model.Tokens{}, model.User{}, model.ErrUnknownProvider
	}

	if err != nil {
		return model.Tokens{}, model.User{}, model.ErrFailedGetLoginFromOAuth
	}

	if oauthData == nil {
		return model.Tokens{}, model.User{}, model.ErrFailedGetLoginFromOAuth
	}

	// Determine login identifier (email or username or phone)
	login := firstNonNil(oauthData)
	if login == "" {
		return model.Tokens{}, model.User{}, model.ErrMissingLoginData
	}

	user, err := s.userRepo.GetByLogin(ctx, login)
	if err != nil && errors.Is(err, model.ErrUserNotFound) {
		// User not found — register new one
		// Generate username if nil
		username := oauthData.Username
		if username == nil {
			gen, err := generateRandomString(8)
			if err != nil {
				return model.Tokens{}, model.User{}, err
			}
			u := "user_" + gen
			username = &u
		}

		// Generate random password
		password, err := generateRandomString(12)
		if err != nil {
			return model.Tokens{}, model.User{}, err
		}

		hashedPassword, err := s.hasher.Hash(password)
		if err != nil {
			return model.Tokens{}, model.User{}, err
		}

		user = model.User{
			UserID:       bson.NewObjectID(),
			Username:     *username,
			Password:     hashedPassword,
			Email:        oauthData.Email,
			Phone:        oauthData.Phone,
			Blocked:      model.UNBLOCKED,
			RegisteredAt: time.Now(),
		}

		if _, err := s.userRepo.Create(ctx, user); err != nil {
			return model.Tokens{}, model.User{}, err
		}
	} else if err != nil {
		return model.Tokens{}, model.User{}, err
	}

	tokens, err := s.createSession(ctx, userSession{
		id:   user.UserID.Hex(),
		role: model.USER_ROLE,
	})
	if err != nil {
		return model.Tokens{}, model.User{}, err
	}

	return tokens, user, nil
}

func (s *AuthService) getGoogleUserData(ctx context.Context, token *oauth2.Token) (*model.OAuthUserData, error) {
	req, _ := http.NewRequestWithContext(ctx, "GET", "https://www.googleapis.com/oauth2/v2/userinfo", nil)
	req.Header.Set("Authorization", "Bearer "+token.AccessToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var data struct {
		Email string `json:"email"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}
	if data.Email == "" {
		return nil, model.ErrMissingLoginData
	}

	return &model.OAuthUserData{Email: &data.Email}, nil
}

func (s *AuthService) getGithubUserData(ctx context.Context, token *oauth2.Token) (*model.OAuthUserData, error) {
	req, _ := http.NewRequestWithContext(ctx, "GET", "https://api.github.com/user", nil)
	req.Header.Set("Authorization", "token "+token.AccessToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var userData struct {
		Login string `json:"login"`
		Email string `json:"email"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&userData); err != nil {
		return nil, err
	}

	// Fallback to verified emails
	if userData.Email == "" {
		reqEmails, _ := http.NewRequestWithContext(ctx, "GET", "https://api.github.com/user/emails", nil)
		reqEmails.Header.Set("Authorization", "token "+token.AccessToken)

		respEmails, err := http.DefaultClient.Do(reqEmails)
		if err != nil {
			return nil, err
		}
		defer respEmails.Body.Close()

		var emails []struct {
			Email    string `json:"email"`
			Primary  bool   `json:"primary"`
			Verified bool   `json:"verified"`
		}
		json.NewDecoder(respEmails.Body).Decode(&emails)
		for _, e := range emails {
			if e.Primary && e.Verified {
				userData.Email = e.Email
				break
			}
		}
		if userData.Email == "" && len(emails) > 0 {
			userData.Email = emails[0].Email
		}
	}

	result := &model.OAuthUserData{}
	if userData.Email != "" {
		result.Email = &userData.Email
	}
	if userData.Login != "" {
		result.Username = &userData.Login
	}
	if result.Email == nil && result.Username == nil {
		return nil, model.ErrMissingLoginData
	}
	return result, nil
}

func (s *AuthService) getFacebookUserData(ctx context.Context, token *oauth2.Token) (*model.OAuthUserData, error) {
	req, _ := http.NewRequestWithContext(ctx, "GET", "https://graph.facebook.com/me?fields=id,name,email", nil)
	req.Header.Set("Authorization", "Bearer "+token.AccessToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var data struct {
		ID    string `json:"id"`
		Email string `json:"email"`
		Name  string `json:"name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}

	if data.Email == "" && data.ID == "" {
		return nil, model.ErrMissingLoginData
	}

	result := &model.OAuthUserData{}
	if data.Email != "" {
		result.Email = &data.Email
	}
	if data.Name != "" {
		result.Username = &data.Name
	} else if data.ID != "" {
		result.Username = &data.ID
	}
	return result, nil
}

func (s *AuthService) exchangeGoogleCode(ctx context.Context, code string) (*oauth2.Token, error) {
	conf := &oauth2.Config{
		ClientID:     s.oAuthConfig.Google.ClientID,
		ClientSecret: s.oAuthConfig.Google.ClientSecret,
		RedirectURL:  s.oAuthConfig.Google.RedirectURL,
		Scopes:       []string{"email", "profile"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://accounts.google.com/o/oauth2/auth",
			TokenURL: "https://oauth2.googleapis.com/token",
		},
	}

	token, err := conf.Exchange(ctx, code)
	if err != nil {
		return nil, err
	}
	return token, nil
}

func (s *AuthService) exchangeFacebookCode(ctx context.Context, code string) (*oauth2.Token, error) {
	conf := &oauth2.Config{
		ClientID:     s.oAuthConfig.Facebook.ClientID,
		ClientSecret: s.oAuthConfig.Facebook.ClientSecret,
		RedirectURL:  s.oAuthConfig.Facebook.RedirectURL,
		Scopes:       []string{"email"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://www.facebook.com/v15.0/dialog/oauth",
			TokenURL: "https://graph.facebook.com/v15.0/oauth/access_token",
		},
	}

	token, err := conf.Exchange(ctx, code)
	if err != nil {
		return nil, err
	}
	return token, nil
}

func (s *AuthService) exchangeGithubCode(ctx context.Context, code string) (*oauth2.Token, error) {
	conf := &oauth2.Config{
		ClientID:     s.oAuthConfig.Github.ClientID,
		ClientSecret: s.oAuthConfig.Github.ClientSecret,
		RedirectURL:  s.oAuthConfig.Github.RedirectURL,
		Scopes:       []string{"user:email"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://github.com/login/oauth/authorize",
			TokenURL: "https://github.com/login/oauth/access_token",
		},
	}

	token, err := conf.Exchange(ctx, code)
	if err != nil {
		return nil, err
	}
	return token, nil
}

func generateRandomString(length int) (string, error) {
	const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, length)

	for i := 0; i < length; i++ {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(chars))))
		if err != nil {
			return "", err
		}
		result[i] = chars[num.Int64()]
	}
	return string(result), nil
}

func firstNonNil(data *model.OAuthUserData) string {
	if data.Email != nil && *data.Email != "" {
		return *data.Email
	}
	if data.Username != nil && *data.Username != "" {
		return *data.Username
	}
	if data.Phone != nil && *data.Phone != "" {
		return *data.Phone
	}
	return ""
}

func (s *AuthService) createSession(ctx context.Context, user userSession) (model.Tokens, error) {
	var (
		res model.Tokens
		err error
	)
	res.AccessToken, err = s.tokenManager.NewJWT(user.id, user.role, s.accessTokenTTL, auth.AccessToken)
	if err != nil {
		return model.Tokens{}, err
	}
	res.RefreshToken, err = s.tokenManager.NewJWT(user.id, user.role, s.refreshTokenTTL, auth.RefreshToken)
	if err != nil {
		return model.Tokens{}, err
	}
	return res, nil
}

func (s *AuthService) GetUserIDByToken(ctx context.Context, accessToken string) (string, error) {
	claims, err := s.tokenManager.ParseToken(accessToken, auth.AccessToken)
	if err != nil {
		return "", err
	}
	return claims.UserID, nil
}
