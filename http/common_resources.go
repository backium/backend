package http

import "github.com/backium/backend/core"

type MoneyRequest struct {
	Value    *int64 `json:"value" validate:"required"`
	Currency string `json:"currency" validate:"required"`
}

type Money struct {
	Value    int64  `json:"value"`
	Currency string `json:"currency"`
}

func NewMoney(m core.Money) Money {
	return Money{
		Value:    m.Value,
		Currency: m.Currency,
	}
}
