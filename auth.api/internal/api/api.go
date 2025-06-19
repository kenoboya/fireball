package api

import (
	"auth-api/internal/config"
	repo "auth-api/internal/repository/mongo"
	"auth-api/internal/repository/mongo/cache"
	http_server "auth-api/internal/server/http"
	"auth-api/internal/service"
	handler "auth-api/internal/transport/http"
	"auth-api/pkg/broker"
	mongodb "auth-api/pkg/database/mongo"
	"auth-api/pkg/database/redis"
	"auth-api/pkg/logger"
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"
)

func Run(configDIR, envDIR string) {
	logger.InitLogger()
	cfg, err := config.Init(configDIR, envDIR)
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

	redisClient := redis.NewClient(cfg.Redis)
	if err := redisClient.Ping(context.Background()).Err(); err != nil {
		logger.Fatal("Failed to connect to redis",
			zap.Error(err),
		)
	}
	defer redisClient.Close()

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

	cacher := cache.NewCashe(redisClient, cfg.Redis.TTL)

	repositories := repo.NewRepositories(db)
	deps, err := service.NewDeps(repositories, rabbitmq, cfg, *cacher)
	if err != nil {
		logger.Error(err)
	}

	services := service.NewServices(deps)
	handler := handler.NewHandler(services)
	httpServer := http_server.NewServer(cfg.Http, handler)

	if err := httpServer.Run(); err != nil {
		logger.Errorf("the server didn't start: %s\n", err)
	}
	logger.Info("Http server started")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)

	<-quit

	const timeout = 5 * time.Second

	ctx, shutdown := context.WithTimeout(context.Background(), timeout)
	defer shutdown()

	if err := httpServer.ShutDown(ctx); err != nil {
		logger.Errorf("failed to shutdown http server: %v", err)
	}

	if err := mongoClient.Disconnect(ctx); err != nil {
		logger.Errorf("failed to stop mongo database: %v", err)
	}

	if err := redisClient.Close(); err != nil {
		logger.Errorf("failed to stop redis: %v", err)
	}

	if err := rabbitmq.CloseConnection(); err != nil {
		logger.Errorf("failed to close connection to rabbitMQ: %v", err)
	}
}
