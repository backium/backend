package repository

import (
	"context"

	"github.com/backium/backend/controller"
	"github.com/backium/backend/entity"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

const (
	customerIDPrefix       = "loc"
	customerCollectionName = "customers"
)

type customerRecord struct {
	ID         string        `bson:"_id"`
	Name       string        `bson:"name"`
	Email      string        `bson:"email"`
	Phone      string        `bson:"phone"`
	Address    AddressRecord `bson:"address"`
	MerchantID string        `bson:"merchant_id"`
}

func customerRecordFrom(c entity.Customer) customerRecord {
	return customerRecord{
		ID:    c.ID,
		Name:  c.Name,
		Email: c.Email,
		Phone: c.Phone,
		Address: AddressRecord{
			Line1:      c.Address.Line1,
			Line2:      c.Address.Line2,
			Province:   c.Address.Province,
			District:   c.Address.District,
			Department: c.Address.Department,
		},
		MerchantID: c.MerchantID,
	}
}

func (c customerRecord) customer() entity.Customer {
	return entity.Customer{
		ID:    c.ID,
		Name:  c.Name,
		Email: c.Email,
		Phone: c.Phone,
		Address: entity.Address{
			Line1:      c.Address.Line1,
			Line2:      c.Address.Line2,
			Province:   c.Address.Province,
			District:   c.Address.District,
			Department: c.Address.Department,
		},
		MerchantID: c.MerchantID,
	}
}

func (c customerRecord) updateQuery() bson.M {
	query := bson.M{}
	if c.Name != "" {
		query["name"] = c.Name
	}
	if c.Email != "" {
		query["email"] = c.Email
	}
	if c.Phone != "" {
		query["phone"] = c.Phone
	}
	if c.Address.Line1 != "" {
		query["address.line1"] = c.Address.Line1
	}
	if c.Address.Line2 != "" {
		query["address.line2"] = c.Address.Line2
	}
	if c.Address.Province != "" {
		query["address.province"] = c.Address.Province
	}
	if c.Address.Line1 != "" {
		query["address.line1"] = c.Address.Line1
	}
	return bson.M{"$set": query}
}

type customerMongoRepository struct {
	collection *mongo.Collection
}

func NewCustomerMongoRepository(db MongoDB) controller.CustomerRepository {
	return &customerMongoRepository{collection: db.Collection(customerCollectionName)}
}

func (r *customerMongoRepository) Create(ctx context.Context, l entity.Customer) (entity.Customer, error) {
	record := customerRecordFrom(l)
	record.ID = generateID(customerIDPrefix)
	res, err := r.collection.InsertOne(ctx, record)
	if err != nil {
		return entity.Customer{}, err
	}
	id := res.InsertedID.(string)
	return r.Retrieve(ctx, id)
}

func (r *customerMongoRepository) Update(ctx context.Context, cus entity.Customer) (entity.Customer, error) {
	query, err := r.updateQuery(ctx, cus)
	if err != nil {
		return entity.Customer{}, nil
	}

	if _, err := r.collection.UpdateByID(ctx, cus.ID, bson.M{"$set": merge}); err != nil {
		return entity.Customer{}, err
	}
	return r.Retrieve(ctx, cus.ID)
}

func (r *customerMongoRepository) UpdateByMerchantID(ctx context.Context, l entity.Customer) (entity.Customer, error) {
	record := customerRecordFrom(l)
	filter := bson.M{"_id": l.ID, "merchant_id": l.MerchantID}
	_, err := r.collection.UpdateOne(ctx, filter, record.updateQuery())
	if err != nil {
		return entity.Customer{}, err
	}
	return r.Retrieve(ctx, l.ID)
}

func (r *customerMongoRepository) Retrieve(ctx context.Context, id string) (entity.Customer, error) {
	res := r.collection.FindOne(ctx, bson.M{"_id": id})
	if err := res.Err(); err != nil {
		return entity.Customer{}, err
	}
	record := customerRecord{}
	if err := res.Decode(&record); err != nil {
		return entity.Customer{}, err
	}
	return record.customer(), nil
}

func (r *customerMongoRepository) ListAll(ctx context.Context, merchantID string) ([]entity.Customer, error) {
	res, err := r.collection.Find(ctx, bson.M{"merchant_id": merchantID})
	if err != nil {
		return nil, err
	}
	var ms []entity.Customer
	for res.Next(context.TODO()) {
		record := customerRecord{}
		if err := res.Decode(&record); err != nil {
			continue
		}
		ms = append(ms, record.customer())
	}
	return ms, nil
}

func (r *customerMongoRepository) Delete(ctx context.Context, id string) (entity.Customer, error) {
	return entity.Customer{}, nil
}

func (r *customerMongoRepository) updateQuery(ctx context.Context, cu entity.Customer) (bson.M, error) {
	record := customerRecordFrom(cu)
	b, err := bson.Marshal(record)
	if err != nil {
		return nil, err
	}
	var merge customerRecord
	res := r.collection.FindOne(ctx, bson.M{"_id": cu.ID})
	if err := res.Decode(&merge); err != nil {
		return nil, err
	}
	if err := bson.Unmarshal(b, &merge); err != nil {
		return nil, err
	}
	return bson.M{"$set": merge}, nil
}
