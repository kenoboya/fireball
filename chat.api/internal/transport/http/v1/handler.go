package v1

import (
	"chat-api/internal/service"

	profile "chat-api/internal/server/grpc/profile/proto"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

type Handler struct {
	services      *service.Services
	upgrader      websocket.Upgrader
	profileClient profile.ProfileServiceClient
}

func NewHandler(services *service.Services, upgrader websocket.Upgrader, profileClient profile.ProfileServiceClient) *Handler {
	return &Handler{
		services:      services,
		upgrader:      upgrader,
		profileClient: profileClient,
	}
}

func (h *Handler) Init(router *gin.RouterGroup) {
	v1 := router.Group("/v1")
	{
		h.initWebSocket(v1)
		h.initChatRoutes(v1)
	}
}
