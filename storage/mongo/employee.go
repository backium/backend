package mongo

import (
	"context"
	"time"

	"github.com/backium/backend/core"
	"github.com/backium/backend/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	employeeCollectionName = "employees"
)

type employeeStorage struct {
	collection *mongo.Collection
	driver     *mongoDriver
	client     *mongo.Client
}

func NewEmployeeStorage(db DB) core.EmployeeStorage {
	coll := db.Collection(employeeCollectionName)
	return &employeeStorage{
		collection: coll,
		driver:     &mongoDriver{Collection: coll},
		client:     db.client,
	}
}

func (s *employeeStorage) Put(ctx context.Context, employee core.Employee) error {
	const op = errors.Op("mongo/employeeStorage.Put")

	now := time.Now().Unix()
	employee.UpdatedAt = now
	filter := bson.M{"_id": employee.ID}
	query := bson.M{"$set": employee}
	opts := options.Update().SetUpsert(true)

	res, err := s.collection.UpdateOne(ctx, filter, query, opts)
	if err != nil {
		return errors.E(op, errors.KindUnexpected, err)
	}

	// Update created_at field if upserted
	if res.UpsertedCount == 1 {
		employee.CreatedAt = now
		query := bson.M{"$set": employee}
		_, err := s.collection.UpdateOne(ctx, filter, query, opts)
		if err != nil {
			return errors.E(op, errors.KindUnexpected, err)
		}
	}

	return nil
}

func (s *employeeStorage) Get(ctx context.Context, id core.ID) (core.Employee, error) {
	const op = errors.Op("mongo/employeeStorage/Get")

	employee := core.Employee{}
	filter := bson.M{"_id": id}

	if err := s.driver.findOneAndDecode(ctx, &employee, filter); err != nil {
		return core.Employee{}, errors.E(op, err)
	}

	return employee, nil
}

func (s *employeeStorage) List(ctx context.Context, f core.EmployeeFilter) ([]core.Employee, int64, error) {
	const op = errors.Op("mongo/employeeStorage.List")

	opts := options.Find().
		SetLimit(f.Limit).
		SetSkip(f.Offset)

	filter := bson.M{"status": bson.M{"$ne": core.StatusShadowDeleted}}
	if f.MerchantID != "" {
		filter["merchant_id"] = f.MerchantID
	}
	if len(f.LocationIDs) != 0 {
		filter["location_ids"] = bson.M{"$in": f.LocationIDs}
	}

	count, err := s.collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, errors.E(op, errors.KindUnexpected, err)
	}

	res, err := s.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, errors.E(op, errors.KindUnexpected, err)
	}

	var employees []core.Employee
	if err := res.All(ctx, &employees); err != nil {
		return nil, 0, errors.E(op, errors.KindUnexpected, err)
	}

	return employees, count, nil
}
