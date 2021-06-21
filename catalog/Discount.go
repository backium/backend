package catalog

import "github.com/backium/backend/base"

type DiscountType string

const (
	Percentage  DiscountType = "percentage"
	Amount      DiscountType = "amount"
)


type DiscountPartial struct {
	Name        *string      	`bson:"name,omitempty"`
	Type 		*DiscountType 	`bson:"discount_type"`
	Amount 		*base.Money  	`bson:"amount"`
	Percentage  *int         	`bson:"percentage,omitempty"`
	LocationIDs *[]string    	`bson:"location_ids,omitempty"`
	Status      *base.Status 	`bson:"status,omitempty"`
}

type Discount struct {
	ID          string      	`bson:"_id"`
	Name        string      	`bson:"name,omitempty"`
	Type 		DiscountType 	`bson:"discount_type"`
	Amount 		base.Money  	`bson:"amount"`
	Percentage  int         	`bson:"percentage"`
	LocationIDs []string    	`bson:"location_ids"`
	MerchantID  string      	`bson:"merchant_id,omitempty"`
	Status      base.Status 	`bson:"status,omitempty"`
}


func NewDiscount() Discount {
	return Discount{
		Amount:		base.Money{
			Amount: 0.00,
			Currency: "USD",
		},
		Percentage:  0,
		LocationIDs: []string{},
		Status:      base.StatusActive,
	}
}
