package model

import (
	"auth-api/pkg/logger"
	"fmt"
	"strings"
	"time"

	"go.uber.org/zap"
)

const (
	PROVIDER_GITHUB   = "github"
	PROVIDER_FACEBOOK = "facebook"
	PROVIDER_GOOGLE   = "google"
)

type SocialMedia struct {
	AccessToken string
	Type        string
	TokenExpiry time.Time
}

type SocialMediaRequest struct {
	Code     string
	Provider string
}

func (r *SocialMediaRequest) Validate() error {
	if r.Code == "" {
		logger.Error("Social code is empty", zap.Error(ErrInvalidUserData))
		return fmt.Errorf("%w: social code is empty", ErrInvalidUserData)
	}

	switch strings.ToUpper(r.Provider) {
	case PROVIDER_GOOGLE, PROVIDER_GITHUB, PROVIDER_FACEBOOK:
		logger.Infof("Provider of social media is: %s", r.Provider)
	default:
		logger.Error("Unknown social media provider", zap.String("provider", r.Provider), zap.Error(ErrInvalidUserData))
		return fmt.Errorf("%w: unknown social media provider: %s", ErrInvalidUserData, r.Provider)
	}

	return nil
}

type SocialMediaData struct {
	ClientID     string
	ClientSecret string
	Code         string
}
