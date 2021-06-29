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
	filter := bson.M{
		"_id":         payment.ID,
		"merchant_id": payment.MerchantID,
	}
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

func (s *paymentStorage) Get(ctx context.Context, id string) (core.Payment, error) {
	const op = errors.Op("mongo/paymentStorage/Get")
	payment := core.Payment{}
	filter := bson.M{
		"_id": id,
	}
	if err := s.driver.findOneAndDecode(ctx, &payment, filter); err != nil {
		return core.Payment{}, errors.E(op, err)
	}
	return payment, nil
}

func (s *paymentStorage) List(ctx context.Context, f core.PaymentFilter) ([]core.Payment, error) {
	const op = errors.Op("mongo/paymentStorage.List")
	opts := options.Find().
		SetLimit(f.Limit).
		SetSkip(f.Offset)

	filter := bson.M{"status": bson.M{"$ne": core.StatusShadowDeleted}}
	if f.MerchantID != "" {
		filter["merchant_id"] = f.MerchantID
	}
	if len(f.IDs) != 0 {
		filter["_id"] = bson.M{"$in": f.IDs}
	}
	if len(f.LocationIDs) != 0 {
		filter["location_ids"] = bson.M{"$in": f.LocationIDs}
	}
	if len(f.OrderIDs) != 0 {
		filter["order_id"] = bson.M{"$in": f.OrderIDs}
	}

	res, err := s.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, errors.E(op, errors.KindUnexpected, err)
	}
	var payments []core.Payment
	if err := res.All(ctx, &payments); err != nil {
		return nil, errors.E(op, errors.KindUnexpected, err)
	}
	return payments, nil
}
