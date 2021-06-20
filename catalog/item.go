package catalog

import "github.com/backium/backend/base"

type PartialItem struct {
	Name        *string      `bson:"name,omitempty"`
	Description *string      `bson:"description,omitempty"`
	CategoryID  *string      `bson:"category_id,omitempty"`
	LocationIDs *[]string    `bson:"location_ids,omitempty"`
	Status      *base.Status `bson:"status,omitempty"`
}

type Item struct {
	ID          string      `bson:"_id"`
	Name        string      `bson:"name,omitempty"`
	Description string      `bson:"description,omitempty"`
	CategoryID  string      `bson:"category_id,omitempty"`
	LocationIDs []string    `bson:"location_ids"`
	MerchantID  string      `bson:"merchant_id,omitempty"`
	Status      base.Status `bson:"status,omitempty"`
}

// Creates an Item with default values
func NewItem() Item {
	return Item{
		LocationIDs: []string{},
		Status:      base.StatusActive,
	}
}
