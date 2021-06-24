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

func (s *customerStorage) Put(ctx context.Context, t core.Customer) error {
	const op = errors.Op("mongo/customerStorage.Put")
	now := time.Now().Unix()
	t.UpdatedAt = now
	f := bson.M{
		"_id":         t.ID,
		"merchant_id": t.MerchantID,
	}
	u := bson.M{"$set": t}
	opts := options.Update().SetUpsert(true)
	res, err := s.collection.UpdateOne(ctx, f, u, opts)
	if err != nil {
		return errors.E(op, errors.KindUnexpected, err)
	}
	// Update created_at field if upserted
	if res.UpsertedCount == 1 {
		t.CreatedAt = now
		query := bson.M{"$set": t}
		_, err := s.collection.UpdateOne(ctx, f, query, opts)
		if err != nil {
			return errors.E(op, errors.KindUnexpected, err)
		}
	}
	return nil
}

func (s *customerStorage) PutBatch(ctx context.Context, batch []core.Customer) error {
	const op = errors.Op("mongo/customerStorage.PutBatch")
	sess, err := s.client.StartSession()
	if err != nil {
		return errors.E(op, errors.KindUnexpected, err)
	}

	_, err = sess.WithTransaction(ctx, func(sessCtx mongo.SessionContext) (interface{}, error) {
		for _, t := range batch {
			if err := s.Put(sessCtx, t); err != nil {
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

func (s *customerStorage) Get(ctx context.Context, id, merchantID string) (core.Customer, error) {
	const op = errors.Op("mongo/customerStorage/Get")
	cust := core.Customer{}
	f := bson.M{
		"_id":         id,
		"merchant_id": merchantID,
	}
	if err := s.driver.findOneAndDecode(ctx, &cust, f); err != nil {
		return core.Customer{}, errors.E(op, err)
	}
	return cust, nil
}

func (s *customerStorage) List(ctx context.Context, fil core.CustomerFilter) ([]core.Customer, error) {
	const op = errors.Op("mongo/customerStorage.List")
	opts := options.Find().
		SetLimit(fil.Limit).
		SetSkip(fil.Offset)

	mfil := bson.M{"status": bson.M{"$ne": core.StatusShadowDeleted}}
	if fil.MerchantID != "" {
		mfil["merchant_id"] = fil.MerchantID
	}
	if fil.IDs != nil {
		mfil["_id"] = bson.M{"$in": fil.IDs}
	}

	res, err := s.collection.Find(ctx, mfil, opts)
	if err != nil {
		return nil, errors.E(op, errors.KindUnexpected, err)
	}
	var custs []core.Customer
	if err := res.All(ctx, &custs); err != nil {
		return nil, errors.E(op, errors.KindUnexpected, err)
	}
	return custs, nil
}
