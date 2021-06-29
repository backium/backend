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
	categoryIDPrefix       = "cat"
	categoryCollectionName = "categories"
)

type categoryStorage struct {
	collection *mongo.Collection
	driver     *mongoDriver
	client     *mongo.Client
}

func NewCategoryStorage(db DB) core.CategoryStorage {
	coll := db.Collection(categoryCollectionName)
	return &categoryStorage{
		collection: coll,
		driver:     &mongoDriver{Collection: coll},
		client:     db.client,
	}
}

func (s *categoryStorage) Put(ctx context.Context, category core.Category) error {
	const op = errors.Op("mongo/categoryStorage.Put")
	now := time.Now().Unix()
	category.UpdatedAt = now
	filter := bson.M{
		"_id":         category.ID,
		"merchant_id": category.MerchantID,
	}
	query := bson.M{"$set": category}
	opts := options.Update().SetUpsert(true)
	res, err := s.collection.UpdateOne(ctx, filter, query, opts)
	if err != nil {
		return errors.E(op, errors.KindUnexpected, err)
	}
	// Update created_at field if upserted
	if res.UpsertedCount == 1 {
		category.CreatedAt = now
		query := bson.M{"$set": category}
		_, err := s.collection.UpdateOne(ctx, filter, query, opts)
		if err != nil {
			return errors.E(op, errors.KindUnexpected, err)
		}
	}
	return nil
}

func (s *categoryStorage) PutBatch(ctx context.Context, batch []core.Category) error {
	const op = errors.Op("mongo/categoryStorage.PutBatch")
	session, err := s.client.StartSession()
	if err != nil {
		return errors.E(op, errors.KindUnexpected, err)
	}

	_, err = session.WithTransaction(ctx, func(sessCtx mongo.SessionContext) (interface{}, error) {
		for _, category := range batch {
			if err := s.Put(sessCtx, category); err != nil {
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

func (s *categoryStorage) Get(ctx context.Context, id, merchantID string, locationIDs []string) (core.Category, error) {
	const op = errors.Op("mongo/categoryStorage/Get")
	category := core.Category{}
	filter := bson.M{
		"_id":         id,
		"merchant_id": merchantID,
	}
	if len(locationIDs) != 0 {
		filter["location_ids"] = bson.M{"$in": locationIDs}
	}
	if err := s.driver.findOneAndDecode(ctx, &category, filter); err != nil {
		return core.Category{}, errors.E(op, err)
	}
	return category, nil
}

func (s *categoryStorage) List(ctx context.Context, f core.CategoryFilter) ([]core.Category, error) {
	const op = errors.Op("mongo/categoryStorage.List")
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
	var categories []core.Category
	if err := res.All(ctx, &categories); err != nil {
		return nil, errors.E(op, errors.KindUnexpected, err)
	}
	return categories, nil
}
