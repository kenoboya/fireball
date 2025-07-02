package mongodb

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

const timeout = time.Minute

type MongoConfig struct {
	Host     string `envconfig:"HOST"`
	Port     int    `envconfig:"PORT"`
	Username string `envconfig:"USERNAME"`
	Password string `envconfig:"PASSWORD"`
	Name     string `envconfig:"NAME"`
}

func (config *MongoConfig) createURI() string {
	if config.Username != "" && config.Password != "" {
		return fmt.Sprintf("mongodb://%s:%s@%s:%d/", config.Username, config.Password, config.Host, config.Port)
	}
	return fmt.Sprintf("mongodb://%s:%d/", config.Host, config.Port)
}

func NewClient(config MongoConfig) (*mongo.Client, error) {
	opts := options.Client().ApplyURI(config.createURI())
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	opts.SetMinPoolSize(5)
	opts.SetMaxPoolSize(50)
	opts.SetMaxConnIdleTime(5 * time.Minute)

	client, err := mongo.Connect(opts)
	if err != nil {
		return nil, err
	}

	err = client.Ping(ctx, nil)
	if err != nil {
		return nil, err
	}
	return client, nil
}

func IsDuplicate(err error) bool {
	var e mongo.WriteException
	if errors.As(err, &e) {
		for _, we := range e.WriteErrors {
			if we.Code == 11000 {
				return true
			}
		}
	}
	return false
}
