package core

import (
	"context"
	"fmt"

	"github.com/backium/backend/errors"
)

type ItemVariation struct {
	ID                   ID              `bson:"_id"`
	Name                 string          `bson:"name"`
	SKU                  string          `bson:"sku"`
	ItemID               ID              `bson:"item_id"`
	Measurement          MeasurementUnit `bson:"measurement"`
	Price                Money           `bson:"price"`
	Cost                 *Money          `bson:"cost"`
	Image                string          `bson:"image"`
	MinimumRequiredStock int64           `bson:"minimum_required_stock"`
	LocationIDs          []ID            `bson:"location_ids"`
	MerchantID           ID              `bson:"merchant_id"`
	CreatedAt            int64           `bson:"created_at"`
	UpdatedAt            int64           `bson:"updated_at"`
	Status               Status          `bson:"status"`
}

// Creates an ItemVariationVariation with default values
func NewItemVariation(name string, itemID, merchantID ID) ItemVariation {
	return ItemVariation{
		ID:          NewID("itemvar"),
		Name:        name,
		ItemID:      itemID,
		LocationIDs: []ID{},
		Status:      StatusActive,
		MerchantID:  merchantID,
	}
}

type ItemVariationStorage interface {
	Put(context.Context, ItemVariation) error
	PutBatch(context.Context, []ItemVariation) error
	Get(context.Context, ID) (ItemVariation, error)
	List(context.Context, ItemVariationQuery) ([]ItemVariation, int64, error)
}

func (s *CatalogService) PutItemVariation(ctx context.Context, variation ItemVariation) (ItemVariation, error) {
	const op = errors.Op("core/CatalogService.PutItemVariation")

	if err := s.ItemVariationStorage.Put(ctx, variation); err != nil {
		return ItemVariation{}, errors.E(op, err)
	}

	variation, err := s.ItemVariationStorage.Get(ctx, variation.ID)
	if err != nil {
		return ItemVariation{}, errors.E(op, err)
	}

	if variation.CreatedAt == variation.UpdatedAt {
		// Initialize inventory counts
		if err := s.initializeInventory(ctx, variation); err != nil {
			fmt.Printf("Problem generating inventory for item %v: %v", variation.ID, err)
		}
	}

	return variation, nil
}

func (s *CatalogService) PutItemVariationVariations(ctx context.Context, variations []ItemVariation) ([]ItemVariation, error) {
	const op = errors.Op("core/CatalogService.PutItemVariationVariations")

	if err := s.ItemVariationStorage.PutBatch(ctx, variations); err != nil {
		return nil, err
	}

	ids := make([]ID, len(variations))
	for i, d := range variations {
		ids[i] = d.ID
	}
	variations, _, err := s.ItemVariationStorage.List(ctx, ItemVariationQuery{
		Filter: ItemVariationFilter{IDs: ids},
	})
	if err != nil {
		return nil, err
	}

	return variations, nil
}

func (s *CatalogService) GetItemVariation(ctx context.Context, id ID) (ItemVariation, error) {
	const op = errors.Op("core/CatalogService.GetItemVariation")

	variation, err := s.ItemVariationStorage.Get(ctx, id)
	if err != nil {
		return ItemVariation{}, errors.E(op, err)
	}

	return variation, nil
}

func (s *CatalogService) ListItemVariation(ctx context.Context, q ItemVariationQuery) ([]ItemVariation, int64, error) {
	const op = errors.Op("core/CatalogService.ListItemVariation")

	variations, count, err := s.ItemVariationStorage.List(ctx, q)
	if err != nil {
		return nil, 0, errors.E(op, err)
	}

	return variations, count, nil
}

func (s *CatalogService) DeleteItemVariation(ctx context.Context, id ID) (ItemVariation, error) {
	const op = errors.Op("core/CatalogService.DeleteItemVariation")

	variation, err := s.ItemVariationStorage.Get(ctx, id)
	if err != nil {
		return ItemVariation{}, errors.E(op, err)
	}

	variation.Status = StatusShadowDeleted
	if err := s.ItemVariationStorage.Put(ctx, variation); err != nil {
		return ItemVariation{}, errors.E(op, err)
	}

	variation, err = s.ItemVariationStorage.Get(ctx, id)
	if err != nil {
		return ItemVariation{}, errors.E(op, err)
	}

	return variation, nil
}

type ItemVariationFilter struct {
	Name        string
	IDs         []ID
	ItemIDs     []ID
	LocationIDs []ID
	MerchantID  ID
}

type ItemVariationSort struct {
	Name SortOrder
}

type ItemVariationQuery struct {
	Limit  int64
	Offset int64
	Filter ItemVariationFilter
	Sort   ItemVariationSort
}
