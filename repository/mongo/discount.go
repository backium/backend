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
	discountIDPrefix       = "disc"
	discountCollectionName = "discounts"
)

type discountRepository struct {
	collection *mongo.Collection
	driver     *mongoDriver
}

func NewDiscountRepository(db DB) catalog.DiscountRepository {
	coll := db.Collection(discountCollectionName)
	return &discountRepository{
		collection: coll,
		driver:     &mongoDriver{Collection: coll},
	}
}

func (r *discountRepository) Create(ctx context.Context, disc catalog.Discount) (string, error) {
	const op = errors.Op("mongo.discountRepository.Create")
	if disc.ID == "" {
		disc.ID = generateID(discountIDPrefix)
	}
	disc.Status = base.StatusActive
	id, err := r.driver.insertOne(ctx, disc)
	if err != nil {
		return "", errors.E(op, err)
	}
	return id, nil
}

func (r *discountRepository) Update(ctx context.Context, disc catalog.Discount) error {
	const op = errors.Op("mongo.discountRepository.Update")
	filter := bson.M{"_id": disc.ID}
	query := bson.M{"$set": disc}
	res, err := r.collection.UpdateOne(ctx, filter, query)
	if err != nil {
		return errors.E(op, errors.KindUnexpected, err)
	}
	if res.MatchedCount == 0 {
		return errors.E(op, errors.KindNotFound, "discount not found")
	}
	return nil
}

func (r *discountRepository) UpdatePartial(ctx context.Context, id string, disc catalog.DiscountPartial) error {
	const op = errors.Op("mongo.discountRepository.Update")
	filter := bson.M{"_id": id}
	query := bson.M{"$set": disc}
	res, err := r.collection.UpdateOne(ctx, filter, query)
	if err != nil {
		return errors.E(op, errors.KindUnexpected, err)
	}
	if res.MatchedCount == 0 {
		return errors.E(op, errors.KindNotFound, "discount not found")
	}
	return nil
}

func (r *discountRepository) Retrieve(ctx context.Context, id string) (catalog.Discount, error) {
	const op = errors.Op("mongo.discountRepository.Retrieve")
	disc := catalog.Discount{}
	filter := bson.M{"_id": id}
	if err := r.driver.findOneAndDecode(ctx, &disc, filter); err != nil {
		return catalog.Discount{}, errors.E(op, err)
	}
	return disc, nil
}

func (r *discountRepository) List(ctx context.Context, fil catalog.DiscountFilter) ([]catalog.Discount, error) {
	const op = errors.Op("mongo.discountRepository.List")
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
	var discounts []catalog.Discount
	if err := res.All(ctx, &discounts); err != nil {
		return nil, errors.E(op, errors.KindUnexpected, err)
	}
	return discounts, nil
}
