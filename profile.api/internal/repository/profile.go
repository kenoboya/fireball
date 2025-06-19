package repo

import (
	"context"
	"profile-api/internal/model"
	mongodb "profile-api/pkg/database/mongo"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type ProfileRepository struct {
	collection *mongo.Collection
}

func NewProfileRepository(collection *mongo.Collection) *ProfileRepository {
	return &ProfileRepository{collection}
}

func (r *ProfileRepository) SetProfile(ctx context.Context, user model.User) error {
	_, err := r.collection.InsertOne(ctx, user)
	if err != nil {
		if mongodb.IsDuplicate(err) {
			return model.ErrProfileAlreadyExists
		}
		return err
	}

	return nil
}

func (r *ProfileRepository) SearchProfile(ctx context.Context, userSearchRequest model.UserSearchRequest) ([]model.UserBriefInfo, error) {
	limit := userSearchRequest.Limit
	if limit <= 0 || limit > 50 {
		limit = 10
	}

	filter := bson.M{
		"username": bson.M{
			"$regex":   "^" + userSearchRequest.Nickname,
			"$options": "i",
		},
	}

	opts := options.Find().SetLimit(limit)

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []model.UserBriefInfo
	for cursor.Next(ctx) {
		var user model.UserBriefInfo
		if err := cursor.Decode(&user); err != nil {
			return nil, err
		}
		results = append(results, user)
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return results, nil
}

func (r *ProfileRepository) GetByUserID(ctx context.Context, userID string) (model.User, error) {
	var user model.User
	err := r.collection.FindOne(ctx, bson.M{"_id": userID}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return user, model.ErrProfileNotFound
		}
		return user, err
	}

	return user, nil
}

func (r *ProfileRepository) GetUserBriefProfile(ctx context.Context, userID string) (model.UserBriefInfo, error) {
	var user model.UserBriefInfo

	err := r.collection.FindOne(ctx, bson.M{"_id": userID}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return model.UserBriefInfo{}, model.ErrUserNotFound
		}
		return model.UserBriefInfo{}, err
	}

	return user, nil
}

func (r *ProfileRepository) UpdateProfile(ctx context.Context, user model.User) error {
	_, err := r.collection.UpdateOne(
		ctx,
		bson.M{"_id": user.UserID},
		bson.M{"$set": user},
	)
	if err != nil {
		return err
	}

	return nil
}

func (r *ProfileRepository) DeleteProfile(ctx context.Context, userID string) error {
	_, err := r.collection.DeleteOne(
		ctx,
		bson.M{"_id": userID},
	)
	if err != nil {
		return err
	}

	return nil
}
