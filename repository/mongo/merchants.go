package mongo

import (
	"context"

	"github.com/backium/backend/controller"
	"github.com/backium/backend/entity"
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

func NewMerchantRepository(db DB) controller.MerchantRepository {
	return &merchantRepository{collection: db.Collection(merchantCollectionName)}
}

func (r *merchantRepository) Create(m entity.Merchant) (entity.Merchant, error) {
	m.ID = generateID(merchantIDPrefix)
	res, err := r.collection.InsertOne(context.TODO(), m)
	if err != nil {
		return entity.Merchant{}, err
	}
	id := res.InsertedID.(string)
	return r.Retrieve(id)
}

func (r *merchantRepository) Update(m entity.Merchant) (entity.Merchant, error) {
	query := bson.M{"$set": m}
	_, err := r.collection.UpdateByID(context.TODO(), m.ID, query)
	if err != nil {
		return entity.Merchant{}, err
	}
	return r.Retrieve(m.ID)
}

func (r *merchantRepository) Retrieve(id string) (entity.Merchant, error) {
	m := entity.Merchant{}
	res := r.collection.FindOne(context.TODO(), bson.M{"_id": id})
	if err := res.Err(); err != nil {
		return entity.Merchant{}, err
	}
	if err := res.Decode(&m); err != nil {
		return entity.Merchant{}, err
	}
	return m, nil
}

func (r *merchantRepository) ListAll() ([]entity.Merchant, error) {
	res, err := r.collection.Find(context.TODO(), bson.M{})
	if err != nil {
		return nil, err
	}
	var ms []entity.Merchant
	if err := res.All(context.TODO(), &ms); err != nil {
		return nil, err
	}
	return ms, nil
}

func (r *merchantRepository) Delete(id string) (entity.Merchant, error) {
	return entity.Merchant{}, nil
}
