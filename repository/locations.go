package repository

import (
	"context"

	"github.com/backium/backend/controller"
	"github.com/backium/backend/entity"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

const (
	locationIDPrefix       = "merch"
	locationCollectionName = "locations"
)

type locationRecord struct {
	ID           string `bson:"_id"`
	Name         string `bson:"name"`
	BusinessName string `bson:"business_name"`
	LocationID   string `bson:"merchant_id"`
}

type locationMongoRepository struct {
	collection *mongo.Collection
}

func NewLocationMongoRepository(db MongoDB) controller.LocationRepository {
	return &locationMongoRepository{collection: db.Collection(locationCollectionName)}
}

func (r *locationMongoRepository) Create(m entity.Location) (entity.Location, error) {
	record := locationRecordFrom(m)
	record.ID = generateID(locationIDPrefix)
	res, err := r.collection.InsertOne(context.TODO(), record)
	if err != nil {
		return entity.Location{}, err
	}
	id := res.InsertedID.(string)
	return r.Retrieve(id)
}

func (r *locationMongoRepository) Update(m entity.Location) (entity.Location, error) {
	record := locationRecordFrom(m)
	_, err := r.collection.UpdateByID(context.TODO(), m.ID, record.updateQuery())
	if err != nil {
		return entity.Location{}, err
	}
	return r.Retrieve(m.ID)
}

func (r *locationMongoRepository) Retrieve(id string) (entity.Location, error) {
	res := r.collection.FindOne(context.TODO(), bson.M{"_id": id})
	if err := res.Err(); err != nil {
		return entity.Location{}, err
	}
	record := locationRecord{}
	if err := res.Decode(&record); err != nil {
		return entity.Location{}, err
	}
	return record.location(), nil
}

func (r *locationMongoRepository) ListAll() ([]entity.Location, error) {
	res, err := r.collection.Find(context.TODO(), bson.M{})
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

func (r *locationMongoRepository) Delete(id string) (entity.Location, error) {
	return entity.Location{}, nil
}

func (m locationRecord) location() entity.Location {
	return entity.Location{
		ID:           m.ID,
		FirstName:    m.FirstName,
		LastName:     m.LastName,
		BusinessName: m.BusinessName,
	}
}

func (m locationRecord) updateQuery() bson.M {
	query := bson.M{}
	if m.FirstName != "" {
		query["first_name"] = m.FirstName
	}
	if m.LastName != "" {
		query["last_name"] = m.LastName
	}
	if m.BusinessName != "" {
		query["business_name"] = m.BusinessName
	}
	return bson.M{"$set": query}
}

func locationRecordFrom(m entity.Location) locationRecord {
	return locationRecord{
		FirstName:    m.FirstName,
		LastName:     m.LastName,
		BusinessName: m.BusinessName,
	}
}
