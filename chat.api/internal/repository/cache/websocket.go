package cache

import (
	"chat-api/internal/model"
	"chat-api/pkg/logger"
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

type WebSocketCache interface {
	SetWebSocket(ctx context.Context, ws model.WebSocket) error
	GetWebSocket(ctx context.Context, userID string) (string, error)
	UpdateWebSocketTTL(ctx context.Context, userID string) error
	DeleteWebSocket(ctx context.Context, userID string) error
	GetClient() *redis.Client
}

func (c *RedisCache) SetWebSocket(ctx context.Context, ws model.WebSocket) error {
	_, err := c.client.Set(ctx, "user_sockets:"+ws.UserID, ws.SocketID, c.ttl).Result()
	if err != nil {
		logger.Error(
			"Error caching websocket to Redis",
			zap.String("userID", ws.UserID),
			zap.String("socketID", ws.SocketID),
			zap.Error(err),
		)
		return fmt.Errorf("error saving websocket data to Redis: %w", err)
	}

	_, err = c.client.Set(ctx, "active_sockets:"+ws.SocketID, ws.UserID, c.ttl).Result()
	if err != nil {
		return fmt.Errorf("error saving websocket data to Redis: %w", err)
	}

	logger.Info("WebSocket successfully saved in Redis", zap.String("userID", ws.UserID))
	return nil
}

func (c *RedisCache) GetWebSocket(ctx context.Context, userID string) (string, error) {
	socketID, err := c.client.Get(ctx, "user_sockets:"+userID).Result()
	if err == redis.Nil {
		return "", model.ErrWebSocketNotFound
	}

	return socketID, nil
}

func (c *RedisCache) UpdateWebSocketTTL(ctx context.Context, userID string) error {
	socketID, err := c.GetWebSocket(ctx, userID)
	if err != nil {
		return err
	}

	_, err = c.client.Expire(ctx, "user_sockets:"+userID, c.ttl).Result()
	if err != nil {
		logger.Error(
			"Error updating TTL for websocket in Redis",
			zap.String("userID", userID),
			zap.Error(err),
		)
		return fmt.Errorf("error updating TTL for websocket data: %w", err)
	}

	_, err = c.client.Expire(ctx, "active_sockets:"+socketID, c.ttl).Result()
	if err != nil {
		return fmt.Errorf("error updating TTL for websocket data: %w", err)
	}

	logger.Info("WebSocket TTL updated successfully", zap.String("userID", userID))
	return nil
}

func (c *RedisCache) DeleteWebSocket(ctx context.Context, userID string) error {
	socketID, err := c.GetWebSocket(ctx, userID)
	if err != nil {
		return err
	}

	_, err = c.client.Del(ctx, "user_sockets:"+userID).Result()
	if err != nil {
		logger.Error(
			"Error deleting websocket from Redis",
			zap.String("userID", userID),
			zap.Error(err),
		)
	}

	_, err = c.client.Del(ctx, "active_sockets:"+socketID).Result()
	if err != nil {
		logger.Error(
			"Error deleting active socket from Redis",
			zap.String("socketID", socketID),
			zap.Error(err),
		)
	}

	logger.Info("WebSocket successfully deleted from Redis", zap.String("userID", userID))
	return nil
}

func (c *RedisCache) GetClient() *redis.Client {
	return c.client
}
