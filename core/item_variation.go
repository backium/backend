package core

import (
	"context"

	"github.com/backium/backend/errors"
)

const maxReturnedItemVariations = 50

type PartialItemVariation struct {
	Name        *string   `bson:"name,omitempty"`
	SKU         *string   `bson:"sku,omitempty"`
	Price       *Money    `bson:"price,omitempty"`
	LocationIDs *[]string `bson:"location_ids,omitempty"`
	Status      *Status   `bson:"status,omitempty"`
}

type ItemVariation struct {
	ID          string   `bson:"_id"`
	Name        string   `bson:"name,omitempty"`
	SKU         string   `bson:"sku,omitempty"`
	ItemID      string   `bson:"item_id,omitempty"`
	Price       Money    `bson:"price"`
	LocationIDs []string `bson:"location_ids"`
	MerchantID  string   `bson:"merchant_id,omitempty"`
	Status      Status   `bson:"status,omitempty"`
}

// Creates an ItemVariation with default values
func NewItemVariation() ItemVariation {
	return ItemVariation{
		LocationIDs: []string{},
		Status:      StatusActive,
	}
}

type ItemVariationRepository interface {
	Create(context.Context, ItemVariation) (string, error)
	Update(context.Context, ItemVariation) error
	UpdatePartial(context.Context, string, PartialItemVariation) error
	Retrieve(context.Context, string) (ItemVariation, error)
	List(context.Context, ItemVariationFilter) ([]ItemVariation, error)
}

type ItemVariationService struct {
	ItemVariationRepository ItemVariationRepository
}

func (c *ItemVariationService) CreateItemVariation(ctx context.Context, itvar ItemVariation) (ItemVariation, error) {
	const op = errors.Op("controller.ItemVariation.Create")
	id, err := c.ItemVariationRepository.Create(ctx, itvar)
	if err != nil {
		return ItemVariation{}, err
	}
	uitvar, err := c.ItemVariationRepository.Retrieve(ctx, id)
	if err != nil {
		return ItemVariation{}, err
	}
	return uitvar, nil
}

func (c *ItemVariationService) UpdateItemVariation(ctx context.Context, id string, itvar PartialItemVariation) (ItemVariation, error) {
	const op = errors.Op("controller.ItemVariation.Update")
	if err := c.ItemVariationRepository.UpdatePartial(ctx, id, itvar); err != nil {
		return ItemVariation{}, errors.E(op, err)
	}
	uitvar, err := c.ItemVariationRepository.Retrieve(ctx, id)
	if err != nil {
		return ItemVariation{}, err
	}
	return uitvar, nil
}

func (c *ItemVariationService) RetrieveItemVariation(ctx context.Context, req ItemVariationRetrieveRequest) (ItemVariation, error) {
	const op = errors.Op("controller.ItemVariation.Retrieve")
	itvar, err := c.ItemVariationRepository.Retrieve(ctx, req.ID)
	if err != nil {
		return ItemVariation{}, errors.E(op, err)
	}

	if itvar.MerchantID != req.MerchantID {
		return ItemVariation{}, errors.E(op, errors.KindNotFound, "trying to retrieve an external itemVariation")
	}
	return itvar, nil
}

func (c *ItemVariationService) ListItemVariation(ctx context.Context, req ItemVariationListRequest) ([]ItemVariation, error) {
	const op = errors.Op("controller.ItemVariation.ListAll")
	limit := int64(maxReturnedItemVariations)
	offset := int64(0)
	if req.Limit != nil {
		limit = *req.Limit
	}
	if req.Offset != nil {
		offset = *req.Offset
	}

	itvars, err := c.ItemVariationRepository.List(ctx, ItemVariationFilter{
		MerchantID: req.MerchantID,
		Limit:      limit,
		Offset:     offset,
	})
	if err != nil {
		return nil, errors.E(op, err)
	}
	return itvars, nil
}

func (c *ItemVariationService) DeleteItemVariation(ctx context.Context, req ItemVariationDeleteRequest) (ItemVariation, error) {
	const op = errors.Op("controller.ItemVariation.Delete")
	itvar, err := c.ItemVariationRepository.Retrieve(ctx, req.ID)
	if err != nil {
		return ItemVariation{}, errors.E(op, err)
	}

	if itvar.MerchantID != req.MerchantID {
		return ItemVariation{}, errors.E(op, errors.KindNotFound, "trying to retrieve an external itemVariation")
	}

	status := StatusShadowDeleted
	update := PartialItemVariation{Status: &status}
	if err := c.ItemVariationRepository.UpdatePartial(ctx, req.ID, update); err != nil {
		return ItemVariation{}, errors.E(op, err)
	}
	ditvar, err := c.ItemVariationRepository.Retrieve(ctx, req.ID)
	if err != nil {
		return ItemVariation{}, errors.E(op, err)
	}
	return ditvar, nil
}

type ItemVariationRetrieveRequest struct {
	ID         string
	MerchantID string
}

type ItemVariationDeleteRequest struct {
	ID         string
	MerchantID string
}

type ItemVariationListRequest struct {
	Limit       *int64
	Offset      *int64
	LocationIDs []string
	MerchantID  string
}

type ItemVariationFilter struct {
	Limit       int64
	Offset      int64
	LocationIDs []string
	MerchantID  string
	IDs         []string
}
