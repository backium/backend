package mongo

import (
	"context"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type DB struct {
	client *mongo.Client
	db     *mongo.Database
}

func New(uri string, dbName string) (DB, error) {
	c, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(uri))
	if err != nil {
		return DB{}, err
	}
	mdb := DB{client: c, db: c.Database(dbName)}
	return mdb, nil
}

func (m DB) Client() *mongo.Client {
	return m.client
}

func (m DB) Collection(name string) *mongo.Collection {
	return m.db.Collection(name)
}

func (m DB) Disconnect() error {
	return m.client.Disconnect(context.TODO())
}
