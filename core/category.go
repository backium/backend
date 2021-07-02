package core

import (
	"context"

	"github.com/backium/backend/errors"
)

type Category struct {
	ID          ID     `bson:"_id"`
	Name        string `bson:"name"`
	Image       string `bson:"image"`
	LocationIDs []ID   `bson:"location_ids"`
	MerchantID  ID     `bson:"merchant_id"`
	CreatedAt   int64  `bson:"created_at"`
	UpdatedAt   int64  `bson:"updated_at"`
	Status      Status `bson:"status"`
}

func NewCategory(name string, merchantID ID) Category {
	return Category{
		ID:          NewID("cat"),
		Name:        name,
		LocationIDs: []ID{},
		Status:      StatusActive,
		MerchantID:  merchantID,
	}
}

type CategoryStorage interface {
	Put(context.Context, Category) error
	PutBatch(context.Context, []Category) error
	Get(context.Context, ID) (Category, error)
	List(context.Context, CategoryFilter) ([]Category, error)
}

func (svc *CatalogService) PutCategory(ctx context.Context, category Category) (Category, error) {
	const op = errors.Op("core/CatalogService.PutCategory")

	if err := svc.CategoryStorage.Put(ctx, category); err != nil {
		return Category{}, err
	}

	category, err := svc.CategoryStorage.Get(ctx, category.ID)
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

	ids := make([]ID, len(categories))
	for i, c := range categories {
		ids[i] = c.ID
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

func (svc *CatalogService) GetCategory(ctx context.Context, id ID) (Category, error) {
	const op = errors.Op("core/CatalogService/GetCategory")

	category, err := svc.CategoryStorage.Get(ctx, id)
	if err != nil {
		return Category{}, errors.E(op, err)
	}

	return category, nil
}

func (svc *CatalogService) ListCategory(ctx context.Context, f CategoryFilter) ([]Category, error) {
	const op = errors.Op("core/CatalogService.ListCategory")

	categories, err := svc.CategoryStorage.List(ctx, CategoryFilter{
		Limit:       f.Limit,
		Offset:      f.Offset,
		IDs:         f.IDs,
		Name:        f.Name,
		LocationIDs: f.LocationIDs,
		MerchantID:  f.MerchantID,
	})
	if err != nil {
		return nil, errors.E(op, err)
	}

	return categories, nil
}

func (svc *CatalogService) DeleteCategory(ctx context.Context, id ID) (Category, error) {
	const op = errors.Op("controller.Category.Delete")

	category, err := svc.CategoryStorage.Get(ctx, id)
	if err != nil {
		return Category{}, errors.E(op, err)
	}

	category.Status = StatusShadowDeleted
	if err := svc.CategoryStorage.Put(ctx, category); err != nil {
		return Category{}, errors.E(op, err)
	}

	category, err = svc.CategoryStorage.Get(ctx, id)
	if err != nil {
		return Category{}, errors.E(op, err)
	}

	return category, nil
}

type CategoryFilter struct {
	Limit       int64
	Offset      int64
	Name        string
	LocationIDs []ID
	MerchantID  ID
	IDs         []ID
}
