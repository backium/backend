package http

import (
	"net/http"

	"github.com/backium/backend/core"
	"github.com/backium/backend/errors"
	"github.com/backium/backend/ptr"
	"github.com/labstack/echo/v4"
)

const (
	PaymentListDefaultSize = 10
	PaymentListMaxSize     = 50
)

func (h *Handler) HandleCreatePayment(c echo.Context) error {
	const op = errors.Op("http/Handler.CreatePayment")

	type request struct {
		OrderID    core.ID          `json:"order_id" validate:"required"`
		Type       core.PaymentType `json:"type" validate:"required"`
		Amount     *MoneyRequest    `json:"amount" validate:"required,dive"`
		TipAmount  *MoneyRequest    `json:"tip_amount" validate:"omitempty,dive"`
		LocationID core.ID          `json:"location_id" validate:"required"`
	}

	ctx := c.Request().Context()

	merchant := core.MerchantFromContext(ctx)
	if merchant == nil {
		return errors.E(op, errors.KindUnexpected, "invalid echo.Context")
	}

	req := request{}
	if err := bindAndValidate(c, &req); err != nil {
		return err
	}

	payment := core.NewPayment(req.Type, req.OrderID, merchant.ID, req.LocationID)
	payment.Amount = core.NewMoney(*req.Amount.Value, req.Amount.Currency)
	payment.TipAmount = core.NewMoney(0, req.Amount.Currency)
	if req.TipAmount != nil {
		payment.TipAmount = core.NewMoney(*req.TipAmount.Value, req.TipAmount.Currency)
	}

	payment, err := h.PaymentService.CreatePayment(ctx, payment)
	if err != nil {
		return errors.E(op, err)
	}

	return c.JSON(http.StatusOK, NewPayment(payment))
}

func (h *Handler) HandleSearchPayment(c echo.Context) error {
	const op = errors.Op("http/Handler.SearchPayments")

	type dateFilter struct {
		Gte int64 `json:"gte" validate:"gte=0"`
		Lte int64 `json:"lte" validate:"gte=0"`
	}

	type filter struct {
		IDs         []core.ID  `json:"ids" validate:"omitempty,dive,id"`
		LocationIDs []core.ID  `json:"location_ids" validate:"omitempty,dive,id"`
		CreatedAt   dateFilter `json:"created_at"`
	}

	type sort struct {
		CreatedAt core.SortOrder `json:"created_at"`
	}

	type request struct {
		Limit  int64  `json:"limit" validate:"gte=0"`
		Offset int64  `json:"offset" validate:"gte=0"`
		Filter filter `json:"filter"`
		Sort   sort   `json:"sort"`
	}

	type response struct {
		Payments []Payment `json:"payments"`
		Total    int64     `json:"total_count"`
	}

	ctx := c.Request().Context()

	merchant := core.MerchantFromContext(ctx)
	if merchant == nil {
		return errors.E(op, errors.KindUnexpected, "invalid echo.Context")
	}

	req := request{}
	if err := bindAndValidate(c, &req); err != nil {
		return err
	}

	var limit int64 = PaymentListDefaultSize
	if req.Limit <= PaymentListMaxSize {
		limit = req.Limit
	} else {
		limit = PaymentListMaxSize
	}

	payments, count, err := h.PaymentService.ListPayment(ctx, core.PaymentQuery{
		Limit:  limit,
		Offset: req.Offset,
		Filter: core.PaymentFilter{
			LocationIDs: req.Filter.LocationIDs,
			MerchantID:  merchant.ID,
			CreatedAt: core.DateFilter{
				Gte: req.Filter.CreatedAt.Gte,
				Lte: req.Filter.CreatedAt.Lte,
			},
		},
		Sort: core.PaymentSort{
			CreatedAt: req.Sort.CreatedAt,
		},
	})
	if err != nil {
		return errors.E(op, err)
	}

	resp := response{
		Payments: make([]Payment, len(payments)),
		Total:    count,
	}
	for i, p := range payments {
		resp.Payments[i] = NewPayment(p)
	}

	return c.JSON(http.StatusOK, resp)
}

type Payment struct {
	ID         core.ID          `json:"id"`
	OrderID    core.ID          `json:"order_id"`
	Type       core.PaymentType `json:"type"`
	Amount     MoneyRequest     `json:"amount"`
	TipAmount  MoneyRequest     `json:"tip_amount"`
	LocationID core.ID          `json:"location_id"`
	CreatedAt  int64            `json:"created_at"`
	UpdatedAt  int64            `json:"updated_at"`
}

func NewPayment(payment core.Payment) Payment {
	return Payment{
		ID:      payment.ID,
		OrderID: payment.OrderID,
		Type:    payment.Type,
		Amount: MoneyRequest{
			Value:    ptr.Int64(payment.Amount.Value),
			Currency: payment.Amount.Currency,
		},
		TipAmount: MoneyRequest{
			Value:    ptr.Int64(payment.TipAmount.Value),
			Currency: payment.TipAmount.Currency,
		},
		LocationID: payment.LocationID,
		CreatedAt:  payment.CreatedAt,
		UpdatedAt:  payment.UpdatedAt,
	}
}
