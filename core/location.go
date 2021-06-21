package core

import (
	"context"

	"github.com/backium/backend/errors"
)

const maxReturnedLocations = 50

type PartialLocation struct {
	Name         *string `bson:"name,omitempty"`
	BusinessName *string `bson:"business_name,omitempty"`
	Status       *Status `bson:"status,omitempty"`
}

type Location struct {
	ID           string `bson:"_id"`
	Name         string `bson:"name,omitempty"`
	BusinessName string `bson:"business_name,omitempty"`
	MerchantID   string `bson:"merchant_id,omitempty"`
	Status       Status `bson:"status,omitempty"`
}

// Creates a Customer with default values
func NewLocation() Location {
	return Location{
		Status: StatusActive,
	}
}

type LocationRepository interface {
	Create(context.Context, Location) (string, error)
	Update(context.Context, Location) error
	UpdatePartial(context.Context, string, PartialLocation) error
	Retrieve(context.Context, string) (Location, error)
	List(context.Context, ListLocationsFilter) ([]Location, error)
}

type LocationService struct {
	LocationRepository LocationRepository
}

func (svc *LocationService) Create(ctx context.Context, loc Location) (Location, error) {
	const op = errors.Op("controller.Location.Create")
	id, err := svc.LocationRepository.Create(ctx, loc)
	if err != nil {
		return Location{}, errors.E(op, err)
	}
	loc, err = svc.LocationRepository.Retrieve(ctx, id)
	if err != nil {
		return Location{}, err
	}
	return loc, nil
}

func (svc *LocationService) Update(ctx context.Context, id string, loc PartialLocation) (Location, error) {
	const op = errors.Op("controller.Location.Update")
	if err := svc.LocationRepository.UpdatePartial(ctx, id, loc); err != nil {
		return Location{}, errors.E(op, err)
	}
	uloc, err := svc.LocationRepository.Retrieve(ctx, id)
	if err != nil {
		return Location{}, err
	}
	return uloc, nil
}

func (svc *LocationService) Retrieve(ctx context.Context, req RetrieveLocationRequest) (Location, error) {
	const op = errors.Op("controller.Location.Retrieve")
	l, err := svc.LocationRepository.Retrieve(ctx, req.ID)
	if err != nil {
		return Location{}, errors.E(op, err)
	}

	if l.MerchantID != req.MerchantID {
		return Location{}, errors.E(op, errors.KindNotFound, "trying to retrieve an external location")
	}
	return l, nil
}

func (svc *LocationService) ListAll(ctx context.Context, req ListAllLocationsRequest) ([]Location, error) {
	const op = errors.Op("controller.Location.ListAll")
	limit := int64(maxReturnedLocations)
	offset := int64(0)
	if req.Limit != nil {
		limit = *req.Limit
	}
	if req.Offset != nil {
		offset = *req.Offset
	}

	locs, err := svc.LocationRepository.List(ctx, ListLocationsFilter{
		MerchantID: req.MerchantID,
		Limit:      limit,
		Offset:     offset,
	})
	if err != nil {
		return nil, errors.E(op, err)
	}
	return locs, nil
}

func (c *LocationService) Delete(ctx context.Context, req DeleteLocationRequest) (Location, error) {
	const op = errors.Op("controller.Location.Delete")
	loc, err := c.LocationRepository.Retrieve(ctx, req.ID)
	if err != nil {
		return Location{}, errors.E(op, err)
	}

	if loc.MerchantID != req.MerchantID {
		return Location{}, errors.E(op, errors.KindNotFound, "trying to retrieve an external location")
	}

	status := StatusShadowDeleted
	update := PartialLocation{Status: &status}
	if err := c.LocationRepository.UpdatePartial(ctx, req.ID, update); err != nil {
		return Location{}, errors.E(op, err)
	}
	dloc, err := c.LocationRepository.Retrieve(ctx, req.ID)
	if err != nil {
		return Location{}, err
	}
	return dloc, nil
}

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
