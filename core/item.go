package core

import (
	"context"

	"github.com/backium/backend/errors"
)

type PartialItem struct {
	Name        *string   `bson:"name,omitempty"`
	Description *string   `bson:"description,omitempty"`
	CategoryID  *string   `bson:"category_id,omitempty"`
	LocationIDs *[]string `bson:"location_ids,omitempty"`
	Status      *Status   `bson:"status,omitempty"`
}

type Item struct {
	ID          string   `bson:"_id"`
	Name        string   `bson:"name,omitempty"`
	Description string   `bson:"description,omitempty"`
	CategoryID  string   `bson:"category_id,omitempty"`
	LocationIDs []string `bson:"location_ids"`
	MerchantID  string   `bson:"merchant_id,omitempty"`
	Status      Status   `bson:"status,omitempty"`
}

// Creates an Item with default values
func NewItem() Item {
	return Item{
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

type ItemRepository interface {
	Create(context.Context, Item) (string, error)
	Update(context.Context, Item) error
	UpdatePartial(context.Context, string, PartialItem) error
	Retrieve(context.Context, string) (Item, error)
	List(context.Context, ItemFilter) ([]Item, error)
}

func (c *CatalogService) CreateItem(ctx context.Context, it Item) (Item, error) {
	const op = errors.Op("controller.Item.Create")
	id, err := c.ItemRepository.Create(ctx, it)
	if err != nil {
		return Item{}, err
	}
	it, err = c.ItemRepository.Retrieve(ctx, id)
	if err != nil {
		return Item{}, err
	}
	return it, nil
}

func (c *CatalogService) UpdateItem(ctx context.Context, id string, it PartialItem) (Item, error) {
	const op = errors.Op("controller.Item.Update")
	if err := c.ItemRepository.UpdatePartial(ctx, id, it); err != nil {
		return Item{}, errors.E(op, err)
	}
	uit, err := c.ItemRepository.Retrieve(ctx, id)
	if err != nil {
		return Item{}, err
	}
	return uit, nil
}

func (c *CatalogService) RetrieveItem(ctx context.Context, req ItemRetrieveRequest) (Item, error) {
	const op = errors.Op("controller.Item.Retrieve")
	it, err := c.ItemRepository.Retrieve(ctx, req.ID)
	if err != nil {
		return Item{}, errors.E(op, err)
	}
	if it.MerchantID != req.MerchantID {
		return Item{}, errors.E(op, errors.KindNotFound, "trying to retrieve an external item")
	}
	return it, nil
}

func (c *CatalogService) ListItem(ctx context.Context, req ItemListRequest) ([]Item, error) {
	const op = errors.Op("controller.Item.ListAll")
	limit := int64(maxReturnedItems)
	offset := int64(0)
	if req.Limit != nil {
		limit = *req.Limit
	}
	if req.Offset != nil {
		offset = *req.Offset
	}

	its, err := c.ItemRepository.List(ctx, ItemFilter{
		MerchantID: req.MerchantID,
		Limit:      limit,
		Offset:     offset,
	})
	if err != nil {
		return nil, errors.E(op, err)
	}
	return its, nil
}

func (c *CatalogService) DeleteItem(ctx context.Context, req ItemDeleteRequest) (Item, error) {
	const op = errors.Op("controller.Item.Delete")
	it, err := c.ItemRepository.Retrieve(ctx, req.ID)
	if err != nil {
		return Item{}, errors.E(op, err)
	}

	if it.MerchantID != req.MerchantID {
		return Item{}, errors.E(op, errors.KindNotFound, "trying to retrieve an external item")
	}

	status := StatusShadowDeleted
	update := PartialItem{Status: &status}
	if err := c.ItemRepository.UpdatePartial(ctx, req.ID, update); err != nil {
		return Item{}, errors.E(op, err)
	}
	dit, err := c.ItemRepository.Retrieve(ctx, req.ID)
	if err != nil {
		return Item{}, errors.E(op, err)
	}
	return dit, nil
}

type ItemRetrieveRequest struct {
	ID         string
	MerchantID string
}

type ItemDeleteRequest struct {
	ID         string
	MerchantID string
}

type ItemListRequest struct {
	Limit       *int64
	Offset      *int64
	LocationIDs []string
	MerchantID  string
}

type ItemFilter struct {
	Limit       int64
	Offset      int64
	LocationIDs []string
	MerchantID  string
	IDs         []string
}
