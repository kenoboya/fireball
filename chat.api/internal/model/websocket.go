package model

import (
	"chat-api/pkg/logger"
	"context"
	"log"
	"strings"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
)

var (
	instanceWebSocketManager *WebSocketManager
	once                     sync.Once
)

type WebSocket struct {
	SocketID string
	UserID   string
}

type WebSocketConnection struct {
	Conn   *websocket.Conn
	UserID string
}

type WebSocketManager struct {
	sockets map[string]*websocket.Conn
	mu      sync.RWMutex
	client  *redis.Client
	ctx     context.Context
}

func NewWebSocketManagerWithRedis(client *redis.Client) *WebSocketManager {
	once.Do(func() {
		manager := &WebSocketManager{
			sockets: make(map[string]*websocket.Conn),
			client:  client,
			ctx:     context.Background(),
		}
		instanceWebSocketManager = manager
		go manager.subscribeToExpiredKeys()
	})
	return instanceWebSocketManager
}

func (m *WebSocketManager) AddConnection(socketID string, conn *websocket.Conn) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.sockets[socketID] = conn
}

func (m *WebSocketManager) RemoveConnection(socketID string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if conn, exists := m.sockets[socketID]; exists {
		conn.Close()
		delete(m.sockets, socketID)
	}
}

func (m *WebSocketManager) GetConnection(socketID string) (*websocket.Conn, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	conn, exists := m.sockets[socketID]
	return conn, exists
}

func (m *WebSocketManager) ClearAllConnections() {
	m.mu.Lock()
	defer m.mu.Unlock()
	for socketID, conn := range m.sockets {
		conn.Close()
		delete(m.sockets, socketID)
	}
	log.Println("All WebSocket connections have been cleared")
}

// Redis Keyspace Events
func (m *WebSocketManager) subscribeToExpiredKeys() {
	// Redis: CONFIG SET notify-keyspace-events Ex
	pubsub := m.client.PSubscribe(m.ctx, "__keyevent@0__:expired")
	log.Println("Subscribed to Redis key expiration events")

	ch := pubsub.Channel()

	for msg := range ch {
		key := msg.Payload
		if strings.HasPrefix(key, "active_sockets:") {
			socketID := strings.TrimPrefix(key, "active_sockets:")
			logger.Infof("Redis key expired for socket %s, removing connection...\n", socketID)
			m.RemoveConnection(socketID)
		}
	}
}
