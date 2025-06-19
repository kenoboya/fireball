package config

import (
	"notification-api/internal/model"
	"notification-api/pkg/broker"
	mySQL "notification-api/pkg/db/MySQL"
	"notification-api/pkg/logger"
	"os"
	"time"

	"github.com/kelseyhightower/envconfig"
	"github.com/spf13/viper"
	"github.com/subosito/gotenv"
	"go.uber.org/zap"
)

type Config struct {
	MySQL    mySQL.MySQLConfig
	RabbitMQ broker.RabbitMQConfig
	Http     HttpConfig
	Email    EmailConfig
	JWT      JWTConfig
	Twilio   TwilioConfig
}

type HttpConfig struct {
	Addr           string        `mapstructure:"port"`
	ReadTimeout    time.Duration `mapstructure:"readTimeout"`
	WriteTimeout   time.Duration `mapstructure:"writeTimeout"`
	MaxHeaderBytes int           `mapstructure:"maxHeaderBytes"`
}

type EmailConfig struct {
	Email string
	Smtp  SmtpConfig
}

type SmtpConfig struct {
	Host     string
	Port     int
	Username string
	Password string
}

type TwilioConfig struct {
	AccountSID  string `envconfig:"ACCOUNT_SID" required:"true"`
	AuthToken   string `envconfig:"AUTH_TOKEN" required:"true"`
	PhoneNumber string `envconfig:"PHONE_NUMBER" required:"true"`
}

type JWTConfig struct {
	SecretAccessKey string
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
			zap.Error(model.ErrEnvFileNotFound),
		)
		return model.ErrEnvFileNotFound
	}

	if err := envconfig.Process("MYSQL", &cfg.MySQL); err != nil {
		logger.Error("Failed to unmarshal environment file",
			zap.String("prefix", "MYSQL"),
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

	if err := envconfig.Process("EMAIL", &cfg.Email); err != nil {
		logger.Error("Failed to unmarshal environment file",
			zap.String("prefix", "EMAIL"),
			zap.String("file", ".env"),
			zap.Error(err),
		)
		return err
	}

	if err := envconfig.Process("SMTP", &cfg.Email.Smtp); err != nil {
		logger.Error("Failed to unmarshal environment file",
			zap.String("prefix", "SMTP"),
			zap.String("file", ".env"),
			zap.Error(err),
		)
		return err
	}

	if err := envconfig.Process("TWILIO", &cfg.Twilio); err != nil {
		logger.Error("Failed to unmarshal environment file",
			zap.String("prefix", "TWILIO"),
			zap.String("file", ".env"),
			zap.Error(err),
		)
		return err
	}

	cfg.JWT.SecretAccessKey = os.Getenv("SECRET_ACCESS_KEY")
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
