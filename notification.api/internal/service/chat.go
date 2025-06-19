package service

import (
	"context"
	"notification-api/internal/model"
	repo "notification-api/internal/repository"
)

type ChatsService struct {
	usersRepo        repo.Users
	messagesRepo     repo.Messages
	chatsRepo        repo.Chats
	notificationRepo repo.Notifications
}

func NewChatsService(user repo.Users, message repo.Messages, chat repo.Chats, notification repo.Notifications) *ChatsService {
	return &ChatsService{
		usersRepo:        user,
		messagesRepo:     message,
		chatsRepo:        chat,
		notificationRepo: notification,
	}
}

func (s *ChatsService) SaveNotificationChat(ctx context.Context, notificationChat model.NotificationChat) error {
	if err := s.usersRepo.SetUser(ctx, notificationChat.Sender); err != nil {
		return err
	}
	internalChatID, err := s.chatsRepo.SetChat(ctx, notificationChat.Chat, &notificationChat.ChatAction)
	if err != nil {
		return err
	}
	if err := s.notificationRepo.SetChat(ctx, internalChatID, notificationChat.RecipientID); err != nil {
		return err
	}
	return nil
}
