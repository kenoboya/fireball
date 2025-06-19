package config

import (
	"auth-api/internal/model"
	"auth-api/pkg/broker"
	mongodb "auth-api/pkg/database/mongo"
	"auth-api/pkg/database/redis"
	"auth-api/pkg/logger"
	"os"
	"time"

	"github.com/kelseyhightower/envconfig"
	"github.com/spf13/viper"
	"github.com/subosito/gotenv"
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
	ClientID     string `env:"CLIENT_ID"`
	ClientSecret string `env:"CLIENT_SECRET"`
	RedirectURL  string `env:"REDIRECT_URL"`
}

type HttpConfig struct {
	Addr           string        `mapstructure:"port"`
	ReadTimeout    time.Duration `mapstructure:"readTimeout"`
	WriteTimeout   time.Duration `mapstructure:"writeTimeout"`
	MaxHeaderBytes int           `mapstructure:"maxHeaderBytes"`
}

type AuthConfig struct {
	JWT          JWTConfig
	PasswordSalt string
}

type JWTConfig struct {
	AccessTokenTTL   time.Duration `mapstructure:"accessTokenTTL"`
	RefreshTokenTTL  time.Duration `mapstructure:"refreshTokenTTL"`
	SecretAccessKey  string
	SecretRefreshKey string
}

func Init(configDIR, envDIR string) (*Config, error) {
	if err := loadViperConfig(configDIR); err != nil {
		return &Config{}, err
	}

	var cfg Config
	if err := unmarshal(&cfg); err != nil {
		return &Config{}, err
	}

	if err := loadFromEnv(&cfg, envDIR); err != nil {
		return &Config{}, err
	}

	return &cfg, nil
}
func unmarshal(config *Config) error {
	if err := viper.UnmarshalKey("http", &config.Http); err != nil {
		logger.Error("Failed to unmarshal config file",
			zap.String("prefix", "http"),
			zap.Error(err),
		)
		return err
	}

	if err := viper.UnmarshalKey("mongo", &config.Mongo); err != nil {
		logger.Error("Failed to unmarshal config file",
			zap.String("prefix", "mongo"),
			zap.Error(err),
		)
		return err
	}

	if err := viper.UnmarshalKey("auth", &config.Auth.JWT); err != nil {
		logger.Error("Failed to unmarshal config file",
			zap.String("prefix", "auth-jwt"),
			zap.Error(err),
		)
		return err
	}

	if err := viper.UnmarshalKey("cache", &config.Redis); err != nil {
		logger.Error("Failed to unmarshal config file",
			zap.String("prefix", "cache"),
			zap.Error(err),
		)
		return err
	}

	if err := viper.UnmarshalKey("rabbitmq", &config.RabbitMQ); err != nil {
		logger.Error("Failed to unmarshal config file",
			zap.String("prefix", "rabbitmq"),
			zap.Error(err),
		)
		return err
	}

	return nil
}

func loadFromEnv(cfg *Config, envDIR string) error {
	if err := gotenv.Load(envDIR); err != nil {
		logger.Error(
			zap.String("file", ".env"),
			zap.Error(model.ErrNotFoundEnvFile),
		)
		return model.ErrNotFoundEnvFile
	}

	if err := envconfig.Process("MONGO", &cfg.Mongo); err != nil {
		logger.Error("Failed to unmarshal environment file",
			zap.String("prefix", "MONGO"),
			zap.String("file", ".env"),
			zap.Error(err),
		)
		return err
	}

	if err := envconfig.Process("REDIS", &cfg.Redis); err != nil {
		logger.Error("Failed to unmarshal environment file",
			zap.String("prefix", "REDIS"),
			zap.String("file", ".env"),
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
			zap.String("file", ".env"),
			zap.Error(err),
		)
		return err
	}

	cfg.Auth.PasswordSalt = os.Getenv("PASSWORD_SALT")
	cfg.Auth.JWT.SecretAccessKey = os.Getenv("SECRET_ACCESS_KEY")
	cfg.Auth.JWT.SecretRefreshKey = os.Getenv("SECRET_REFRESH_KEY")

	return nil
}

func loadViperConfig(path string) error {
	viper.SetConfigName("server")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(path)

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			logger.Error(
				zap.String("file", "server.yaml"),
				zap.String("path", path),
				zap.Error(model.ErrNotFoundConfigFile),
			)
			return model.ErrNotFoundConfigFile
		} else {
			return err
		}
	}
	return viper.MergeInConfig()
}
