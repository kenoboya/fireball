package http_server

import (
	"context"
	"net/http"
	"notification-api/internal/config"
	handler "notification-api/internal/transport/http"
	"time"
)

type server struct {
	server  *http.Server
	handler *handler.Handler
}

func NewServer(config config.HttpConfig, handler *handler.Handler) *server {
	return &server{
		handler: handler,
		server: &http.Server{
			Addr:           config.Addr,
			Handler:        handler.Init(),
			ReadTimeout:    config.ReadTimeout,
			WriteTimeout:   config.WriteTimeout,
			MaxHeaderBytes: config.MaxHeaderBytes,
		},
	}
}

func (s *server) Run(ctx context.Context) error {
	errCh := make(chan error, 1)

	go s.handler.HandlerV1.StartConsumers(ctx)

	go func() {
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return s.server.Shutdown(shutdownCtx)
	case err := <-errCh:
		return err
	}
}
