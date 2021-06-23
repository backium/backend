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
	orderCollectionName = "orders"
)

type orderStorage struct {
	collection *mongo.Collection
	driver     *mongoDriver
}

func NewOrderStorage(db DB) core.OrderStorage {
	coll := db.Collection(orderCollectionName)
	return &orderStorage{
		collection: coll,
		driver:     &mongoDriver{Collection: coll},
	}
}

func (s *orderStorage) Put(ctx context.Context, order core.Order) error {
	const op = errors.Op("mongo/orderStorage.Put")
	now := time.Now().Unix()
	order.UpdatedAt = now
	f := bson.M{"_id": order.ID}
	u := bson.M{"$set": order}
	opts := options.Update().SetUpsert(true)
	res, err := s.collection.UpdateOne(ctx, f, u, opts)
	if err != nil {
		return errors.E(op, errors.KindUnexpected, err)
	}
	// Update created_at field if upserted
	if res.UpsertedCount == 1 {
		order.CreatedAt = now
		query := bson.M{"$set": order}
		_, err := s.collection.UpdateOne(ctx, f, query, opts)
		if err != nil {
			return errors.E(op, errors.KindUnexpected, err)
		}
	}
	return nil
}

func (r *orderStorage) Get(ctx context.Context, id string) (core.Order, error) {
	const op = errors.Op("mongo/orderStorage.Get")
	order := core.Order{}
	filter := bson.M{"_id": id}
	if err := r.driver.findOneAndDecode(ctx, &order, filter); err != nil {
		return core.Order{}, errors.E(op, err)
	}
	return order, nil
}
