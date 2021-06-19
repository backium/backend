package entity

type Category struct {
	ID          string   `bson:"_id"`
	Name        string   `bson:"name,omitempty"`
	LocationIDs []string `bson:"location_ids"`
	MerchantID  string   `bson:"merchant_id,omitempty"`
	Status      Status   `bson:"status,omitempty"`
}

func NewCategory() Category {
	return Category{
		LocationIDs: []string{},
		Status:      StatusActive,
	}
}
