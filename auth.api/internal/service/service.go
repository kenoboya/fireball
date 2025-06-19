package service

import (
	"auth-api/internal/config"
	"auth-api/internal/model"
	repo "auth-api/internal/repository/mongo"
	"auth-api/internal/repository/mongo/cache"
	"auth-api/pkg/auth"
	"auth-api/pkg/broker"
	"auth-api/pkg/hash"
	"context"
	"fmt"
	"time"
)

//go:generate mockgen -source=service.go -destination=mocks/mock.go install go.uber.org/mock/mockgen@latest

type Services struct {
	Auth          Auth
	Notifications Notifications
}

func NewServices(devs *Deps) *Services {
	return &Services{
		Auth: NewAuthService(devs.repo.Users,
			devs.hasher, devs.tokenManager, devs.cacher, devs.oAuthConfig,
			devs.accessTokenTTL, devs.refreshTokenTTL),
		Notifications: NewNotificationService(devs.rabbitMQ),
	}
}

type Deps struct {
	repo            *repo.Repositories
	rabbitMQ        *broker.RabbitMQ
	hasher          hash.PasswordHasher
	tokenManager    auth.TokenManager
	cacher          cache.Cache
	oAuthConfig     config.OAuthConfig
	accessTokenTTL  time.Duration
	refreshTokenTTL time.Duration
}

func NewDeps(repo *repo.Repositories, rabbitMQ *broker.RabbitMQ, config *config.Config, cacher cache.Cache) (*Deps, error) {
	hasher := hash.NewSHA256Hasher(config.Auth.PasswordSalt)
	tokenManager, err := auth.NewManager(config.Auth.JWT.SecretAccessKey, config.Auth.JWT.SecretRefreshKey)
	if err != nil {
		return nil, fmt.Errorf("tokenManager: %w", err)
	}
	return &Deps{
		repo:            repo,
		rabbitMQ:        rabbitMQ,
		hasher:          hasher,
		tokenManager:    tokenManager,
		cacher:          cacher,
		oAuthConfig:     config.OAuth,
		accessTokenTTL:  config.Auth.JWT.AccessTokenTTL,
		refreshTokenTTL: config.Auth.JWT.RefreshTokenTTL,
	}, nil
}

//go:generate mockgen -source=service.go -destination=mocks/mock.go install go.uber.org/mock/mockgen@latest
type Auth interface {
	SignUp(ctx context.Context, userSignUp model.UserSignUp) (model.Tokens, model.User, error)
	SignIn(ctx context.Context, requestSignIn model.UserSignIn) (model.Tokens, model.User, error)
	Refresh(ctx context.Context, refreshToken string) (model.Tokens, error)
	VerifyToken(ctx context.Context, accessToken string) error
	VerifyCode(ctx context.Context, login string) (model.VerifyCodeInput, error)
	EntranceViaSocialMedia(ctx context.Context, request model.SocialMediaRequest) (model.Tokens, model.User, error)
	GetAccessTokenTTL() time.Duration
	GetRefreshTokenTTL() time.Duration
}

type Notifications interface {
	SendNotification(ctx context.Context, notRMQ model.NotificationRabbitMQ, data any) error
}
