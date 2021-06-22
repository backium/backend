package mongo

import (
	"context"

	"github.com/backium/backend/core"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

const (
	merchantIDPrefix       = "merch"
	merchantCollectionName = "merchants"
)

type merchantRepository struct {
	collection *mongo.Collection
}

func NewMerchantRepository(db DB) core.MerchantRepository {
	return &merchantRepository{collection: db.Collection(merchantCollectionName)}
}

func (r *merchantRepository) Create(m core.Merchant) (core.Merchant, error) {
	m.ID = generateID(merchantIDPrefix)
	res, err := r.collection.InsertOne(context.TODO(), m)
	if err != nil {
		return core.Merchant{}, err
	}
	id := res.InsertedID.(string)
	return r.Retrieve(id)
}

func (r *merchantRepository) Update(m core.Merchant) (core.Merchant, error) {
	query := bson.M{"$set": m}
	_, err := r.collection.UpdateByID(context.TODO(), m.ID, query)
	if err != nil {
		return core.Merchant{}, err
	}
	return r.Retrieve(m.ID)
}

func (r *merchantRepository) Retrieve(id string) (core.Merchant, error) {
	m := core.Merchant{}
	res := r.collection.FindOne(context.TODO(), bson.M{"_id": id})
	if err := res.Err(); err != nil {
		return core.Merchant{}, err
	}
	if err := res.Decode(&m); err != nil {
		return core.Merchant{}, err
	}
	return m, nil
}

func (r *merchantRepository) ListAll() ([]core.Merchant, error) {
	res, err := r.collection.Find(context.TODO(), bson.M{})
	if err != nil {
		return nil, err
	}
	var ms []core.Merchant
	if err := res.All(context.TODO(), &ms); err != nil {
		return nil, err
	}
	return ms, nil
}

func (r *merchantRepository) Delete(id string) (core.Merchant, error) {
	return core.Merchant{}, nil
}
