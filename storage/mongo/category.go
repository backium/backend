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

func (s *categoryStorage) Put(ctx context.Context, cat core.Category) error {
	const op = errors.Op("mongo/categoryStorage.Put")
	now := time.Now().Unix()
	cat.UpdatedAt = now
	f := bson.M{
		"_id":         cat.ID,
		"merchant_id": cat.MerchantID,
	}
	u := bson.M{"$set": cat}
	opts := options.Update().SetUpsert(true)
	res, err := s.collection.UpdateOne(ctx, f, u, opts)
	if err != nil {
		return errors.E(op, errors.KindUnexpected, err)
	}
	// Update created_at field if upserted
	if res.UpsertedCount == 1 {
		cat.CreatedAt = now
		query := bson.M{"$set": cat}
		_, err := s.collection.UpdateOne(ctx, f, query, opts)
		if err != nil {
			return errors.E(op, errors.KindUnexpected, err)
		}
	}
	return nil
}

func (s *categoryStorage) PutBatch(ctx context.Context, batch []core.Category) error {
	const op = errors.Op("mongo/categoryStorage.PutBatch")
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

func (s *categoryStorage) Get(ctx context.Context, id, merchantID string, locationIDs []string) (core.Category, error) {
	const op = errors.Op("mongo/categoryStorage/Get")
	cat := core.Category{}
	f := bson.M{
		"_id":         id,
		"merchant_id": merchantID,
	}
	if len(locationIDs) != 0 {
		f["location_ids"] = bson.M{"$in": locationIDs}
	}
	if err := s.driver.findOneAndDecode(ctx, &cat, f); err != nil {
		return core.Category{}, errors.E(op, err)
	}
	return cat, nil
}

func (s *categoryStorage) List(ctx context.Context, fil core.CategoryFilter) ([]core.Category, error) {
	const op = errors.Op("mongo/categoryStorage.List")
	fo := options.Find().
		SetLimit(fil.Limit).
		SetSkip(fil.Offset)

	mfil := bson.M{"status": bson.M{"$ne": core.StatusShadowDeleted}}
	if fil.MerchantID != "" {
		mfil["merchant_id"] = fil.MerchantID
	}
	if fil.IDs != nil {
		mfil["_id"] = bson.M{"$in": fil.IDs}
	}
	if fil.LocationIDs != nil {
		mfil["location_ids"] = bson.M{"$in": fil.LocationIDs}
	}

	res, err := s.collection.Find(ctx, mfil, fo)
	if err != nil {
		return nil, errors.E(op, errors.KindUnexpected, err)
	}
	var cats []core.Category
	if err := res.All(ctx, &cats); err != nil {
		return nil, errors.E(op, errors.KindUnexpected, err)
	}
	return cats, nil
}
