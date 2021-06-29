package http

import (
	"net/http"

	"github.com/backium/backend/core"
	"github.com/backium/backend/errors"
	"github.com/backium/backend/ptr"
	"github.com/labstack/echo/v4"
)

func (h *Handler) CreatePayment(c echo.Context) error {
	const op = errors.Op("http/Handler.CreatePayment")
	ac, ok := c.(*AuthContext)
	if !ok {
		return errors.E(op, errors.KindUnexpected, "invalid echo.Context")
	}
	req := PaymentCreateRequest{}
	if err := bindAndValidate(c, &req); err != nil {
		return err
	}

	payment := core.NewPayment(ac.MerchantID, req.LocationID)
	payment.Type = req.Type
	payment.OrderID = req.OrderID
	payment.Amount = core.NewMoney(*req.Amount.Value, req.Amount.Currency)
	payment.TipAmount = core.NewMoney(0, req.Amount.Currency)
	if req.TipAmount != nil {
		payment.TipAmount = core.NewMoney(*req.TipAmount.Value, req.TipAmount.Currency)
	}

	ctx := c.Request().Context()
	payment, err := h.PaymentService.CreatePayment(ctx, payment)
	if err != nil {
		return errors.E(op, err)
	}
	return c.JSON(http.StatusOK, NewPayment(payment))
}

type Payment struct {
	ID         string           `json:"id"`
	OrderID    string           `json:"order_id"`
	Type       core.PaymentType `json:"type"`
	Amount     Money            `json:"amount"`
	TipAmount  Money            `json:"tip_amount"`
	LocationID string           `json:"location_id"`
	CreatedAt  int64            `json:"created_at"`
	UpdatedAt  int64            `json:"updated_at"`
}

func NewPayment(payment core.Payment) Payment {
	return Payment{
		ID:      payment.ID,
		OrderID: payment.OrderID,
		Type:    payment.Type,
		Amount: Money{
			Value:    ptr.Int64(payment.Amount.Value),
			Currency: payment.Amount.Currency,
		},
		TipAmount: Money{
			Value:    ptr.Int64(payment.TipAmount.Value),
			Currency: payment.TipAmount.Currency,
		},
		LocationID: payment.LocationID,
		CreatedAt:  payment.CreatedAt,
		UpdatedAt:  payment.UpdatedAt,
	}
}

type PaymentCreateRequest struct {
	OrderID    string           `json:"order_id" validate:"required"`
	Type       core.PaymentType `json:"type" validate:"required"`
	Amount     *Money           `json:"amount" validate:"required,dive"`
	TipAmount  *Money           `json:"tip_amount" validate:"omitempty,dive"`
	LocationID string           `json:"location_id" validate:"required"`
}
