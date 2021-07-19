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
	filter := bson.M{"_id": count.ID}
	query := bson.M{"$set": count}
	opts := options.Update().SetUpsert(true)

	_, err := s.countCollection.UpdateOne(ctx, filter, query, opts)
	if err != nil {
		return errors.E(op, errors.KindUnexpected, err)
	}

	return nil
}

func (s *inventoryStorage) PutAdj(ctx context.Context, adj core.InventoryAdjustment) error {
	const op = errors.Op("mongo/inventoryStorage.Put")

	now := time.Now().Unix()
	filter := bson.M{"_id": adj.ID}
	query := bson.M{"$set": adj}
	opts := options.Update().SetUpsert(true)

	res, err := s.adjCollection.UpdateOne(ctx, filter, query, opts)
	if err != nil {
		return errors.E(op, errors.KindUnexpected, err)
	}

	if res.UpsertedCount == 1 {
		query := bson.M{"$set": bson.M{"created_at": now}}
		_, err := s.adjCollection.UpdateOne(ctx, filter, query, opts)
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

func (s *inventoryStorage) ListCount(ctx context.Context, f core.InventoryFilter) ([]core.InventoryCount, int64, error) {
	const op = errors.Op("mongo/inventoryStorage.List")

	opts := options.Find().
		SetLimit(f.Limit).
		SetSkip(f.Offset)

	filter := bson.M{"status": bson.M{"$ne": core.StatusShadowDeleted}}
	if f.MerchantID != "" {
		filter["merchant_id"] = f.MerchantID
	}
	if len(f.LocationIDs) != 0 {
		filter["location_id"] = bson.M{"$in": f.LocationIDs}
	}
	if len(f.ItemVariationIDs) != 0 {
		filter["item_variation_id"] = bson.M{"$in": f.ItemVariationIDs}
	}
	if len(f.IDs) != 0 {
		filter["_id"] = bson.M{"$in": f.IDs}
	}

	count, err := s.countCollection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, errors.E(op, errors.KindUnexpected, err)
	}

	res, err := s.countCollection.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, errors.E(op, errors.KindUnexpected, err)
	}

	var counts []core.InventoryCount
	if err := res.All(ctx, &counts); err != nil {
		return nil, 0, errors.E(op, errors.KindUnexpected, err)
	}

	return counts, count, nil
}

func (s *inventoryStorage) ListAdjustment(ctx context.Context, f core.InventoryFilter) ([]core.InventoryAdjustment, int64, error) {
	const op = errors.Op("mongo/inventoryStorage.ListAdjustment")

	opts := options.Find().
		SetLimit(f.Limit).
		SetSkip(f.Offset)

	filter := bson.M{"status": bson.M{"$ne": core.StatusShadowDeleted}}
	if f.MerchantID != "" {
		filter["merchant_id"] = f.MerchantID
	}
	if len(f.LocationIDs) != 0 {
		filter["location_id"] = bson.M{"$in": f.LocationIDs}
	}
	if len(f.ItemVariationIDs) != 0 {
		filter["item_variation_id"] = bson.M{"$in": f.ItemVariationIDs}
	}
	if len(f.EmployeeIDs) != 0 {
		filter["employee_id"] = bson.M{"$in": f.EmployeeIDs}
	}
	if len(f.IDs) != 0 {
		filter["_id"] = bson.M{"$in": f.IDs}
	}
	if f.CreatedAt.Gte != 0 {
		filter["created_at"] = bson.M{"$gte": f.CreatedAt.Gte}
	}
	if f.CreatedAt.Lte != 0 {
		filter["created_at"] = bson.M{"$gte": f.CreatedAt.Gte, "$lte": f.CreatedAt.Lte}
	}
	filter["auto_generated"] = f.AutoGenerated

	count, err := s.adjCollection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, errors.E(op, errors.KindUnexpected, err)
	}

	res, err := s.adjCollection.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, errors.E(op, errors.KindUnexpected, err)
	}

	var adjs []core.InventoryAdjustment
	if err := res.All(ctx, &adjs); err != nil {
		return nil, 0, errors.E(op, errors.KindUnexpected, err)
	}

	return adjs, count, nil
}
