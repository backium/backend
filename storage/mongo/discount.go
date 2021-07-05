package mongo

import (
	"context"
	"time"

	"github.com/backium/backend/core"
	"github.com/backium/backend/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
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
	filter := bson.M{"_id": discount.ID}
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

func (s *discountStorage) Get(ctx context.Context, id core.ID) (core.Discount, error) {
	const op = errors.Op("mongo/discountStorage/Get")

	discount := core.Discount{}
	filter := bson.M{"_id": id}

	if err := s.driver.findOneAndDecode(ctx, &discount, filter); err != nil {
		return core.Discount{}, errors.E(op, err)
	}

	return discount, nil
}

func (s *discountStorage) List(ctx context.Context, q core.DiscountQuery) ([]core.Discount, int64, error) {
	const op = errors.Op("mongo/discountStorage.List")

	opts := options.Find().
		SetLimit(q.Limit).
		SetSkip(q.Offset)

	if q.Sort.Name != core.SortNone {
		opts.SetSort(bson.M{"name": sortOrder(q.Sort.Name)})
	}

	filter := bson.M{"status": bson.M{"$ne": core.StatusShadowDeleted}}
	if q.Filter.MerchantID != "" {
		filter["merchant_id"] = q.Filter.MerchantID
	}
	if len(q.Filter.IDs) != 0 {
		filter["_id"] = bson.M{"$in": q.Filter.IDs}
	}
	if len(q.Filter.LocationIDs) != 0 {
		filter["location_ids"] = bson.M{"$in": q.Filter.LocationIDs}
	}
	if q.Filter.Name != "" {
		filter["name"] = bson.M{"$regex": primitive.Regex{Pattern: q.Filter.Name, Options: "i"}}
	}

	count, err := s.collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, errors.E(op, errors.KindUnexpected, err)
	}

	res, err := s.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, errors.E(op, errors.KindUnexpected, err)
	}

	var discounts []core.Discount
	if err := res.All(ctx, &discounts); err != nil {
		return nil, 0, errors.E(op, errors.KindUnexpected, err)
	}

	return discounts, count, nil
}
