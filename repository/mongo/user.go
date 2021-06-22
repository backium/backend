package mongo

import (
	"context"

	"github.com/backium/backend/core"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

const (
	userIDPrefix       = "user"
	userCollectionName = "users"
)

type userRepository struct {
	collection *mongo.Collection
	driver     *mongoDriver
}

func NewUserRepository(db DB) core.UserRepository {
	coll := db.Collection(userCollectionName)
	return &userRepository{
		collection: coll,
		driver:     &mongoDriver{Collection: coll},
	}
}

func (r *userRepository) Create(ctx context.Context, u core.User) (string, error) {
	u.ID = generateID(userIDPrefix)
	res, err := r.collection.InsertOne(ctx, u)
	if err != nil {
		return "", err
	}
	id := res.InsertedID.(string)
	return id, nil
}

func (r *userRepository) Retrieve(ctx context.Context, id string) (core.User, error) {
	ur := core.User{}
	fil := bson.M{"_id": id}
	if err := r.driver.findOneAndDecode(ctx, &ur, fil); err != nil {
		return core.User{}, err
	}
	return ur, nil
}

func (r *userRepository) RetrieveByEmail(ctx context.Context, email string) (core.User, error) {
	ur := core.User{}
	fil := bson.M{"email": email}
	if err := r.driver.findOneAndDecode(ctx, &ur, fil); err != nil {
		return core.User{}, err
	}
	return ur, nil
}
