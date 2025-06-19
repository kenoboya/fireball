package http_server

import (
	"chat-api/internal/config"
	handler "chat-api/internal/transport/http"
	"context"
	"net/http"
)

type server struct {
	server *http.Server
}

func NewServer(config config.HttpConfig, handler *handler.Handler) *server {
	return &server{
		server: &http.Server{
			Addr:           config.Addr,
			Handler:        handler.Init(),
			ReadTimeout:    config.ReadTimeout,
			WriteTimeout:   config.WriteTimeout,
			MaxHeaderBytes: config.MaxHeaderBytes,
		},
	}
}
func (s *server) Run() error {
	if err := s.server.ListenAndServe(); err != nil {
		return err
	}
	return nil
}

func (s *server) ShutDown(ctx context.Context) error {
	if err := s.server.Shutdown(ctx); err != nil {
		return err
	}
	return nil
}
