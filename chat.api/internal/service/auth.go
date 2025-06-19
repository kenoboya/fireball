package service

import (
	"chat-api/internal/model"
	"chat-api/internal/repository/cache"
	"chat-api/pkg/auth"
	"context"
	"math/rand"
	"strings"
	"time"
)

const (
	letterBytes   = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	letterIdxMax  = 63
	letterIdxMask = 0x3f //  Mask for 6 junior bits
	letterIdxBits = 6    // Number of bits for indexing (63 bits)
	optimalLength = 16
)

var src = rand.NewSource(time.Now().UnixNano())

type AuthService struct {
	tokenManeger auth.TokenManager
	cache        *cache.Cache
}

func NewAuthService(tkManager auth.TokenManager, cache *cache.Cache) *AuthService {
	return &AuthService{
		tokenManeger: tkManager,
		cache:        cache,
	}
}

func (s *AuthService) ValidateToken(token string) (string, error) {
	claims, err := s.tokenManeger.ParseToken(token)
	if err != nil {
		return "", err
	}
	return claims.UserID, err
}

func (s *AuthService) SetWebSocket(ctx context.Context, ws model.WebSocketConnection) error {
	socketID := randStringBytesMaskImprSrcSB(optimalLength)
	socketManager := model.NewWebSocketManagerWithRedis(s.cache.WebSocketCache.GetClient())
	socketManager.AddConnection(socketID, ws.Conn)
	return s.cache.WebSocketCache.SetWebSocket(ctx, model.WebSocket{UserID: ws.UserID, SocketID: socketID})
}

func (s *AuthService) GetWebSocket(ctx context.Context, userID string) (*model.WebSocketConnection, error) {
	socketID, err := s.cache.WebSocketCache.GetWebSocket(ctx, userID)
	if err != nil {
		return nil, err
	}
	socketManager := model.NewWebSocketManagerWithRedis(s.cache.WebSocketCache.GetClient())
	conn, exists := socketManager.GetConnection(socketID)
	if !exists {
		return nil, model.ErrWebSocketNotFound
	}
	return &model.WebSocketConnection{
		Conn:   conn,
		UserID: userID,
	}, nil
}

func (s *AuthService) UpdateWebSocket(ctx context.Context, userID string) error {
	return s.cache.WebSocketCache.UpdateWebSocketTTL(ctx, userID)
}

func (s *AuthService) DeleteWebSocket(ctx context.Context, userID string) error {
	socketID, err := s.cache.WebSocketCache.GetWebSocket(ctx, userID)
	if err != nil {
		return err
	}
	socketManager := model.NewWebSocketManagerWithRedis(s.cache.WebSocketCache.GetClient())
	socketManager.RemoveConnection(socketID)
	if err := s.cache.WebSocketCache.DeleteWebSocket(ctx, userID); err != nil {
		return err
	}
	return nil
}

func randStringBytesMaskImprSrcSB(n int) string {
	sb := strings.Builder{}
	sb.Grow(n)

	for i, cache, remain := n-1, src.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = src.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			sb.WriteByte(letterBytes[idx])
			i--
		}
		cache >>= letterIdxBits
		remain--
	}

	return sb.String()
}
