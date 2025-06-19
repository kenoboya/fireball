package model

import (
	"errors"
)

var (
	ErrInvalidContactData   = errors.New("adding yourself to contact")
	ErrInvalidUserData      = errors.New("invalid user data")
	ErrConfigFileNotFound   = errors.New("failed to find config file")
	ErrEnvFileNotFound      = errors.New("failed to load environment file")
	ErrProfileAlreadyExists = errors.New("profile already exists")
	ErrProfileNotFound      = errors.New("failed to find profile")
	ErrAliasAlreadyExists   = errors.New("alias already exists")
	ErrContactNotFound      = errors.New("failed to find contact")
	ErrAliasNotFound        = errors.New("failed to find alias")
	ErrUserNotFound         = errors.New("failed to find user")
)
