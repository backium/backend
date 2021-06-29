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

func (s *merchantStorage) Put(ctx context.Context, merchant core.Merchant) error {
	const op = errors.Op("mongo/merchantStorage.Put")
	now := time.Now().Unix()
	merchant.UpdatedAt = now
	filter := bson.M{"_id": merchant.ID}
	query := bson.M{"$set": merchant}
	opts := options.Update().SetUpsert(true)
	res, err := s.collection.UpdateOne(ctx, filter, query, opts)
	if err != nil {
		return errors.E(op, errors.KindUnexpected, err)
	}
	// Update created_at field if upserted
	if res.UpsertedCount == 1 {
		merchant.CreatedAt = now
		query := bson.M{"$set": merchant}
		_, err := s.collection.UpdateOne(ctx, filter, query, opts)
		if err != nil {
			return errors.E(op, errors.KindUnexpected, err)
		}
	}
	return nil
}

func (s *merchantStorage) PutKey(ctx context.Context, merchantID string, key core.Key) error {
	const op = errors.Op("mongo/merchantStorage/PutKey")
	merchant, err := s.Get(ctx, merchantID)
	if err != nil {
		return errors.E(op, err)
	}
	newKey := true
	for i, k := range merchant.Keys {
		if k.Token == key.Token {
			merchant.Keys[i].Name = key.Name
			newKey = false
			break
		}
	}
	if newKey {
		merchant.Keys = append(merchant.Keys, key)
	}
	if err := s.Put(ctx, merchant); err != nil {
		return errors.E(op, err)
	}
	return nil
}

func (s *merchantStorage) Get(ctx context.Context, id string) (core.Merchant, error) {
	const op = errors.Op("mongo/merchantStorage/Get")
	merchant := core.Merchant{}
	filter := bson.M{"_id": id}
	if err := s.driver.findOneAndDecode(ctx, &merchant, filter); err != nil {
		return core.Merchant{}, errors.E(op, err)
	}
	return merchant, nil
}

func (s *merchantStorage) GetByKey(ctx context.Context, key string) (core.Merchant, error) {
	const op = errors.Op("mongo/merchantStorage/GetByKey")
	merchant := core.Merchant{}
	filter := bson.M{"keys.token": key}
	if err := s.driver.findOneAndDecode(ctx, &merchant, filter); err != nil {
		return core.Merchant{}, errors.E(op, err)
	}
	return merchant, nil
}
