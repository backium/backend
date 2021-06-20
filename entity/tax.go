package entity

type Tax struct {
	ID          string   `bson:"_id"`
	Name        string   `bson:"name,omitempty"`
	Percentage  int      `bson:"percentage"`
	LocationIDs []string `bson:"location_ids"`
	MerchantID  string   `bson:"merchant_id,omitempty"`
	Status      Status   `bson:"status,omitempty"`
}

func NewTax() Tax {
	return Tax{
		LocationIDs: []string{},
		Status:      StatusActive,
	}
}
