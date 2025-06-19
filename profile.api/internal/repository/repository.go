package repo

import (
	"context"
	"profile-api/internal/model"
	"profile-api/pkg/logger"

	"go.mongodb.org/mongo-driver/v2/mongo"
)

type Repositories struct {
	Profile  Profile
	Contacts Contacts
}

func NewRepositories(db *mongo.Database) *Repositories {
	profileCollection := db.Collection(profileCollection)
	if err := ensureIndexes(context.Background(), profileCollection); err != nil {
		logger.Errorf("failed to create indexes for profile collection")
	}

	return &Repositories{
		Profile:  NewProfileRepository(profileCollection),
		Contacts: NewContactsRepository(db.Collection(aliasCollection)),
	}
}

type Profile interface {
	SetProfile(ctx context.Context, user model.User) error
	SearchProfile(ctx context.Context, userSearchRequest model.UserSearchRequest) ([]model.UserBriefInfo, error)
	GetUserBriefProfile(ctx context.Context, userID string) (model.UserBriefInfo, error)
	GetByUserID(ctx context.Context, userID string) (model.User, error)
	UpdateProfile(ctx context.Context, user model.User) error
	DeleteProfile(ctx context.Context, userID string) error
}

type Contacts interface {
	SetContact(ctx context.Context, contactRequest model.Contact) error
	GetContact(ctx context.Context, request model.UserRequest) (model.Contact, error)
	GetContacts(ctx context.Context, senderID string) ([]model.Contact, error)
	DeleteContact(ctx context.Context, request model.UserRequest) error

	GetAlias(ctx context.Context, request model.UserRequest) (string, error)
	UpdateAlias(ctx context.Context, contactRequest model.Contact) error
	DeleteAlias(ctx context.Context, request model.UserRequest) error
}
