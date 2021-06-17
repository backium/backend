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
	locationIDPrefix       = "loc"
	locationCollectionName = "locations"
)

type locationRepository struct {
	collection *mongo.Collection
	driver     *mongoDriver
}

func NewLocationRepository(db DB) controller.LocationRepository {
	coll := db.Collection(locationCollectionName)
	return &locationRepository{
		collection: coll,
		driver:     &mongoDriver{Collection: coll},
	}
}

func (r *locationRepository) Create(ctx context.Context, loc entity.Location) (entity.Location, error) {
	const op = errors.Op("mongo.locationRepository.Create")
	loc.ID = generateID(locationIDPrefix)
	loc.Status = entity.StatusActive
	id, err := r.driver.insertOne(ctx, loc)
	if err != nil {
		return entity.Location{}, errors.E(op, err)
	}
	loc, err = r.Retrieve(ctx, id)
	if err != nil {
		return entity.Location{}, errors.E(op, err)
	}
	return loc, err
}

func (r *locationRepository) Update(ctx context.Context, loc entity.Location) (entity.Location, error) {
	const op = errors.Op("mongo.locationRepository.Update")
	fil := bson.M{"_id": loc.ID}
	query := bson.M{"$set": loc}
	res, err := r.collection.UpdateOne(ctx, fil, query)
	if err != nil {
		return entity.Location{}, errors.E(op, errors.KindUnexpected, err)
	}
	if res.MatchedCount == 0 {
		return entity.Location{}, errors.E(op, errors.KindNotFound, "location not found")
	}
	loc, err = r.Retrieve(ctx, loc.ID)
	if err != nil {
		return loc, errors.E(op, err)
	}
	return loc, nil
}

func (r *locationRepository) Retrieve(ctx context.Context, id string) (entity.Location, error) {
	const op = errors.Op("mongo.locationRepository.Retrieve")
	locr := entity.Location{}
	filter := bson.M{"_id": id}
	if err := r.driver.findOneAndDecode(ctx, &locr, filter); err != nil {
		return entity.Location{}, errors.E(op, err)
	}
	return locr, nil
}

func (r *locationRepository) List(ctx context.Context, fil controller.ListLocationsFilter) ([]entity.Location, error) {
	const op = errors.Op("mongo.locationRepository.List")
	fo := options.Find().
		SetLimit(fil.Limit).
		SetSkip(fil.Offset)

	mfil := bson.M{"status": bson.M{"$ne": entity.StatusShadowDeleted}}
	if fil.MerchantID != "" {
		mfil["merchant_id"] = fil.MerchantID
	}

	res, err := r.collection.Find(ctx, mfil, fo)
	if err != nil {
		return nil, errors.E(op, errors.KindUnexpected, err)
	}
	locs := []entity.Location{}
	if err := res.All(ctx, &locs); err != nil {
		return nil, errors.E(op, errors.KindUnexpected, err)
	}
	return locs, nil
}

func (r *locationRepository) Delete(ctx context.Context, id string) (entity.Location, error) {
	const op = errors.Op("mongo.locationRepository.Delete")
	loc, err := r.Update(ctx, entity.Location{ID: id, Status: entity.StatusShadowDeleted})
	if err != nil {
		return loc, errors.E(op, err)
	}
	return loc, nil
}