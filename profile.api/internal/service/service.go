package service

import (
	"context"
	"profile-api/internal/model"
	repo "profile-api/internal/repository"
	"profile-api/pkg/auth"
)

type Services struct {
	Profiles Profiles
	Contacts Contacts
	Auth     Auth
}

type Deps struct {
	repo      *repo.Repositories
	tkManager auth.Manager
}

func NewServices(deps *Deps) *Services {
	return &Services{
		Profiles: NewProfileService(deps.repo.Profile, deps.repo.Contacts),
		Contacts: NewContactsService(deps.repo.Contacts),
		Auth:     NewAuthService(&deps.tkManager),
	}
}

func NewDeps(repo *repo.Repositories, tkManager auth.Manager) *Deps {
	return &Deps{
		repo:      repo,
		tkManager: tkManager,
	}
}

type Profiles interface {
	SetProfile(ctx context.Context, user model.User) error
	SearchProfile(ctx context.Context, userSearchRequest model.UserSearchRequest) ([]model.UserBriefInfo, error)
	GetByUserID(ctx context.Context, userID string) (model.User, error)
	GetUserBriefProfile(ctx context.Context, request model.UserRequest) (model.UserBriefInfo, error)
	GetUserBriefProfileForNotification(ctx context.Context, request model.UserRequest) (model.UserBriefInfo, error)
	GetUserProfiles(ctx context.Context, senderID string, recipientIDs []string) ([]model.User, error)
	UpdateProfile(ctx context.Context, user model.User) error
	DeleteProfile(ctx context.Context, userID string) error
	GetContacts(ctx context.Context, senderID string) ([]model.User, error)
	GetContact(ctx context.Context, request model.UserRequest) (model.User, error)
}

type Contacts interface {
	SetContact(ctx context.Context, request model.Contact) error
	GetAlias(ctx context.Context, request model.UserRequest) (string, error)
	UpdateAlias(ctx context.Context, request model.Contact) error
	DeleteAlias(ctx context.Context, request model.UserRequest) error
}

type Auth interface {
	ValidateToken(token string) (string, error)
}
