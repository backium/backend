package core

const (
	maxReturnedItems          = 50
	maxReturnedItemVariations = 50
	maxReturnedCategories     = 50
	maxReturnedTaxes          = 50
)

type CatalogService struct {
	CategoryRepository      CategoryRepository
	ItemRepository          ItemRepository
	ItemVariationRepository ItemVariationRepository
	TaxRepository           TaxRepository
}
