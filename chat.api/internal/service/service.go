package service

import (
	"chat-api/internal/model"
	"chat-api/internal/repository/cache"
	repo "chat-api/internal/repository/psql"
	"chat-api/pkg/auth"
	"chat-api/pkg/broker"
	"chat-api/pkg/crypto"
	"context"
)

type Services struct {
	Chats            Chats
	Messages         Messages
	Auth             Auth
	Notifications    Notifications
	MessageEncrypter crypto.MessageEncrypter
}

type Deps struct {
	repositories     *repo.Repositories
	tokenManager     auth.TokenManager
	rabbitMQ         *broker.RabbitMQ
	messageEncrypter crypto.MessageEncrypter
	cache            *cache.Cache
}

func NewServices(deps *Deps) *Services {
	return &Services{
		Chats:            NewChatService(deps.repositories.Messages, deps.repositories.Chats, deps.repositories.Media, deps.repositories.Files, deps.repositories.Locations, deps.repositories.Pinned),
		Messages:         NewMessageService(deps.repositories.Messages, deps.repositories.Files, deps.repositories.Media, deps.repositories.Locations),
		Auth:             NewAuthService(deps.tokenManager, deps.cache),
		Notifications:    NewNotificationService(deps.rabbitMQ),
		MessageEncrypter: deps.messageEncrypter,
	}
}

func NewDeps(repo *repo.Repositories, tkManager auth.TokenManager, rabbit *broker.RabbitMQ, messageEncrypter crypto.MessageEncrypter, cache *cache.Cache) *Deps {
	return &Deps{
		repositories:     repo,
		tokenManager:     tkManager,
		rabbitMQ:         rabbit,
		messageEncrypter: messageEncrypter,
		cache:            cache,
	}
}

type Auth interface {
	ValidateToken(token string) (string, error)
	SetWebSocket(ctx context.Context, ws model.WebSocketConnection) error
	GetWebSocket(ctx context.Context, userID string) (*model.WebSocketConnection, error)
	UpdateWebSocket(ctx context.Context, userID string) error
	DeleteWebSocket(ctx context.Context, userID string) error
}

type Chats interface {
	SetChatRole(ctx context.Context, chatRole model.ChatRole) error
	SetBlockChat(ctx context.Context, blockChat model.BlockChat) error
	CreatePrivateChat(ctx context.Context, request model.CreatePrivateChatRequest) (model.CreatePrivateChatResponse, error)
	CreateGroupChat(ctx context.Context, request *model.CreateGroupChatRequest) error
	GetParticipantsOfChat(ctx context.Context, chatID int64) ([]string, error)
	GetChatByChatID(ctx context.Context, chatID int64) (model.ChatDB, error)
	UpdatePinnedChat(ctx context.Context, pinnedChatWithFlag model.PinnedChatWithFlag) error
	InitializeChatsForMessenger(ctx context.Context, userID string) ([]model.Chat, error)
	InitializePinnedChatsForMessenger(ctx context.Context, userID string) ([]model.PinnedChatInit, error)
}

type Messages interface {
	SendMessage(ctx context.Context, createMessageRequest *model.CreateMessageRequest) error
}

type Notifications interface {
	SendNotification(ctx context.Context, notRMQ model.NotificationRabbitMQ, data any) error
}
