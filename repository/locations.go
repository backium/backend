package repository

import (
	"context"

	"github.com/backium/backend/controller"
	"github.com/backium/backend/entity"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

const (
	locationIDPrefix       = "loc"
	locationCollectionName = "locations"
)

type locationRecord struct {
	ID           string `bson:"_id"`
	Name         string `bson:"name"`
	BusinessName string `bson:"business_name"`
	MerchantID   string `bson:"merchant_id"`
}

func newLocationRecord(loc entity.Location) locationRecord {
	return locationRecord{
		ID:           loc.ID,
		Name:         loc.Name,
		BusinessName: loc.BusinessName,
		MerchantID:   loc.MerchantID,
	}
}

func (loc locationRecord) location() entity.Location {
	return entity.Location{
		ID:           loc.ID,
		Name:         loc.Name,
		BusinessName: loc.BusinessName,
		MerchantID:   loc.MerchantID,
	}
}

func (loc locationRecord) updateQuery() bson.M {
	query := bson.M{}
	if loc.Name != "" {
		query["name"] = loc.Name
	}
	if loc.BusinessName != "" {
		query["business_name"] = loc.BusinessName
	}
	return bson.M{"$set": query}
}

type locationMongoRepo struct {
	collection *mongo.Collection
	driver     *mongoDriver
}

func NewLocationMongoRepository(db MongoDB) controller.LocationRepository {
	coll := db.Collection(locationCollectionName)
	return &locationMongoRepo{
		collection: coll,
		driver:     &mongoDriver{Collection: coll},
	}
}

func (r *locationMongoRepo) Create(ctx context.Context, loc entity.Location) (entity.Location, error) {
	locr := newLocationRecord(loc)
	locr.ID = generateID(locationIDPrefix)
	id, err := r.driver.insertOne(ctx, locr)
	if err != nil {
		return entity.Location{}, err
	}
	return r.Retrieve(ctx, id)
}

func (r *locationMongoRepo) Update(ctx context.Context, loc entity.Location) (entity.Location, error) {
	locr := locationRecord{}
	filter := bson.M{"_id": loc.ID}
	if err := r.driver.findOneAndDecode(ctx, &locr, filter); err != nil {
		return entity.Location{}, err
	}
	cusUpdate := newLocationRecord(loc)
	if err := updateFields(&locr, cusUpdate); err != nil {
		return entity.Location{}, err
	}
	query := bson.M{"$set": locr}
	if _, err := r.collection.UpdateOne(ctx, filter, query); err != nil {
		return entity.Location{}, err
	}
	return r.Retrieve(ctx, loc.ID)
}

func (r *locationMongoRepo) Retrieve(ctx context.Context, id string) (entity.Location, error) {
	locr := locationRecord{}
	filter := bson.M{"_id": id}
	if err := r.driver.findOneAndDecode(ctx, &locr, filter); err != nil {
		return entity.Location{}, err
	}
	return locr.location(), nil
}

func (r *locationMongoRepo) ListAll(ctx context.Context, merchantID string) ([]entity.Location, error) {
	res, err := r.collection.Find(ctx, bson.M{"merchant_id": merchantID})
	if err != nil {
		return nil, err
	}
	var ms []entity.Location
	for res.Next(context.TODO()) {
		record := locationRecord{}
		if err := res.Decode(&record); err != nil {
			continue
		}
		ms = append(ms, record.location())
	}
	return ms, nil
}

func (r *locationMongoRepo) Delete(ctx context.Context, id string) (entity.Location, error) {
	return entity.Location{}, nil
}
