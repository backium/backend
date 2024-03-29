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
	filter := bson.M{"_id": tax.ID}
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

func (s *taxStorage) Get(ctx context.Context, id core.ID) (core.Tax, error) {
	const op = errors.Op("mongo/taxStorage/Get")

	tax := core.Tax{}
	filter := bson.M{"_id": id}

	if err := s.driver.findOneAndDecode(ctx, &tax, filter); err != nil {
		return core.Tax{}, errors.E(op, err)
	}

	return tax, nil
}

func (s *taxStorage) List(ctx context.Context, q core.TaxQuery) ([]core.Tax, int64, error) {
	const op = errors.Op("mongo/taxStorage.List")

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

	var taxes []core.Tax
	if err := res.All(ctx, &taxes); err != nil {
		return nil, 0, errors.E(op, errors.KindUnexpected, err)
	}

	return taxes, count, nil
}
