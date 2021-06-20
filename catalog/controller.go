package catalog

import (
	"context"

	"github.com/backium/backend/base"
	"github.com/backium/backend/errors"
)

const (
	maxReturnedCategories     = 50
	maxReturnedItems          = 50
	maxReturnedItemVariations = 50
	maxReturnedTaxes          = 50
)

type Controller struct {
	CategoryRepository      CategoryRepository
	ItemRepository          ItemRepository
	ItemVariationRepository ItemVariationRepository
	TaxRepository           TaxRepository
}

func (c *Controller) CreateCategory(ctx context.Context, cat Category) (Category, error) {
	const op = errors.Op("controller.Category.Create")
	id, err := c.CategoryRepository.Create(ctx, cat)
	if err != nil {
		return Category{}, errors.E(op, err)
	}
	ncat, err := c.CategoryRepository.Retrieve(ctx, id)
	if err != nil {
		return Category{}, errors.E(op, err)
	}
	return ncat, nil
}

func (c *Controller) UpdateCategory(ctx context.Context, id string, cat CategoryPartial) (Category, error) {
	const op = errors.Op("controller.Category.Update")
	if err := c.CategoryRepository.UpdatePartial(ctx, id, cat); err != nil {
		return Category{}, errors.E(op, err)
	}
	ucat, err := c.CategoryRepository.Retrieve(ctx, id)
	if err != nil {
		return Category{}, errors.E(op, err)
	}
	return ucat, nil
}

func (c *Controller) RetrieveCategory(ctx context.Context, req CategoryRetrieveRequest) (Category, error) {
	const op = errors.Op("controller.Category.Retrieve")
	cat, err := c.CategoryRepository.Retrieve(ctx, req.ID)
	if err != nil {
		return Category{}, errors.E(op, err)
	}
	if cat.MerchantID != req.MerchantID {
		return Category{}, errors.E(op, errors.KindNotFound, "trying to retrieve an external category")
	}
	return cat, nil
}

func (c *Controller) ListCategory(ctx context.Context, req CategoryListRequest) ([]Category, error) {
	const op = errors.Op("controller.Category.ListAll")
	limit := int64(maxReturnedCategories)
	offset := int64(0)
	if req.Limit != nil {
		limit = *req.Limit
	}
	if req.Offset != nil {
		offset = *req.Offset
	}

	cuss, err := c.CategoryRepository.List(ctx, CategoryFilter{
		MerchantID: req.MerchantID,
		Limit:      limit,
		Offset:     offset,
	})
	if err != nil {
		return nil, errors.E(op, err)
	}
	return cuss, nil
}

func (c *Controller) DeleteCategory(ctx context.Context, req CategoryDeleteRequest) (Category, error) {
	const op = errors.Op("controller.Category.Delete")
	cat, err := c.CategoryRepository.Retrieve(ctx, req.ID)
	if err != nil {
		return Category{}, errors.E(op, err)
	}

	if cat.MerchantID != req.MerchantID {
		return Category{}, errors.E(op, errors.KindNotFound, "trying to retrieve an external category")
	}

	status := base.StatusShadowDeleted
	update := CategoryPartial{Status: &status}
	if err := c.CategoryRepository.UpdatePartial(ctx, req.ID, update); err != nil {
		return Category{}, errors.E(op, err)
	}
	dcat, err := c.CategoryRepository.Retrieve(ctx, req.ID)
	if err != nil {
		return Category{}, errors.E(op, err)
	}
	return dcat, nil
}

func (c *Controller) CreateItem(ctx context.Context, it Item) (Item, error) {
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

func (c *Controller) UpdateItem(ctx context.Context, id string, it PartialItem) (Item, error) {
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

func (c *Controller) RetrieveItem(ctx context.Context, req ItemRetrieveRequest) (Item, error) {
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

func (c *Controller) ListItem(ctx context.Context, req ItemListRequest) ([]Item, error) {
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

func (c *Controller) DeleteItem(ctx context.Context, req ItemDeleteRequest) (Item, error) {
	const op = errors.Op("controller.Item.Delete")
	it, err := c.ItemRepository.Retrieve(ctx, req.ID)
	if err != nil {
		return Item{}, errors.E(op, err)
	}

	if it.MerchantID != req.MerchantID {
		return Item{}, errors.E(op, errors.KindNotFound, "trying to retrieve an external item")
	}

	status := base.StatusShadowDeleted
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

func (c *Controller) CreateItemVariation(ctx context.Context, itvar ItemVariation) (ItemVariation, error) {
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

func (c *Controller) UpdateItemVariation(ctx context.Context, id string, itvar PartialItemVariation) (ItemVariation, error) {
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

func (c *Controller) RetrieveItemVariation(ctx context.Context, req ItemVariationRetrieveRequest) (ItemVariation, error) {
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

func (c *Controller) ListItemVariation(ctx context.Context, req ItemVariationListRequest) ([]ItemVariation, error) {
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

func (c *Controller) DeleteItemVariation(ctx context.Context, req ItemVariationDeleteRequest) (ItemVariation, error) {
	const op = errors.Op("controller.ItemVariation.Delete")
	itvar, err := c.ItemVariationRepository.Retrieve(ctx, req.ID)
	if err != nil {
		return ItemVariation{}, errors.E(op, err)
	}

	if itvar.MerchantID != req.MerchantID {
		return ItemVariation{}, errors.E(op, errors.KindNotFound, "trying to retrieve an external itemVariation")
	}

	status := base.StatusShadowDeleted
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

func (c *Controller) CreateTax(ctx context.Context, it Tax) (Tax, error) {
	const op = errors.Op("controller.Tax.Create")
	id, err := c.TaxRepository.Create(ctx, it)
	if err != nil {
		return Tax{}, err
	}
	it, err = c.TaxRepository.Retrieve(ctx, id)
	if err != nil {
		return Tax{}, err
	}
	return it, nil
}

func (c *Controller) UpdateTax(ctx context.Context, id string, it TaxPartial) (Tax, error) {
	const op = errors.Op("controller.Tax.Update")
	if err := c.TaxRepository.UpdatePartial(ctx, id, it); err != nil {
		return Tax{}, errors.E(op, err)
	}
	uit, err := c.TaxRepository.Retrieve(ctx, id)
	if err != nil {
		return Tax{}, err
	}
	return uit, nil
}

func (c *Controller) RetrieveTax(ctx context.Context, req TaxRetrieveRequest) (Tax, error) {
	const op = errors.Op("controller.Tax.Retrieve")
	it, err := c.TaxRepository.Retrieve(ctx, req.ID)
	if err != nil {
		return Tax{}, errors.E(op, err)
	}
	if it.MerchantID != req.MerchantID {
		return Tax{}, errors.E(op, errors.KindNotFound, "trying to retrieve an external tax")
	}
	return it, nil
}

func (c *Controller) ListTax(ctx context.Context, req TaxListRequest) ([]Tax, error) {
	const op = errors.Op("controller.Tax.ListAll")
	limit := int64(maxReturnedTaxes)
	offset := int64(0)
	if req.Limit != nil {
		limit = *req.Limit
	}
	if req.Offset != nil {
		offset = *req.Offset
	}

	its, err := c.TaxRepository.List(ctx, TaxFilter{
		MerchantID: req.MerchantID,
		Limit:      limit,
		Offset:     offset,
	})
	if err != nil {
		return nil, errors.E(op, err)
	}
	return its, nil
}

func (c *Controller) DeleteTax(ctx context.Context, req TaxDeleteRequest) (Tax, error) {
	const op = errors.Op("controller.Tax.Delete")
	it, err := c.TaxRepository.Retrieve(ctx, req.ID)
	if err != nil {
		return Tax{}, errors.E(op, err)
	}

	if it.MerchantID != req.MerchantID {
		return Tax{}, errors.E(op, errors.KindNotFound, "trying to retrieve an external tax")
	}

	status := base.StatusShadowDeleted
	update := TaxPartial{Status: &status}
	if err := c.TaxRepository.UpdatePartial(ctx, req.ID, update); err != nil {
		return Tax{}, errors.E(op, err)
	}
	dit, err := c.TaxRepository.Retrieve(ctx, req.ID)
	if err != nil {
		return Tax{}, errors.E(op, err)
	}
	return dit, nil
}

type CategoryRetrieveRequest struct {
	ID         string
	MerchantID string
}

type CategoryDeleteRequest struct {
	ID         string
	MerchantID string
}

type CategoryListRequest struct {
	Limit       *int64
	Offset      *int64
	LocationIDs []string
	MerchantID  string
}

type CategoryFilter struct {
	Limit       int64
	Offset      int64
	LocationIDs []string
	MerchantID  string
	IDs         []string
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

type TaxRetrieveRequest struct {
	ID         string
	MerchantID string
}

type TaxDeleteRequest struct {
	ID         string
	MerchantID string
}

type TaxListRequest struct {
	Limit       *int64
	Offset      *int64
	LocationIDs []string
	MerchantID  string
}

type TaxFilter struct {
	Limit       int64
	Offset      int64
	LocationIDs []string
	MerchantID  string
	IDs         []string
}
