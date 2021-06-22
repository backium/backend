package core

import (
	"context"

	"github.com/backium/backend/errors"
	"golang.org/x/crypto/bcrypt"
)

type UserKind string

const (
	UserKindOwner    UserKind = "owner"
	UserKindEmployee UserKind = "employee"
	UserKindSuper    UserKind = "super"
)

type User struct {
	ID           string   `bson:"_id"`
	Email        string   `bson:"email,omitempty"`
	PasswordHash string   `bson:"password_hash,omitempty"`
	Kind         UserKind `bson:"kind,omitempty"`
	MerchantID   string   `bson:"merchant_id,omitempty"`
}

func (u *User) PasswordEquals(password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)) == nil
}

func HashUserPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(hash), err
}

type UserRepository interface {
	Create(context.Context, User) (string, error)
	Retrieve(context.Context, string) (User, error)
	RetrieveByEmail(context.Context, string) (User, error)
}

type UserService struct {
	UserRepository     UserRepository
	MerchantRepository MerchantRepository
	LocationRepository LocationRepository
}

func (svc *UserService) Create(ctx context.Context, req UserCreateRequest) (User, error) {
	const op = errors.Op("controller.User.Create")
	_, err := svc.UserRepository.RetrieveByEmail(ctx, req.Email)
	if err == nil {
		return User{}, errors.E(op, errors.KindUserExist, "user email used")
	}
	hash, err := HashUserPassword(req.Password)
	if err != nil {
		return User{}, errors.E(op, errors.KindUnexpected, err)
	}

	// Create an owner user with merchant, locations, etc
	user := User{}
	m, err := svc.MerchantRepository.Create(Merchant{
		BusinessName: "My Business",
	})
	if err != nil {
		return user, errors.E(op, errors.KindUnexpected, err)
	}
	_, err = svc.LocationRepository.Create(ctx, Location{
		Name:         "My Business",
		BusinessName: "My Business",
		MerchantID:   m.ID,
	})
	if err != nil {
		return user, errors.E(op, errors.KindUnexpected, err)
	}
	id, err := svc.UserRepository.Create(ctx, User{
		Email:        req.Email,
		PasswordHash: hash,
		Kind:         UserKindOwner,
		MerchantID:   m.ID,
	})
	if err != nil {
		return User{}, errors.E(op, errors.KindUnexpected, err)
	}
	user, err = svc.UserRepository.Retrieve(ctx, id)
	if err != nil {
		return User{}, errors.E(op, err)
	}
	return user, nil
}

func (svc *UserService) Login(ctx context.Context, req UserLoginRequest) (User, error) {
	const op = errors.Op("controller.User.Login")
	user, err := svc.UserRepository.RetrieveByEmail(ctx, req.Email)
	if err != nil {
		return User{}, errors.E(op, errors.KindInvalidCredentials, err)
	}
	if !user.PasswordEquals(req.Password) {
		return User{}, errors.E(op, errors.KindInvalidCredentials, "invalid password")
	}
	return user, nil
}

type UserCreateRequest struct {
	Email    string
	Password string
	IsOwner  bool
}

type UserLoginRequest struct {
	Email    string
	Password string
}