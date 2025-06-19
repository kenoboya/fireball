package mySQL

import (
	"fmt"
	"notification-api/pkg/logger"
	"time"

	"github.com/jmoiron/sqlx"
)

type MySQLConfig struct {
	Username string
	Password string
	Host     string
	Port     int
	Name     string
}

func (db MySQLConfig) getDatabaseConnectionString() string {
	return fmt.Sprintf("%s:%s@(%s:%d)/%s", db.Username, db.Password, db.Host, db.Port, db.Name)
}

func MySQLConnection(cfg MySQLConfig) (*sqlx.DB, error) {
	var db *sqlx.DB
	var err error
	maxRetries := 10
	delay := 5 * time.Second

	dsn := cfg.getDatabaseConnectionString()

	for i := 0; i < maxRetries; i++ {
		db, err = sqlx.Connect("mysql", dsn)
		if err == nil {
			logger.Info("Successfully connected to MySQL")
			return db, nil
		}

		logger.Infof("Failed to connect to MySQL: %v", err)
		logger.Infof("Retrying in %v seconds... (%d/%d)", delay.Seconds(), i+1, maxRetries)
		time.Sleep(delay)
	}

	return nil, err
}
