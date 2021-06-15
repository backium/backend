package controller

import (
	"context"
	"errors"

	"github.com/backium/backend/entity"
)

type User struct {
	Repository         UserRepository
	MerchantRepository MerchantRepository
}

func (uc *User) Create(ctx context.Context, req CreateUserRequest) (entity.User, error) {
	hash, err := entity.HashUserPassword(req.Password)
	if err != nil {
		return entity.User{}, err
	}

	// Create an owner user with merchant, locations, etc
	// TODO: Create a default location and other entities
	m, err := uc.MerchantRepository.Create(entity.Merchant{
		BusinessName: "My Business",
	})
	if err != nil {
		return entity.User{}, err
	}
	user, err := uc.Repository.Create(ctx, entity.User{
		Email:        req.Email,
		PasswordHash: hash,
		IsOwner:      true,
		MerchantID:   m.ID,
	})
	if err != nil {
		return entity.User{}, err
	}
	return user, nil
}

func (uc *User) Login(ctx context.Context, req LoginUserRequest) (entity.User, error) {
	user, err := uc.Repository.RetrieveByEmail(ctx, req.Email)
	if err != nil {
		return entity.User{}, err
	}

	if !user.PasswordEquals(req.Password) {
		return entity.User{}, errors.New("wrong password")
	}

	return user, err
}

type UserRepository interface {
	Create(context.Context, entity.User) (entity.User, error)
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
