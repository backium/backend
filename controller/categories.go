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
type PartialCategory struct {
	Name        *string        `bson:"name,omitempty"`
	LocationIDs *[]string      `bson:"location_ids,omitempty"`
	Status      *entity.Status `bson:"status,omitempty"`
}

type CategoryRepository interface {
	Create(context.Context, entity.Category) (string, error)
	Update(context.Context, entity.Category) error
	UpdatePartial(context.Context, string, PartialCategory) error
	Retrieve(context.Context, string) (entity.Category, error)
	List(context.Context, ListCategoriesFilter) ([]entity.Category, error)
}

type Category struct {
	Repository CategoryRepository
}

func (c *Category) Create(ctx context.Context, cat entity.Category) (entity.Category, error) {
	const op = errors.Op("controller.Category.Create")
	id, err := c.Repository.Create(ctx, cat)
	if err != nil {
		return entity.Category{}, errors.E(op, err)
	}
	ncat, err := c.Repository.Retrieve(ctx, id)
	if err != nil {
		return entity.Category{}, errors.E(op, err)
	}
	return ncat, nil
}

func (c *Category) Update(ctx context.Context, id string, cat PartialCategory) (entity.Category, error) {
	const op = errors.Op("controller.Category.Update")
	if err := c.Repository.UpdatePartial(ctx, id, cat); err != nil {
		return entity.Category{}, errors.E(op, err)
	}
	ucat, err := c.Repository.Retrieve(ctx, id)
	if err != nil {
		return entity.Category{}, errors.E(op, err)
	}
	return ucat, nil
}

func (c *Category) Retrieve(ctx context.Context, req RetrieveCategoryRequest) (entity.Category, error) {
	const op = errors.Op("controller.Category.Retrieve")
	cat, err := c.Repository.Retrieve(ctx, req.ID)
	if err != nil {
		return entity.Category{}, errors.E(op, err)
	}
	if cat.MerchantID != req.MerchantID {
		return entity.Category{}, errors.E(op, errors.KindNotFound, "trying to retrieve an external category")
	}
	return cat, nil
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
	cat, err := c.Repository.Retrieve(ctx, req.ID)
	if err != nil {
		return entity.Category{}, errors.E(op, err)
	}

	if cat.MerchantID != req.MerchantID {
		return entity.Category{}, errors.E(op, errors.KindNotFound, "trying to retrieve an external category")
	}

	status := entity.StatusShadowDeleted
	update := PartialCategory{Status: &status}
	if err := c.Repository.UpdatePartial(ctx, req.ID, update); err != nil {
		return entity.Category{}, errors.E(op, err)
	}
	dcat, err := c.Repository.Retrieve(ctx, req.ID)
	if err != nil {
		return entity.Category{}, errors.E(op, err)
	}
	return dcat, nil
}
