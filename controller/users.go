package controller

import (
	"context"

	"github.com/backium/backend/entity"
	"github.com/backium/backend/errors"
)

type UserRepository interface {
	Create(context.Context, entity.User) (string, error)
	Retrieve(context.Context, string) (entity.User, error)
	RetrieveByEmail(context.Context, string) (entity.User, error)
}

type CreateUserRequest struct {
	Email    string
	Password string
	IsOwner  bool
}

type LoginUserRequest struct {
	Email    string
	Password string
}

type User struct {
	Repository         UserRepository
	MerchantRepository MerchantRepository
	LocationRepository LocationRepository
}

func (uc *User) Create(ctx context.Context, req CreateUserRequest) (entity.User, error) {
	const op = errors.Op("controller.User.Create")
	_, err := uc.Repository.RetrieveByEmail(ctx, req.Email)
	if err == nil {
		return entity.User{}, errors.E(op, errors.KindUserExist, "user email used")
	}
	hash, err := entity.HashUserPassword(req.Password)
	if err != nil {
		return entity.User{}, errors.E(op, errors.KindUnexpected, err)
	}

	// Create an owner user with merchant, locations, etc
	user := entity.User{}
	m, err := uc.MerchantRepository.Create(entity.Merchant{
		BusinessName: "My Business",
	})
	if err != nil {
		return user, errors.E(op, errors.KindUnexpected, err)
	}
	_, err = uc.LocationRepository.Create(ctx, entity.Location{
		Name:         "My Business",
		BusinessName: "My Business",
		MerchantID:   m.ID,
	})
	if err != nil {
		return user, errors.E(op, errors.KindUnexpected, err)
	}
	id, err := uc.Repository.Create(ctx, entity.User{
		Email:        req.Email,
		PasswordHash: hash,
		Kind:         entity.UserKindOwner,
		MerchantID:   m.ID,
	})
	if err != nil {
		return entity.User{}, errors.E(op, errors.KindUnexpected, err)
	}
	user, err = uc.Repository.Retrieve(ctx, id)
	if err != nil {
		return entity.User{}, errors.E(op, err)
	}
	return user, nil
}

func (uc *User) Login(ctx context.Context, req LoginUserRequest) (entity.User, error) {
	const op = errors.Op("controller.User.Login")
	user, err := uc.Repository.RetrieveByEmail(ctx, req.Email)
	if err != nil {
		return entity.User{}, errors.E(op, errors.KindInvalidCredentials, err)
	}
	if !user.PasswordEquals(req.Password) {
		return entity.User{}, errors.E(op, errors.KindInvalidCredentials, "invalid password")
	}
	return user, nil
}
