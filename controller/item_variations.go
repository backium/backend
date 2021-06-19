package controller

import (
	"context"

	"github.com/backium/backend/entity"
	"github.com/backium/backend/errors"
)

const (
	maxReturnedItemVariations = 50
)

type RetrieveItemVariationRequest struct {
	ID         string
	MerchantID string
}

type DeleteItemVariationRequest struct {
	ID         string
	MerchantID string
}

type ListAllItemVariationsRequest struct {
	Limit       *int64
	Offset      *int64
	LocationIDs []string
	MerchantID  string
}

type ListItemVariationsFilter struct {
	Limit       int64
	Offset      int64
	LocationIDs []string
	MerchantID  string
	IDs         []string
}

// PartialItemVariation is a partial representation of an ItemVariation, useful for partial updates.
type PartialItemVariation struct {
	Name        *string        `bson:"name,omitempty"`
	SKU         *string        `bson:"sku,omitempty"`
	Price       *entity.Money  `bson:"price,omitempty"`
	LocationIDs *[]string      `bson:"location_ids,omitempty"`
	Status      *entity.Status `bson:"status,omitempty"`
}

type ItemVariationRepository interface {
	Create(context.Context, entity.ItemVariation) (string, error)
	Update(context.Context, entity.ItemVariation) error
	UpdatePartial(context.Context, string, PartialItemVariation) error
	Retrieve(context.Context, string) (entity.ItemVariation, error)
	List(context.Context, ListItemVariationsFilter) ([]entity.ItemVariation, error)
}

type ItemVariation struct {
	Repository ItemVariationRepository
}

func (c *ItemVariation) Create(ctx context.Context, itvar entity.ItemVariation) (entity.ItemVariation, error) {
	const op = errors.Op("controller.ItemVariation.Create")
	id, err := c.Repository.Create(ctx, itvar)
	if err != nil {
		return entity.ItemVariation{}, err
	}
	uitvar, err := c.Repository.Retrieve(ctx, id)
	if err != nil {
		return entity.ItemVariation{}, err
	}
	return uitvar, nil
}

func (c *ItemVariation) Update(ctx context.Context, id string, itvar PartialItemVariation) (entity.ItemVariation, error) {
	const op = errors.Op("controller.ItemVariation.Update")
	if err := c.Repository.UpdatePartial(ctx, id, itvar); err != nil {
		return entity.ItemVariation{}, errors.E(op, err)
	}
	uitvar, err := c.Repository.Retrieve(ctx, id)
	if err != nil {
		return entity.ItemVariation{}, err
	}
	return uitvar, nil
}

func (c *ItemVariation) Retrieve(ctx context.Context, req RetrieveItemVariationRequest) (entity.ItemVariation, error) {
	const op = errors.Op("controller.ItemVariation.Retrieve")
	itvar, err := c.Repository.Retrieve(ctx, req.ID)
	if err != nil {
		return entity.ItemVariation{}, errors.E(op, err)
	}

	if itvar.MerchantID != req.MerchantID {
		return entity.ItemVariation{}, errors.E(op, errors.KindNotFound, "trying to retrieve an external itemVariation")
	}
	return itvar, nil
}

func (c *ItemVariation) ListAll(ctx context.Context, req ListAllItemVariationsRequest) ([]entity.ItemVariation, error) {
	const op = errors.Op("controller.ItemVariation.ListAll")
	limit := int64(maxReturnedItemVariations)
	offset := int64(0)
	if req.Limit != nil {
		limit = *req.Limit
	}
	if req.Offset != nil {
		offset = *req.Offset
	}

	itvars, err := c.Repository.List(ctx, ListItemVariationsFilter{
		MerchantID: req.MerchantID,
		Limit:      limit,
		Offset:     offset,
	})
	if err != nil {
		return nil, errors.E(op, err)
	}
	return itvars, nil
}

func (c *ItemVariation) Delete(ctx context.Context, req DeleteItemVariationRequest) (entity.ItemVariation, error) {
	const op = errors.Op("controller.ItemVariation.Delete")
	itvar, err := c.Repository.Retrieve(ctx, req.ID)
	if err != nil {
		return entity.ItemVariation{}, errors.E(op, err)
	}

	if itvar.MerchantID != req.MerchantID {
		return entity.ItemVariation{}, errors.E(op, errors.KindNotFound, "trying to retrieve an external itemVariation")
	}

	status := entity.StatusShadowDeleted
	update := PartialItemVariation{Status: &status}
	if err := c.Repository.UpdatePartial(ctx, req.ID, update); err != nil {
		return entity.ItemVariation{}, errors.E(op, err)
	}
	ditvar, err := c.Repository.Retrieve(ctx, req.ID)
	if err != nil {
		return entity.ItemVariation{}, errors.E(op, err)
	}
	return ditvar, nil
}
