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
	customerIDPrefix       = "cus"
	customerCollectionName = "customers"
)

type customerRepository struct {
	collection *mongo.Collection
	driver     *mongoDriver
}

func NewCustomerRepository(db DB) controller.CustomerRepository {
	coll := db.Collection(customerCollectionName)
	return &customerRepository{
		collection: coll,
		driver:     &mongoDriver{Collection: coll},
	}
}

func (r *customerRepository) Create(ctx context.Context, cus entity.Customer) (string, error) {
	const op = errors.Op("mongo.customerRepository.Create")
	if cus.ID == "" {
		cus.ID = generateID(customerIDPrefix)
	}
	cus.Status = entity.StatusActive
	id, err := r.driver.insertOne(ctx, cus)
	if err != nil {
		return "", errors.E(op, err)
	}
	return id, nil
}

func (r *customerRepository) Update(ctx context.Context, cus entity.Customer) error {
	const op = errors.Op("mongo.customerRepository.Update")
	filter := bson.M{"_id": cus.ID}
	query := bson.M{"$set": cus}
	res, err := r.collection.UpdateOne(ctx, filter, query)
	if err != nil {
		return errors.E(op, errors.KindUnexpected, err)
	}
	if res.MatchedCount == 0 {
		return errors.E(op, errors.KindNotFound, "customer not found")
	}
	return nil
}

func (r *customerRepository) UpdatePartial(ctx context.Context, id string, cus controller.PartialCustomer) error {
	const op = errors.Op("mongo.customerRepository.Update")
	filter := bson.M{"_id": id}
	query := bson.M{"$set": cus}
	res, err := r.collection.UpdateOne(ctx, filter, query)
	if err != nil {
		return errors.E(op, errors.KindUnexpected, err)
	}
	if res.MatchedCount == 0 {
		return errors.E(op, errors.KindNotFound, "customer not found")
	}
	return nil
}

func (r *customerRepository) Retrieve(ctx context.Context, id string) (entity.Customer, error) {
	const op = errors.Op("mongo.customerRepository.Retrieve")
	cusr := entity.Customer{}
	filter := bson.M{"_id": id}
	if err := r.driver.findOneAndDecode(ctx, &cusr, filter); err != nil {
		return entity.Customer{}, errors.E(op, err)
	}
	return cusr, nil
}

func (r *customerRepository) List(ctx context.Context, fil controller.ListCustomersFilter) ([]entity.Customer, error) {
	const op = errors.Op("mongo.customerRepository.List")
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

	res, err := r.collection.Find(ctx, mfil, fo)
	if err != nil {
		return nil, errors.E(op, errors.KindUnexpected, err)
	}
	var cuss []entity.Customer
	if err := res.All(ctx, &cuss); err != nil {
		return nil, errors.E(op, errors.KindUnexpected, err)
	}
	return cuss, nil
}
