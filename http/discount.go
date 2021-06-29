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

func (h *Handler) CreateDiscount(c echo.Context) error {
	const op = errors.Op("http/Handler.CreateDiscount")
	ac, ok := c.(*AuthContext)
	if !ok {
		return errors.E(op, errors.KindUnexpected, "invalid echo.Context")
	}
	req := DiscountCreateRequest{}
	if err := bindAndValidate(c, &req); err != nil {
		return err
	}
	discount := core.NewDiscount(ac.MerchantID)
	discount.Name = req.Name
	discount.Type = req.Type
	if req.LocationIDs != nil {
		discount.LocationIDs = *req.LocationIDs
	}
	if req.Type == core.DiscountTypePercentage {
		discount.Percentage = ptr.GetFloat64(req.Percentage)
	}
	if req.Type == core.DiscountTypeFixed && req.Fixed != nil {
		discount.Fixed = core.Money{
			Amount:   *req.Fixed.Amount,
			Currency: req.Fixed.Currency,
		}
	}

	ctx := c.Request().Context()
	discount, err := h.CatalogService.PutDiscount(ctx, discount)
	if err != nil {
		return errors.E(op, err)
	}
	return c.JSON(http.StatusOK, NewDiscount(discount))
}

func (h *Handler) UpdateDiscount(c echo.Context) error {
	const op = errors.Op("http/Handler.UpdateDiscount")
	ac, ok := c.(*AuthContext)
	if !ok {
		return errors.E(op, errors.KindUnexpected, "invalid echo.Context")
	}
	req := DiscountUpdateRequest{}
	if err := bindAndValidate(c, &req); err != nil {
		return err
	}
	ctx := c.Request().Context()
	discount, err := h.CatalogService.GetDiscount(ctx, req.ID, ac.MerchantID, nil)
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
	if req.Fixed != nil {
		discount.Fixed = core.Money{
			Amount:   *req.Fixed.Amount,
			Currency: req.Fixed.Currency,
		}
	}

	discount, err = h.CatalogService.PutDiscount(ctx, discount)
	if err != nil {
		return errors.E(op, err)
	}
	return c.JSON(http.StatusOK, NewDiscount(discount))
}

func (h *Handler) BatchCreateDiscount(c echo.Context) error {
	const op = errors.Op("http/Handler.BatchCreateDiscount")
	ac, ok := c.(*AuthContext)
	if !ok {
		return errors.E(op, errors.KindUnexpected, "invalid echo.Context")
	}
	req := DiscountBatchCreateRequest{}
	if err := bindAndValidate(c, &req); err != nil {
		return err
	}
	discounts := make([]core.Discount, len(req.Discounts))
	for i, discount := range req.Discounts {
		discounts[i] = core.NewDiscount(ac.MerchantID)
		if discount.LocationIDs != nil {
			discounts[i].LocationIDs = *discount.LocationIDs
		}
		discounts[i].Name = discount.Name
		discounts[i].Percentage = *discount.Percentage
	}

	ctx := c.Request().Context()
	discounts, err := h.CatalogService.PutDiscounts(ctx, discounts)
	if err != nil {
		return errors.E(op, err)
	}
	resp := DiscountListResponse{Discounts: make([]Discount, len(discounts))}
	for i, t := range discounts {
		resp.Discounts[i] = NewDiscount(t)
	}
	return c.JSON(http.StatusOK, resp)
}

func (h *Handler) RetrieveDiscount(c echo.Context) error {
	const op = errors.Op("http/Handler.RetrieveDiscount")
	ac, ok := c.(*AuthContext)
	if !ok {
		return errors.E(op, errors.KindUnexpected, "invalid echo.Context")
	}
	ctx := c.Request().Context()
	discount, err := h.CatalogService.GetDiscount(ctx, c.Param("id"), ac.MerchantID, nil)
	if err != nil {
		return errors.E(op, err)
	}
	return c.JSON(http.StatusOK, NewDiscount(discount))
}

func (h *Handler) ListDiscounts(c echo.Context) error {
	const op = errors.Op("http/Handler.ListDiscounts")
	ac, ok := c.(*AuthContext)
	if !ok {
		return errors.E(op, errors.KindUnexpected, "invalid echo.Context")
	}
	req := DiscountListRequest{}
	if err := bindAndValidate(c, &req); err != nil {
		return err
	}

	var limit, offset int64 = DiscountListDefaultSize, 0
	if req.Limit <= DiscountListMaxSize {
		limit = req.Limit
	}
	if req.Limit > DiscountListMaxSize {
		limit = DiscountListMaxSize
	}
	if req.Offset != 0 {
		offset = req.Offset
	}

	ctx := c.Request().Context()
	its, err := h.CatalogService.ListDiscount(ctx, core.DiscountFilter{
		Limit:      limit,
		Offset:     offset,
		MerchantID: ac.MerchantID,
	})
	if err != nil {
		return errors.E(op, err)
	}
	res := make([]Discount, len(its))
	for i, it := range its {
		res[i] = NewDiscount(it)
	}
	return c.JSON(http.StatusOK, DiscountListResponse{res})
}

func (h *Handler) DeleteDiscount(c echo.Context) error {
	const op = errors.Op("http/Handler.DeleteDiscount")
	ac, ok := c.(*AuthContext)
	if !ok {
		return errors.E(op, errors.KindUnexpected, "invalid echo.Context")
	}
	ctx := c.Request().Context()
	discount, err := h.CatalogService.DeleteDiscount(ctx, c.Param("id"), ac.MerchantID, nil)
	if err != nil {
		return errors.E(op, err)
	}
	return c.JSON(http.StatusOK, NewDiscount(discount))
}

type Discount struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Type        core.DiscountType `json:"type"`
	Fixed       *Money            `json:"fixed,omitempty"`
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
	if discount.Type == core.DiscountTypeFixed {
		resp.Fixed = &Money{
			Amount:   &discount.Fixed.Amount,
			Currency: discount.Fixed.Currency,
		}
	} else {
		resp.Percentage = ptr.Float64(discount.Percentage)
	}
	return resp
}

type DiscountCreateRequest struct {
	Name        string            `json:"name" validate:"required"`
	Type        core.DiscountType `json:"type" validate:"required"`
	Fixed       *Money            `json:"fixed" validate:"omitempty"`
	Percentage  *float64          `json:"percentage" validate:"omitempty,min=0,max=100"`
	LocationIDs *[]string         `json:"location_ids" validate:"omitempty,dive,required"`
}

type DiscountUpdateRequest struct {
	ID          string             `param:"id" validate:"required"`
	Name        *string            `json:"name" validate:"omitempty,min=1"`
	Type        *core.DiscountType `json:"type"`
	Fixed       *Money             `json:"fixed" validate:"omitempty"`
	Percentage  *float64           `json:"percentage" validate:"omitempty,min=0,max=100"`
	LocationIDs *[]string          `json:"location_ids" validate:"omitempty,dive,required"`
}

type DiscountBatchCreateRequest struct {
	Discounts []DiscountCreateRequest
}

type DiscountListRequest struct {
	Limit  int64 `query:"limit" validate:"gte=0"`
	Offset int64 `query:"offset" validate:"gte=0"`
}

type DiscountListResponse struct {
	Discounts []Discount `json:"discounts"`
}
