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
	inventoryCountCollectionName = "inventorycounts"
	inventoryAdjCollectionName   = "inventoryadjusments"
)

type inventoryStorage struct {
	countCollection *mongo.Collection
	adjCollection   *mongo.Collection
	countDriver     *mongoDriver
	adjDriver       *mongoDriver
	client          *mongo.Client
}

func NewInventoryStorage(db DB) core.InventoryStorage {
	count := db.Collection(inventoryCountCollectionName)
	adj := db.Collection(inventoryAdjCollectionName)
	return &inventoryStorage{
		countCollection: count,
		adjCollection:   adj,
		countDriver:     &mongoDriver{Collection: count},
		adjDriver:       &mongoDriver{Collection: adj},
		client:          db.client,
	}
}

func (s *inventoryStorage) PutCount(ctx context.Context, count core.InventoryCount) error {
	const op = errors.Op("mongo/inventoryStorage.Put")
	count.CalculatedAt = time.Now().Unix()
	f := bson.M{
		"_id":         count.ID,
		"merchant_id": count.MerchantID,
	}
	u := bson.M{"$set": count}
	opts := options.Update().SetUpsert(true)
	_, err := s.countCollection.UpdateOne(ctx, f, u, opts)
	if err != nil {
		return errors.E(op, errors.KindUnexpected, err)
	}
	return nil
}

func (s *inventoryStorage) PutAdj(ctx context.Context, adj core.InventoryAdjustment) error {
	const op = errors.Op("mongo/inventoryStorage.Put")
	now := time.Now().Unix()
	f := bson.M{
		"_id":         adj.ID,
		"merchant_id": adj.MerchantID,
	}
	u := bson.M{"$set": adj}
	opts := options.Update().SetUpsert(true)
	res, err := s.adjCollection.UpdateOne(ctx, f, u, opts)
	if err != nil {
		return errors.E(op, errors.KindUnexpected, err)
	}
	if res.UpsertedCount == 1 {
		query := bson.M{"$set": bson.M{"created_at": now}}
		_, err := s.adjCollection.UpdateOne(ctx, f, query, opts)
		if err != nil {
			return errors.E(op, errors.KindUnexpected, err)
		}
	}
	return nil
}

func (s *inventoryStorage) PutBatchCount(ctx context.Context, batch []core.InventoryCount) error {
	const op = errors.Op("mongo/inventoryStorage.PutBatch")
	sess, err := s.client.StartSession()
	if err != nil {
		return errors.E(op, errors.KindUnexpected, err)
	}

	_, err = sess.WithTransaction(ctx, func(sessCtx mongo.SessionContext) (interface{}, error) {
		for _, count := range batch {
			if err := s.PutCount(sessCtx, count); err != nil {
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

func (s *inventoryStorage) PutBatchAdj(ctx context.Context, batch []core.InventoryAdjustment) error {
	const op = errors.Op("mongo/inventoryStorage.PutBatch")
	sess, err := s.client.StartSession()
	if err != nil {
		return errors.E(op, errors.KindUnexpected, err)
	}

	_, err = sess.WithTransaction(ctx, func(sessCtx mongo.SessionContext) (interface{}, error) {
		for _, adj := range batch {
			if err := s.PutAdj(sessCtx, adj); err != nil {
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

func (s *inventoryStorage) ListCount(ctx context.Context, f core.InventoryFilter) ([]core.InventoryCount, error) {
	const op = errors.Op("mongo/inventoryStorage.List")
	opts := options.Find().
		SetLimit(f.Limit).
		SetSkip(f.Offset)

	fil := bson.M{"status": bson.M{"$ne": core.StatusShadowDeleted}}
	if f.MerchantID != "" {
		fil["merchant_id"] = f.MerchantID
	}
	if len(f.LocationIDs) != 0 {
		fil["location_id"] = bson.M{"$in": f.LocationIDs}
	}
	if len(f.ItemVariationIDs) != 0 {
		fil["item_variation_id"] = bson.M{"$in": f.ItemVariationIDs}
	}
	if len(f.IDs) != 0 {
		fil["_id"] = bson.M{"$in": f.IDs}
	}

	res, err := s.countCollection.Find(ctx, fil, opts)
	if err != nil {
		return nil, errors.E(op, errors.KindUnexpected, err)
	}
	var counts []core.InventoryCount
	if err := res.All(ctx, &counts); err != nil {
		return nil, errors.E(op, errors.KindUnexpected, err)
	}
	return counts, nil
}
