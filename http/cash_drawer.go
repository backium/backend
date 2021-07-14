package http

import (
	"net/http"

	"github.com/backium/backend/core"
	"github.com/backium/backend/errors"
	"github.com/backium/backend/ptr"
	"github.com/labstack/echo/v4"
)

const (
	CashDrawerCountListDefaultSize = 10
	CashDrawerCountListMaxSize     = 50
)

func (h *Handler) HandleChangeCashDrawer(c echo.Context) error {
	const op = errors.Op("http/Handler.ChangeCashDrawer")

	type request struct {
		ID     core.ID           `param:"id" validate:"required"`
		Op     core.CashDrawerOp `json:"op" validate:"required"`
		Amount *MoneyRequest     `json:"amount" validate:"required"`
		Note   string            `json:"note"`
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

	adj := core.NewCashDrawerAdjustment(req.ID, merchant.ID)
	adj.Op = req.Op
	adj.Note = req.Note
	adj.Amount = core.NewMoney(ptr.GetInt64(req.Amount.Value), req.Amount.Currency)

	cash, err := h.LocationService.AdjustCashDrawer(ctx, adj)
	if err != nil {
		return errors.E(op, err)
	}

	resp := NewCashDrawer(cash)

	return c.JSON(http.StatusOK, resp)
}

func (h *Handler) HandleSearchCashDrawer(c echo.Context) error {
	const op = errors.Op("http/Handler.HandleSearchCashDrawer")

	type filter struct {
		LocationIDs []core.ID `json:"location_ids" validate:"omitempty,dive,id"`
		Name        string    `json:"name"`
	}

	type request struct {
		Limit  int64  `json:"limit" validate:"gte=0"`
		Offset int64  `json:"offset" validate:"gte=0"`
		Filter filter `json:"filter"`
	}

	type response struct {
		CashDrawers []CashDrawer `json:"cash_drawers"`
		Total       int64        `json:"total_count"`
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

	var limit int64 = CashDrawerCountListDefaultSize
	if req.Limit <= CashDrawerCountListMaxSize {
		limit = req.Limit
	} else {
		limit = CashDrawerCountListMaxSize
	}

	drawers, totalCount, err := h.LocationService.ListCashDrawer(ctx, core.CashDrawerQuery{
		Limit:  limit,
		Offset: req.Offset,
		Filter: core.CashDrawerFilter{
			LocationIDs: req.Filter.LocationIDs,
		},
	})
	if err != nil {
		return errors.E(op, err)
	}

	resp := response{
		CashDrawers: NewCashDrawers(drawers),
		Total:       totalCount,
	}

	return c.JSON(http.StatusOK, resp)
}

type CashDrawer struct {
	ID           core.ID `json:"id"`
	Amount       Money   `json:"amount"`
	CalculatedAt int64   `json:"calculated_at"`
	LocationID   core.ID `json:"location_id"`
}

type CashDrawerAdjustment struct {
	CashDrawerID core.ID           `json:"cash_drawer_id"`
	Amount       Money             `json:"amount"`
	Op           core.CashDrawerOp `json:"operation"`
	Note         string            `json:"note"`
	LocationID   core.ID           `json:"location_id"`
	CreatedAt    int64             `json:"created_at"`
}

func NewCashDrawerAdjustment(adj core.CashDrawerAdjustment) CashDrawerAdjustment {
	return CashDrawerAdjustment{
		CashDrawerID: adj.CashDrawerID,
		Amount:       NewMoney(adj.Amount),
		Op:           adj.Op,
		Note:         adj.Note,
		CreatedAt:    adj.CreatedAt,
	}
}

func NewCashDrawer(count core.CashDrawer) CashDrawer {
	return CashDrawer{
		ID:           count.ID,
		Amount:       NewMoney(count.Amount),
		CalculatedAt: count.CalculatedAt,
		LocationID:   count.LocationID,
	}
}

func NewCashDrawers(drawers []core.CashDrawer) []CashDrawer {
	resp := make([]CashDrawer, len(drawers))
	for i, d := range drawers {
		resp[i] = NewCashDrawer(d)
	}
	return resp
}
