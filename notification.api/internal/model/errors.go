package model

import "errors"

var (
	ErrInvalidUserData          = errors.New("invalid user data")
	ErrConfigFileNotFound       = errors.New("failed to find config file")
	ErrEnvFileNotFound          = errors.New("failed to load environment file")
	ErrUserNotificationNotFound = errors.New("user notification not found")
	ErrAlreadyRead              = errors.New("message already read")

	ErrChannelNotInitialized = errors.New("channel not initialized")
)
