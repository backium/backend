package mongo

import (
	"context"

	"github.com/backium/backend/base"
	"github.com/backium/backend/catalog"
	"github.com/backium/backend/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	taxIDPrefix       = "tax"
	taxCollectionName = "taxes"
)

type taxRepository struct {
	collection *mongo.Collection
	driver     *mongoDriver
}

func NewTaxRepository(db DB) catalog.TaxRepository {
	coll := db.Collection(taxCollectionName)
	return &taxRepository{
		collection: coll,
		driver:     &mongoDriver{Collection: coll},
	}
}

func (r *taxRepository) Create(ctx context.Context, tax catalog.Tax) (string, error) {
	const op = errors.Op("mongo.taxRepository.Create")
	if tax.ID == "" {
		tax.ID = generateID(taxIDPrefix)
	}
	tax.Status = base.StatusActive
	id, err := r.driver.insertOne(ctx, tax)
	if err != nil {
		return "", errors.E(op, err)
	}
	return id, nil
}

func (r *taxRepository) Update(ctx context.Context, it catalog.Tax) error {
	const op = errors.Op("mongo.taxRepository.Update")
	filter := bson.M{"_id": it.ID}
	query := bson.M{"$set": it}
	res, err := r.collection.UpdateOne(ctx, filter, query)
	if err != nil {
		return errors.E(op, errors.KindUnexpected, err)
	}
	if res.MatchedCount == 0 {
		return errors.E(op, errors.KindNotFound, "tax not found")
	}
	return nil
}

func (r *taxRepository) UpdatePartial(ctx context.Context, id string, it catalog.TaxPartial) error {
	const op = errors.Op("mongo.taxRepository.Update")
	filter := bson.M{"_id": id}
	query := bson.M{"$set": it}
	res, err := r.collection.UpdateOne(ctx, filter, query)
	if err != nil {
		return errors.E(op, errors.KindUnexpected, err)
	}
	if res.MatchedCount == 0 {
		return errors.E(op, errors.KindNotFound, "tax not found")
	}
	return nil
}

func (r *taxRepository) Retrieve(ctx context.Context, id string) (catalog.Tax, error) {
	const op = errors.Op("mongo.taxRepository.Retrieve")
	tax := catalog.Tax{}
	filter := bson.M{"_id": id}
	if err := r.driver.findOneAndDecode(ctx, &tax, filter); err != nil {
		return catalog.Tax{}, errors.E(op, err)
	}
	return tax, nil
}

func (r *taxRepository) List(ctx context.Context, fil catalog.TaxFilter) ([]catalog.Tax, error) {
	const op = errors.Op("mongo.taxRepository.List")
	fo := options.Find().
		SetLimit(fil.Limit).
		SetSkip(fil.Offset)

	mfil := bson.M{"status": bson.M{"$ne": base.StatusShadowDeleted}}
	if fil.MerchantID != "" {
		mfil["merchant_id"] = fil.MerchantID
	}
	if fil.IDs != nil {
		mfil["_id"] = bson.M{"$in": fil.IDs}
	}
	if fil.LocationIDs != nil {
		mfil["location_ids"] = bson.M{"$in": fil.LocationIDs}
	}

	res, err := r.collection.Find(ctx, mfil, fo)
	if err != nil {
		return nil, errors.E(op, errors.KindUnexpected, err)
	}
	var taxes []catalog.Tax
	if err := res.All(ctx, &taxes); err != nil {
		return nil, errors.E(op, errors.KindUnexpected, err)
	}
	return taxes, nil
}