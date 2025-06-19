package handler

import (
	"notification-api/internal/service"
	v1 "notification-api/internal/transport/http/v1"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	services  *service.Services
	HandlerV1 *v1.Handler
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
	h.HandlerV1 = v1.NewHandler(h.services)
	api := router.Group("/api")
	{
		h.HandlerV1.Init(api)
	}

}
