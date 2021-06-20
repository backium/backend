package controller

import (
	"context"

	"github.com/backium/backend/base"
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

type PartialLocation struct {
	Name         *string      `bson:"name,omitempty"`
	BusinessName *string      `bson:"business_name,omitempty"`
	Status       *base.Status `bson:"status,omitempty"`
}

type LocationRepository interface {
	Create(context.Context, entity.Location) (string, error)
	Update(context.Context, entity.Location) error
	UpdatePartial(context.Context, string, PartialLocation) error
	Retrieve(context.Context, string) (entity.Location, error)
	List(context.Context, ListLocationsFilter) ([]entity.Location, error)
}

type Location struct {
	Repository LocationRepository
}

func (c *Location) Create(ctx context.Context, loc entity.Location) (entity.Location, error) {
	const op = errors.Op("controller.Location.Create")
	id, err := c.Repository.Create(ctx, loc)
	if err != nil {
		return entity.Location{}, errors.E(op, err)
	}
	loc, err = c.Repository.Retrieve(ctx, id)
	if err != nil {
		return entity.Location{}, err
	}
	return loc, nil
}

func (c *Location) Update(ctx context.Context, id string, loc PartialLocation) (entity.Location, error) {
	const op = errors.Op("controller.Location.Update")
	if err := c.Repository.UpdatePartial(ctx, id, loc); err != nil {
		return entity.Location{}, errors.E(op, err)
	}
	uloc, err := c.Repository.Retrieve(ctx, id)
	if err != nil {
		return entity.Location{}, err
	}
	return uloc, nil
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
	loc, err := c.Repository.Retrieve(ctx, req.ID)
	if err != nil {
		return entity.Location{}, errors.E(op, err)
	}

	if loc.MerchantID != req.MerchantID {
		return entity.Location{}, errors.E(op, errors.KindNotFound, "trying to retrieve an external location")
	}

	status := base.StatusShadowDeleted
	update := PartialLocation{Status: &status}
	if err := c.Repository.UpdatePartial(ctx, req.ID, update); err != nil {
		return entity.Location{}, errors.E(op, err)
	}
	dloc, err := c.Repository.Retrieve(ctx, req.ID)
	if err != nil {
		return entity.Location{}, err
	}
	return dloc, nil
}
