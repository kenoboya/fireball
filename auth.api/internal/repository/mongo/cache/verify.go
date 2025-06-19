package cache

import (
	"auth-api/internal/model"
	"auth-api/pkg/logger"
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

type VerifyCodeCache interface {
	SetVerifyCode(ctx context.Context, vc model.VerifyCodeInput) error
	GetVerifyCode(ctx context.Context, login string) (string, error)
	DeleteVerifyCode(ctx context.Context, login string) error
	GetClient() *redis.Client
}

func (c *RedisCache) SetVerifyCode(ctx context.Context, vc model.VerifyCodeInput) error {
	_, err := c.client.Set(ctx, "verify_code:"+vc.Recipient, vc.Code, c.ttl).Result()
	if err != nil {
		logger.Error(
			"Error caching verify-code to Redis",
			zap.String("login", vc.Recipient),
			zap.Error(err),
		)
		return fmt.Errorf("error saving verify-code to Redis: %w", err)
	}

	logger.Info("Verify code successfully saved in Redis", zap.String("login", vc.Recipient))
	return nil
}

func (c *RedisCache) GetVerifyCode(ctx context.Context, login string) (string, error) {
	key := "verify_code:" + login
	code, err := c.client.Get(ctx, key).Result()

	if err == redis.Nil {
		logger.Info("Verify code not found in Redis", zap.String("login", login))
		return "", model.ErrVerifyCodeNotFound
	} else if err != nil {
		logger.Error("Error getting verify code from Redis", zap.String("login", login), zap.Error(err))
		return "", model.ErrVerifyCodeGetError
	}

	logger.Info("Successfully got verify code from Redis", zap.String("login", login))
	return code, nil
}

func (c *RedisCache) DeleteVerifyCode(ctx context.Context, login string) error {
	key := "verify_code:" + login
	_, err := c.client.Del(ctx, key).Result()
	if err != nil {
		logger.Error("Error deleting verify code from Redis", zap.String("login", login), zap.Error(err))
		return fmt.Errorf("error deleting verify code from Redis: %w", err)
	}

	logger.Info("Successfully deleted verify code from Redis", zap.String("login", login))
	return nil
}

func (c *RedisCache) GetClient() *redis.Client {
	return c.client
}
