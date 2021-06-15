package controller

import (
	"context"
	"errors"

	"github.com/backium/backend/entity"
)

type RetrieveLocationRequest struct {
	ID         string
	MerchantID string
}

type LocationRepository interface {
	Create(context.Context, entity.Location) (entity.Location, error)
	Update(context.Context, entity.Location) (entity.Location, error)
	UpdateByMerchantID(context.Context, entity.Location) (entity.Location, error)
	Retrieve(context.Context, string) (entity.Location, error)
	ListAll(context.Context, string) ([]entity.Location, error)
	Delete(context.Context, string) (entity.Location, error)
}

type Location struct {
	Repository LocationRepository
}

func (c *Location) Create(ctx context.Context, l entity.Location) (entity.Location, error) {
	return c.Repository.Create(ctx, l)
}

func (c *Location) Update(ctx context.Context, l entity.Location) (entity.Location, error) {
	return c.Repository.UpdateByMerchantID(ctx, l)
}

func (c *Location) Retrieve(ctx context.Context, req RetrieveLocationRequest) (entity.Location, error) {
	l, err := c.Repository.Retrieve(ctx, req.ID)
	if err != nil {
		return entity.Location{}, err
	}

	if l.MerchantID != req.MerchantID {
		return entity.Location{}, errors.New("external location")
	}
	return l, nil
}

func (c *Location) ListAll(ctx context.Context, merchantID string) ([]entity.Location, error) {
	return c.Repository.ListAll(ctx, merchantID)
}
