package config

import (
	"chat-api/pkg/broker"
	"chat-api/pkg/db/psql"
	"chat-api/pkg/db/redis"
	"chat-api/pkg/logger"
	"os"
	"time"

	"github.com/kelseyhightower/envconfig"
	"go.uber.org/zap"
)

type Config struct {
	Redis     redis.RedisConfig
	Psql      psql.PSQlConfig
	RabbitMQ  broker.RabbitMQConfig
	Http      HttpConfig
	WebSocket WebSocketConfig
	Grpc      GrpcConfig
	Auth      AuthConfig
}

type AuthConfig struct {
	JWT         JWTConfig
	MessageSalt string `envconfig:"MESSAGE_SALT"`
}

type JWTConfig struct {
	SecretAccessKey string `envconfig:"SECRET_ACCESS_KEY"`
}

type WebSocketConfig struct {
	ReadBufferSize  int `envconfig:"READ_BUFFER_SIZE"`
	WriteBufferSize int `envconfig:"WRITE_BUFFER_SIZE"`
}

type HttpConfig struct {
	Addr           string        `envconfig:"PORT"`
	ReadTimeout    time.Duration `envconfig:"READ_TIME_OUT"`
	WriteTimeout   time.Duration `envconfig:"WRITE_TIME_OUT"`
	MaxHeaderBytes int           `envconfig:"MAX_HEADER_BYTES"`
}

type GrpcConfig struct {
	GrpcProfileConfig GrpcProfileConfig
}

type GrpcProfileConfig struct {
	Addr string `envconfig:"PORT"`
}

func Init() (*Config, error) {
	var cfg Config
	if err := loadFromEnv(&cfg); err != nil {
		return &Config{}, err
	}

	return &cfg, nil
}

func loadFromEnv(cfg *Config) error {
	if err := envconfig.Process("HTTP", &cfg.Http); err != nil {
		logger.Error("Failed to unmarshal environment file",
			zap.String("prefix", "HTTP"),
			zap.String("file", "config-app"),
			zap.Error(err),
		)
		return err
	}

	if err := envconfig.Process("WEBSOCKET", &cfg.WebSocket); err != nil {
		logger.Error("Failed to unmarshal environment file",
			zap.String("prefix", "WEBSOCKET"),
			zap.String("file", "config-app"),
			zap.Error(err),
		)
		return err
	}

	if err := envconfig.Process("PROFILE", &cfg.Grpc.GrpcProfileConfig); err != nil {
		logger.Error("Failed to unmarshal environment file",
			zap.String("prefix", "PROFILE"),
			zap.String("file", "config-app"),
			zap.Error(err),
		)
		return err
	}

	if err := envconfig.Process("POSTGRES", &cfg.Psql); err != nil {
		logger.Error("Failed to unmarshal environment file",
			zap.String("prefix", "POSTGRES"),
			zap.String("file", "config-app/.env"),
			zap.Error(err),
		)
		return err
	}

	if err := envconfig.Process("REDIS", &cfg.Redis); err != nil {
		logger.Error("Failed to unmarshal environment file",
			zap.String("prefix", "REDIS"),
			zap.String("file", "config-app/.env"),
			zap.Error(err),
		)
		return err
	}

	if err := envconfig.Process("RABBITMQ", &cfg.RabbitMQ); err != nil {
		logger.Error("Failed to unmarshal environment file",
			zap.String("prefix", "RABBITMQ"),
			zap.String("file", "config-app/.env"),
			zap.Error(err),
		)
		return err
	}

	cfg.Auth.MessageSalt = os.Getenv("MESSAGE_SALT")
	cfg.Auth.JWT.SecretAccessKey = os.Getenv("SECRET_ACCESS_KEY")
	return nil
}
