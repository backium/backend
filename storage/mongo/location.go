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

func (s *locationStorage) Put(ctx context.Context, location core.Location) error {
	const op = errors.Op("mongo/locationStorage.Put")
	now := time.Now().Unix()
	location.UpdatedAt = now
	filter := bson.M{
		"_id":         location.ID,
		"merchant_id": location.MerchantID,
	}
	query := bson.M{"$set": location}
	opts := options.Update().SetUpsert(true)
	res, err := s.collection.UpdateOne(ctx, filter, query, opts)
	if err != nil {
		return errors.E(op, errors.KindUnexpected, err)
	}
	// Update created_at field if upserted
	if res.UpsertedCount == 1 {
		location.CreatedAt = now
		query := bson.M{"$set": location}
		_, err := s.collection.UpdateOne(ctx, filter, query, opts)
		if err != nil {
			return errors.E(op, errors.KindUnexpected, err)
		}
	}
	return nil
}

func (s *locationStorage) PutBatch(ctx context.Context, batch []core.Location) error {
	const op = errors.Op("mongo/locationStorage.PutBatch")
	session, err := s.client.StartSession()
	if err != nil {
		return errors.E(op, errors.KindUnexpected, err)
	}

	_, err = session.WithTransaction(ctx, func(sessCtx mongo.SessionContext) (interface{}, error) {
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
	location := core.Location{}
	filter := bson.M{
		"_id":         id,
		"merchant_id": merchantID,
	}
	if err := s.driver.findOneAndDecode(ctx, &location, filter); err != nil {
		return core.Location{}, errors.E(op, err)
	}
	return location, nil
}

func (s *locationStorage) List(ctx context.Context, f core.LocationFilter) ([]core.Location, error) {
	const op = errors.Op("mongo/locationStorage.List")
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

	res, err := s.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, errors.E(op, errors.KindUnexpected, err)
	}
	var locations []core.Location
	if err := res.All(ctx, &locations); err != nil {
		return nil, errors.E(op, errors.KindUnexpected, err)
	}
	return locations, nil
}
