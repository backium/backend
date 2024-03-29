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
	itemCollectionName = "items"
)

type itemStorage struct {
	collection *mongo.Collection
	driver     *mongoDriver
	client     *mongo.Client
}

func NewItemRepository(db DB) core.ItemStorage {
	coll := db.Collection(itemCollectionName)
	return &itemStorage{
		collection: coll,
		driver:     &mongoDriver{Collection: coll},
		client:     db.client,
	}
}

func (s *itemStorage) Put(ctx context.Context, item core.Item) error {
	const op = errors.Op("mongo/itemStorage.Put")

	now := time.Now().Unix()
	item.UpdatedAt = now
	filter := bson.M{"_id": item.ID}
	query := bson.M{"$set": item}
	opts := options.Update().SetUpsert(true)

	res, err := s.collection.UpdateOne(ctx, filter, query, opts)
	if err != nil {
		return errors.E(op, errors.KindUnexpected, err)
	}

	// Update created_at field if upserted
	if res.UpsertedCount == 1 {
		item.CreatedAt = now
		query := bson.M{"$set": item}
		_, err := s.collection.UpdateOne(ctx, filter, query, opts)
		if err != nil {
			return errors.E(op, errors.KindUnexpected, err)
		}
	}

	return nil
}

func (s *itemStorage) PutBatch(ctx context.Context, batch []core.Item) error {
	const op = errors.Op("mongo/itemStorage.PutBatch")

	session, err := s.client.StartSession()
	if err != nil {
		return errors.E(op, errors.KindUnexpected, err)
	}

	_, err = session.WithTransaction(ctx, func(sessCtx mongo.SessionContext) (interface{}, error) {
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

func (s *itemStorage) Get(ctx context.Context, id core.ID) (core.Item, error) {
	const op = errors.Op("mongo/itemStorage/Get")

	item := core.Item{}
	filter := bson.M{"_id": id}

	if err := s.driver.findOneAndDecode(ctx, &item, filter); err != nil {
		return core.Item{}, errors.E(op, err)
	}

	return item, nil
}

func (s *itemStorage) List(ctx context.Context, q core.ItemQuery) ([]core.Item, int64, error) {
	const op = errors.Op("mongo/itemStorage.List")

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
	if len(q.Filter.CategoryIDs) != 0 {
		filter["category_id"] = bson.M{"$in": q.Filter.CategoryIDs}
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

	var items []core.Item
	if err := res.All(ctx, &items); err != nil {
		return nil, 0, errors.E(op, errors.KindUnexpected, err)
	}

	return items, count, nil
}
