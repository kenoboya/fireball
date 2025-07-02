package config

import (
	"auth-api/pkg/broker"
	mongodb "auth-api/pkg/database/mongo"
	"auth-api/pkg/database/redis"
	"auth-api/pkg/logger"
	"os"
	"time"

	"github.com/kelseyhightower/envconfig"
	"go.uber.org/zap"
)

type Config struct {
	Http     HttpConfig
	RabbitMQ broker.RabbitMQConfig
	Redis    redis.RedisConfig
	Auth     AuthConfig
	Mongo    mongodb.MongoConfig
	OAuth    OAuthConfig
}

type OAuthConfig struct {
	Google   OAuth
	Github   OAuth
	Facebook OAuth
}

type OAuth struct {
	ClientID     string `envconfig:"CLIENT_ID"`
	ClientSecret string `envconfig:"CLIENT_SECRET"`
	RedirectURL  string `envconfig:"REDIRECT_URL"`
}

type HttpConfig struct {
	Addr           string        `envconfig:"PORT"`
	ReadTimeout    time.Duration `envconfig:"READ_TIME_OUT"`
	WriteTimeout   time.Duration `envconfig:"WRITE_TIME_OUT"`
	MaxHeaderBytes int           `envconfig:"MAX_HEADER_BYTES"`
}

type AuthConfig struct {
	JWT          JWTConfig
	PasswordSalt string `envconfig:"PASSWORD_SALT"`
}

type JWTConfig struct {
	AccessTokenTTL   time.Duration `envconfig:"ACCESS_TOKEN_TTL"`
	RefreshTokenTTL  time.Duration `envconfig:"REFRESH_TOKEN_TTL"`
	SecretAccessKey  string        `envconfig:"SECRET_ACCESS_KEY"`
	SecretRefreshKey string        `envconfig:"SECRET_REFRESH_KEY"`
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

	if err := envconfig.Process("AUTH", &cfg.Auth.JWT); err != nil {
		logger.Error("Failed to unmarshal environment file",
			zap.String("prefix", "AUTH"),
			zap.String("file", "config-app/.env"),
			zap.Error(err),
		)
		return err
	}

	if err := envconfig.Process("MONGO", &cfg.Mongo); err != nil {
		logger.Error("Failed to unmarshal environment file",
			zap.String("prefix", "MONGO"),
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

	if err := envconfig.Process("GOOGLE", &cfg.OAuth.Google); err != nil {
		logger.Error("Failed to unmarshal environment file",
			zap.String("prefix", "GOOGLE"),
			zap.String("file", ".env"),
			zap.Error(err),
		)
		return err
	}

	if err := envconfig.Process("FACEBOOK", &cfg.OAuth.Facebook); err != nil {
		logger.Error("Failed to unmarshal environment file",
			zap.String("prefix", "FACEBOOK"),
			zap.String("file", ".env"),
			zap.Error(err),
		)
		return err
	}

	if err := envconfig.Process("GITHUB", &cfg.OAuth.Github); err != nil {
		logger.Error("Failed to unmarshal environment file",
			zap.String("prefix", "GITHUB"),
			zap.String("file", ".env"),
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

	cfg.Auth.PasswordSalt = os.Getenv("PASSWORD_SALT")

	return nil
}
