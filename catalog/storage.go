package catalog

import (
	"context"
)

type CategoryRepository interface {
	Create(context.Context, Category) (string, error)
	Update(context.Context, Category) error
	UpdatePartial(context.Context, string, CategoryPartial) error
	Retrieve(context.Context, string) (Category, error)
	List(context.Context, CategoryFilter) ([]Category, error)
}

type ItemRepository interface {
	Create(context.Context, Item) (string, error)
	Update(context.Context, Item) error
	UpdatePartial(context.Context, string, PartialItem) error
	Retrieve(context.Context, string) (Item, error)
	List(context.Context, ItemFilter) ([]Item, error)
}

type ItemVariationRepository interface {
	Create(context.Context, ItemVariation) (string, error)
	Update(context.Context, ItemVariation) error
	UpdatePartial(context.Context, string, PartialItemVariation) error
	Retrieve(context.Context, string) (ItemVariation, error)
	List(context.Context, ItemVariationFilter) ([]ItemVariation, error)
}

type TaxRepository interface {
	Create(context.Context, Tax) (string, error)
	Update(context.Context, Tax) error
	UpdatePartial(context.Context, string, TaxPartial) error
	Retrieve(context.Context, string) (Tax, error)
	List(context.Context, TaxFilter) ([]Tax, error)
}
