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
	paymentCollectionName = "payments"
)

type paymentStorage struct {
	collection *mongo.Collection
	client     *mongo.Client
	driver     *mongoDriver
}

func NewPaymentStorage(db DB) core.PaymentStorage {
	coll := db.Collection(paymentCollectionName)
	return &paymentStorage{
		collection: coll,
		client:     db.client,
		driver:     &mongoDriver{Collection: coll},
	}
}

func (s *paymentStorage) Put(ctx context.Context, payment core.Payment) error {
	const op = errors.Op("mongo/paymentStorage.Put")

	now := time.Now().Unix()
	payment.UpdatedAt = now
	filter := bson.M{"_id": payment.ID}
	query := bson.M{"$set": payment}
	opts := options.Update().SetUpsert(true)

	res, err := s.collection.UpdateOne(ctx, filter, query, opts)
	if err != nil {
		return errors.E(op, errors.KindUnexpected, err)
	}

	// Update created_at field if upserted
	if res.UpsertedCount == 1 {
		payment.CreatedAt = now
		query := bson.M{"$set": payment}
		_, err := s.collection.UpdateOne(ctx, filter, query, opts)
		if err != nil {
			return errors.E(op, errors.KindUnexpected, err)
		}
	}

	return nil
}

func (s *paymentStorage) Get(ctx context.Context, id core.ID) (core.Payment, error) {
	const op = errors.Op("mongo/paymentStorage/Get")

	payment := core.Payment{}
	filter := bson.M{"_id": id}
	if err := s.driver.findOneAndDecode(ctx, &payment, filter); err != nil {
		return core.Payment{}, errors.E(op, err)
	}

	return payment, nil
}

func (s *paymentStorage) List(ctx context.Context, q core.PaymentQuery) ([]core.Payment, int64, error) {
	const op = errors.Op("mongo/paymentStorage.List")

	opts := options.Find().
		SetLimit(q.Limit).
		SetSkip(q.Offset)

	if q.Sort.CreatedAt != core.SortNone {
		opts.SetSort(bson.M{"created_at": sortOrder(q.Sort.CreatedAt)})
	}

	filter := bson.M{"status": bson.M{"$ne": core.StatusShadowDeleted}}
	if q.Filter.MerchantID != "" {
		filter["merchant_id"] = q.Filter.MerchantID
	}
	if len(q.Filter.OrderIDs) != 0 {
		filter["order_id"] = bson.M{"$in": q.Filter.OrderIDs}
	}
	if len(q.Filter.IDs) != 0 {
		filter["_id"] = bson.M{"$in": q.Filter.IDs}
	}
	if len(q.Filter.LocationIDs) != 0 {
		filter["location_id"] = bson.M{"$in": q.Filter.LocationIDs}
	}
	if q.Filter.CreatedAt.Gte != 0 {
		filter["created_at"] = bson.M{"$gte": q.Filter.CreatedAt.Gte}
	}
	if q.Filter.CreatedAt.Lte != 0 {
		filter["created_at"] = bson.M{"$gte": q.Filter.CreatedAt.Gte, "$lte": q.Filter.CreatedAt.Lte}
	}

	count, err := s.collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, errors.E(op, errors.KindUnexpected, err)
	}

	res, err := s.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, errors.E(op, errors.KindUnexpected, err)
	}

	var payments []core.Payment
	if err := res.All(ctx, &payments); err != nil {
		return nil, 0, errors.E(op, errors.KindUnexpected, err)
	}

	return payments, count, nil
}
