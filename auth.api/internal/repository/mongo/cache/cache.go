package cache

import (
	"time"

	"github.com/redis/go-redis/v9"
)

type Cache struct {
	VerifyCodeCache VerifyCodeCache
}

type RedisCache struct {
	client *redis.Client
	ttl    time.Duration
}

func NewCashe(client *redis.Client, ttl time.Duration) *Cache {
	return &Cache{VerifyCodeCache: NewRedisCache(client, ttl)}
}

func NewRedisCache(client *redis.Client, ttl time.Duration) *RedisCache {
	return &RedisCache{
		client: client,
		ttl:    ttl,
	}
}
