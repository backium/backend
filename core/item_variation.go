package core

import (
	"context"
	"fmt"

	"github.com/backium/backend/errors"
)

type ItemVariation struct {
	ID          string   `bson:"_id"`
	Name        string   `bson:"name"`
	SKU         string   `bson:"sku"`
	ItemID      string   `bson:"item_id"`
	Price       Money    `bson:"price"`
	Image       string   `bson:"image"`
	LocationIDs []string `bson:"location_ids"`
	MerchantID  string   `bson:"merchant_id"`
	CreatedAt   int64    `bson:"created_at"`
	UpdatedAt   int64    `bson:"updated_at"`
	Status      Status   `bson:"status"`
}

// Creates an ItemVariationVariation with default values
func NewItemVariation(name, itemID, merchantID string) ItemVariation {
	return ItemVariation{
		ID:          NewID("itemvar"),
		Name:        name,
		ItemID:      itemID,
		LocationIDs: []string{},
		Status:      StatusActive,
		MerchantID:  merchantID,
	}
}

type ItemVariationStorage interface {
	Put(context.Context, ItemVariation) error
	PutBatch(context.Context, []ItemVariation) error
	Get(context.Context, string, string, []string) (ItemVariation, error)
	List(context.Context, ItemVariationFilter) ([]ItemVariation, error)
}

func (s *CatalogService) PutItemVariation(ctx context.Context, variation ItemVariation) (ItemVariation, error) {
	const op = errors.Op("core/CatalogService.PutItemVariation")

	if err := s.ItemVariationStorage.Put(ctx, variation); err != nil {
		return ItemVariation{}, errors.E(op, err)
	}

	variation, err := s.ItemVariationStorage.Get(ctx, variation.ID, variation.MerchantID, nil)
	if err != nil {
		return ItemVariation{}, errors.E(op, err)
	}

	// Initialize inventory counts
	if err := s.initializeInventory(ctx, variation); err != nil {
		fmt.Printf("Problem generating inventory for item %v: %v", variation.ID, err)
	}

	return variation, nil
}

func (s *CatalogService) PutItemVariationVariations(ctx context.Context, variations []ItemVariation) ([]ItemVariation, error) {
	const op = errors.Op("core/CatalogService.PutItemVariationVariations")

	if err := s.ItemVariationStorage.PutBatch(ctx, variations); err != nil {
		return nil, err
	}

	ids := make([]string, len(variations))
	for i, d := range variations {
		ids[i] = d.ID
	}
	variations, err := s.ItemVariationStorage.List(ctx, ItemVariationFilter{
		Limit: int64(len(variations)),
		IDs:   ids,
	})
	if err != nil {
		return nil, err
	}

	return variations, nil
}

func (s *CatalogService) GetItemVariation(ctx context.Context, id, merchantID string, locationIDs []string) (ItemVariation, error) {
	const op = errors.Op("core/CatalogService.GetItemVariation")

	variation, err := s.ItemVariationStorage.Get(ctx, id, merchantID, locationIDs)
	if err != nil {
		return ItemVariation{}, errors.E(op, err)
	}

	return variation, nil
}

func (s *CatalogService) ListItemVariation(ctx context.Context, f ItemVariationFilter) ([]ItemVariation, error) {
	const op = errors.Op("core/CatalogService.ListItemVariation")

	variations, err := s.ItemVariationStorage.List(ctx, ItemVariationFilter{
		MerchantID: f.MerchantID,
		Limit:      f.Limit,
		Offset:     f.Offset,
		ItemIDs:    f.ItemIDs,
	})
	if err != nil {
		return nil, errors.E(op, err)
	}

	return variations, nil
}

func (s *CatalogService) DeleteItemVariation(ctx context.Context, id, merchantID string, locationIDs []string) (ItemVariation, error) {
	const op = errors.Op("core/CatalogService.DeleteItemVariation")

	variation, err := s.ItemVariationStorage.Get(ctx, id, merchantID, locationIDs)
	if err != nil {
		return ItemVariation{}, errors.E(op, err)
	}

	variation.Status = StatusShadowDeleted
	if err := s.ItemVariationStorage.Put(ctx, variation); err != nil {
		return ItemVariation{}, errors.E(op, err)
	}

	variation, err = s.ItemVariationStorage.Get(ctx, id, merchantID, locationIDs)
	if err != nil {
		return ItemVariation{}, errors.E(op, err)
	}

	return variation, nil
}

type ItemVariationFilter struct {
	Limit       int64
	Offset      int64
	LocationIDs []string
	MerchantID  string
	ItemIDs     []string
	IDs         []string
}
