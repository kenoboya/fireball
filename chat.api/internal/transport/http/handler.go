package handler

import (
	profile "chat-api/internal/server/grpc/profile/proto"
	"chat-api/internal/service"
	v1 "chat-api/internal/transport/http/v1"

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

func (h *Handler) Init() *gin.Engine {
	router := gin.Default()
	router.Use(
		gin.Recovery(),
		gin.Logger(),
	)
	h.initAPI(router)
	return router
}

func (h *Handler) initAPI(router *gin.Engine) {
	handlerV1 := v1.NewHandler(h.services, h.upgrader, h.profileClient)
	api := router.Group("/api")
	{
		handlerV1.Init(api)
	}
}
