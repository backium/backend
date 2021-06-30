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
	taxCollectionName = "taxes"
)

type taxStorage struct {
	collection *mongo.Collection
	client     *mongo.Client
	driver     *mongoDriver
}

func NewTaxStorage(db DB) core.TaxStorage {
	coll := db.Collection(taxCollectionName)
	return &taxStorage{
		collection: coll,
		client:     db.client,
		driver:     &mongoDriver{Collection: coll},
	}
}

func (s *taxStorage) Put(ctx context.Context, tax core.Tax) error {
	const op = errors.Op("mongo/taxStorage.Put")

	now := time.Now().Unix()
	tax.UpdatedAt = now
	filter := bson.M{
		"_id":         tax.ID,
		"merchant_id": tax.MerchantID,
	}
	query := bson.M{"$set": tax}
	opts := options.Update().SetUpsert(true)

	res, err := s.collection.UpdateOne(ctx, filter, query, opts)
	if err != nil {
		return errors.E(op, errors.KindUnexpected, err)
	}

	// Update created_at field if upserted
	if res.UpsertedCount == 1 {
		tax.CreatedAt = now
		query := bson.M{"$set": tax}
		_, err := s.collection.UpdateOne(ctx, filter, query, opts)
		if err != nil {
			return errors.E(op, errors.KindUnexpected, err)
		}
	}

	return nil
}

func (s *taxStorage) PutBatch(ctx context.Context, batch []core.Tax) error {
	const op = errors.Op("mongo/taxStorage.PutBatch")

	session, err := s.client.StartSession()
	if err != nil {
		return errors.E(op, errors.KindUnexpected, err)
	}

	_, err = session.WithTransaction(ctx, func(sessCtx mongo.SessionContext) (interface{}, error) {
		for _, tax := range batch {
			if err := s.Put(sessCtx, tax); err != nil {
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

func (s *taxStorage) Get(ctx context.Context, id, merchantID string, locationIDs []string) (core.Tax, error) {
	const op = errors.Op("mongo/taxStorage/Get")

	tax := core.Tax{}
	filter := bson.M{
		"_id":         id,
		"merchant_id": merchantID,
	}
	if len(locationIDs) != 0 {
		filter["location_ids"] = bson.M{"$in": locationIDs}
	}

	if err := s.driver.findOneAndDecode(ctx, &tax, filter); err != nil {
		return core.Tax{}, errors.E(op, err)
	}

	return tax, nil
}

func (s *taxStorage) List(ctx context.Context, f core.TaxFilter) ([]core.Tax, error) {
	const op = errors.Op("mongo/taxStorage.List")

	opts := options.Find().
		SetLimit(f.Limit).
		SetSkip(f.Offset)

	filter := bson.M{"status": bson.M{"$ne": core.StatusShadowDeleted}}
	if f.MerchantID != "" {
		filter["merchant_id"] = f.MerchantID
	}
	if len(f.IDs) != 0 {
		filter["_id"] = bson.M{"$in": f.IDs}
	}
	if len(f.LocationIDs) != 0 {
		filter["location_ids"] = bson.M{"$in": f.LocationIDs}
	}

	res, err := s.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, errors.E(op, errors.KindUnexpected, err)
	}

	var taxes []core.Tax
	if err := res.All(ctx, &taxes); err != nil {
		return nil, errors.E(op, errors.KindUnexpected, err)
	}

	return taxes, nil
}
