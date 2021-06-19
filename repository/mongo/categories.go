package mongo

import (
	"context"

	"github.com/backium/backend/controller"
	"github.com/backium/backend/entity"
	"github.com/backium/backend/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	categoryIDPrefix       = "cat"
	categoryCollectionName = "categories"
)

type categoryRepository struct {
	collection *mongo.Collection
	driver     *mongoDriver
}

func NewCategoryRepository(db DB) controller.CategoryRepository {
	coll := db.Collection(categoryCollectionName)
	return &categoryRepository{
		collection: coll,
		driver:     &mongoDriver{Collection: coll},
	}
}

func (r *categoryRepository) Create(ctx context.Context, cat entity.Category) (string, error) {
	const op = errors.Op("mongo.categoryRepository.Create")
	if cat.ID == "" {
		cat.ID = generateID(categoryIDPrefix)
	}
	cat.Status = entity.StatusActive
	id, err := r.driver.insertOne(ctx, cat)
	if err != nil {
		return "", errors.E(op, err)
	}
	return id, nil
}

func (r *categoryRepository) Update(ctx context.Context, cat entity.Category) error {
	const op = errors.Op("mongo.categoryRepository.Update")
	filter := bson.M{"_id": cat.ID}
	query := bson.M{"$set": cat}
	res, err := r.collection.UpdateOne(ctx, filter, query)
	if err != nil {
		return errors.E(op, errors.KindUnexpected, err)
	}
	if res.MatchedCount == 0 {
		return errors.E(op, errors.KindNotFound, "category not found")
	}
	return nil
}

func (r *categoryRepository) UpdatePartial(ctx context.Context, id string, cat controller.PartialCategory) error {
	const op = errors.Op("mongo.categoryRepository.Update")
	filter := bson.M{"_id": id}
	query := bson.M{"$set": cat}
	res, err := r.collection.UpdateOne(ctx, filter, query)
	if err != nil {
		return errors.E(op, errors.KindUnexpected, err)
	}
	if res.MatchedCount == 0 {
		return errors.E(op, errors.KindNotFound, "category not found")
	}
	return nil
}

func (r *categoryRepository) Retrieve(ctx context.Context, id string) (entity.Category, error) {
	const op = errors.Op("mongo.categoryRepository.Retrieve")
	cat := entity.Category{}
	filter := bson.M{"_id": id}
	if err := r.driver.findOneAndDecode(ctx, &cat, filter); err != nil {
		return entity.Category{}, errors.E(op, err)
	}
	return cat, nil
}

func (r *categoryRepository) List(ctx context.Context, fil controller.ListCategoriesFilter) ([]entity.Category, error) {
	const op = errors.Op("mongo.categoryRepository.List")
	fo := options.Find().
		SetLimit(fil.Limit).
		SetSkip(fil.Offset)

	mfil := bson.M{"status": bson.M{"$ne": entity.StatusShadowDeleted}}
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
	var cats []entity.Category
	if err := res.All(ctx, &cats); err != nil {
		return nil, errors.E(op, errors.KindUnexpected, err)
	}
	return cats, nil
}
