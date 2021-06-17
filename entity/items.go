package entity

type Item struct {
	ID          string   `bson:"_id"`
	Name        string   `bson:"name,omitempty"`
	Description string   `bson:"description,omitempty"`
	CategoryID  string   `bson:"category_id,omitempty"`
	LocationIDs []string `bson:"location_ids,omitempty"`
	MerchantID  string   `bson:"merchant_id,omitempty"`
	Status      Status   `bson:"status,omitempty"`
}
