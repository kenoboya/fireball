package v1

import (
	"chat-api/internal/model"
	"chat-api/pkg/logger"
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

const (
	WEBSOCKET_TYPE_CREATE_PRIVATE_CHAT = "create private chat"
	WEBSOCKET_TYPE_CREATE_GROUP_CHAT   = "create group chat"
	WEBSOCKET_TYPE_SEND_MESSAGE        = "send message"

	WEBSOCKET_TYPE_MESSAGE_ACTION = "action on message"
	WEBSOCKET_TYPE_CHAT_ACTION    = "action on chat"
)

type WSMessage struct {
	Type string `json:"type"`
}

func (h *Handler) initWebSocket(router *gin.RouterGroup) {
	ws := router.Group("/ws")
	{
		ws.GET("", h.InitializeWebSocket)
	}
}

func (h *Handler) InitializeWebSocket(c *gin.Context) {
	userID := h.extractUserIDFromToken(c)
	if userID == "" {
		logger.Error("User ID is empty, unauthorized access")
		newResponse(c, http.StatusUnauthorized, "Unauthorized: token is missing or invalid")
		return
	}

	ws, err := h.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		logger.Warn("Failed to update websocket",
			zap.Error(err),
		)
		return
	}
	defer ws.Close()

	if err = h.services.Auth.SetWebSocket(c.Request.Context(), model.WebSocketConnection{
		Conn:   ws,
		UserID: userID,
	}); err != nil {
		logger.Warn("Failed to set websocket",
			zap.Error(err),
		)
		return
	}

	for {
		_, msg, err := ws.ReadMessage()
		if err != nil {
			logger.Warn("Failed to read from websocket", zap.Error(err))
			if authErr := h.services.Auth.DeleteWebSocket(c.Request.Context(), userID); authErr != nil {
				logger.Warn("Failed to remove websocket connection", zap.Error(authErr))
			}
			break
		}

		var wsMsg WSMessage
		if err := json.Unmarshal(msg, &wsMsg); err != nil {
			logger.Warn("Failed to parse JSON", zap.Error(err))
			continue
		}

		switch wsMsg.Type {
		case WEBSOCKET_TYPE_SEND_MESSAGE:
			var request model.CreateMessageRequest
			if err := json.Unmarshal(msg, &request); err != nil {
				logger.Warn("Failed to parse JSON", zap.Error(err))
				continue
			}
			h.sendMessage(ws, request)
		case WEBSOCKET_TYPE_CREATE_PRIVATE_CHAT:
			var request model.CreatePrivateChatRequest
			if err := json.Unmarshal(msg, &request); err != nil {
				logger.Warn("Failed to parse JSON", zap.Error(err))
				continue
			}
			h.createPrivateChat(ws, request)
		case WEBSOCKET_TYPE_CREATE_GROUP_CHAT:
			var request model.CreateGroupChatRequest
			if err := json.Unmarshal(msg, &request); err != nil {
				logger.Warn("Failed to parse JSON", zap.Error(err))
				continue
			}
			h.createGroupChat(ws, request)
		// case WEBSOCKET_TYPE_MESSAGE_ACTION:
		// 	var
		// case WEBSOCKET_TYPE_CHAT_ACTION:
		default:
			logger.Warn("Unknown WebSocket message type", zap.String("type", wsMsg.Type))
		}
	}
}

func (h *Handler) extractUserIDFromToken(c *gin.Context) string {
	token := c.Query("token")

	if token == "" {
		cookieToken, err := c.Cookie("access_token")
		if err != nil || cookieToken == "" {
			newResponse(c, http.StatusUnauthorized, "Authorization token is missing")
			return ""
		}
		token = cookieToken
	}
	userID, err := h.services.Auth.ValidateToken(token)
	if err != nil {
		newResponse(c, http.StatusUnauthorized, "Invalid token")
		return ""
	}

	return userID
}
