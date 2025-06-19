package model

import (
	"errors"
)

var (
	ErrInvalidUserData        = errors.New("invalid user data")
	ErrConfigFileNotFound     = errors.New("failed to find config file")
	ErrEnvFileNotFound        = errors.New("failed to load environment file")
	ErrInvalidSaltKey         = errors.New("key must be 16, 24, or 32 bytes long")
	ErrInvalidParamsOfChat    = errors.New("params of chat is invalid")
	ErrInvalidParamsOfMessage = errors.New("params of message is invalid")
	ErrUploadFile             = errors.New("failed to upload file")
	ErrUploadLocation         = errors.New("failed to upload location")
	ErrFilesIsEmpty           = errors.New("files is empty or null")
	ErrMediaIsEmpty           = errors.New("media is empty or null")
	ErrLocationIsEmpty        = errors.New("location is empty or null")
	ErrFailedToEncryptMessage = errors.New("failed to encrypting message")

	ErrWebSocketNotFound                 = errors.New("websocket not found for the specified user")
	ErrConvertWebSocketCacheToRedisCache = errors.New("failed to convert interface web socket cache to type RedisCache")
)
