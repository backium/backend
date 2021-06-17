package controller

import (
	"context"

	"github.com/backium/backend/entity"
	"github.com/backium/backend/errors"
)

const (
	maxReturnedLocations = 50
)

type RetrieveLocationRequest struct {
	ID         string
	MerchantID string
}

type DeleteLocationRequest struct {
	ID         string
	MerchantID string
}

type ListAllLocationsRequest struct {
	Limit      *int64
	Offset     *int64
	MerchantID string
}

type ListLocationsFilter struct {
	Limit      int64
	Offset     int64
	MerchantID string
	IDs        []string
}

type LocationRepository interface {
	Create(context.Context, entity.Location) (entity.Location, error)
	Update(context.Context, entity.Location) (entity.Location, error)
	Retrieve(context.Context, string) (entity.Location, error)
	List(context.Context, ListLocationsFilter) ([]entity.Location, error)
	Delete(context.Context, string) (entity.Location, error)
}

type Location struct {
	Repository LocationRepository
}

func (c *Location) Create(ctx context.Context, l entity.Location) (entity.Location, error) {
	const op = errors.Op("controller.Location.Create")
	loc, err := c.Repository.Create(ctx, l)
	if err != nil {
		return loc, errors.E(op, err)
	}
	return loc, nil
}

func (c *Location) Update(ctx context.Context, l entity.Location) (entity.Location, error) {
	const op = errors.Op("controller.Location.Update")
	loc, err := c.Repository.Update(ctx, l)
	if err != nil {
		return loc, errors.E(op, err)
	}
	return loc, nil
}

func (c *Location) Retrieve(ctx context.Context, req RetrieveLocationRequest) (entity.Location, error) {
	const op = errors.Op("controller.Location.Retrieve")
	l, err := c.Repository.Retrieve(ctx, req.ID)
	if err != nil {
		return entity.Location{}, errors.E(op, err)
	}

	if l.MerchantID != req.MerchantID {
		return entity.Location{}, errors.E(op, errors.KindNotFound, "trying to retrieve an external location")
	}
	return l, nil
}

func (c *Location) ListAll(ctx context.Context, req ListAllLocationsRequest) ([]entity.Location, error) {
	const op = errors.Op("controller.Location.ListAll")
	limit := int64(maxReturnedLocations)
	offset := int64(0)
	if req.Limit != nil {
		limit = *req.Limit
	}
	if req.Offset != nil {
		offset = *req.Offset
	}

	locs, err := c.Repository.List(ctx, ListLocationsFilter{
		MerchantID: req.MerchantID,
		Limit:      limit,
		Offset:     offset,
	})
	if err != nil {
		return nil, errors.E(op, err)
	}
	return locs, nil
}

func (c *Location) Delete(ctx context.Context, req DeleteLocationRequest) (entity.Location, error) {
	const op = errors.Op("controller.Location.Delete")
	l, err := c.Repository.Retrieve(ctx, req.ID)
	if err != nil {
		return entity.Location{}, errors.E(op, err)
	}

	if l.MerchantID != req.MerchantID {
		return entity.Location{}, errors.E(op, errors.KindNotFound, "trying to retrieve an external location")
	}

	loc, err := c.Repository.Update(ctx, entity.Location{
		ID:     req.ID,
		Status: entity.StatusShadowDeleted,
	})
	if err != nil {
		return loc, errors.E(op, err)
	}
	return loc, nil
}
