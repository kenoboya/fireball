package repo

import (
	"auth-api/internal/model"
	mongodb "auth-api/pkg/database/mongo"
	"context"
	"errors"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type UsersRepository struct {
	collection *mongo.Collection
}

func NewUsersRepository(collection *mongo.Collection) *UsersRepository {
	return &UsersRepository{collection}
}

func (r *UsersRepository) Create(ctx context.Context, user model.User) (bson.ObjectID, error) {
	orConditions := []bson.M{
		{"username": user.Username},
	}

	if user.Email != nil && *user.Email != "" {
		orConditions = append(orConditions, bson.M{"email": user.Email})
	}

	if user.Phone != nil && *user.Phone != "" {
		orConditions = append(orConditions, bson.M{"phone": user.Phone})
	}

	filter := bson.M{"$or": orConditions}

	var existingUser model.User
	err := r.collection.FindOne(ctx, filter).Decode(&existingUser)
	if err == nil {
		if existingUser.Username == user.Username {
			return bson.NilObjectID, fmt.Errorf("%w: username already exists", model.ErrUserAlreadyExists)
		}
		if user.Email != nil && existingUser.Email != nil && *existingUser.Email == *user.Email {
			return bson.NilObjectID, fmt.Errorf("%w: email already exists", model.ErrUserAlreadyExists)
		}
		if user.Phone != nil && existingUser.Phone != nil && *existingUser.Phone == *user.Phone {
			return bson.NilObjectID, fmt.Errorf("%w: phone already exists", model.ErrUserAlreadyExists)
		}
		return bson.NilObjectID, model.ErrUserAlreadyExists
	} else if err != mongo.ErrNoDocuments {
		return bson.NilObjectID, fmt.Errorf("failed to check user existence: %w", err)
	}

	user.RegisteredAt = time.Now()

	result, err := r.collection.InsertOne(ctx, user)
	if err != nil {
		if mongodb.IsDuplicate(err) {
			return bson.NilObjectID, model.ErrUserAlreadyExists
		}
		return bson.NilObjectID, fmt.Errorf("failed to insert user: %w", err)
	}

	oid, ok := result.InsertedID.(bson.ObjectID)
	if !ok {
		return bson.NilObjectID, model.ErrFailedConvertID
	}

	return oid, nil
}

func (r *UsersRepository) GetByLogin(ctx context.Context, login string) (model.User, error) {
	var user model.User

	filter := bson.M{
		"$or": []bson.M{
			{"username": login},
			{"email": login},
			{"phone": login},
		},
	}

	if err := r.collection.FindOne(ctx, filter).Decode(&user); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return model.User{}, model.ErrUserNotFound
		}
		return model.User{}, err
	}

	return user, nil
}
