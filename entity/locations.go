package entity

import "github.com/backium/backend/base"

type Location struct {
	ID           string      `bson:"_id"`
	Name         string      `bson:"name,omitempty"`
	BusinessName string      `bson:"business_name,omitempty"`
	MerchantID   string      `bson:"merchant_id,omitempty"`
	Status       base.Status `bson:"status,omitempty"`
}

// Creates a Customer with default values
func NewLocation() Location {
	return Location{
		Status: base.StatusActive,
	}
}
