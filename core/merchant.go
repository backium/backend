package core

import (
	"context"

	"github.com/backium/backend/errors"
)

type Merchant struct {
	ID           string `bson:"_id"`
	FirstName    string `bson:"first_name"`
	LastName     string `bson:"last_name"`
	BusinessName string `bson:"business_name"`
	CreatedAt    int64  `bson:"created_at"`
	UpdatedAt    int64  `bson:"updated_at"`
	Keys         []Key  `bson:"keys"`
}

func NewMerchant() Merchant {
	return Merchant{
		ID:   NewID("merch"),
		Keys: []Key{},
	}
}

type Key struct {
	Name  string `bson:"name"`
	Token string `bson:"token"`
}

func NewKey(name string) Key {
	return Key{
		Name:  name,
		Token: NewIDWithSize("sk", 25),
	}
}

type MerchantStorage interface {
	Put(context.Context, Merchant) error
	PutKey(context.Context, string, Key) error
	Get(context.Context, string) (Merchant, error)
	GetByKey(context.Context, string) (Merchant, error)
}

type MerchantService struct {
	MerchantStorage MerchantStorage
}

func (svc *MerchantService) PutMerchant(ctx context.Context, merchant Merchant) (Merchant, error) {
	const op = errors.Op("controller.Merchant.Create")
	if err := svc.MerchantStorage.Put(ctx, merchant); err != nil {
		return Merchant{}, err
	}
	merchant, err := svc.MerchantStorage.Get(ctx, merchant.ID)
	if err != nil {
		return Merchant{}, err
	}
	return merchant, nil
}

func (svc *MerchantService) GetMerchant(ctx context.Context, id string) (Merchant, error) {
	const op = errors.Op("core/MerchantService.GetMerchant")
	merchant, err := svc.MerchantStorage.Get(ctx, id)
	if err != nil {
		return Merchant{}, errors.E(op, err)
	}
	return merchant, nil
}

func (svc *MerchantService) CreateKey(ctx context.Context, keyName, merchantID string) (Key, error) {
	const op = errors.Op("core/MerchantService.CreateKey")
	key := NewKey(keyName)
	if err := svc.MerchantStorage.PutKey(ctx, merchantID, key); err != nil {
		return Key{}, errors.E(op, err)
	}
	return key, nil
}

func (svc *MerchantService) GetMerchantByKey(ctx context.Context, key string) (Merchant, error) {
	const op = errors.Op("core/MerchantService.GetMerchantByKey")
	merchant, err := svc.MerchantStorage.GetByKey(ctx, key)
	if err != nil {
		return Merchant{}, errors.E(op, err)
	}
	return merchant, nil
}
