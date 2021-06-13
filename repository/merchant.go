package repository

import (
	"context"

	"github.com/backium/backend/merchants"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

const (
	merchantIDPrefix       = "merch"
	merchantCollectionName = "merchants"
)

type merchantRecord struct {
	ID           string `bson:"_id"`
	FirstName    string `bson:"first_name"`
	LastName     string `bson:"last_name"`
	BusinessName string `bson:"business_name"`
}

type merchantMongoRepository struct {
	collection *mongo.Collection
}

func NewMerchantMongoRepository(db MongoDB) merchants.MerchantRepository {
	return &merchantMongoRepository{collection: db.Collection(merchantCollectionName)}
}

func (r *merchantMongoRepository) Create(m merchants.Merchant) (merchants.Merchant, error) {
	record := merchantRecordFrom(m)
	record.ID = generateID(merchantIDPrefix)
	res, err := r.collection.InsertOne(context.TODO(), record)
	if err != nil {
		return merchants.Merchant{}, err
	}
	id := res.InsertedID.(string)
	return r.Retrieve(id)
}

func (r *merchantMongoRepository) Update(m merchants.Merchant) (merchants.Merchant, error) {
	record := merchantRecordFrom(m)
	_, err := r.collection.UpdateByID(context.TODO(), m.ID, record.updateQuery())
	if err != nil {
		return merchants.Merchant{}, err
	}
	return r.Retrieve(m.ID)
}

func (r *merchantMongoRepository) Retrieve(id string) (merchants.Merchant, error) {
	res := r.collection.FindOne(context.TODO(), bson.M{"_id": id})
	if err := res.Err(); err != nil {
		return merchants.Merchant{}, err
	}
	record := merchantRecord{}
	if err := res.Decode(&record); err != nil {
		return merchants.Merchant{}, err
	}
	return record.merchant(), nil
}

func (r *merchantMongoRepository) ListAll() ([]merchants.Merchant, error) {
	res, err := r.collection.Find(context.TODO(), bson.M{})
	if err != nil {
		return nil, err
	}
	var ms []merchants.Merchant
	for res.Next(context.TODO()) {
		record := merchantRecord{}
		if err := res.Decode(&record); err != nil {
			continue
		}
		ms = append(ms, record.merchant())
	}
	return ms, nil
}

func (r *merchantMongoRepository) Delete(id string) (merchants.Merchant, error) {
	return merchants.Merchant{}, nil
}

func (m merchantRecord) merchant() merchants.Merchant {
	return merchants.Merchant{
		ID:           m.ID,
		FirstName:    m.FirstName,
		LastName:     m.LastName,
		BusinessName: m.BusinessName,
	}
}

func (m merchantRecord) updateQuery() bson.M {
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

func merchantRecordFrom(m merchants.Merchant) merchantRecord {
	return merchantRecord{
		FirstName:    m.FirstName,
		LastName:     m.LastName,
		BusinessName: m.BusinessName,
	}
}
