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
	discountCollectionName = "discounts"
)

type discountStorage struct {
	collection *mongo.Collection
	client     *mongo.Client
	driver     *mongoDriver
}

func NewDiscountStorage(db DB) core.DiscountStorage {
	coll := db.Collection(discountCollectionName)
	return &discountStorage{
		collection: coll,
		client:     db.client,
		driver:     &mongoDriver{Collection: coll},
	}
}

func (s *discountStorage) Put(ctx context.Context, d core.Discount) error {
	const op = errors.Op("mongo/discountStorage.Put")
	now := time.Now().Unix()
	d.UpdatedAt = now
	f := bson.M{"_id": d.ID}
	u := bson.M{"$set": d}
	opts := options.Update().SetUpsert(true)
	res, err := s.collection.UpdateOne(ctx, f, u, opts)
	if err != nil {
		return errors.E(op, errors.KindUnexpected, err)
	}
	// Update created_at field if upserted
	if res.UpsertedCount == 1 {
		d.CreatedAt = now
		query := bson.M{"$set": d}
		_, err := s.collection.UpdateOne(ctx, f, query, opts)
		if err != nil {
			return errors.E(op, errors.KindUnexpected, err)
		}
	}
	return nil
}

func (s *discountStorage) PutBatch(ctx context.Context, batch []core.Discount) error {
	const op = errors.Op("mongo/discountStorage.PutBatch")
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

func (s *discountStorage) Get(ctx context.Context, id string) (core.Discount, error) {
	const op = errors.Op("mongo/discountStorage/Get")
	d := core.Discount{}
	f := bson.M{"_id": id}
	if err := s.driver.findOneAndDecode(ctx, &d, f); err != nil {
		return core.Discount{}, errors.E(op, err)
	}
	return d, nil
}

func (s *discountStorage) List(ctx context.Context, f core.DiscountFilter) ([]core.Discount, error) {
	const op = errors.Op("mongo/discountStorage.List")
	opts := options.Find().
		SetLimit(f.Limit).
		SetSkip(f.Offset)

	fil := bson.M{"status": bson.M{"$ne": core.StatusShadowDeleted}}
	if f.MerchantID != "" {
		fil["merchant_id"] = f.MerchantID
	}
	if f.IDs != nil {
		fil["_id"] = bson.M{"$in": f.IDs}
	}
	if f.LocationIDs != nil {
		fil["location_ids"] = bson.M{"$in": f.LocationIDs}
	}

	res, err := s.collection.Find(ctx, fil, opts)
	if err != nil {
		return nil, errors.E(op, errors.KindUnexpected, err)
	}
	var dd []core.Discount
	if err := res.All(ctx, &dd); err != nil {
		return nil, errors.E(op, errors.KindUnexpected, err)
	}
	return dd, nil
}
