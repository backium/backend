package repository

import (
	"context"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoDB struct {
	client *mongo.Client
	db     *mongo.Database
}

func NewMongoDB(uri string, dbName string) (MongoDB, error) {
	c, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(uri))
	if err != nil {
		return MongoDB{}, err
	}
	mdb := MongoDB{client: c, db: c.Database(dbName)}
	return mdb, nil
}

func (m MongoDB) Collection(name string) *mongo.Collection {
	return m.db.Collection(name)
}

func (m MongoDB) Disconnect() error {
	return m.client.Disconnect(context.TODO())
}
