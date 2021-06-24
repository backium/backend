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

func (s *itemVariationStorage) Put(ctx context.Context, item core.ItemVariation) error {
	const op = errors.Op("mongo/itemVariationStorage.Put")
	now := time.Now().Unix()
	item.UpdatedAt = now
	f := bson.M{
		"_id":         item.ID,
		"merchant_id": item.MerchantID,
	}
	u := bson.M{"$set": item}
	opts := options.Update().SetUpsert(true)
	res, err := s.collection.UpdateOne(ctx, f, u, opts)
	if err != nil {
		return errors.E(op, errors.KindUnexpected, err)
	}
	// Update created_at field if upserted
	if res.UpsertedCount == 1 {
		item.CreatedAt = now
		query := bson.M{"$set": item}
		_, err := s.collection.UpdateOne(ctx, f, query, opts)
		if err != nil {
			return errors.E(op, errors.KindUnexpected, err)
		}
	}
	return nil
}

func (s *itemVariationStorage) PutBatch(ctx context.Context, batch []core.ItemVariation) error {
	const op = errors.Op("mongo/itemVariationStorage.PutBatch")
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

func (s *itemVariationStorage) Get(ctx context.Context, id, merchantID string, locationIDs []string) (core.ItemVariation, error) {
	const op = errors.Op("mongo/itemVariationStorage/Get")
	v := core.ItemVariation{}
	f := bson.M{
		"_id":         id,
		"merchant_id": merchantID,
	}
	if len(locationIDs) != 0 {
		f["location_ids"] = bson.M{"$in": locationIDs}
	}
	if err := s.driver.findOneAndDecode(ctx, &v, f); err != nil {
		return core.ItemVariation{}, errors.E(op, err)
	}
	return v, nil
}

func (s *itemVariationStorage) List(ctx context.Context, f core.ItemVariationFilter) ([]core.ItemVariation, error) {
	const op = errors.Op("mongo/itemVariationStorage.List")
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
	if f.ItemIDs != nil {
		fil["item_id"] = bson.M{"$in": f.ItemIDs}
	}

	res, err := s.collection.Find(ctx, fil, opts)
	if err != nil {
		return nil, errors.E(op, errors.KindUnexpected, err)
	}
	var vv []core.ItemVariation
	if err := res.All(ctx, &vv); err != nil {
		return nil, errors.E(op, errors.KindUnexpected, err)
	}
	return vv, nil
}
