package http

import (
	"net/http"

	"github.com/backium/backend/core"
	"github.com/backium/backend/errors"
	"github.com/backium/backend/ptr"
	"github.com/labstack/echo/v4"
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
	d := core.NewDiscount()
	d.Name = req.Name
	d.MerchantID = ac.MerchantID
	d.Type = req.Type
	if req.LocationIDs != nil {
		d.LocationIDs = *req.LocationIDs
	}
	if req.Type == core.DiscountTypePercentage {
		d.Percentage = ptr.GetInt64(req.Percentage)
	}
	if req.Type == core.DiscountTypeFixed && req.Fixed != nil {
		d.Fixed = core.Money{
			Amount:   *req.Fixed.Amount,
			Currency: req.Fixed.Currency,
		}
	}

	ctx := c.Request().Context()
	d, err := h.CatalogService.PutDiscount(ctx, d)
	if err != nil {
		return errors.E(op, err)
	}
	return c.JSON(http.StatusOK, NewDiscount(d))
}

func (h *Handler) UpdateDiscount(c echo.Context) error {
	const op = errors.Op("http/Handler.UpdateDiscount")
	req := DiscountUpdateRequest{}
	if err := bindAndValidate(c, &req); err != nil {
		return err
	}
	ctx := c.Request().Context()
	d, err := h.CatalogService.GetDiscount(ctx, req.ID)
	if err != nil {
		return errors.E(op, err)
	}

	if req.Name != nil {
		d.Name = *req.Name
	}
	if req.Type != nil {
		d.Type = *req.Type
	}
	if req.Percentage != nil {
		d.Percentage = *req.Percentage
	}
	if req.LocationIDs != nil {
		d.LocationIDs = *req.LocationIDs
	}
	if req.Fixed != nil {
		d.Fixed = core.Money{
			Amount:   *req.Fixed.Amount,
			Currency: req.Fixed.Currency,
		}
	}

	d, err = h.CatalogService.PutDiscount(ctx, d)
	if err != nil {
		return errors.E(op, err)
	}
	return c.JSON(http.StatusOK, NewDiscount(d))
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
	dd := make([]core.Discount, len(req.Discounts))
	for i, d := range req.Discounts {
		dd[i] = core.NewDiscount()
		if d.LocationIDs != nil {
			dd[i].LocationIDs = *d.LocationIDs
		}
		dd[i].Name = d.Name
		dd[i].Percentage = *d.Percentage
		dd[i].MerchantID = ac.MerchantID
	}

	ctx := c.Request().Context()
	dd, err := h.CatalogService.PutDiscounts(ctx, dd)
	if err != nil {
		return errors.E(op, err)
	}
	res := make([]Discount, len(dd))
	for i, t := range dd {
		res[i] = NewDiscount(t)
	}
	return c.JSON(http.StatusOK, DiscountListResponse{Discounts: res})
}

func (h *Handler) RetrieveDiscount(c echo.Context) error {
	const op = errors.Op("http/Handler.RetrieveDiscount")
	ctx := c.Request().Context()
	it, err := h.CatalogService.GetDiscount(ctx, c.Param("id"))
	if err != nil {
		return errors.E(op, err)
	}
	return c.JSON(http.StatusOK, NewDiscount(it))
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
	ctx := c.Request().Context()
	its, err := h.CatalogService.ListDiscount(ctx, core.DiscountFilter{
		Limit:      ptr.GetInt64(req.Limit),
		Offset:     ptr.GetInt64(req.Offset),
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
	ctx := c.Request().Context()
	d, err := h.CatalogService.DeleteDiscount(ctx, c.Param("id"))
	if err != nil {
		return errors.E(op, err)
	}
	return c.JSON(http.StatusOK, NewDiscount(d))
}

type Discount struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Type        core.DiscountType `json:"type"`
	Fixed       *Money            `json:"fixed,omitempty"`
	Percentage  *int64            `json:"percentage,omitempty"`
	LocationIDs []string          `json:"location_ids"`
	MerchantID  string            `json:"merchant_id"`
	CreatedAt   int64             `json:"created_at"`
	UpdatedAt   int64             `json:"updated_at"`
	Status      core.Status       `json:"status"`
}

func NewDiscount(d core.Discount) Discount {
	dr := Discount{
		ID:          d.ID,
		Name:        d.Name,
		Type:        d.Type,
		LocationIDs: d.LocationIDs,
		MerchantID:  d.MerchantID,
		CreatedAt:   d.CreatedAt,
		UpdatedAt:   d.UpdatedAt,
		Status:      d.Status,
	}
	if d.Type == core.DiscountTypeFixed {
		dr.Fixed = &Money{
			Amount:   &d.Fixed.Amount,
			Currency: d.Fixed.Currency,
		}
	} else {
		dr.Percentage = ptr.Int64(d.Percentage)
	}
	return dr
}

type DiscountCreateRequest struct {
	Name        string            `json:"name" validate:"required"`
	Type        core.DiscountType `json:"type" validate:"required"`
	Fixed       *Money            `json:"fixed" validate:"omitempty"`
	Percentage  *int64            `json:"percentage" validate:"omitempty,min=0,max=100"`
	LocationIDs *[]string         `json:"location_ids" validate:"omitempty,dive,required"`
}

type DiscountUpdateRequest struct {
	ID          string             `param:"id" validate:"required"`
	Name        *string            `json:"name" validate:"omitempty,min=1"`
	Type        *core.DiscountType `json:"type"`
	Fixed       *Money             `json:"fixed" validate:"omitempty"`
	Percentage  *int64             `json:"percentage" validate:"omitempty,min=0,max=100"`
	LocationIDs *[]string          `json:"location_ids" validate:"omitempty,dive,required"`
}

type DiscountBatchCreateRequest struct {
	Discounts []DiscountCreateRequest
}

type DiscountListRequest struct {
	Limit  *int64 `query:"limit" validate:"omitempty,gte=1"`
	Offset *int64 `query:"offset"`
}

type DiscountListResponse struct {
	Discounts []Discount `json:"discounts"`
}
