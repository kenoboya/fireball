package api

import (
	"chat-api/internal/config"
	"chat-api/internal/repository/cache"
	repo "chat-api/internal/repository/psql"
	grpc_profile_server "chat-api/internal/server/grpc/profile"
	http_server "chat-api/internal/server/http"
	"chat-api/internal/service"
	handler "chat-api/internal/transport/http"
	"chat-api/pkg/auth"
	"chat-api/pkg/broker"
	"chat-api/pkg/crypto"
	"chat-api/pkg/db/psql"
	"chat-api/pkg/db/redis"
	"chat-api/pkg/logger"
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/gorilla/websocket"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
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

	upgrader := websocket.Upgrader{
		ReadBufferSize:  cfg.WebSocket.ReadBufferSize,
		WriteBufferSize: cfg.WebSocket.WriteBufferSize,
		CheckOrigin:     func(r *http.Request) bool { return true }, // Для CORS, в проде — аккуратно
	}

	redisClient := redis.NewClient(cfg.Redis)
	if err := redisClient.Ping(context.Background()).Err(); err != nil {
		logger.Fatal("Failed to connect to redis",
			zap.Error(err),
		)
	}
	defer redisClient.Close()

	db := connectToDatabase(cfg)
	defer db.Close()

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

	cache := cache.NewCashe(redisClient, cfg.Redis.TTL)
	messangeCrypter, err := crypto.NewAESCipher(cfg.Auth.PasswordSalt)
	if err != nil {
		logger.Fatal("Failed to initialize to message crypter",
			zap.Error(err),
		)
	}

	repositories := repo.NewRepositories(db)
	tokenManager, err := auth.NewManager(cfg.Auth.JWT.SecretAccessKey)
	if err != nil {
		logger.Fatal("Failed to create tokenManager",
			zap.Error(err),
		)
	}
	deps := service.NewDeps(repositories, tokenManager, rabbitmq, messangeCrypter, cache)

	services := service.NewServices(deps)

	profileServer := grpc_profile_server.NewProfileServer(cfg.Grpc.GrpcProfileConfig)
	if err := profileServer.Run(); err != nil {
		logger.Warn("Failed to connect to profile server",
			zap.Error(err),
		)
	}

	logger.Info("profile server started")

	httpHandler := handler.NewHandler(services, upgrader, profileServer.ProfileClient)
	httpServer := http_server.NewServer(cfg.Http, httpHandler)

	if err := httpServer.Run(); err != nil {
		logger.Errorf("the server didn't start: %s\n", err)
	}

	logger.Info("http server started")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)

	<-quit

	if err := httpServer.ShutDown(context.Background()); err != nil {
		logger.Errorf("failed to shutdown http server: %v", err)
	}

	profileServer.Stop()

	if err := db.Close(); err != nil {
		logger.Errorf("failed to stop postgres: %v", err)
	}

	if err := redisClient.Close(); err != nil {
		logger.Errorf("failed to stop redis: %v", err)
	}
	if err := rabbitmq.CloseConnection(); err != nil {
		logger.Errorf("failed to close connection to rabbitMQ: %v", err)
	}
}

func connectToDatabase(cfg *config.Config) *sqlx.DB {
	db, err := psql.NewPostgresConnection(cfg.Psql)
	if err != nil {
		logger.Fatal(
			zap.String("package", "internal/api"),
			zap.String("file", "api.go"),
			zap.String("function", "connectToDatabase()"),
			zap.Error(err),
		)
	}
	return db
}
