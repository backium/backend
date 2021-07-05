package mongo

import (
	"context"
	"time"

	"github.com/backium/backend/core"
	"github.com/backium/backend/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
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

func (s *employeeStorage) List(ctx context.Context, q core.EmployeeQuery) ([]core.Employee, int64, error) {
	const op = errors.Op("mongo/employeeStorage.List")

	opts := options.Find().
		SetLimit(q.Limit).
		SetSkip(q.Offset)

	if q.Sort.Name != core.SortNone {
		opts.SetSort(bson.M{"first_name": sortOrder(q.Sort.Name)})
	}

	filter := bson.M{"status": bson.M{"$ne": core.StatusShadowDeleted}}
	if q.Filter.MerchantID != "" {
		filter["merchant_id"] = q.Filter.MerchantID
	}
	if len(q.Filter.IDs) != 0 {
		filter["_id"] = bson.M{"$in": q.Filter.IDs}
	}
	if len(q.Filter.LocationIDs) != 0 {
		filter["location_ids"] = bson.M{"$in": q.Filter.LocationIDs}
	}
	if q.Filter.Name != "" {
		reg := primitive.Regex{Pattern: q.Filter.Name, Options: "i"}
		filter["$or"] = bson.A{
			bson.M{"first_name": bson.M{"$regex": reg}},
			bson.M{"last_name": bson.M{"$regex": reg}},
		}
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
