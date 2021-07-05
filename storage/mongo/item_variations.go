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
	itemVariationIDPrefix       = "itemvar"
	itemVariationCollectionName = "itemvariations"
)

type itemVariationStorage struct {
	collection *mongo.Collection
	driver     *mongoDriver
	client     *mongo.Client
}

func NewItemVariationStorage(db DB) core.ItemVariationStorage {
	coll := db.Collection(itemVariationCollectionName)
	return &itemVariationStorage{
		collection: coll,
		driver:     &mongoDriver{Collection: coll},
		client:     db.client,
	}
}

func (s *itemVariationStorage) Put(ctx context.Context, variation core.ItemVariation) error {
	const op = errors.Op("mongo/itemVariationStorage.Put")

	now := time.Now().Unix()
	variation.UpdatedAt = now

	filter := bson.M{"_id": variation.ID}
	query := bson.M{"$set": variation}
	opts := options.Update().SetUpsert(true)
	res, err := s.collection.UpdateOne(ctx, filter, query, opts)
	if err != nil {
		return errors.E(op, errors.KindUnexpected, err)
	}

	// Update created_at field if upserted
	if res.UpsertedCount == 1 {
		variation.CreatedAt = now
		query := bson.M{"$set": variation}
		_, err := s.collection.UpdateOne(ctx, filter, query, opts)
		if err != nil {
			return errors.E(op, errors.KindUnexpected, err)
		}
	}

	return nil
}

func (s *itemVariationStorage) PutBatch(ctx context.Context, batch []core.ItemVariation) error {
	const op = errors.Op("mongo/itemVariationStorage.PutBatch")

	session, err := s.client.StartSession()
	if err != nil {
		return errors.E(op, errors.KindUnexpected, err)
	}

	_, err = session.WithTransaction(ctx, func(sessCtx mongo.SessionContext) (interface{}, error) {
		for _, variation := range batch {
			if err := s.Put(sessCtx, variation); err != nil {
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

func (s *itemVariationStorage) Get(ctx context.Context, id core.ID) (core.ItemVariation, error) {
	const op = errors.Op("mongo/itemVariationStorage/Get")

	variation := core.ItemVariation{}
	filter := bson.M{"_id": id}

	if err := s.driver.findOneAndDecode(ctx, &variation, filter); err != nil {
		return core.ItemVariation{}, errors.E(op, err)
	}

	return variation, nil
}

func (s *itemVariationStorage) List(ctx context.Context, q core.ItemVariationQuery) ([]core.ItemVariation, int64, error) {
	const op = errors.Op("mongo/itemVariationStorage.List")

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
	if len(q.Filter.ItemIDs) != 0 {
		filter["item_id"] = bson.M{"$in": q.Filter.ItemIDs}
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

	var variations []core.ItemVariation
	if err := res.All(ctx, &variations); err != nil {
		return nil, 0, errors.E(op, errors.KindUnexpected, err)
	}

	return variations, count, nil
}
