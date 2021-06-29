package core

type Currency string

const (
	PEN Currency = "PEN"
	USD Currency = "USD"
)

type Money struct {
	Value    int64  `bson:"value"`
	Currency string `bson:"currency"`
}

func NewMoney(value int64, currency string) Money {
	return Money{
		Value:    value,
		Currency: currency,
	}
}
