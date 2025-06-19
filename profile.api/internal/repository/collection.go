package repo

import (
	"context"
	"profile-api/pkg/logger"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

const (
	profileCollection = "profiles"
	aliasCollection   = "contacts"
)

func ensureIndexes(ctx context.Context, collection *mongo.Collection) error {
	indexModel := mongo.IndexModel{
		Keys: bson.D{{Key: "username", Value: 1}},
	}

	_, err := collection.Indexes().CreateOne(ctx, indexModel)
	if err != nil {
		logger.Errorf("failed to create index")
		return err
	}

	return nil
}
