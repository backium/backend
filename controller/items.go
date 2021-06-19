package controller

import (
	"context"

	"github.com/backium/backend/entity"
	"github.com/backium/backend/errors"
)

const (
	maxReturnedItems = 50
)

type RetrieveItemRequest struct {
	ID         string
	MerchantID string
}

type DeleteItemRequest struct {
	ID         string
	MerchantID string
}

type ListAllItemsRequest struct {
	Limit       *int64
	Offset      *int64
	LocationIDs []string
	MerchantID  string
}

type ListItemsFilter struct {
	Limit       int64
	Offset      int64
	LocationIDs []string
	MerchantID  string
	IDs         []string
}

type PartialItem struct {
	Name        *string        `bson:"name,omitempty"`
	Description *string        `bson:"description,omitempty"`
	CategoryID  *string        `bson:"category_id,omitempty"`
	LocationIDs *[]string      `bson:"location_ids,omitempty"`
	Status      *entity.Status `bson:"status,omitempty"`
}

type ItemRepository interface {
	Create(context.Context, entity.Item) (string, error)
	Update(context.Context, entity.Item) error
	UpdatePartial(context.Context, string, PartialItem) error
	Retrieve(context.Context, string) (entity.Item, error)
	List(context.Context, ListItemsFilter) ([]entity.Item, error)
}

type Item struct {
	Repository ItemRepository
}

func (c *Item) Create(ctx context.Context, it entity.Item) (entity.Item, error) {
	const op = errors.Op("controller.Item.Create")
	id, err := c.Repository.Create(ctx, it)
	if err != nil {
		return entity.Item{}, err
	}
	it, err = c.Repository.Retrieve(ctx, id)
	if err != nil {
		return entity.Item{}, err
	}
	return it, nil
}

func (c *Item) Update(ctx context.Context, id string, it PartialItem) (entity.Item, error) {
	const op = errors.Op("controller.Item.Update")
	if err := c.Repository.UpdatePartial(ctx, id, it); err != nil {
		return entity.Item{}, errors.E(op, err)
	}
	uit, err := c.Repository.Retrieve(ctx, id)
	if err != nil {
		return entity.Item{}, err
	}
	return uit, nil
}

func (c *Item) Retrieve(ctx context.Context, req RetrieveItemRequest) (entity.Item, error) {
	const op = errors.Op("controller.Item.Retrieve")
	it, err := c.Repository.Retrieve(ctx, req.ID)
	if err != nil {
		return entity.Item{}, errors.E(op, err)
	}
	if it.MerchantID != req.MerchantID {
		return entity.Item{}, errors.E(op, errors.KindNotFound, "trying to retrieve an external item")
	}
	return it, nil
}

func (c *Item) ListAll(ctx context.Context, req ListAllItemsRequest) ([]entity.Item, error) {
	const op = errors.Op("controller.Item.ListAll")
	limit := int64(maxReturnedItems)
	offset := int64(0)
	if req.Limit != nil {
		limit = *req.Limit
	}
	if req.Offset != nil {
		offset = *req.Offset
	}

	its, err := c.Repository.List(ctx, ListItemsFilter{
		MerchantID: req.MerchantID,
		Limit:      limit,
		Offset:     offset,
	})
	if err != nil {
		return nil, errors.E(op, err)
	}
	return its, nil
}

func (c *Item) Delete(ctx context.Context, req DeleteItemRequest) (entity.Item, error) {
	const op = errors.Op("controller.Item.Delete")
	it, err := c.Repository.Retrieve(ctx, req.ID)
	if err != nil {
		return entity.Item{}, errors.E(op, err)
	}

	if it.MerchantID != req.MerchantID {
		return entity.Item{}, errors.E(op, errors.KindNotFound, "trying to retrieve an external item")
	}

	status := entity.StatusShadowDeleted
	update := PartialItem{Status: &status}
	if err := c.Repository.UpdatePartial(ctx, req.ID, update); err != nil {
		return entity.Item{}, errors.E(op, err)
	}
	dit, err := c.Repository.Retrieve(ctx, req.ID)
	if err != nil {
		return entity.Item{}, errors.E(op, err)
	}
	return dit, nil
}
