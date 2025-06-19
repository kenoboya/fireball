package psql

import (
	"chat-api/pkg/logger"
	"fmt"

	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

type PSQlConfig struct {
	Host     string
	Port     int
	Username string
	Name     string
	SSLmode  string
	Password string
}

func (db *PSQlConfig) getDatabaseConnectionString() string {
	return fmt.Sprintf("host=%s port=%d user=%s dbname=%s sslmode=%s password=%s",
		db.Host, db.Port, db.Username, db.Name, db.SSLmode, db.Password)
}

func NewPostgresConnection(cfg PSQlConfig) (*sqlx.DB, error) {
	db, err := sqlx.Connect("postgres", cfg.getDatabaseConnectionString())
	if err != nil {
		logger.Error("Failed to connect to PostgreSQL with the provided configuration",
			zap.String("host", cfg.Host),
			zap.Int("port", cfg.Port),
			zap.String("user", cfg.Username),
			zap.String("dbname", cfg.Name),
			zap.String("sslmode", cfg.SSLmode),
			zap.Error(err),
		)
		return nil, err
	}
	return db, nil
}
