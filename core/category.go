package core

import (
	"context"

	"github.com/backium/backend/errors"
)

const (
	maxReturnedCategories = 50
)

type CategoryRepository interface {
	Create(context.Context, Category) (string, error)
	Update(context.Context, Category) error
	UpdatePartial(context.Context, string, CategoryPartial) error
	Retrieve(context.Context, string) (Category, error)
	List(context.Context, CategoryFilter) ([]Category, error)
}

type CategoryPartial struct {
	Name        *string   `bson:"name,omitempty"`
	LocationIDs *[]string `bson:"location_ids,omitempty"`
	Status      *Status   `bson:"status,omitempty"`
}

type Category struct {
	ID          string   `bson:"_id"`
	Name        string   `bson:"name,omitempty"`
	LocationIDs []string `bson:"location_ids"`
	MerchantID  string   `bson:"merchant_id,omitempty"`
	Status      Status   `bson:"status,omitempty"`
}

func NewCategory() Category {
	return Category{
		LocationIDs: []string{},
		Status:      StatusActive,
	}
}

type CategoryService struct {
	CategoryRepository CategoryRepository
}

func (c *CategoryService) CreateCategory(ctx context.Context, cat Category) (Category, error) {
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

func (c *CategoryService) UpdateCategory(ctx context.Context, id string, cat CategoryPartial) (Category, error) {
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

func (c *CategoryService) RetrieveCategory(ctx context.Context, req CategoryRetrieveRequest) (Category, error) {
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

func (c *CategoryService) ListCategory(ctx context.Context, req CategoryListRequest) ([]Category, error) {
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

func (c *CategoryService) DeleteCategory(ctx context.Context, req CategoryDeleteRequest) (Category, error) {
	const op = errors.Op("controller.Category.Delete")
	cat, err := c.CategoryRepository.Retrieve(ctx, req.ID)
	if err != nil {
		return Category{}, errors.E(op, err)
	}

	if cat.MerchantID != req.MerchantID {
		return Category{}, errors.E(op, errors.KindNotFound, "trying to retrieve an external category")
	}

	status := StatusShadowDeleted
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
