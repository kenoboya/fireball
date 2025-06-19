package model

import (
	"errors"
)

var (
	ErrNotFoundConfigFile           = errors.New("failed to find config file")
	ErrNotFoundEnvFile              = errors.New("failed to load environment file")
	ErrInvalidUserData              = errors.New("user data is invalid")
	ErrUserNotFound                 = errors.New("user not found")
	ErrUserBlocked                  = errors.New("user is blocked")
	ErrInvalidPassword              = errors.New("invalid password")
	ErrInvalidLogin                 = errors.New("invalid login")
	ErrLoginEmpty                   = errors.New("login cannot be empty")
	ErrUserAlreadyExists            = errors.New("user already exists")
	ErrFailedConvertID              = errors.New("failed to convert inserted ID to ObjectID")
	ErrFailedGetMetadataFromContext = errors.New("failed to get metadata from context")
	ErrTokenNotFound                = errors.New("failed to find token")
	ErrTokenIsEmpty                 = errors.New("token is empty")
	ErrTokenIsExpired               = errors.New("token is expired")
	ErrTokenNotHaveExpirationTime   = errors.New("token does not have an expiration time")
	ErrInvalidToken                 = errors.New("token is invalid")
	ErrVerifyCodeNotFound           = errors.New("verify code not found")
	ErrVerifyCodeGetError           = errors.New("error getting verify code from Redis")
	ErrVerifyCodeInvalid            = errors.New("invalid verify code")

	ErrUnknownProvider         = errors.New("unknown social provider")
	ErrFailedTokenExchange     = errors.New("failed to exchange token")
	ErrFailedGetLoginFromOAuth = errors.New("failed to get login from social provider")
	ErrMissingLoginData        = errors.New("email or username is required")
)
