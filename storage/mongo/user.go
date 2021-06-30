package mongo

import (
	"context"
	"fmt"

	"github.com/backium/backend/core"
	"github.com/backium/backend/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	userIDPrefix       = "user"
	userCollectionName = "users"
)

type userStorage struct {
	collection *mongo.Collection
	driver     *mongoDriver
}

func NewUserRepository(db DB) core.UserStorage {
	coll := db.Collection(userCollectionName)
	return &userStorage{
		collection: coll,
		driver:     &mongoDriver{Collection: coll},
	}
}

func (s *userStorage) Put(ctx context.Context, user core.User) error {
	const op = errors.Op("mongo/userStorage.Put")
	filter := bson.M{
		"_id":         user.ID,
		"merchant_id": user.MerchantID,
	}
	query := bson.M{"$set": user}
	opts := options.Update().SetUpsert(true)
	_, err := s.collection.UpdateOne(ctx, filter, query, opts)
	if err != nil {
		return errors.E(op, errors.KindUnexpected, err)
	}
	return nil
}

func (s *userStorage) Get(ctx context.Context, id string) (core.User, error) {
	const op = errors.Op("mongo/userStorage/Get")
	user := core.User{}
	filter := bson.M{"_id": id}
	if err := s.driver.findOneAndDecode(ctx, &user, filter); err != nil {
		return core.User{}, errors.E(op, err)
	}
	return user, nil
}

func (s *userStorage) GetByEmail(ctx context.Context, email string) (core.User, error) {
	const op = errors.Op("mongo/userStorage/Get")
	user := core.User{}
	filter := bson.M{"email": email}
	fmt.Println("filter", filter)
	if err := s.driver.findOneAndDecode(ctx, &user, filter); err != nil {
		fmt.Println("err", err)
		return core.User{}, errors.E(op, err)
	}
	return user, nil
}