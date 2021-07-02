package core

type Currency string

const (
	PEN Currency = "pen"
	USD Currency = "usd"
)

type Money struct {
	Value    int64    `bson:"value"`
	Currency Currency `bson:"currency"`
}

func NewMoney(value int64, currency Currency) Money {
	return Money{
		Value:    value,
		Currency: currency,
	}
}
