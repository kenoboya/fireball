package redis

import (
	"chat-api/pkg/logger"
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisConfig struct {
	Host     string        `envconfig:"HOST"`
	Port     int           `envconfig:"PORT"`
	Password string        `envconfig:"PASSWORD"`
	DB       int           `envconfig:"DB"`
	TTL      time.Duration `envconfig:"TTL"`
}

func NewClient(config RedisConfig) *redis.Client {
	addr := fmt.Sprintf("%s:%d", config.Host, config.Port)

	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: config.Password,
		DB:       config.DB,
	})

	_, err := client.ConfigSet(context.Background(), "notify-keyspace-events", "Ex").Result()
	if err != nil {
		logger.Fatalf("Error enabling keyspace notifications: %v", err)
	}

	return client
}
