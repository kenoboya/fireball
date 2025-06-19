package repo

import (
	"context"
	"profile-api/internal/model"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type ContactsRepository struct {
	collection *mongo.Collection
}

func NewContactsRepository(collection *mongo.Collection) *ContactsRepository {
	return &ContactsRepository{collection: collection}
}

func (r *ContactsRepository) SetContact(ctx context.Context, contactRequest model.Contact) error {
	filter := bson.M{
		"sender":    contactRequest.UserRequest.SenderID,
		"recipient": contactRequest.UserRequest.RecipientID,
	}

	update := bson.M{
		"$set": bson.M{
			"alias": contactRequest.Alias,
		},
	}

	opts := options.UpdateOne()
	opts.SetUpsert(true)

	_, err := r.collection.UpdateOne(ctx, filter, update, opts)
	return err
}

func (r *ContactsRepository) GetAlias(ctx context.Context, request model.UserRequest) (string, error) {
	filter := bson.M{
		"sender":    request.SenderID,
		"recipient": request.RecipientID,
	}

	var result struct {
		Alias string `bson:"alias"`
	}

	err := r.collection.FindOne(ctx, filter).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return "", model.ErrAliasNotFound
		}
		return "", err
	}
	return result.Alias, nil
}

func (r *ContactsRepository) UpdateAlias(ctx context.Context, contactRequest model.Contact) error {
	filter := bson.M{
		"sender":    contactRequest.UserRequest.SenderID,
		"recipient": contactRequest.UserRequest.RecipientID,
	}

	update := bson.M{"$set": bson.M{"alias": contactRequest.Alias}}

	res, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}
	if res.MatchedCount == 0 {
		return model.ErrAliasNotFound
	}
	return nil
}

func (r *ContactsRepository) GetContacts(ctx context.Context, senderID string) ([]model.Contact, error) {
	filter := bson.M{"sender": senderID}

	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var contacts []model.Contact

	for cursor.Next(ctx) {
		var c struct {
			Sender    string `bson:"sender"`
			Recipient string `bson:"recipient"`
			Alias     string `bson:"alias"`
		}

		if err := cursor.Decode(&c); err != nil {
			return nil, err
		}

		contact := model.Contact{
			UserRequest: model.UserRequest{
				SenderID:    c.Sender,
				RecipientID: c.Recipient,
			},
			Alias: c.Alias,
		}

		contacts = append(contacts, contact)
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return contacts, nil
}

func (r *ContactsRepository) GetContact(ctx context.Context, request model.UserRequest) (model.Contact, error) {
	filter := bson.M{
		"sender":    request.SenderID,
		"recipient": request.RecipientID,
	}

	var c struct {
		Sender    string `bson:"sender"`
		Recipient string `bson:"recipient"`
		Alias     string `bson:"alias"`
	}

	err := r.collection.FindOne(ctx, filter).Decode(&c)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return model.Contact{}, model.ErrContactNotFound
		}
		return model.Contact{}, err
	}

	contact := model.Contact{
		UserRequest: model.UserRequest{
			SenderID:    c.Sender,
			RecipientID: c.Recipient,
		},
		Alias: c.Alias,
	}

	return contact, nil
}

func (r *ContactsRepository) DeleteContact(ctx context.Context, request model.UserRequest) error {
	filter := bson.M{
		"sender":    request.SenderID,
		"recipient": request.RecipientID,
	}

	res, err := r.collection.DeleteOne(ctx, filter)
	if err != nil {
		return err
	}
	if res.DeletedCount == 0 {
		return model.ErrContactNotFound
	}
	return nil
}

func (r *ContactsRepository) DeleteAlias(ctx context.Context, request model.UserRequest) error {
	filter := bson.M{
		"sender":    request.SenderID,
		"recipient": request.RecipientID,
	}

	update := bson.M{
		"$set": bson.M{
			"alias": "",
		},
	}

	res, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}
	if res.MatchedCount == 0 {
		return model.ErrAliasNotFound
	}
	return nil
}
