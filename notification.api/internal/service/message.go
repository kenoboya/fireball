package service

import (
	"context"
	"notification-api/internal/model"
	repo "notification-api/internal/repository"
)

type MessagesService struct {
	usersRepo        repo.Users
	messagesRepo     repo.Messages
	chatsRepo        repo.Chats
	notificationRepo repo.Notifications
}

func NewMessagesService(user repo.Users, message repo.Messages, chat repo.Chats, notification repo.Notifications) *MessagesService {
	return &MessagesService{
		usersRepo:        user,
		messagesRepo:     message,
		chatsRepo:        chat,
		notificationRepo: notification,
	}
}

func (s *MessagesService) SaveNotificationMessage(ctx context.Context, notificationMessage model.NotificationMessage) error {
	if err := s.usersRepo.SetUser(ctx, notificationMessage.Sender); err != nil {
		return err
	}
	internalMessageID, err := s.messagesRepo.SetMessage(ctx, notificationMessage.Message, notificationMessage.MessageAction)
	if err != nil {
		return err
	}
	if err := s.notificationRepo.SetMessage(ctx, internalMessageID, notificationMessage.RecipientID); err != nil {
		return err
	}
	internalChatID, err := s.chatsRepo.SetChat(ctx, notificationMessage.Chat, nil)
	if err != nil {
		return err
	}
	if err := s.chatsRepo.AddMessageToChat(ctx, internalChatID, internalMessageID); err != nil {
		return err
	}
	return nil
}
