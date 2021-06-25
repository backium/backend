package core

import (
	"context"

	"github.com/backium/backend/errors"
)

const (
	maxReturnedCategories     = 50
	defaultReturnedCategories = 10
)

type Category struct {
	ID          string   `bson:"_id"`
	Name        string   `bson:"name"`
	Image       string   `bson:"image"`
	LocationIDs []string `bson:"location_ids"`
	MerchantID  string   `bson:"merchant_id"`
	CreatedAt   int64    `bson:"created_at"`
	UpdatedAt   int64    `bson:"updated_at"`
	Status      Status   `bson:"status"`
}

func NewCategory() Category {
	return Category{
		ID:          generateID("cat"),
		LocationIDs: []string{},
		Status:      StatusActive,
	}
}

type CategoryStorage interface {
	Put(context.Context, Category) error
	PutBatch(context.Context, []Category) error
	Get(context.Context, string, string, []string) (Category, error)
	List(context.Context, CategoryFilter) ([]Category, error)
}

func (svc *CatalogService) PutCategory(ctx context.Context, t Category) (Category, error) {
	const op = errors.Op("core/CatalogService.PutCategory")
	if err := svc.CategoryStorage.Put(ctx, t); err != nil {
		return Category{}, err
	}
	t, err := svc.CategoryStorage.Get(ctx, t.ID, t.MerchantID, t.LocationIDs)
	if err != nil {
		return Category{}, err
	}
	return t, nil
}

func (svc *CatalogService) PutCategories(ctx context.Context, tt []Category) ([]Category, error) {
	const op = errors.Op("core/CatalogService.PutCategories")
	if err := svc.CategoryStorage.PutBatch(ctx, tt); err != nil {
		return nil, err
	}
	ids := make([]string, len(tt))
	for i, t := range tt {
		ids[i] = t.ID
	}
	tt, err := svc.CategoryStorage.List(ctx, CategoryFilter{
		Limit: int64(len(tt)),
		IDs:   ids,
	})
	if err != nil {
		return nil, err
	}
	return tt, nil
}

func (svc *CatalogService) GetCategory(ctx context.Context, id, merchantID string, locationIDs []string) (Category, error) {
	const op = errors.Op("core/CatalogService/GetCategory")
	it, err := svc.CategoryStorage.Get(ctx, id, merchantID, locationIDs)
	if err != nil {
		return Category{}, errors.E(op, err)
	}
	return it, nil
}

func (svc *CatalogService) ListCategory(ctx context.Context, f CategoryFilter) ([]Category, error) {
	const op = errors.Op("core/CatalogService.ListCategory")
	limit, offset := int64(defaultReturnedCategories), int64(0)
	if f.Limit != 0 && f.Limit < maxReturnedCategories {
		limit = f.Limit
	}
	if f.Offset != 0 {
		offset = f.Offset
	}

	its, err := svc.CategoryStorage.List(ctx, CategoryFilter{
		MerchantID: f.MerchantID,
		Limit:      limit,
		Offset:     offset,
	})
	if err != nil {
		return nil, errors.E(op, err)
	}
	return its, nil
}

func (svc *CatalogService) DeleteCategory(ctx context.Context, id, merchantID string, locationIDs []string) (Category, error) {
	const op = errors.Op("controller.Category.Delete")
	category, err := svc.CategoryStorage.Get(ctx, id, merchantID, locationIDs)
	if err != nil {
		return Category{}, errors.E(op, err)
	}

	category.Status = StatusShadowDeleted
	if err := svc.CategoryStorage.Put(ctx, category); err != nil {
		return Category{}, errors.E(op, err)
	}
	resp, err := svc.CategoryStorage.Get(ctx, id, merchantID, locationIDs)
	if err != nil {
		return Category{}, errors.E(op, err)
	}
	return resp, nil
}

type CategoryFilter struct {
	Limit       int64
	Offset      int64
	LocationIDs []string
	MerchantID  string
	IDs         []string
}
