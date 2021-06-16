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

type customerMongoRepository struct {
	collection *mongo.Collection
	driver     *mongoDriver
}

func NewCustomerMongoRepository(db MongoDB) controller.CustomerRepository {
	collection := db.Collection(customerCollectionName)
	return &customerMongoRepository{
		collection: collection,
		driver: &mongoDriver{
			Collection: collection,
		},
	}
}

func (r *customerMongoRepository) Create(ctx context.Context, cus entity.Customer) (entity.Customer, error) {
	record := newCustomerRecord(cus)
	record.ID = generateID(customerIDPrefix)
	res, err := r.collection.InsertOne(ctx, record)
	if err != nil {
		return entity.Customer{}, err
	}
	id := res.InsertedID.(string)
	return r.Retrieve(ctx, id)
}

func (r *customerMongoRepository) Update(ctx context.Context, cus entity.Customer) (entity.Customer, error) {
	rec := customerRecord{}
	filter := bson.M{"_id": cus.ID}
	if err := r.driver.findOneAndDecode(ctx, &rec, filter); err != nil {
		return entity.Customer{}, err
	}
	if err := updateFields(&rec, newCustomerRecord(cus)); err != nil {
		return entity.Customer{}, err
	}
	query := bson.M{"$set": rec}
	if _, err := r.collection.UpdateOne(ctx, filter, query); err != nil {
		return entity.Customer{}, err
	}
	return r.Retrieve(ctx, cus.ID)
}

func (r *customerMongoRepository) Retrieve(ctx context.Context, id string) (entity.Customer, error) {
	rec := customerRecord{}
	filter := bson.M{"_id": id}
	if err := r.driver.findOneAndDecode(ctx, &rec, filter); err != nil {
		return entity.Customer{}, err
	}
	return rec.customer(), nil
}

func (r *customerMongoRepository) ListAll(ctx context.Context, merchantID string) ([]entity.Customer, error) {
	mfil := bson.M{
		"merchant_id": merchantID,
	}
	res, err := r.collection.Find(ctx, mfil)
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

func (r *customerMongoRepository) Search(ctx context.Context, fil controller.SearchCustomersFilter) ([]entity.Customer, error) {
	mfil := filterQuery(fil)
	fo := options.Find()
	fo.SetLimit(fil.Limit)
	fo.SetSkip(fil.Offset)
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

func (r *customerMongoRepository) Delete(ctx context.Context, id string) (entity.Customer, error) {
	return entity.Customer{}, nil
}
