package catalog

import "github.com/backium/backend/base"

type TaxPartial struct {
	Name        *string      `bson:"name,omitempty"`
	Percentage  *int         `bson:"percentage,omitempty"`
	LocationIDs *[]string    `bson:"location_ids,omitempty"`
	Status      *base.Status `bson:"status,omitempty"`
}

type Tax struct {
	ID          string      `bson:"_id"`
	Name        string      `bson:"name,omitempty"`
	Percentage  int         `bson:"percentage"`
	LocationIDs []string    `bson:"location_ids"`
	MerchantID  string      `bson:"merchant_id,omitempty"`
	Status      base.Status `bson:"status,omitempty"`
}

func NewTax() Tax {
	return Tax{
		LocationIDs: []string{},
		Status:      base.StatusActive,
	}
}
