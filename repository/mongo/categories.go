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
	categoryCollectionName = "categorys"
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

func (r *categoryRepository) Create(ctx context.Context, cus entity.Category) (entity.Category, error) {
	const op = errors.Op("mongo.categoryRepository.Create")
	if cus.ID == "" {
		cus.ID = generateID(categoryIDPrefix)
	}
	cus.Status = entity.StatusActive
	id, err := r.driver.insertOne(ctx, cus)
	if err != nil {
		return entity.Category{}, errors.E(op, err)
	}
	return r.Retrieve(ctx, id)
}

func (r *categoryRepository) Update(ctx context.Context, cus entity.Category) (entity.Category, error) {
	const op = errors.Op("mongo.categoryRepository.Update")
	filter := bson.M{"_id": cus.ID}
	query := bson.M{"$set": cus}
	res, err := r.collection.UpdateOne(ctx, filter, query)
	if err != nil {
		return entity.Category{}, errors.E(op, errors.KindUnexpected, err)
	}
	if res.MatchedCount == 0 {
		return entity.Category{}, errors.E(op, errors.KindNotFound, "category not found")
	}
	cus, err = r.Retrieve(ctx, cus.ID)
	if err != nil {
		return entity.Category{}, errors.E(op, err)
	}
	return cus, nil
}

func (r *categoryRepository) Retrieve(ctx context.Context, id string) (entity.Category, error) {
	const op = errors.Op("mongo.categoryRepository.Retrieve")
	cusr := entity.Category{}
	filter := bson.M{"_id": id}
	if err := r.driver.findOneAndDecode(ctx, &cusr, filter); err != nil {
		return entity.Category{}, errors.E(op, err)
	}
	return cusr, nil
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
	var cuss []entity.Category
	if err := res.All(ctx, &cuss); err != nil {
		return nil, errors.E(op, errors.KindUnexpected, err)
	}
	return cuss, nil
}

func (r *categoryRepository) Delete(ctx context.Context, id string) (entity.Category, error) {
	const op = errors.Op("mongo.categoryRepository.Delete")
	loc, err := r.Update(ctx, entity.Category{
		ID:     id,
		Status: entity.StatusShadowDeleted,
	})
	if err != nil {
		return loc, errors.E(op, err)
	}
	return loc, nil
}
