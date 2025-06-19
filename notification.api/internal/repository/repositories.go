package repo

import (
	"context"
	"notification-api/internal/model"

	"github.com/jmoiron/sqlx"
)

type Repositories struct {
	Verification  Verification
	Users         Users
	Messages      Messages
	Chats         Chats
	Notifications Notifications
}

func NewRepositories(db *sqlx.DB) *Repositories {
	return &Repositories{
		Verification:  NewVerificationRepository(db),
		Users:         NewUsersRepository(db),
		Messages:      NewMessagesRepository(db),
		Chats:         NewChatsRepository(db),
		Notifications: NewNotificationsRepository(db),
	}
}

type Users interface {
	SetUser(ctx context.Context, uBriefInfo model.UserBriefInfo) error
	SetUserNotification(ctx context.Context, userNotification model.UserNotification) error
	GetUserNotifications(ctx context.Context, userID string) ([]model.UserNotification, error)
	UpdateUserNotification(ctx context.Context, userNotification model.UserNotification) error
	UpdateUserNotificationForChat(ctx context.Context, userNotification model.UserNotification) error
	DeleteUserNotification(ctx context.Context, userID string) error

	GetUserMutedChat(ctx context.Context, userID string) ([]model.ChatBriefInfo, error)
}

type Messages interface {
	SetMessage(ctx context.Context, mBriefInfo model.MessageBriefInfo, action string) (int64, error)
	GetMessage(ctx context.Context, externalMessageID int64) (model.MessageBriefInfo, string, error)
	UpdateMessage(ctx context.Context, mBriefInfo model.MessageBriefInfo, action string) error
	DeleteMessage(ctx context.Context, externalMessageID int64) error
}

type Chats interface {
	AddMessageToChat(ctx context.Context, internalChatID, internalMessageID int64) error
	SetChat(ctx context.Context, cBriefInfo model.ChatBriefInfo, action *string) (int64, error)
	GetChat(ctx context.Context, externalChatID int64) (model.ChatBriefInfo, string, error)
	UpdateChat(ctx context.Context, cBriefInfo model.ChatBriefInfo, action *string) error
	DeleteChat(ctx context.Context, externalChatID int64) error
}

type Notifications interface {
	SetMessage(ctx context.Context, internalMessageID int64, recipientID string) error
	GetMessages(ctx context.Context, internalMessageID int64) ([]model.NotificationMessage, error)
	GetMessage(ctx context.Context, internalMessageID int64, recipientID string) (model.NotificationMessage, error)
	GetMessagesForRecipient(ctx context.Context, recipientID string) ([]model.NotificationMessage, error)

	SetChat(ctx context.Context, internalChatID int64, recipientID string) error
	GetChats(ctx context.Context, internalChatID int64) ([]model.NotificationChat, error)
	GetChat(ctx context.Context, internalChatID int64, recipientID string) (model.NotificationChat, error)
	GetChatsForRecipient(ctx context.Context, recipientID string) ([]model.NotificationChat, error)
}

type Verification interface {
	SetRecordVerificationLog(ctx context.Context, vc model.VerifyCodeInput, method string) error
}
