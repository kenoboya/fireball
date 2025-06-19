package config

import (
	"os"
	"profile-api/internal/model"
	mongodb "profile-api/pkg/database/mongo"
	"profile-api/pkg/logger"
	"time"

	"github.com/kelseyhightower/envconfig"
	"github.com/spf13/viper"
	"github.com/subosito/gotenv"
	"go.uber.org/zap"
)

type Config struct {
	Grpc  GrpcConfig
	Mongo mongodb.MongoConfig
	Http  HttpConfig
	Auth  AuthConfig
}

type HttpConfig struct {
	Addr           string        `mapstructure:"port"`
	ReadTimeout    time.Duration `mapstructure:"readTimeout"`
	WriteTimeout   time.Duration `mapstructure:"writeTimeout"`
	MaxHeaderBytes int           `mapstructure:"maxHeaderBytes"`
}

type AuthConfig struct {
	JWT JWTConfig
}

type JWTConfig struct {
	SecretAccessKey string
}

type GrpcConfig struct {
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

	if err := viper.UnmarshalKey("grpc", &config.Grpc); err != nil {
		logger.Error("Failed to unmarshal config file",
			zap.String("prefix", "grpc"),
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

	if err := envconfig.Process("MONGO", &cfg.Mongo); err != nil {
		logger.Error("Failed to unmarshal environment file",
			zap.String("prefix", "MONGO"),
			zap.String("file", ".env"),
			zap.Error(err),
		)
		return err
	}

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
