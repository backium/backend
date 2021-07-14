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
	cashDrawerIDPrefix          = "cat"
	cashDrawerCollectionName    = "cashdrawers"
	cashDrawerAdjCollectionName = "cashdraweradjustments"
)

type cashDrawerStorage struct {
	collection    *mongo.Collection
	adjCollection *mongo.Collection
	driver        *mongoDriver
	client        *mongo.Client
}

func NewCashDrawerStorage(db DB) core.CashDrawerStorage {
	coll := db.Collection(cashDrawerCollectionName)
	adj := db.Collection(cashDrawerAdjCollectionName)
	return &cashDrawerStorage{
		collection:    coll,
		adjCollection: adj,
		driver:        &mongoDriver{Collection: coll},
		client:        db.client,
	}
}

func (s *cashDrawerStorage) Put(ctx context.Context, drawer core.CashDrawer) error {
	const op = errors.Op("mongo/cashDrawerStorage.Put")

	now := time.Now().Unix()
	drawer.CalculatedAt = now
	filter := bson.M{"_id": drawer.ID}
	query := bson.M{"$set": drawer}
	opts := options.Update().SetUpsert(true)

	_, err := s.collection.UpdateOne(ctx, filter, query, opts)
	if err != nil {
		return errors.E(op, errors.KindUnexpected, err)
	}

	return nil
}

func (s *cashDrawerStorage) PutAdj(ctx context.Context, adj core.CashDrawerAdjustment) error {
	const op = errors.Op("mongo/cashDrawerStorage.PutAdj")

	filter := bson.M{"_id": adj.ID}
	query := bson.M{"$set": adj}
	opts := options.Update().SetUpsert(true)
	now := time.Now().Unix()

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

func (s *cashDrawerStorage) Get(ctx context.Context, id core.ID) (core.CashDrawer, error) {
	const op = errors.Op("mongo/cashDrawerStorage/Get")

	drawer := core.CashDrawer{}
	filter := bson.M{"_id": id}

	if err := s.driver.findOneAndDecode(ctx, &drawer, filter); err != nil {
		return core.CashDrawer{}, errors.E(op, err)
	}

	return drawer, nil
}

func (s *cashDrawerStorage) List(ctx context.Context, q core.CashDrawerQuery) ([]core.CashDrawer, int64, error) {
	const op = errors.Op("mongo/cashDrawerStorage.List")

	opts := options.Find().
		SetLimit(q.Limit).
		SetSkip(q.Offset)

	filter := bson.M{"status": bson.M{"$ne": core.StatusShadowDeleted}}
	if q.Filter.MerchantID != "" {
		filter["merchant_id"] = q.Filter.MerchantID
	}
	if len(q.Filter.IDs) != 0 {
		filter["_id"] = bson.M{"$in": q.Filter.IDs}
	}
	if len(q.Filter.LocationIDs) != 0 {
		filter["location_id"] = bson.M{"$in": q.Filter.LocationIDs}
	}

	count, err := s.collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, errors.E(op, errors.KindUnexpected, err)
	}

	res, err := s.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, errors.E(op, errors.KindUnexpected, err)
	}

	var drawers []core.CashDrawer
	if err := res.All(ctx, &drawers); err != nil {
		return nil, 0, errors.E(op, errors.KindUnexpected, err)
	}

	return drawers, count, nil
}
