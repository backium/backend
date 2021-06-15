package repository

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

type userRecord struct {
	ID           string `bson:"_id"`
	Email        string `bson:"email"`
	PasswordHash string `bson:"password_hash"`
	IsOwner      bool   `bson:"is_owner"`
	IsSuper      bool   `bson:"is_super"`
	MerchantID   string `bson:"merchant_id"`
}

type userMongoRepository struct {
	collection *mongo.Collection
}

func NewUserMongoRepository(db MongoDB) controller.UserRepository {
	return &userMongoRepository{collection: db.Collection(userCollectionName)}
}

func (r *userMongoRepository) Create(ctx context.Context, u entity.User) (entity.User, error) {
	record := userRecordFrom(u)
	record.ID = generateID(userIDPrefix)
	res, err := r.collection.InsertOne(ctx, record)
	if err != nil {
		return entity.User{}, err
	}
	id := res.InsertedID.(string)
	return r.Retrieve(ctx, id)
}

func (r *userMongoRepository) Retrieve(ctx context.Context, id string) (entity.User, error) {
	res := r.collection.FindOne(context.TODO(), bson.M{"_id": id})
	if err := res.Err(); err != nil {
		return entity.User{}, err
	}
	record := userRecord{}
	if err := res.Decode(&record); err != nil {
		return entity.User{}, err
	}
	return record.user(), nil
}

func (r *userMongoRepository) RetrieveByEmail(ctx context.Context, email string) (entity.User, error) {
	res := r.collection.FindOne(context.TODO(), bson.M{"email": email})
	if err := res.Err(); err != nil {
		return entity.User{}, err
	}
	record := userRecord{}
	if err := res.Decode(&record); err != nil {
		return entity.User{}, err
	}
	return record.user(), nil
}

func (ur userRecord) user() entity.User {
	return entity.User{
		ID:           ur.ID,
		Email:        ur.ID,
		PasswordHash: ur.PasswordHash,
		IsOwner:      ur.IsOwner,
		IsSuper:      ur.IsSuper,
		MerchantID:   ur.MerchantID,
	}
}

func userRecordFrom(u entity.User) userRecord {
	return userRecord{
		ID:           u.ID,
		Email:        u.Email,
		PasswordHash: u.PasswordHash,
		IsOwner:      u.IsOwner,
		IsSuper:      u.IsSuper,
		MerchantID:   u.MerchantID,
	}
}
