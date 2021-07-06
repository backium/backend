package core

import (
	"context"

	"github.com/backium/backend/errors"
)

type Location struct {
	ID           ID     `bson:"_id"`
	Name         string `bson:"name"`
	BusinessName string `bson:"business_name"`
	Image        string `bson:"image"`
	MerchantID   ID     `bson:"merchant_id"`
	CreatedAt    int64  `bson:"created_at"`
	UpdatedAt    int64  `bson:"updated_at"`
	Status       Status `bson:"status"`
}

// Creates a Location with default values
func NewLocation(name string, merchantID ID) Location {
	return Location{
		ID:         NewID("loc"),
		Name:       name,
		Status:     StatusActive,
		MerchantID: merchantID,
	}
}

type LocationStorage interface {
	Put(context.Context, Location) error
	PutBatch(context.Context, []Location) error
	Get(context.Context, ID) (Location, error)
	List(context.Context, LocationQuery) ([]Location, int64, error)
}

type LocationService struct {
	LocationStorage LocationStorage
}

func (svc *LocationService) PutLocation(ctx context.Context, location Location) (Location, error) {
	const op = errors.Op("controller.Location.Create")

	if err := svc.LocationStorage.Put(ctx, location); err != nil {
		return Location{}, err
	}

	location, err := svc.LocationStorage.Get(ctx, location.ID)
	if err != nil {
		return Location{}, err
	}

	return location, nil
}

func (svc *LocationService) PutLocations(ctx context.Context, locations []Location) ([]Location, error) {
	const op = errors.Op("core/LocationService.PutLocations")

	if err := svc.LocationStorage.PutBatch(ctx, locations); err != nil {
		return nil, err
	}

	ids := make([]ID, len(locations))
	for i, t := range locations {
		ids[i] = t.ID
	}
	locations, _, err := svc.LocationStorage.List(ctx, LocationQuery{
		Filter: LocationFilter{IDs: ids},
	})
	if err != nil {
		return nil, err
	}

	return locations, nil
}

func (svc *LocationService) GetLocation(ctx context.Context, id ID) (Location, error) {
	const op = errors.Op("core/LocationService.GetLocation")

	location, err := svc.LocationStorage.Get(ctx, id)
	if err != nil {
		return Location{}, errors.E(op, err)
	}

	return location, nil
}

func (svc *LocationService) ListLocation(ctx context.Context, q LocationQuery) ([]Location, int64, error) {
	const op = errors.Op("core/LocationService.ListLocation")

	locations, count, err := svc.LocationStorage.List(ctx, q)
	if err != nil {
		return nil, 0, errors.E(op, err)
	}

	return locations, count, nil
}

func (svc *LocationService) DeleteLocation(ctx context.Context, id ID) (Location, error) {
	const op = errors.Op("core/LocationService.DeleteLocation")

	location, err := svc.LocationStorage.Get(ctx, id)
	if err != nil {
		return Location{}, errors.E(op, err)
	}

	location.Status = StatusShadowDeleted
	if err := svc.LocationStorage.Put(ctx, location); err != nil {
		return Location{}, errors.E(op, err)
	}

	location, err = svc.LocationStorage.Get(ctx, id)
	if err != nil {
		return Location{}, errors.E(op, err)
	}

	return location, nil
}

type LocationFilter struct {
	Name       string
	IDs        []ID
	MerchantID ID
}

type LocationSort struct {
	Name SortOrder
}

type LocationQuery struct {
	Limit  int64
	Offset int64
	Filter LocationFilter
	Sort   LocationSort
}
