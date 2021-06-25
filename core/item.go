package core

import (
	"context"

	"github.com/backium/backend/errors"
)

const (
	maxReturnedItems     = 50
	defaultReturnedItems = 10
)

type Item struct {
	ID          string   `bson:"_id"`
	Name        string   `bson:"name,omitempty"`
	Description string   `bson:"description,omitempty"`
	CategoryID  string   `bson:"category_id,omitempty"`
	LocationIDs []string `bson:"location_ids"`
	MerchantID  string   `bson:"merchant_id,omitempty"`
	CreatedAt   int64    `bson:"created_at"`
	UpdatedAt   int64    `bson:"updated_at"`
	Status      Status   `bson:"status,omitempty"`
}

// Creates an Item with default values
func NewItem() Item {
	return Item{
		ID:          NewID("item"),
		LocationIDs: []string{},
		Status:      StatusActive,
	}
}

func (it *Item) ItemVariations(vars []ItemVariation) []ItemVariation {
	itemVars := []ItemVariation{}
	for _, itvar := range vars {
		if itvar.ItemID == it.ID {
			itemVars = append(itemVars, itvar)
		}
	}
	return itemVars
}

type ItemStorage interface {
	Put(context.Context, Item) error
	PutBatch(context.Context, []Item) error
	Get(context.Context, string, string, []string) (Item, error)
	List(context.Context, ItemFilter) ([]Item, error)
}

func (s *CatalogService) PutItem(ctx context.Context, it Item) (Item, error) {
	const op = errors.Op("core/CatalogService.PutItem")
	if err := s.ItemStorage.Put(ctx, it); err != nil {
		return Item{}, err
	}
	it, err := s.ItemStorage.Get(ctx, it.ID, it.MerchantID, nil)
	if err != nil {
		return Item{}, err
	}
	return it, nil
}

func (s *CatalogService) PutItems(ctx context.Context, ii []Item) ([]Item, error) {
	const op = errors.Op("core/CatalogService.PutItems")
	if err := s.ItemStorage.PutBatch(ctx, ii); err != nil {
		return nil, err
	}
	ids := make([]string, len(ii))
	for i, d := range ii {
		ids[i] = d.ID
	}
	ii, err := s.ItemStorage.List(ctx, ItemFilter{
		Limit: int64(len(ii)),
		IDs:   ids,
	})
	if err != nil {
		return nil, err
	}
	return ii, nil
}

func (s *CatalogService) GetItem(ctx context.Context, id, merchantID string, locationIDs []string) (Item, error) {
	const op = errors.Op("core/CatalogService.GetItem")
	it, err := s.ItemStorage.Get(ctx, id, merchantID, locationIDs)
	if err != nil {
		return Item{}, errors.E(op, err)
	}
	return it, nil
}

func (s *CatalogService) ListItem(ctx context.Context, f ItemFilter) ([]Item, error) {
	const op = errors.Op("core/CatalogService.ListItem")
	limit, offset := int64(defaultReturnedItems), int64(0)
	if f.Limit != 0 && f.Limit < maxReturnedItems {
		limit = f.Limit
	}
	if f.Offset != 0 {
		offset = f.Offset
	}

	dd, err := s.ItemStorage.List(ctx, ItemFilter{
		MerchantID: f.MerchantID,
		Limit:      limit,
		Offset:     offset,
	})
	if err != nil {
		return nil, errors.E(op, err)
	}
	return dd, nil
}

func (s *CatalogService) DeleteItem(ctx context.Context, id, merchantID string, locationIDs []string) (Item, error) {
	const op = errors.Op("core/CatalogService.DeleteItem")
	d, err := s.ItemStorage.Get(ctx, id, merchantID, locationIDs)
	if err != nil {
		return Item{}, errors.E(op, err)
	}

	d.Status = StatusShadowDeleted
	if err := s.ItemStorage.Put(ctx, d); err != nil {
		return Item{}, errors.E(op, err)
	}
	resp, err := s.ItemStorage.Get(ctx, id, merchantID, locationIDs)
	if err != nil {
		return Item{}, errors.E(op, err)
	}
	return resp, nil
}

type ItemFilter struct {
	Limit       int64
	Offset      int64
	LocationIDs []string
	MerchantID  string
	IDs         []string
}
