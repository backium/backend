package catalog

import "github.com/backium/backend/base"

type PartialItemVariation struct {
	Name        *string      `bson:"name,omitempty"`
	SKU         *string      `bson:"sku,omitempty"`
	Price       *base.Money  `bson:"price,omitempty"`
	LocationIDs *[]string    `bson:"location_ids,omitempty"`
	Status      *base.Status `bson:"status,omitempty"`
}

type ItemVariation struct {
	ID          string      `bson:"_id"`
	Name        string      `bson:"name,omitempty"`
	SKU         string      `bson:"sku,omitempty"`
	ItemID      string      `bson:"item_id,omitempty"`
	Price       base.Money  `bson:"price"`
	LocationIDs []string    `bson:"location_ids"`
	MerchantID  string      `bson:"merchant_id,omitempty"`
	Status      base.Status `bson:"status,omitempty"`
}

// Creates an ItemVariation with default values
func NewItemVariation() ItemVariation {
	return ItemVariation{
		LocationIDs: []string{},
		Status:      base.StatusActive,
	}
}
