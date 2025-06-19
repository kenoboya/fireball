package service

import (
	"context"
	"notification-api/internal/model"
	repo "notification-api/internal/repository"
)

type NotificationService struct {
	notificationRepo repo.Notifications
	userRepo         repo.Users
}

func NewNotificationService(notification repo.Notifications, user repo.Users) *NotificationService {
	return &NotificationService{
		notificationRepo: notification,
		userRepo:         user,
	}
}

func (s *NotificationService) GetNotifications(ctx context.Context, userID string) (model.NotificationResponse, error) {
	var notificationResponse model.NotificationResponse
	var err error

	if notificationResponse.NotificationMessages, err = s.notificationRepo.GetMessagesForRecipient(ctx, userID); err != nil {
		return model.NotificationResponse{}, err
	}

	if notificationResponse.NotificationChat, err = s.notificationRepo.GetChatsForRecipient(ctx, userID); err != nil {
		return model.NotificationResponse{}, err
	}

	if notificationResponse.MutedChat, err = s.userRepo.GetUserMutedChat(ctx, userID); err != nil {
		return model.NotificationResponse{}, err
	}

	return notificationResponse, nil
}

func (s *NotificationService) SetUserChatNotificationStatus(ctx context.Context, userNotification model.UserNotification) error {
	return s.userRepo.SetUserNotification(ctx, userNotification)
}

func (s *NotificationService) GetUserMutedChat(ctx context.Context, userID string) ([]model.UserNotification, error) {
	return s.userRepo.GetUserNotifications(ctx, userID)
}
