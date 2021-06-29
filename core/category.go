package core

import (
	"context"

	"github.com/backium/backend/errors"
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
		ID:          NewID("cat"),
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

func (svc *CatalogService) PutCategory(ctx context.Context, category Category) (Category, error) {
	const op = errors.Op("core/CatalogService.PutCategory")
	if err := svc.CategoryStorage.Put(ctx, category); err != nil {
		return Category{}, err
	}
	category, err := svc.CategoryStorage.Get(ctx, category.ID, category.MerchantID, category.LocationIDs)
	if err != nil {
		return Category{}, err
	}
	return category, nil
}

func (svc *CatalogService) PutCategories(ctx context.Context, categories []Category) ([]Category, error) {
	const op = errors.Op("core/CatalogService.PutCategories")
	if err := svc.CategoryStorage.PutBatch(ctx, categories); err != nil {
		return nil, err
	}
	ids := make([]string, len(categories))
	for i, t := range categories {
		ids[i] = t.ID
	}
	categories, err := svc.CategoryStorage.List(ctx, CategoryFilter{
		Limit: int64(len(categories)),
		IDs:   ids,
	})
	if err != nil {
		return nil, err
	}
	return categories, nil
}

func (svc *CatalogService) GetCategory(ctx context.Context, id, merchantID string, locationIDs []string) (Category, error) {
	const op = errors.Op("core/CatalogService/GetCategory")
	category, err := svc.CategoryStorage.Get(ctx, id, merchantID, locationIDs)
	if err != nil {
		return Category{}, errors.E(op, err)
	}
	return category, nil
}

func (svc *CatalogService) ListCategory(ctx context.Context, f CategoryFilter) ([]Category, error) {
	const op = errors.Op("core/CatalogService.ListCategory")
	categories, err := svc.CategoryStorage.List(ctx, CategoryFilter{
		MerchantID: f.MerchantID,
		Limit:      f.Limit,
		Offset:     f.Offset,
	})
	if err != nil {
		return nil, errors.E(op, err)
	}
	return categories, nil
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
	category, err = svc.CategoryStorage.Get(ctx, id, merchantID, locationIDs)
	if err != nil {
		return Category{}, errors.E(op, err)
	}
	return category, nil
}

type CategoryFilter struct {
	Limit       int64
	Offset      int64
	LocationIDs []string
	MerchantID  string
	IDs         []string
}
