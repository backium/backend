package handler

type MoneyResponse struct {
	Amount   *int64 `json:"amount" validate:"required"`
	Currency string `json:"currency" validate:"required"`
}
