package entity

type Money struct {
	Amount   int64  `bson:"amount"`
	Currency string `bson:"currency"`
}
