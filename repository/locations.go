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

func locationRecordFrom(loc entity.Location) locationRecord {
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

type locationMongoRepository struct {
	collection *mongo.Collection
}

func NewLocationMongoRepository(db MongoDB) controller.LocationRepository {
	return &locationMongoRepository{collection: db.Collection(locationCollectionName)}
}

func (r *locationMongoRepository) Create(ctx context.Context, loc entity.Location) (entity.Location, error) {
	record := locationRecordFrom(loc)
	record.ID = generateID(locationIDPrefix)
	res, err := r.collection.InsertOne(ctx, record)
	if err != nil {
		return entity.Location{}, err
	}
	id := res.InsertedID.(string)
	return r.Retrieve(ctx, id)
}

func (r *locationMongoRepository) Update(ctx context.Context, loc entity.Location) (entity.Location, error) {
	record := locationRecordFrom(loc)
	_, err := r.collection.UpdateByID(ctx, loc.ID, record.updateQuery())
	if err != nil {
		return entity.Location{}, err
	}
	return r.Retrieve(ctx, loc.ID)
}

func (r *locationMongoRepository) UpdateByMerchantID(ctx context.Context, loc entity.Location) (entity.Location, error) {
	record := locationRecordFrom(loc)
	filter := bson.M{"_id": loc.ID, "merchant_id": loc.MerchantID}
	_, err := r.collection.UpdateOne(ctx, filter, record.updateQuery())
	if err != nil {
		return entity.Location{}, err
	}
	return r.Retrieve(ctx, loc.ID)
}

func (r *locationMongoRepository) Retrieve(ctx context.Context, id string) (entity.Location, error) {
	res := r.collection.FindOne(ctx, bson.M{"_id": id})
	if err := res.Err(); err != nil {
		return entity.Location{}, err
	}
	record := locationRecord{}
	if err := res.Decode(&record); err != nil {
		return entity.Location{}, err
	}
	return record.location(), nil
}

func (r *locationMongoRepository) ListAll(ctx context.Context, merchantID string) ([]entity.Location, error) {
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

func (r *locationMongoRepository) Delete(ctx context.Context, id string) (entity.Location, error) {
	return entity.Location{}, nil
}
