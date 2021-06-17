package mongo

import (
	"context"

	"github.com/backium/backend/controller"
	"github.com/backium/backend/entity"
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

func NewUserRepository(db DB) controller.UserRepository {
	coll := db.Collection(userCollectionName)
	return &userRepository{
		collection: coll,
		driver:     &mongoDriver{Collection: coll},
	}
}

func (r *userRepository) Create(ctx context.Context, u entity.User) (entity.User, error) {
	u.ID = generateID(userIDPrefix)
	res, err := r.collection.InsertOne(ctx, u)
	if err != nil {
		return entity.User{}, err
	}
	id := res.InsertedID.(string)
	return r.Retrieve(ctx, id)
}

func (r *userRepository) Retrieve(ctx context.Context, id string) (entity.User, error) {
	ur := entity.User{}
	fil := bson.M{"_id": id}
	if err := r.driver.findOneAndDecode(ctx, &ur, fil); err != nil {
		return entity.User{}, err
	}
	return ur, nil
}

func (r *userRepository) RetrieveByEmail(ctx context.Context, email string) (entity.User, error) {
	ur := entity.User{}
	fil := bson.M{"email": email}
	if err := r.driver.findOneAndDecode(ctx, &ur, fil); err != nil {
		return entity.User{}, err
	}
	return ur, nil
}
