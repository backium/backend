package mongo

import (
	"context"
	"time"

	"github.com/backium/backend/core"
	"github.com/backium/backend/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	merchantIDPrefix       = "merch"
	merchantCollectionName = "merchants"
)

type merchantStorage struct {
	collection *mongo.Collection
	client     *mongo.Client
	driver     *mongoDriver
}

func NewMerchantStorage(db DB) core.MerchantStorage {
	coll := db.Collection(merchantCollectionName)
	return &merchantStorage{
		collection: coll,
		client:     db.client,
		driver:     &mongoDriver{Collection: coll},
	}
}

func (s *merchantStorage) Put(ctx context.Context, merch core.Merchant) error {
	const op = errors.Op("mongo/merchantStorage.Put")
	now := time.Now().Unix()
	merch.UpdatedAt = now
	f := bson.M{"_id": merch.ID}
	u := bson.M{"$set": merch}
	opts := options.Update().SetUpsert(true)
	res, err := s.collection.UpdateOne(ctx, f, u, opts)
	if err != nil {
		return errors.E(op, errors.KindUnexpected, err)
	}
	// Update created_at field if upserted
	if res.UpsertedCount == 1 {
		merch.CreatedAt = now
		query := bson.M{"$set": merch}
		_, err := s.collection.UpdateOne(ctx, f, query, opts)
		if err != nil {
			return errors.E(op, errors.KindUnexpected, err)
		}
	}
	return nil
}

func (s *merchantStorage) PutKey(ctx context.Context, merchantID string, key core.Key) error {
	const op = errors.Op("mongo/merchantStorage/PutKey")
	merch, err := s.Get(ctx, merchantID)
	if err != nil {
		return errors.E(op, err)
	}
	newKey := true
	for i, k := range merch.Keys {
		if k.Token == key.Token {
			merch.Keys[i].Name = key.Name
			newKey = false
			break
		}
	}
	if newKey {
		merch.Keys = append(merch.Keys, key)
	}
	if err := s.Put(ctx, merch); err != nil {
		return errors.E(op, err)
	}
	return nil
}

func (s *merchantStorage) Get(ctx context.Context, id string) (core.Merchant, error) {
	const op = errors.Op("mongo/merchantStorage/Get")
	cust := core.Merchant{}
	f := bson.M{"_id": id}
	if err := s.driver.findOneAndDecode(ctx, &cust, f); err != nil {
		return core.Merchant{}, errors.E(op, err)
	}
	return cust, nil
}
