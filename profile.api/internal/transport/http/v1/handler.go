package v1

import (
	"profile-api/internal/service"

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

func (h *Handler) Init(router *gin.RouterGroup) {
	v1 := router.Group("/v1")
	{
		h.initAuthRoutes(v1)
	}
}
