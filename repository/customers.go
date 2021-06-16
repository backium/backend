package repository

import (
	"context"

	"github.com/backium/backend/controller"
	"github.com/backium/backend/entity"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	customerIDPrefix       = "cus"
	customerCollectionName = "customers"
)

type customerRecord struct {
	ID         string         `bson:"_id"`
	Name       string         `bson:"name,omitempty"`
	Email      string         `bson:"email,omitempty"`
	Phone      string         `bson:"phone,omitempty"`
	Address    *addressRecord `bson:"address,omitempty"`
	MerchantID string         `bson:"merchant_id,omitempty"`
}

func newCustomerRecord(c entity.Customer) customerRecord {
	var addr *addressRecord
	if c.Address != nil {
		addr = &addressRecord{
			Line1:      c.Address.Line1,
			Line2:      c.Address.Line2,
			Province:   c.Address.Province,
			District:   c.Address.District,
			Department: c.Address.Department,
		}
	}
	return customerRecord{
		ID:         c.ID,
		Name:       c.Name,
		Email:      c.Email,
		Phone:      c.Phone,
		Address:    addr,
		MerchantID: c.MerchantID,
	}
}

func (c customerRecord) customer() entity.Customer {
	var addr *entity.Address
	if c.Address != nil {
		addr = &entity.Address{
			Line1:      c.Address.Line1,
			Line2:      c.Address.Line2,
			Province:   c.Address.Province,
			District:   c.Address.District,
			Department: c.Address.Department,
		}
	}
	return entity.Customer{
		ID:         c.ID,
		Name:       c.Name,
		Email:      c.Email,
		Phone:      c.Phone,
		Address:    addr,
		MerchantID: c.MerchantID,
	}
}

func filterQuery(fil controller.SearchCustomersFilter) bson.M {
	q := bson.M{}
	if fil.MerchantID != "" {
		q["merchant_id"] = fil.MerchantID
	}
	if fil.IDs != nil {
		q["_id"] = bson.M{"$in": fil.IDs}
	}
	return q
}

type customerMongoRepo struct {
	collection *mongo.Collection
	driver     *mongoDriver
}

func NewCustomerMongoRepository(db MongoDB) controller.CustomerRepository {
	coll := db.Collection(customerCollectionName)
	return &customerMongoRepo{
		collection: coll,
		driver:     &mongoDriver{Collection: coll},
	}
}

func (r *customerMongoRepo) Create(ctx context.Context, cus entity.Customer) (entity.Customer, error) {
	cusr := newCustomerRecord(cus)
	cusr.ID = generateID(customerIDPrefix)
	id, err := r.driver.insertOne(ctx, cusr)
	if err != nil {
		return entity.Customer{}, err
	}
	return r.Retrieve(ctx, id)
}

func (r *customerMongoRepo) Update(ctx context.Context, cus entity.Customer) (entity.Customer, error) {
	cusr := customerRecord{}
	filter := bson.M{"_id": cus.ID}
	if err := r.driver.findOneAndDecode(ctx, &cusr, filter); err != nil {
		return entity.Customer{}, err
	}
	cusUpdate := newCustomerRecord(cus)
	if err := updateFields(&cusr, cusUpdate); err != nil {
		return entity.Customer{}, err
	}
	query := bson.M{"$set": cusr}
	if _, err := r.collection.UpdateOne(ctx, filter, query); err != nil {
		return entity.Customer{}, err
	}
	return r.Retrieve(ctx, cus.ID)
}

func (r *customerMongoRepo) Retrieve(ctx context.Context, id string) (entity.Customer, error) {
	cusr := customerRecord{}
	filter := bson.M{"_id": id}
	if err := r.driver.findOneAndDecode(ctx, &cusr, filter); err != nil {
		return entity.Customer{}, err
	}
	return cusr.customer(), nil
}

func (r *customerMongoRepo) Search(ctx context.Context, fil controller.SearchCustomersFilter) ([]entity.Customer, error) {
	mfil := filterQuery(fil)
	fo := options.Find().SetLimit(fil.Limit).SetSkip(fil.Offset)
	res, err := r.collection.Find(ctx, mfil, fo)
	if err != nil {
		return nil, err
	}
	var cuss []entity.Customer
	for res.Next(ctx) {
		record := customerRecord{}
		if err := res.Decode(&record); err != nil {
			continue
		}
		cuss = append(cuss, record.customer())
	}
	return cuss, nil
}

func (r *customerMongoRepo) Delete(ctx context.Context, id string) (entity.Customer, error) {
	return entity.Customer{}, nil
}
