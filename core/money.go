package core

type Currency string

const (
	PEN Currency = "PEN"
	USD Currency = "USD"
)

type Money struct {
	Amount   int64  `bson:"amount"`
	Currency string `bson:"currency"`
}

func NewMoney(amount int64, currency string) Money {
	return Money{
		Amount:   amount,
		Currency: currency,
	}
}
