package core

import (
	"context"

	"github.com/backium/backend/errors"
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
func NewItem(merchantID string) Item {
	return Item{
		ID:          NewID("item"),
		LocationIDs: []string{},
		Status:      StatusActive,
		MerchantID:  merchantID,
	}
}

// ItemVariations returns the variations that belong to the item
func (it *Item) ItemVariations(variations []ItemVariation) []ItemVariation {
	itemVariations := []ItemVariation{}
	for _, variation := range variations {
		if variation.ItemID == it.ID {
			itemVariations = append(itemVariations, variation)
		}
	}
	return itemVariations
}

type ItemStorage interface {
	Put(context.Context, Item) error
	PutBatch(context.Context, []Item) error
	Get(context.Context, string, string, []string) (Item, error)
	List(context.Context, ItemFilter) ([]Item, error)
}

func (s *CatalogService) PutItem(ctx context.Context, item Item) (Item, error) {
	const op = errors.Op("core/CatalogService.PutItem")
	if err := s.ItemStorage.Put(ctx, item); err != nil {
		return Item{}, err
	}
	item, err := s.ItemStorage.Get(ctx, item.ID, item.MerchantID, nil)
	if err != nil {
		return Item{}, err
	}
	return item, nil
}

func (s *CatalogService) PutItems(ctx context.Context, items []Item) ([]Item, error) {
	const op = errors.Op("core/CatalogService.PutItems")
	if err := s.ItemStorage.PutBatch(ctx, items); err != nil {
		return nil, err
	}
	ids := make([]string, len(items))
	for i, d := range items {
		ids[i] = d.ID
	}
	items, err := s.ItemStorage.List(ctx, ItemFilter{
		Limit: int64(len(items)),
		IDs:   ids,
	})
	if err != nil {
		return nil, err
	}
	return items, nil
}

func (s *CatalogService) GetItem(ctx context.Context, id, merchantID string, locationIDs []string) (Item, error) {
	const op = errors.Op("core/CatalogService.GetItem")
	item, err := s.ItemStorage.Get(ctx, id, merchantID, locationIDs)
	if err != nil {
		return Item{}, errors.E(op, err)
	}
	return item, nil
}

func (s *CatalogService) ListItem(ctx context.Context, f ItemFilter) ([]Item, error) {
	const op = errors.Op("core/CatalogService.ListItem")
	items, err := s.ItemStorage.List(ctx, ItemFilter{
		MerchantID: f.MerchantID,
		Limit:      f.Limit,
		Offset:     f.Offset,
	})
	if err != nil {
		return nil, errors.E(op, err)
	}
	return items, nil
}

func (s *CatalogService) DeleteItem(ctx context.Context, id, merchantID string, locationIDs []string) (Item, error) {
	const op = errors.Op("core/CatalogService.DeleteItem")
	item, err := s.ItemStorage.Get(ctx, id, merchantID, locationIDs)
	if err != nil {
		return Item{}, errors.E(op, err)
	}

	item.Status = StatusShadowDeleted
	if err := s.ItemStorage.Put(ctx, item); err != nil {
		return Item{}, errors.E(op, err)
	}
	item, err = s.ItemStorage.Get(ctx, id, merchantID, locationIDs)
	if err != nil {
		return Item{}, errors.E(op, err)
	}
	return item, nil
}

type ItemFilter struct {
	Limit       int64
	Offset      int64
	LocationIDs []string
	MerchantID  string
	IDs         []string
}
