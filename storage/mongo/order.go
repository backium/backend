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
	filter := bson.M{
		"_id":         order.ID,
		"merchant_id": order.MerchantID,
	}
	query := bson.M{"$set": order}
	opts := options.Update().SetUpsert(true)

	res, err := s.collection.UpdateOne(ctx, filter, query, opts)
	if err != nil {
		return errors.E(op, errors.KindUnexpected, err)
	}

	// Update created_at field if upserted
	if res.UpsertedCount == 1 {
		order.CreatedAt = now
		query := bson.M{"$set": order}
		_, err := s.collection.UpdateOne(ctx, filter, query, opts)
		if err != nil {
			return errors.E(op, errors.KindUnexpected, err)
		}
	}

	return nil
}

func (r *orderStorage) Get(ctx context.Context, id, merchantID string, locationIDs []string) (core.Order, error) {
	const op = errors.Op("mongo/orderStorage.Get")

	order := core.Order{}
	filter := bson.M{
		"_id":         id,
		"merchant_id": merchantID,
	}
	if len(locationIDs) != 0 {
		filter["location_ids"] = bson.M{"$in": locationIDs}
	}

	if err := r.driver.findOneAndDecode(ctx, &order, filter); err != nil {
		return core.Order{}, errors.E(op, err)
	}

	return order, nil
}

func (s *orderStorage) List(ctx context.Context, f core.OrderFilter) ([]core.Order, error) {
	const op = errors.Op("mongo/orderStorage.List")

	opts := options.Find().
		SetLimit(f.Limit).
		SetSkip(f.Offset)

	filter := bson.M{"status": bson.M{"$ne": core.StatusShadowDeleted}}
	if f.MerchantID != "" {
		filter["merchant_id"] = f.MerchantID
	}
	if len(f.LocationIDs) != 0 {
		filter["location_id"] = bson.M{"$in": f.LocationIDs}
	}

	res, err := s.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, errors.E(op, errors.KindUnexpected, err)
	}

	var orders []core.Order
	if err := res.All(ctx, &orders); err != nil {
		return nil, errors.E(op, errors.KindUnexpected, err)
	}

	return orders, nil

}
