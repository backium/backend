package core

type CatalogService struct {
	CategoryStorage      CategoryStorage
	ItemStorage          ItemStorage
	ItemVariationStorage ItemVariationStorage
	TaxStorage           TaxStorage
	DiscountStorage      DiscountStorage
}
