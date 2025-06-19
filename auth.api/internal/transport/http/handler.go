package handler

import (
	"auth-api/internal/service"
	v1 "auth-api/internal/transport/http/v1"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	services *service.Services
}

func NewHandler(services *service.Services) *Handler {
	return &Handler{
		services: services,
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
	handlerV1 := v1.NewHandler(h.services)
	api := router.Group("/api")
	{
		handlerV1.Init(api)
	}
}
