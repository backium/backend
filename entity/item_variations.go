package entity

type ItemVariation struct {
	ID          string   `bson:"_id"`
	Name        string   `bson:"name,omitempty"`
	SKU         string   `bson:"sku,omitempty"`
	ItemID      string   `bson:"item_id,omitempty"`
	Price       Money    `bson:"price"`
	LocationIDs []string `bson:"location_ids"`
	MerchantID  string   `bson:"merchant_id,omitempty"`
	Status      Status   `bson:"status,omitempty"`
}

// Creates an ItemVariation with default values
func NewItemVariation() ItemVariation {
	return ItemVariation{
		LocationIDs: []string{},
		Status:      StatusActive,
	}
}
