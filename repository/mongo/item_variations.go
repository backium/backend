package mongo

import (
	"context"
	"fmt"

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

type itemVariationRepository struct {
	collection *mongo.Collection
	driver     *mongoDriver
}

func NewItemVariationRepository(db DB) core.ItemVariationRepository {
	coll := db.Collection(itemVariationCollectionName)
	return &itemVariationRepository{
		collection: coll,
		driver:     &mongoDriver{Collection: coll},
	}
}

func (r *itemVariationRepository) Create(ctx context.Context, itvar core.ItemVariation) (string, error) {
	const op = errors.Op("mongo.itemVariationRepository.Create")
	if itvar.ID == "" {
		itvar.ID = generateID(itemVariationIDPrefix)
	}
	itvar.Status = core.StatusActive
	id, err := r.driver.insertOne(ctx, itvar)
	if err != nil {
		return "", errors.E(op, err)
	}
	return id, nil
}

func (r *itemVariationRepository) Update(ctx context.Context, itvar core.ItemVariation) error {
	const op = errors.Op("mongo.itemVariationRepository.Update")
	filter := bson.M{"_id": itvar.ID}
	query := bson.M{"$set": itvar}
	res, err := r.collection.UpdateOne(ctx, filter, query)
	if err != nil {
		return errors.E(op, errors.KindUnexpected, err)
	}
	if res.MatchedCount == 0 {
		return errors.E(op, errors.KindNotFound, "itemVariation not found")
	}
	return nil
}

func (r *itemVariationRepository) UpdatePartial(ctx context.Context, id string, itvar core.PartialItemVariation) error {
	const op = errors.Op("mongo.itemVariationRepository.Update")
	filter := bson.M{"_id": id}
	query := bson.M{"$set": itvar}
	res, err := r.collection.UpdateOne(ctx, filter, query)
	if err != nil {
		fmt.Printf("%T", err)
		return errors.E(op, errors.KindUnexpected, err)
	}
	if res.MatchedCount == 0 {
		return errors.E(op, errors.KindNotFound, "itemVariation not found")
	}
	return nil
}

func (r *itemVariationRepository) Retrieve(ctx context.Context, id string) (core.ItemVariation, error) {
	const op = errors.Op("mongo.itemVariationRepository.Retrieve")
	cusr := core.ItemVariation{}
	filter := bson.M{"_id": id}
	if err := r.driver.findOneAndDecode(ctx, &cusr, filter); err != nil {
		return core.ItemVariation{}, errors.E(op, err)
	}
	return cusr, nil
}

func (r *itemVariationRepository) List(ctx context.Context, fil core.ItemVariationFilter) ([]core.ItemVariation, error) {
	const op = errors.Op("mongo.itemVariationRepository.List")
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

	res, err := r.collection.Find(ctx, mfil, fo)
	if err != nil {
		return nil, errors.E(op, errors.KindUnexpected, err)
	}
	var itvars []core.ItemVariation
	if err := res.All(ctx, &itvars); err != nil {
		return nil, errors.E(op, errors.KindUnexpected, err)
	}
	return itvars, nil
}
