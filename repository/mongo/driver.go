package mongo

import (
	"context"

	"github.com/backium/backend/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type mongoDriver struct {
	*mongo.Collection
}

func (md *mongoDriver) insertOne(ctx context.Context, document interface{}) (string, error) {
	const op = errors.Op("mongo.mongoDriver.insertOne")
	res, err := md.InsertOne(ctx, document)
	if err != nil {
		return "", errors.E(op, errors.KindUnexpected, err)
	}
	return res.InsertedID.(string), nil
}

func (md *mongoDriver) findOneAndDecode(ctx context.Context, val interface{},
	filter interface{}, opts ...*options.FindOneOptions) error {
	const op = errors.Op("mongo.mongoDriver.findOneAndDecode")
	res := md.FindOne(ctx, filter)
	if err := res.Err(); err != nil {
		return errors.E(op, errors.KindNotFound, err)
	}
	if err := res.Decode(val); err != nil {
		return errors.E(op, errors.KindUnexpected, err)
	}
	return nil
}

func updateFields(base interface{}, diff interface{}) error {
	b, err := bson.Marshal(diff)
	if err != nil {
		return err
	}
	if err := bson.Unmarshal(b, base); err != nil {
		return err
	}
	return nil
}
