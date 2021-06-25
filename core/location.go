package core

import (
	"context"

	"github.com/backium/backend/errors"
)

const (
	maxReturnedLocations     = 50
	defaultReturnedLocations = 10
)

type Location struct {
	ID           string `bson:"_id"`
	Name         string `bson:"name"`
	BusinessName string `bson:"business_name"`
	Image        string `bson:"image"`
	MerchantID   string `bson:"merchant_id"`
	CreatedAt    int64  `bson:"created_at"`
	UpdatedAt    int64  `bson:"updated_at"`
	Status       Status `bson:"status"`
}

// Creates a Location with default values
func NewLocation() Location {
	return Location{
		ID:     NewID("loc"),
		Status: StatusActive,
	}
}

type LocationStorage interface {
	Put(context.Context, Location) error
	PutBatch(context.Context, []Location) error
	Get(context.Context, string, string) (Location, error)
	List(context.Context, LocationFilter) ([]Location, error)
}

type LocationService struct {
	LocationStorage LocationStorage
}

func (svc *LocationService) PutLocation(ctx context.Context, location Location) (Location, error) {
	const op = errors.Op("controller.Location.Create")
	if err := svc.LocationStorage.Put(ctx, location); err != nil {
		return Location{}, err
	}
	location, err := svc.LocationStorage.Get(ctx, location.ID, location.MerchantID)
	if err != nil {
		return Location{}, err
	}
	return location, nil
}

func (svc *LocationService) PutLocations(ctx context.Context, cc []Location) ([]Location, error) {
	const op = errors.Op("core/LocationService.PutLocations")
	if err := svc.LocationStorage.PutBatch(ctx, cc); err != nil {
		return nil, err
	}
	ids := make([]string, len(cc))
	for i, t := range cc {
		ids[i] = t.ID
	}
	cc, err := svc.LocationStorage.List(ctx, LocationFilter{
		Limit: int64(len(cc)),
		IDs:   ids,
	})
	if err != nil {
		return nil, err
	}
	return cc, nil
}

func (svc *LocationService) GetLocation(ctx context.Context, id, merchantID string) (Location, error) {
	const op = errors.Op("core/LocationService.GetLocation")
	cust, err := svc.LocationStorage.Get(ctx, id, merchantID)
	if err != nil {
		return Location{}, errors.E(op, err)
	}
	return cust, nil
}

func (svc *LocationService) ListLocation(ctx context.Context, f LocationFilter) ([]Location, error) {
	const op = errors.Op("core/LocationService.ListLocation")
	limit, offset := int64(defaultReturnedLocations), int64(0)
	if f.Limit != 0 && f.Limit < maxReturnedLocations {
		limit = f.Limit
	}
	if f.Offset != 0 {
		offset = f.Offset
	}

	cc, err := svc.LocationStorage.List(ctx, LocationFilter{
		MerchantID: f.MerchantID,
		Limit:      limit,
		Offset:     offset,
	})
	if err != nil {
		return nil, errors.E(op, err)
	}
	return cc, nil
}

func (svc *LocationService) DeleteLocation(ctx context.Context, id, merchantID string) (Location, error) {
	const op = errors.Op("core/LocationService.DeleteLocation")
	location, err := svc.LocationStorage.Get(ctx, id, merchantID)
	if err != nil {
		return Location{}, errors.E(op, err)
	}

	location.Status = StatusShadowDeleted
	if err := svc.LocationStorage.Put(ctx, location); err != nil {
		return Location{}, errors.E(op, err)
	}
	resp, err := svc.LocationStorage.Get(ctx, id, merchantID)
	if err != nil {
		return Location{}, errors.E(op, err)
	}
	return resp, nil
}

type LocationFilter struct {
	Limit      int64
	Offset     int64
	MerchantID string
	IDs        []string
}
