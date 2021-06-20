package catalog

import "github.com/backium/backend/base"

type CategoryPartial struct {
	Name        *string      `bson:"name,omitempty"`
	LocationIDs *[]string    `bson:"location_ids,omitempty"`
	Status      *base.Status `bson:"status,omitempty"`
}

type Category struct {
	ID          string      `bson:"_id"`
	Name        string      `bson:"name,omitempty"`
	LocationIDs []string    `bson:"location_ids"`
	MerchantID  string      `bson:"merchant_id,omitempty"`
	Status      base.Status `bson:"status,omitempty"`
}

func NewCategory() Category {
	return Category{
		LocationIDs: []string{},
		Status:      base.StatusActive,
	}
}
