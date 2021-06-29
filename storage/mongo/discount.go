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

func (s *discountStorage) Put(ctx context.Context, discount core.Discount) error {
	const op = errors.Op("mongo/discountStorage.Put")
	now := time.Now().Unix()
	discount.UpdatedAt = now
	filter := bson.M{
		"_id":         discount.ID,
		"merchant_id": discount.MerchantID,
	}
	query := bson.M{"$set": discount}
	opts := options.Update().SetUpsert(true)
	res, err := s.collection.UpdateOne(ctx, filter, query, opts)
	if err != nil {
		return errors.E(op, errors.KindUnexpected, err)
	}
	// Update created_at field if upserted
	if res.UpsertedCount == 1 {
		discount.CreatedAt = now
		query := bson.M{"$set": discount}
		_, err := s.collection.UpdateOne(ctx, filter, query, opts)
		if err != nil {
			return errors.E(op, errors.KindUnexpected, err)
		}
	}
	return nil
}

func (s *discountStorage) PutBatch(ctx context.Context, batch []core.Discount) error {
	const op = errors.Op("mongo/discountStorage.PutBatch")
	session, err := s.client.StartSession()
	if err != nil {
		return errors.E(op, errors.KindUnexpected, err)
	}

	_, err = session.WithTransaction(ctx, func(sessCtx mongo.SessionContext) (interface{}, error) {
		for _, discount := range batch {
			if err := s.Put(sessCtx, discount); err != nil {
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

func (s *discountStorage) Get(ctx context.Context, id, merchantID string, locationIDs []string) (core.Discount, error) {
	const op = errors.Op("mongo/discountStorage/Get")
	discount := core.Discount{}
	filter := bson.M{
		"_id":         id,
		"merchant_id": merchantID,
	}
	if len(locationIDs) != 0 {
		filter["location_ids"] = bson.M{"$in": locationIDs}
	}
	if err := s.driver.findOneAndDecode(ctx, &discount, filter); err != nil {
		return core.Discount{}, errors.E(op, err)
	}
	return discount, nil
}

func (s *discountStorage) List(ctx context.Context, f core.DiscountFilter) ([]core.Discount, error) {
	const op = errors.Op("mongo/discountStorage.List")
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
	if f.LocationIDs != nil {
		filter["location_ids"] = bson.M{"$in": f.LocationIDs}
	}

	res, err := s.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, errors.E(op, errors.KindUnexpected, err)
	}
	var discounts []core.Discount
	if err := res.All(ctx, &discounts); err != nil {
		return nil, errors.E(op, errors.KindUnexpected, err)
	}
	return discounts, nil
}
