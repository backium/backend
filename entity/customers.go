package entity

import "github.com/backium/backend/base"

type Customer struct {
	ID         string        `bson:"_id"`
	Name       string        `bson:"name,omitempty"`
	Email      string        `bson:"email,omitempty"`
	Phone      string        `bson:"phone,omitempty"`
	Address    *base.Address `bson:"address,omitempty"`
	MerchantID string        `bson:"merchant_id,omitempty"`
	Status     base.Status   `bson:"status,omitempty"`
}

// Creates a Customer with default values
func NewCustomer() Customer {
	return Customer{
		Status: base.StatusActive,
	}
}
