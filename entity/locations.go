package entity

type Location struct {
	ID           string `bson:"_id"`
	Name         string `bson:"name,omitempty"`
	BusinessName string `bson:"business_name,omitempty"`
	MerchantID   string `bson:"merchant_id,omitempty"`
	Status       Status `bson:"status,omitempty"`
}

// Creates a Customer with default values
func NewLocation() Location {
	return Location{
		Status: StatusActive,
	}
}
