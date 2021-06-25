package core

import (
	"context"

	"github.com/backium/backend/errors"
)

const (
	maxReturnedItemVariations     = 50
	defaultReturnedItemVariations = 10
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
func NewItemVariation() ItemVariation {
	return ItemVariation{
		ID:          NewID("itvar"),
		LocationIDs: []string{},
		Status:      StatusActive,
	}
}

type ItemVariationStorage interface {
	Put(context.Context, ItemVariation) error
	PutBatch(context.Context, []ItemVariation) error
	Get(context.Context, string, string, []string) (ItemVariation, error)
	List(context.Context, ItemVariationFilter) ([]ItemVariation, error)
}

func (s *CatalogService) PutItemVariation(ctx context.Context, it ItemVariation) (ItemVariation, error) {
	const op = errors.Op("core/CatalogService.PutItemVariation")
	if err := s.ItemVariationStorage.Put(ctx, it); err != nil {
		return ItemVariation{}, err
	}
	it, err := s.ItemVariationStorage.Get(ctx, it.ID, it.MerchantID, nil)
	if err != nil {
		return ItemVariation{}, err
	}
	return it, nil
}

func (s *CatalogService) PutItemVariationVariations(ctx context.Context, ii []ItemVariation) ([]ItemVariation, error) {
	const op = errors.Op("core/CatalogService.PutItemVariationVariations")
	if err := s.ItemVariationStorage.PutBatch(ctx, ii); err != nil {
		return nil, err
	}
	ids := make([]string, len(ii))
	for i, d := range ii {
		ids[i] = d.ID
	}
	ii, err := s.ItemVariationStorage.List(ctx, ItemVariationFilter{
		Limit: int64(len(ii)),
		IDs:   ids,
	})
	if err != nil {
		return nil, err
	}
	return ii, nil
}

func (s *CatalogService) GetItemVariation(ctx context.Context, id, merchantID string, locationIDs []string) (ItemVariation, error) {
	const op = errors.Op("core/CatalogService.GetItemVariation")
	it, err := s.ItemVariationStorage.Get(ctx, id, merchantID, locationIDs)
	if err != nil {
		return ItemVariation{}, errors.E(op, err)
	}
	return it, nil
}

func (s *CatalogService) ListItemVariation(ctx context.Context, f ItemVariationFilter) ([]ItemVariation, error) {
	const op = errors.Op("core/CatalogService.ListItemVariation")
	limit, offset := int64(defaultReturnedItemVariations), int64(0)
	if f.Limit != 0 && f.Limit < maxReturnedItemVariations {
		limit = f.Limit
	}
	if f.Offset != 0 {
		offset = f.Offset
	}

	dd, err := s.ItemVariationStorage.List(ctx, ItemVariationFilter{
		MerchantID: f.MerchantID,
		Limit:      limit,
		Offset:     offset,
		ItemIDs:    f.ItemIDs,
	})
	if err != nil {
		return nil, errors.E(op, err)
	}
	return dd, nil
}

func (s *CatalogService) DeleteItemVariation(ctx context.Context, id, merchantID string, locationIDs []string) (ItemVariation, error) {
	const op = errors.Op("core/CatalogService.DeleteItemVariation")
	d, err := s.ItemVariationStorage.Get(ctx, id, merchantID, locationIDs)
	if err != nil {
		return ItemVariation{}, errors.E(op, err)
	}

	d.Status = StatusShadowDeleted
	if err := s.ItemVariationStorage.Put(ctx, d); err != nil {
		return ItemVariation{}, errors.E(op, err)
	}
	resp, err := s.ItemVariationStorage.Get(ctx, id, merchantID, locationIDs)
	if err != nil {
		return ItemVariation{}, errors.E(op, err)
	}
	return resp, nil
}

type ItemVariationFilter struct {
	Limit       int64
	Offset      int64
	LocationIDs []string
	MerchantID  string
	ItemIDs     []string
	IDs         []string
}
