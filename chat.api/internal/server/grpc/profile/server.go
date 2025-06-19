package grpc_profile_server

import (
	"chat-api/internal/config"
	"chat-api/pkg/logger"

	profile "chat-api/internal/server/grpc/profile/proto"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type profileServer struct {
	srv           *grpc.Server
	addr          string
	connection    *grpc.ClientConn
	ProfileClient profile.ProfileServiceClient
}

func NewProfileServer(config config.GrpcProfileConfig) *profileServer {
	return &profileServer{
		srv:  grpc.NewServer(),
		addr: config.Addr,
	}
}

func (s *profileServer) Run() error {
	var err error
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}

	s.connection, err = grpc.NewClient(s.addr, opts...)
	if err != nil {
		logger.Error("Failed to connect to profile server",
			zap.String("server", "profile"),
			zap.Error(err),
		)
		return err
	}

	s.ProfileClient = profile.NewProfileServiceClient(s.connection)
	return nil
}

func (s *profileServer) Stop() {
	if s.connection != nil {
		if err := s.connection.Close(); err != nil {
			logger.Error("Failed to close connection",
				zap.Error(err),
			)
		} else {
			logger.Info("Connection closed successfully")
		}
	} else {
		logger.Warn("No active connection to close")
	}

	logger.Info("ProfileServer stopped gracefully")
}
