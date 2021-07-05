package core

import (
	"context"

	"github.com/backium/backend/errors"
)

type Item struct {
	ID          ID     `bson:"_id"`
	Name        string `bson:"name,omitempty"`
	Description string `bson:"description,omitempty"`
	CategoryID  ID     `bson:"category_id,omitempty"`
	LocationIDs []ID   `bson:"location_ids"`
	MerchantID  ID     `bson:"merchant_id,omitempty"`
	CreatedAt   int64  `bson:"created_at"`
	UpdatedAt   int64  `bson:"updated_at"`
	Status      Status `bson:"status,omitempty"`
}

func NewItem(name string, categoryID, merchantID ID) Item {
	return Item{
		ID:          NewID("item"),
		Name:        name,
		CategoryID:  categoryID,
		LocationIDs: []ID{},
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
	Get(context.Context, ID) (Item, error)
	List(context.Context, ItemQuery) ([]Item, error)
}

func (s *CatalogService) PutItem(ctx context.Context, item Item) (Item, error) {
	const op = errors.Op("core/CatalogService.PutItem")

	if err := s.ItemStorage.Put(ctx, item); err != nil {
		return Item{}, err
	}

	item, err := s.ItemStorage.Get(ctx, item.ID)
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

	ids := make([]ID, len(items))
	for i, d := range items {
		ids[i] = d.ID
	}
	items, err := s.ItemStorage.List(ctx, ItemQuery{
		Filter: ItemFilter{IDs: ids},
	})
	if err != nil {
		return nil, err
	}

	return items, nil
}

func (s *CatalogService) GetItem(ctx context.Context, id ID) (Item, error) {
	const op = errors.Op("core/CatalogService.GetItem")

	item, err := s.ItemStorage.Get(ctx, id)
	if err != nil {
		return Item{}, errors.E(op, err)
	}

	return item, nil
}

func (s *CatalogService) ListItem(ctx context.Context, q ItemQuery) ([]Item, error) {
	const op = errors.Op("core/CatalogService.ListItem")

	items, err := s.ItemStorage.List(ctx, q)
	if err != nil {
		return nil, errors.E(op, err)
	}

	return items, nil
}

func (s *CatalogService) DeleteItem(ctx context.Context, id ID) (Item, error) {
	const op = errors.Op("core/CatalogService.DeleteItem")

	item, err := s.ItemStorage.Get(ctx, id)
	if err != nil {
		return Item{}, errors.E(op, err)
	}

	item.Status = StatusShadowDeleted
	if err := s.ItemStorage.Put(ctx, item); err != nil {
		return Item{}, errors.E(op, err)
	}

	item, err = s.ItemStorage.Get(ctx, id)
	if err != nil {
		return Item{}, errors.E(op, err)
	}

	return item, nil
}

type ItemFilter struct {
	Name        string
	IDs         []ID
	CategoryIDs []ID
	LocationIDs []ID
	MerchantID  ID
}

type ItemSort struct {
	Name SortOrder
}

type ItemQuery struct {
	Limit  int64
	Offset int64
	Filter ItemFilter
	Sort   ItemSort
}
