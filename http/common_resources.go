package http

type Money struct {
	Value    *int64 `json:"value" validate:"required"`
	Currency string `json:"currency" validate:"required"`
}
