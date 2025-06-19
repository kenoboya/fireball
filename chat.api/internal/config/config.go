package config

import (
	"chat-api/internal/model"
	"chat-api/pkg/broker"
	"chat-api/pkg/db/psql"
	"chat-api/pkg/db/redis"
	"chat-api/pkg/logger"
	"os"
	"time"

	"github.com/kelseyhightower/envconfig"
	"github.com/spf13/viper"
	"github.com/subosito/gotenv"
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
	JWT          JWTConfig
	PasswordSalt string
}

type JWTConfig struct {
	SecretAccessKey string
}

type WebSocketConfig struct {
	ReadBufferSize  int `mapstructure:"readBufferSize"`
	WriteBufferSize int `mapstructure:"writeBufferSize"`
}

type HttpConfig struct {
	Addr           string        `mapstructure:"port"`
	ReadTimeout    time.Duration `mapstructure:"readTimeout"`
	WriteTimeout   time.Duration `mapstructure:"writeTimeout"`
	MaxHeaderBytes int           `mapstructure:"maxHeaderBytes"`
}

type GrpcConfig struct {
	GrpcProfileConfig GrpcProfileConfig
}

type GrpcProfileConfig struct {
	Addr string `mapstructure:"port"`
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

	if err := viper.UnmarshalKey("websocket", &config.WebSocket); err != nil {
		logger.Error("Failed to unmarshal config file",
			zap.String("prefix", "websocket"),
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

	if err := viper.UnmarshalKey("profile", &config.Grpc.GrpcProfileConfig); err != nil {
		logger.Error("Failed to unmarshal config file",
			zap.String("prefix", "profile"),
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
			zap.Error(model.ErrEnvFileNotFound),
		)
		return model.ErrEnvFileNotFound
	}

	if err := envconfig.Process("REDIS", &cfg.Redis); err != nil {
		logger.Error("Failed to unmarshal environment file",
			zap.String("prefix", "REDIS"),
			zap.String("file", ".env"),
			zap.Error(err),
		)
		return err
	}

	if err := envconfig.Process("DB", &cfg.Psql); err != nil {
		logger.Error("Failed to unmarshal environment file",
			zap.String("prefix", "DB"),
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
				zap.Error(model.ErrConfigFileNotFound),
			)
			return model.ErrConfigFileNotFound
		} else {
			return err
		}
	}
	return viper.MergeInConfig()
}
