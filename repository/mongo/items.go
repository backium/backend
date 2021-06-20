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
	itemIDPrefix       = "item"
	itemCollectionName = "items"
)

type itemRepository struct {
	collection *mongo.Collection
	driver     *mongoDriver
}

func NewItemRepository(db DB) catalog.ItemRepository {
	coll := db.Collection(itemCollectionName)
	return &itemRepository{
		collection: coll,
		driver:     &mongoDriver{Collection: coll},
	}
}

func (r *itemRepository) Create(ctx context.Context, cus catalog.Item) (string, error) {
	const op = errors.Op("mongo.itemRepository.Create")
	if cus.ID == "" {
		cus.ID = generateID(itemIDPrefix)
	}
	cus.Status = base.StatusActive
	id, err := r.driver.insertOne(ctx, cus)
	if err != nil {
		return "", errors.E(op, err)
	}
	return id, nil
}

func (r *itemRepository) Update(ctx context.Context, it catalog.Item) error {
	const op = errors.Op("mongo.itemRepository.Update")
	filter := bson.M{"_id": it.ID}
	query := bson.M{"$set": it}
	res, err := r.collection.UpdateOne(ctx, filter, query)
	if err != nil {
		return errors.E(op, errors.KindUnexpected, err)
	}
	if res.MatchedCount == 0 {
		return errors.E(op, errors.KindNotFound, "item not found")
	}
	return nil
}

func (r *itemRepository) UpdatePartial(ctx context.Context, id string, it catalog.PartialItem) error {
	const op = errors.Op("mongo.itemRepository.Update")
	filter := bson.M{"_id": id}
	query := bson.M{"$set": it}
	res, err := r.collection.UpdateOne(ctx, filter, query)
	if err != nil {
		return errors.E(op, errors.KindUnexpected, err)
	}
	if res.MatchedCount == 0 {
		return errors.E(op, errors.KindNotFound, "item not found")
	}
	return nil
}

func (r *itemRepository) Retrieve(ctx context.Context, id string) (catalog.Item, error) {
	const op = errors.Op("mongo.itemRepository.Retrieve")
	cusr := catalog.Item{}
	filter := bson.M{"_id": id}
	if err := r.driver.findOneAndDecode(ctx, &cusr, filter); err != nil {
		return catalog.Item{}, errors.E(op, err)
	}
	return cusr, nil
}

func (r *itemRepository) List(ctx context.Context, fil catalog.ItemFilter) ([]catalog.Item, error) {
	const op = errors.Op("mongo.itemRepository.List")
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
	var cuss []catalog.Item
	if err := res.All(ctx, &cuss); err != nil {
		return nil, errors.E(op, errors.KindUnexpected, err)
	}
	return cuss, nil
}
