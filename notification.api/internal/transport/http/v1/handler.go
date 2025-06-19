package v1

import (
	"context"
	"notification-api/internal/service"
	"notification-api/pkg/logger"
	"sync"

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
		h.initNotificationRoutes(v1)
	}
}

func (h *Handler) StartConsumers(ctx context.Context) {
	var wg sync.WaitGroup

	logger.Info("Starting RabbitMQ consumers...")
	h.runConsumers(ctx, &wg)
	logger.Info("All consumers have been launched")

	<-ctx.Done()
	logger.Info("Shutdown signal received, stopping consumers...")

	wg.Wait()
	logger.Info("All consumers stopped gracefully")
}

func (h *Handler) runConsumers(ctx context.Context, wg *sync.WaitGroup) {
	consumers := []struct {
		name string
		fn   func(context.Context)
	}{
		{"VerifyCodeEmail", h.consumeVerifyCodeEmail},
		{"VerifyCodePhone", h.consumeVerifyCodePhone},
		{"SendMessage", h.consumeSendMessage},
		{"CreateChat", h.consumeCreateChat},
	}

	for _, consumer := range consumers {
		wg.Add(1)
		go func(name string, consumerFn func(context.Context)) {
			defer wg.Done()
			logger.Infof("Starting consumer: %s", name)
			consumerFn(ctx)
			logger.Infof("Consumer stopped: %s", name)
		}(consumer.name, consumer.fn)
	}
}
