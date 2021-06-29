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
	filter := bson.M{
		"_id":         variation.ID,
		"merchant_id": variation.MerchantID,
	}
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

func (s *itemVariationStorage) Get(ctx context.Context, id, merchantID string, locationIDs []string) (core.ItemVariation, error) {
	const op = errors.Op("mongo/itemVariationStorage/Get")
	variation := core.ItemVariation{}
	filter := bson.M{
		"_id":         id,
		"merchant_id": merchantID,
	}
	if len(locationIDs) != 0 {
		filter["location_ids"] = bson.M{"$in": locationIDs}
	}
	if err := s.driver.findOneAndDecode(ctx, &variation, filter); err != nil {
		return core.ItemVariation{}, errors.E(op, err)
	}
	return variation, nil
}

func (s *itemVariationStorage) List(ctx context.Context, f core.ItemVariationFilter) ([]core.ItemVariation, error) {
	const op = errors.Op("mongo/itemVariationStorage.List")
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
	if f.ItemIDs != nil {
		filter["item_id"] = bson.M{"$in": f.ItemIDs}
	}

	res, err := s.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, errors.E(op, errors.KindUnexpected, err)
	}
	var variations []core.ItemVariation
	if err := res.All(ctx, &variations); err != nil {
		return nil, errors.E(op, errors.KindUnexpected, err)
	}
	return variations, nil
}
