package service

import (
	"context"
	"profile-api/internal/model"
	repo "profile-api/internal/repository"
)

type ContactsService struct {
	repo repo.Contacts
}

func NewContactsService(repo repo.Contacts) *ContactsService {
	return &ContactsService{repo: repo}
}

func (s *ContactsService) SetContact(ctx context.Context, request model.Contact) error {
	if request.UserRequest.SenderID == request.UserRequest.RecipientID {
		return model.ErrInvalidContactData
	}
	return s.repo.SetContact(ctx, request)
}

func (s *ContactsService) GetAlias(ctx context.Context, request model.UserRequest) (string, error) {
	return s.repo.GetAlias(ctx, request)
}

func (s *ContactsService) UpdateAlias(ctx context.Context, request model.Contact) error {
	return s.repo.UpdateAlias(ctx, request)
}

func (s *ContactsService) DeleteAlias(ctx context.Context, request model.UserRequest) error {
	return s.repo.DeleteAlias(ctx, request)
}
