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
	locationCollectionName = "locations"
)

type locationStorage struct {
	collection *mongo.Collection
	client     *mongo.Client
	driver     *mongoDriver
}

func NewLocationStorage(db DB) core.LocationStorage {
	coll := db.Collection(locationCollectionName)
	return &locationStorage{
		collection: coll,
		client:     db.client,
		driver:     &mongoDriver{Collection: coll},
	}
}

func (s *locationStorage) Put(ctx context.Context, loc core.Location) error {
	const op = errors.Op("mongo/locationStorage.Put")
	now := time.Now().Unix()
	loc.UpdatedAt = now
	f := bson.M{
		"_id":         loc.ID,
		"merchant_id": loc.MerchantID,
	}
	u := bson.M{"$set": loc}
	opts := options.Update().SetUpsert(true)
	res, err := s.collection.UpdateOne(ctx, f, u, opts)
	if err != nil {
		return errors.E(op, errors.KindUnexpected, err)
	}
	// Update created_at field if upserted
	if res.UpsertedCount == 1 {
		loc.CreatedAt = now
		query := bson.M{"$set": loc}
		_, err := s.collection.UpdateOne(ctx, f, query, opts)
		if err != nil {
			return errors.E(op, errors.KindUnexpected, err)
		}
	}
	return nil
}

func (s *locationStorage) PutBatch(ctx context.Context, batch []core.Location) error {
	const op = errors.Op("mongo/locationStorage.PutBatch")
	sess, err := s.client.StartSession()
	if err != nil {
		return errors.E(op, errors.KindUnexpected, err)
	}

	_, err = sess.WithTransaction(ctx, func(sessCtx mongo.SessionContext) (interface{}, error) {
		for _, loc := range batch {
			if err := s.Put(sessCtx, loc); err != nil {
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

func (s *locationStorage) Get(ctx context.Context, id, merchantID string) (core.Location, error) {
	const op = errors.Op("mongo/locationStorage/Get")
	loc := core.Location{}
	f := bson.M{
		"_id":         id,
		"merchant_id": merchantID,
	}
	if err := s.driver.findOneAndDecode(ctx, &loc, f); err != nil {
		return core.Location{}, errors.E(op, err)
	}
	return loc, nil
}

func (s *locationStorage) List(ctx context.Context, fil core.LocationFilter) ([]core.Location, error) {
	const op = errors.Op("mongo/locationStorage.List")
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

	res, err := s.collection.Find(ctx, mfil, fo)
	if err != nil {
		return nil, errors.E(op, errors.KindUnexpected, err)
	}
	var locs []core.Location
	if err := res.All(ctx, &locs); err != nil {
		return nil, errors.E(op, errors.KindUnexpected, err)
	}
	return locs, nil
}
