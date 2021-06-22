package mongo

import (
	"context"

	"github.com/backium/backend/core"
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

func NewLocationRepository(db DB) core.LocationRepository {
	coll := db.Collection(locationCollectionName)
	return &locationRepository{
		collection: coll,
		driver:     &mongoDriver{Collection: coll},
	}
}

func (r *locationRepository) Create(ctx context.Context, loc core.Location) (string, error) {
	const op = errors.Op("mongo.locationRepository.Create")
	loc.ID = generateID(locationIDPrefix)
	loc.Status = core.StatusActive
	id, err := r.driver.insertOne(ctx, loc)
	if err != nil {
		return "", errors.E(op, err)
	}
	return id, err
}

func (r *locationRepository) Update(ctx context.Context, loc core.Location) error {
	const op = errors.Op("mongo.locationRepository.Update")
	fil := bson.M{"_id": loc.ID}
	query := bson.M{"$set": loc}
	res, err := r.collection.UpdateOne(ctx, fil, query)
	if err != nil {
		return errors.E(op, errors.KindUnexpected, err)
	}
	if res.MatchedCount == 0 {
		return errors.E(op, errors.KindNotFound, "location not found")
	}
	return nil
}

func (r *locationRepository) UpdatePartial(ctx context.Context, id string, loc core.LocationPartial) error {
	const op = errors.Op("mongo.locationRepository.Update")
	fil := bson.M{"_id": id}
	query := bson.M{"$set": loc}
	res, err := r.collection.UpdateOne(ctx, fil, query)
	if err != nil {
		return errors.E(op, errors.KindUnexpected, err)
	}
	if res.MatchedCount == 0 {
		return errors.E(op, errors.KindNotFound, "location not found")
	}
	return nil
}

func (r *locationRepository) Retrieve(ctx context.Context, id string) (core.Location, error) {
	const op = errors.Op("mongo.locationRepository.Retrieve")
	locr := core.Location{}
	filter := bson.M{"_id": id}
	if err := r.driver.findOneAndDecode(ctx, &locr, filter); err != nil {
		return core.Location{}, errors.E(op, err)
	}
	return locr, nil
}

func (r *locationRepository) List(ctx context.Context, fil core.ListLocationsFilter) ([]core.Location, error) {
	const op = errors.Op("mongo.locationRepository.List")
	fo := options.Find().
		SetLimit(fil.Limit).
		SetSkip(fil.Offset)

	mfil := bson.M{"status": bson.M{"$ne": core.StatusShadowDeleted}}
	if fil.MerchantID != "" {
		mfil["merchant_id"] = fil.MerchantID
	}

	res, err := r.collection.Find(ctx, mfil, fo)
	if err != nil {
		return nil, errors.E(op, errors.KindUnexpected, err)
	}
	locs := []core.Location{}
	if err := res.All(ctx, &locs); err != nil {
		return nil, errors.E(op, errors.KindUnexpected, err)
	}
	return locs, nil
}