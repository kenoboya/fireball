package api

import (
	"context"
	"os"
	"os/signal"
	"profile-api/internal/config"
	repo "profile-api/internal/repository"
	grpc_server "profile-api/internal/server/grpc"
	http_server "profile-api/internal/server/http"
	"profile-api/internal/service"
	grpc_handler "profile-api/internal/transport/grpc"
	http_handler "profile-api/internal/transport/http"
	"profile-api/pkg/auth"
	mongodb "profile-api/pkg/database/mongo"
	"profile-api/pkg/logger"
	"syscall"
	"time"

	"go.uber.org/zap"
)

func Run() {
	logger.InitLogger()
	cfg, err := config.Init()
	if err != nil {
		logger.Fatal("Failed to initialize config",
			zap.Error(err),
			zap.String("context", "Initializing application"),
			zap.String("version", "1.0.0"),
		)
	}

	mongoClient, err := mongodb.NewClient(cfg.Mongo)
	if err != nil {
		logger.Fatal("Failed to connect to mongo",
			zap.Error(err),
			zap.String("context", "Initializing application"),
		)
	}

	if err := mongoClient.Ping(context.Background(), nil); err != nil {
		logger.Fatal("Mongo ping failed", zap.Error(err))
	} else {
		logger.Info("Mongo is alive and responding")
	}

	db := mongoClient.Database(cfg.Mongo.Name)

	tokenManager, err := auth.NewManager(cfg.Auth.JWT.SecretAccessKey)
	if err != nil {
		logger.Fatal("Failed to create tokenManager",
			zap.Error(err),
		)
	}

	repositories := repo.NewRepositories(db)
	deps := service.NewDeps(repositories, *tokenManager)

	services := service.NewServices(deps)
	grpcHandler := grpc_handler.NewProfileHandler(services)
	httpHandler := http_handler.NewHandler(services)
	grpcServer := grpc_server.NewServer(cfg.Grpc, grpcHandler)
	httpServer := http_server.NewServer(cfg.Http, httpHandler)

	go func() {
		if err := grpcServer.Run(); err != nil {
			logger.Fatalf("The grpc server didn't start: %s\n", err)
		}
	}()

	logger.Info("Grpc server started")

	go func() {
		if err := httpServer.Run(); err != nil {
			logger.Fatalf("The http server didn't start: %s\n", err)
		}
	}()

	logger.Info("Http server started")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)

	<-quit

	grpcServer.Stop()

	const timeout = 5 * time.Second

	ctx, shutdown := context.WithTimeout(context.Background(), timeout)
	defer shutdown()

	if err := httpServer.ShutDown(ctx); err != nil {
		logger.Errorf("failed to shutdown http server: %v", err)
	}

	if err := mongoClient.Disconnect(ctx); err != nil {
		logger.Errorf("failed to stop mongo database: %v", err)
	}
}
