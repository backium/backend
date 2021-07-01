package http

import (
	"net/http"

	"github.com/backium/backend/core"
	"github.com/backium/backend/errors"
	"github.com/backium/backend/ptr"
	"github.com/labstack/echo/v4"
)

const (
	DiscountListDefaultSize = 10
	DiscountListMaxSize     = 50
)

func (h *Handler) HandleCreateDiscount(c echo.Context) error {
	const op = errors.Op("http/Handler.CreateDiscount")

	type request struct {
		Name        string            `json:"name" validate:"required"`
		Type        core.DiscountType `json:"type" validate:"required"`
		Amount      *MoneyRequest     `json:"amount" validate:"omitempty"`
		Percentage  *float64          `json:"percentage" validate:"omitempty,min=0,max=100"`
		LocationIDs *[]string         `json:"location_ids" validate:"omitempty,dive,required"`
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

	discount := core.NewDiscount(req.Name, req.Type, merchant.ID)
	if req.LocationIDs != nil {
		discount.LocationIDs = *req.LocationIDs
	}
	if req.Type == core.DiscountPercentage {
		discount.Percentage = ptr.GetFloat64(req.Percentage)
	}
	if req.Type == core.DiscountFixed && req.Amount != nil {
		discount.Amount = core.Money{
			Value:    *req.Amount.Value,
			Currency: req.Amount.Currency,
		}
	}

	discount, err := h.CatalogService.PutDiscount(ctx, discount)
	if err != nil {
		return errors.E(op, err)
	}

	return c.JSON(http.StatusOK, NewDiscount(discount))
}

func (h *Handler) HandleUpdateDiscount(c echo.Context) error {
	const op = errors.Op("http/Handler.UpdateDiscount")

	type request struct {
		ID          string             `param:"id" validate:"required"`
		Name        *string            `json:"name" validate:"omitempty,min=1"`
		Type        *core.DiscountType `json:"type"`
		Amount      *MoneyRequest      `json:"amount" validate:"omitempty"`
		Percentage  *float64           `json:"percentage" validate:"omitempty,min=0,max=100"`
		LocationIDs *[]string          `json:"location_ids" validate:"omitempty,dive,required"`
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

	discount, err := h.CatalogService.GetDiscount(ctx, req.ID, merchant.ID, nil)
	if err != nil {
		return errors.E(op, err)
	}

	if req.Name != nil {
		discount.Name = *req.Name
	}
	if req.Type != nil {
		discount.Type = *req.Type
	}
	if req.Percentage != nil {
		discount.Percentage = *req.Percentage
	}
	if req.LocationIDs != nil {
		discount.LocationIDs = *req.LocationIDs
	}
	if req.Amount != nil {
		discount.Amount = core.Money{
			Value:    *req.Amount.Value,
			Currency: req.Amount.Currency,
		}
	}

	discount, err = h.CatalogService.PutDiscount(ctx, discount)
	if err != nil {
		return errors.E(op, err)
	}

	return c.JSON(http.StatusOK, NewDiscount(discount))
}

func (h *Handler) HandleRetrieveDiscount(c echo.Context) error {
	const op = errors.Op("http/Handler.RetrieveDiscount")

	ctx := c.Request().Context()

	merchant := core.MerchantFromContext(ctx)
	if merchant == nil {
		return errors.E(op, errors.KindUnexpected, "invalid echo.Context")
	}

	discount, err := h.CatalogService.GetDiscount(ctx, c.Param("id"), merchant.ID, nil)
	if err != nil {
		return errors.E(op, err)
	}

	return c.JSON(http.StatusOK, NewDiscount(discount))
}

func (h *Handler) HandleListDiscounts(c echo.Context) error {
	const op = errors.Op("http/Handler.ListDiscounts")

	type request struct {
		Limit  int64 `query:"limit" validate:"gte=0"`
		Offset int64 `query:"offset" validate:"gte=0"`
	}

	type response struct {
		Discounts []Discount `json:"discounts"`
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

	var limit, offset int64 = DiscountListDefaultSize, req.Offset
	if req.Limit <= DiscountListMaxSize {
		limit = req.Limit
	} else {
		limit = DiscountListMaxSize
	}

	discounts, err := h.CatalogService.ListDiscount(ctx, core.DiscountFilter{
		Limit:      limit,
		Offset:     offset,
		MerchantID: merchant.ID,
	})
	if err != nil {
		return errors.E(op, err)
	}

	res := make([]Discount, len(discounts))
	for i, it := range discounts {
		res[i] = NewDiscount(it)
	}

	return c.JSON(http.StatusOK, response{res})
}

func (h *Handler) HandleDeleteDiscount(c echo.Context) error {
	const op = errors.Op("http/Handler.DeleteDiscount")

	ctx := c.Request().Context()

	merchant := core.MerchantFromContext(ctx)
	if merchant == nil {
		return errors.E(op, errors.KindUnexpected, "invalid echo.Context")
	}

	discount, err := h.CatalogService.DeleteDiscount(ctx, c.Param("id"), merchant.ID, nil)
	if err != nil {
		return errors.E(op, err)
	}

	return c.JSON(http.StatusOK, NewDiscount(discount))
}

type Discount struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Type        core.DiscountType `json:"type"`
	Amount      *MoneyRequest     `json:"amount,omitempty"`
	Percentage  *float64          `json:"percentage,omitempty"`
	LocationIDs []string          `json:"location_ids"`
	MerchantID  string            `json:"merchant_id"`
	CreatedAt   int64             `json:"created_at"`
	UpdatedAt   int64             `json:"updated_at"`
	Status      core.Status       `json:"status"`
}

func NewDiscount(discount core.Discount) Discount {
	resp := Discount{
		ID:          discount.ID,
		Name:        discount.Name,
		Type:        discount.Type,
		LocationIDs: discount.LocationIDs,
		MerchantID:  discount.MerchantID,
		CreatedAt:   discount.CreatedAt,
		UpdatedAt:   discount.UpdatedAt,
		Status:      discount.Status,
	}
	if discount.Type == core.DiscountFixed {
		resp.Amount = &MoneyRequest{
			Value:    &discount.Amount.Value,
			Currency: discount.Amount.Currency,
		}
	} else {
		resp.Percentage = ptr.Float64(discount.Percentage)
	}
	return resp
}
