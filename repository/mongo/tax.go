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

func (s *taxStorage) Put(ctx context.Context, t core.Tax) error {
	const op = errors.Op("mongo/taxStorage.Put")
	now := time.Now().Unix()
	t.UpdatedAt = now
	f := bson.M{"_id": t.ID}
	u := bson.M{"$set": t}
	opts := options.Update().SetUpsert(true)
	res, err := s.collection.UpdateOne(ctx, f, u, opts)
	if err != nil {
		return errors.E(op, errors.KindUnexpected, err)
	}
	// Update created_at field if upserted
	if res.UpsertedCount == 1 {
		t.CreatedAt = now
		query := bson.M{"$set": t}
		_, err := s.collection.UpdateOne(ctx, f, query, opts)
		if err != nil {
			return errors.E(op, errors.KindUnexpected, err)
		}
	}
	return nil
}

func (s *taxStorage) PutBatch(ctx context.Context, batch []core.Tax) error {
	const op = errors.Op("mongo/taxStorage.PutBatch")
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

func (s *taxStorage) Get(ctx context.Context, id string) (core.Tax, error) {
	const op = errors.Op("mongo/taxStorage/Get")
	tax := core.Tax{}
	filter := bson.M{"_id": id}
	if err := s.driver.findOneAndDecode(ctx, &tax, filter); err != nil {
		return core.Tax{}, errors.E(op, err)
	}
	return tax, nil
}

func (s *taxStorage) List(ctx context.Context, fil core.TaxFilter) ([]core.Tax, error) {
	const op = errors.Op("mongo/taxStorage.List")
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
	if fil.LocationIDs != nil {
		mfil["location_ids"] = bson.M{"$in": fil.LocationIDs}
	}

	res, err := s.collection.Find(ctx, mfil, fo)
	if err != nil {
		return nil, errors.E(op, errors.KindUnexpected, err)
	}
	var taxes []core.Tax
	if err := res.All(ctx, &taxes); err != nil {
		return nil, errors.E(op, errors.KindUnexpected, err)
	}
	return taxes, nil
}
