package grpc_server

import (
	"net"
	"profile-api/internal/config"
	"profile-api/internal/server/grpc/proto"
	grpc_handler "profile-api/internal/transport/grpc"
	"profile-api/pkg/logger"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type server struct {
	srv     *grpc.Server
	addr    string
	handler *grpc_handler.ProfileHandler
}

func NewServer(config config.GrpcConfig, handler *grpc_handler.ProfileHandler) *server {
	return &server{
		srv:     grpc.NewServer(),
		addr:    config.Addr,
		handler: handler,
	}
}

func (s *server) Run() error {
	lis, err := net.Listen("tcp", s.addr)
	if err != nil {
		logger.Error("Failed to listen on TCP",
			zap.String("server", "grpc"),
			zap.Error(err),
		)
		return err
	}

	logger.Info("Starting gRPC server",
		zap.String("address", s.addr),
	)

	proto.RegisterProfileServiceServer(s.srv, s.handler)

	reflection.Register(s.srv)

	if err := s.srv.Serve(lis); err != nil {
		logger.Error("Failed to serve gRPC server",
			zap.String("server", "grpc"),
			zap.String("address", s.addr),
			zap.Error(err),
		)
		return err
	}

	return nil
}

func (s *server) Stop() {
	s.srv.GracefulStop()
	logger.Info("gRPC server has stopped gracefully",
		zap.String("server", "grpc"),
	)
}
