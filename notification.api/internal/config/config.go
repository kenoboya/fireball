package config

import (
	"notification-api/pkg/broker"
	mySQL "notification-api/pkg/db/MySQL"
	"notification-api/pkg/logger"
	"os"
	"time"

	"github.com/kelseyhightower/envconfig"
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
	Addr           string        `envconfig:"PORT"`
	ReadTimeout    time.Duration `envconfig:"READ_TIME_OUT"`
	WriteTimeout   time.Duration `envconfig:"WRITE_TIME_OUT"`
	MaxHeaderBytes int           `envconfig:"MAX_HEADER_BYTES"`
}

type EmailConfig struct {
	Email string `envconfig:"EMAIL"`
	Smtp  SmtpConfig
}

type SmtpConfig struct {
	Host     string `envconfig:"HOST"`
	Port     int    `envconfig:"PORT"`
	Username string `envconfig:"USERNAME"`
	Password string `envconfig:"PASSWORD"`
}

type TwilioConfig struct {
	AccountSID  string `envconfig:"ACCOUNT_SID" required:"true"`
	AuthToken   string `envconfig:"AUTH_TOKEN" required:"true"`
	PhoneNumber string `envconfig:"PHONE_NUMBER" required:"true"`
}

type JWTConfig struct {
	SecretAccessKey string `envconfig:"SECRET_ACCESS_KEY"`
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

	if err := envconfig.Process("MYSQL", &cfg.MySQL); err != nil {
		logger.Error("Failed to unmarshal environment file",
			zap.String("prefix", "MYSQL"),
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

	if err := envconfig.Process("EMAIL", &cfg.Email); err != nil {
		logger.Error("Failed to unmarshal environment file",
			zap.String("prefix", "EMAIL"),
			zap.String("file", "config-app/.env"),
			zap.Error(err),
		)
		return err
	}

	if err := envconfig.Process("SMTP", &cfg.Email.Smtp); err != nil {
		logger.Error("Failed to unmarshal environment file",
			zap.String("prefix", "SMTP"),
			zap.String("file", "config-app/.env"),
			zap.Error(err),
		)
		return err
	}

	if err := envconfig.Process("TWILIO", &cfg.Twilio); err != nil {
		logger.Error("Failed to unmarshal environment file",
			zap.String("prefix", "TWILIO"),
			zap.String("file", "config-app/.env"),
			zap.Error(err),
		)
		return err
	}

	cfg.JWT.SecretAccessKey = os.Getenv("SECRET_ACCESS_KEY")
	return nil
}
