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
	customerCollectionName = "customers"
)

type customerStorage struct {
	collection *mongo.Collection
	client     *mongo.Client
	driver     *mongoDriver
}

func NewCustomerStorage(db DB) core.CustomerStorage {
	coll := db.Collection(customerCollectionName)
	return &customerStorage{
		collection: coll,
		client:     db.client,
		driver:     &mongoDriver{Collection: coll},
	}
}

func (s *customerStorage) Put(ctx context.Context, customer core.Customer) error {
	const op = errors.Op("mongo/customerStorage.Put")

	now := time.Now().Unix()
	customer.UpdatedAt = now
	filter := bson.M{"_id": customer.ID}
	query := bson.M{"$set": customer}
	opts := options.Update().SetUpsert(true)

	res, err := s.collection.UpdateOne(ctx, filter, query, opts)
	if err != nil {
		return errors.E(op, errors.KindUnexpected, err)
	}

	// Update created_at field if upserted
	if res.UpsertedCount == 1 {
		customer.CreatedAt = now
		query := bson.M{"$set": customer}
		_, err := s.collection.UpdateOne(ctx, filter, query, opts)
		if err != nil {
			return errors.E(op, errors.KindUnexpected, err)
		}
	}

	return nil
}

func (s *customerStorage) PutBatch(ctx context.Context, batch []core.Customer) error {
	const op = errors.Op("mongo/customerStorage.PutBatch")

	session, err := s.client.StartSession()
	if err != nil {
		return errors.E(op, errors.KindUnexpected, err)
	}

	_, err = session.WithTransaction(ctx, func(sessCtx mongo.SessionContext) (interface{}, error) {
		for _, customer := range batch {
			if err := s.Put(sessCtx, customer); err != nil {
				return nil, err
			}
		}
		return nil, nil
	})

	if err != nil {
		return errors.E(op, errors.KindUnexpected, err)
	}

	return nil
}

func (s *customerStorage) Get(ctx context.Context, id core.ID) (core.Customer, error) {
	const op = errors.Op("mongo/customerStorage/Get")

	customer := core.Customer{}
	filter := bson.M{"_id": id}

	if err := s.driver.findOneAndDecode(ctx, &customer, filter); err != nil {
		return core.Customer{}, errors.E(op, err)
	}

	return customer, nil
}

func (s *customerStorage) List(ctx context.Context, f core.CustomerFilter) ([]core.Customer, error) {
	const op = errors.Op("mongo/customerStorage.List")

	opts := options.Find().
		SetLimit(f.Limit).
		SetSkip(f.Offset)

	filter := bson.M{"status": bson.M{"$ne": core.StatusShadowDeleted}}
	if f.MerchantID != "" {
		filter["merchant_id"] = f.MerchantID
	}
	if f.IDs != nil {
		filter["_id"] = bson.M{"$in": f.IDs}
	}

	res, err := s.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, errors.E(op, errors.KindUnexpected, err)
	}

	var customers []core.Customer
	if err := res.All(ctx, &customers); err != nil {
		return nil, errors.E(op, errors.KindUnexpected, err)
	}

	return customers, nil
}
