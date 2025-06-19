package model

type Tokens struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type OAuthUserData struct {
	Email    *string
	Username *string
	Phone    *string
}
