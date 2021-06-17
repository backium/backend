package controller

import (
	"context"

	"github.com/backium/backend/entity"
	"github.com/backium/backend/errors"
)

const (
	maxReturnedCategories = 50
)

type RetrieveCategoryRequest struct {
	ID         string
	MerchantID string
}

type DeleteCategoryRequest struct {
	ID         string
	MerchantID string
}

type ListAllCategoriesRequest struct {
	Limit       *int64
	Offset      *int64
	LocationIDs []string
	MerchantID  string
}

type ListCategoriesFilter struct {
	Limit       int64
	Offset      int64
	LocationIDs []string
	MerchantID  string
	IDs         []string
}

type CategoryRepository interface {
	Create(context.Context, entity.Category) (entity.Category, error)
	Update(context.Context, entity.Category) (entity.Category, error)
	Retrieve(context.Context, string) (entity.Category, error)
	List(context.Context, ListCategoriesFilter) ([]entity.Category, error)
	Delete(context.Context, string) (entity.Category, error)
}

type Category struct {
	Repository CategoryRepository
}

func (c *Category) Create(ctx context.Context, cus entity.Category) (entity.Category, error) {
	const op = errors.Op("controller.Category.Create")
	cus, err := c.Repository.Create(ctx, cus)
	if err != nil {
		return cus, errors.E(op, err)
	}
	return cus, nil
}

func (c *Category) Update(ctx context.Context, cus entity.Category) (entity.Category, error) {
	const op = errors.Op("controller.Category.Update")
	loc, err := c.Repository.Update(ctx, cus)
	if err != nil {
		return loc, errors.E(op, err)
	}
	return loc, nil
}

func (c *Category) Retrieve(ctx context.Context, req RetrieveCategoryRequest) (entity.Category, error) {
	const op = errors.Op("controller.Category.Retrieve")
	cus, err := c.Repository.Retrieve(ctx, req.ID)
	if err != nil {
		return entity.Category{}, errors.E(op, err)
	}

	if cus.MerchantID != req.MerchantID {
		return entity.Category{}, errors.E(op, errors.KindNotFound, "trying to retrieve an external category")
	}
	return cus, nil
}

func (c *Category) ListAll(ctx context.Context, req ListAllCategoriesRequest) ([]entity.Category, error) {
	const op = errors.Op("controller.Category.ListAll")
	limit := int64(maxReturnedCategories)
	offset := int64(0)
	if req.Limit != nil {
		limit = *req.Limit
	}
	if req.Offset != nil {
		offset = *req.Offset
	}

	cuss, err := c.Repository.List(ctx, ListCategoriesFilter{
		MerchantID: req.MerchantID,
		Limit:      limit,
		Offset:     offset,
	})
	if err != nil {
		return nil, errors.E(op, err)
	}
	return cuss, nil
}

func (c *Category) Delete(ctx context.Context, req DeleteCategoryRequest) (entity.Category, error) {
	const op = errors.Op("controller.Category.Delete")
	cus, err := c.Repository.Retrieve(ctx, req.ID)
	if err != nil {
		return entity.Category{}, errors.E(op, err)
	}

	if cus.MerchantID != req.MerchantID {
		return entity.Category{}, errors.E(op, errors.KindNotFound, "trying to retrieve an external category")
	}

	loc, err := c.Repository.Update(ctx, entity.Category{
		ID:     req.ID,
		Status: entity.StatusShadowDeleted,
	})
	if err != nil {
		return loc, errors.E(op, err)
	}
	return loc, nil
}
