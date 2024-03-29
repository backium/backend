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
	filter := bson.M{"_id": order.ID}
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

func (r *orderStorage) Get(ctx context.Context, id core.ID) (core.Order, error) {
	const op = errors.Op("mongo/orderStorage.Get")

	order := core.Order{}
	filter := bson.M{"_id": id}

	if err := r.driver.findOneAndDecode(ctx, &order, filter); err != nil {
		return core.Order{}, errors.E(op, err)
	}

	return order, nil
}

func (s *orderStorage) List(ctx context.Context, q core.OrderQuery) ([]core.Order, int64, error) {
	const op = errors.Op("mongo/orderStorage.List")

	opts := options.Find().
		SetLimit(q.Limit).
		SetSkip(q.Offset)

	if q.Sort.CreatedAt != core.SortNone {
		opts.SetSort(bson.M{"created_at": sortOrder(q.Sort.CreatedAt)})
	}

	if q.Sort.UpdatedAt != core.SortNone {
		opts.SetSort(bson.M{"updated_at": sortOrder(q.Sort.UpdatedAt)})
	}

	filter := bson.M{"status": bson.M{"$ne": core.StatusShadowDeleted}}
	if q.Filter.MerchantID != "" {
		filter["merchant_id"] = q.Filter.MerchantID
	}
	if len(q.Filter.IDs) != 0 {
		filter["_id"] = bson.M{"$in": q.Filter.IDs}
	}
	if len(q.Filter.PaymentTypes) != 0 {
		filter["payment_types"] = bson.M{"$in": q.Filter.PaymentTypes}
	}
	if len(q.Filter.LocationIDs) != 0 {
		filter["location_id"] = bson.M{"$in": q.Filter.LocationIDs}
	}
	if len(q.Filter.EmployeeIDs) != 0 {
		filter["employee_id"] = bson.M{"$in": q.Filter.EmployeeIDs}
	}
	if len(q.Filter.CustomerIDs) != 0 {
		filter["customer_id"] = bson.M{"$in": q.Filter.CustomerIDs}
	}
	if len(q.Filter.States) != 0 {
		filter["state"] = bson.M{"$in": q.Filter.States}
	}
	if q.Filter.CreatedAt.Gte != 0 {
		filter["created_at"] = bson.M{"$gte": q.Filter.CreatedAt.Gte}
	}
	if q.Filter.CreatedAt.Lte != 0 {
		filter["created_at"] = bson.M{"$gte": q.Filter.CreatedAt.Gte, "$lte": q.Filter.CreatedAt.Lte}
	}
	if q.Filter.UpdatedAt.Gte != 0 {
		filter["updated_at"] = bson.M{"$gte": q.Filter.UpdatedAt.Gte}
	}
	if q.Filter.UpdatedAt.Lte != 0 {
		filter["updated_at"] = bson.M{"$gte": q.Filter.UpdatedAt.Gte, "$lte": q.Filter.UpdatedAt.Lte}
	}

	count, err := s.collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, errors.E(op, errors.KindUnexpected, err)
	}

	res, err := s.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, errors.E(op, errors.KindUnexpected, err)
	}

	var orders []core.Order
	if err := res.All(ctx, &orders); err != nil {
		return nil, 0, errors.E(op, errors.KindUnexpected, err)
	}

	return orders, count, nil
}
