package api

import (
	"context"
	"notification-api/internal/config"
	repo "notification-api/internal/repository"
	http_server "notification-api/internal/server/http"
	"notification-api/internal/service"
	handler "notification-api/internal/transport/http"
	"notification-api/pkg/auth"
	"notification-api/pkg/broker"
	mySQL "notification-api/pkg/db/MySQL"
	"notification-api/pkg/logger"
	"os"
	"os/signal"
	"syscall"

	_ "github.com/go-sql-driver/mysql"
	"go.uber.org/zap"
)

func Run(configDIR string, envDIR string) {
	logger.InitLogger()

	cfg, err := config.Init(configDIR, envDIR)
	if err != nil {
		logger.Fatal("Failed to initialize config",
			zap.Error(err),
			zap.String("context", "Initializing application"),
			zap.String("version", "1.0.0"),
		)
	}

	rabbitmq, err := broker.NewRabbitMQ(cfg.RabbitMQ)
	if err != nil {
		logger.Fatal("Failed to connect to rabbitMQ",
			zap.Error(err),
		)
	}

	if err := rabbitmq.InitializationOfChannels(); err != nil {
		logger.Fatal("Failed to initialize of channels in rabbitMQ",
			zap.Error(err),
		)
	}

	db, err := mySQL.MySQLConnection(cfg.MySQL)
	if err != nil {
		logger.Error(err)
	}
	defer db.Close()

	tkManager, err := auth.NewManager(cfg.JWT.SecretAccessKey)
	if err != nil {
		logger.Fatal("Failed to create new token manager",
			zap.Error(err),
		)
	}

	repositories := repo.NewRepositories(db)
	deps := service.NewDeps(cfg.Email, cfg.Twilio, rabbitmq, repositories, tkManager)
	services := service.NewServices(deps)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	h := handler.NewHandler(services)

	httpServer := http_server.NewServer(cfg.Http, h)

	go func() {
		if err := httpServer.Run(ctx); err != nil {
			logger.Errorf("HTTP server error: %s", err)
			cancel()
		}
	}()
	logger.Info("HTTP server started")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)

	<-quit
	logger.Info("Shutdown signal received...")

	cancel()

	if err := rabbitmq.CloseConnection(); err != nil {
		logger.Errorf("Failed to close connection to RabbitMQ: %v", err)
	}

	logger.Info("Server stopped gracefully")
}
