package entity

type Merchant struct {
	ID           string `bson:"_id"`
	FirstName    string `bson:"first_name,omitempty"`
	LastName     string `bson:"last_name,omitempty"`
	BusinessName string `bson:"business_name,omitempty"`
}
