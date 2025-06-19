package service

import (
	"context"
	"notification-api/internal/config"
	"notification-api/internal/model"
	repo "notification-api/internal/repository"
	"notification-api/pkg/auth"
	"notification-api/pkg/broker"
)

type Deps struct {
	emailConfig  config.EmailConfig
	twilioConfig config.TwilioConfig
	rabbitMQ     *broker.RabbitMQ
	repositories *repo.Repositories
	tkManager    auth.TokenManager
}

func NewDeps(emailConfig config.EmailConfig, twilioConfig config.TwilioConfig, rabbit *broker.RabbitMQ, repositories *repo.Repositories, tkManager auth.TokenManager) *Deps {
	return &Deps{
		emailConfig:  emailConfig,
		twilioConfig: twilioConfig,
		rabbitMQ:     rabbit,
		repositories: repositories,
		tkManager:    tkManager,
	}
}

type Services struct {
	Emails        Emails
	Phones        Phones
	Messages      Messages
	Chats         Chats
	Auth          Auth
	Notifications Notifications
	RabbitMQ      *broker.RabbitMQ
}

func NewServices(deps *Deps) *Services {
	return &Services{
		Emails:        NewEmailService(deps.emailConfig, deps.repositories.Verification),
		Phones:        NewPhoneService(deps.twilioConfig, deps.repositories.Verification),
		Messages:      NewMessagesService(deps.repositories.Users, deps.repositories.Messages, deps.repositories.Chats, deps.repositories.Notifications),
		Chats:         NewChatsService(deps.repositories.Users, deps.repositories.Messages, deps.repositories.Chats, deps.repositories.Notifications),
		Auth:          NewAuthService(deps.tkManager),
		Notifications: NewNotificationService(deps.repositories.Notifications, deps.repositories.Users),
		RabbitMQ:      deps.rabbitMQ,
	}
}

type Notifications interface {
	GetNotifications(ctx context.Context, userID string) (model.NotificationResponse, error)

	SetUserChatNotificationStatus(ctx context.Context, userNotification model.UserNotification) error
	GetUserMutedChat(ctx context.Context, userID string) ([]model.UserNotification, error)
}

type Auth interface {
	ValidateToken(token string) (string, error)
}

type Emails interface {
	SendVerifyCodeToEmail(ctx context.Context, vc model.VerifyCodeInput) error
}

type Phones interface {
	SendVerifyCodeToPhone(ctx context.Context, vc model.VerifyCodeInput) error
}

type Messages interface {
	SaveNotificationMessage(ctx context.Context, notificationMessage model.NotificationMessage) error
}

type Chats interface {
	SaveNotificationChat(ctx context.Context, notificationChat model.NotificationChat) error
}
