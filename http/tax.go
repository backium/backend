package http

import (
	"net/http"

	"github.com/backium/backend/core"
	"github.com/backium/backend/errors"
	"github.com/labstack/echo/v4"
)

func (h *Handler) CreateTax(c echo.Context) error {
	const op = errors.Op("handler.Tax.Create")
	ac, ok := c.(*AuthContext)
	if !ok {
		return errors.E(op, errors.KindUnexpected, "invalid echo.Context")
	}
	req := TaxCreateRequest{}
	if err := bindAndValidate(c, &req); err != nil {
		return err
	}
	t := core.NewTax()
	if req.LocationIDs != nil {
		t.LocationIDs = *req.LocationIDs
	}
	t.Name = req.Name
	t.Percentage = *req.Percentage
	t.MerchantID = ac.MerchantID

	t, err := h.CatalogService.CreateTax(c.Request().Context(), t)
	if err != nil {
		return errors.E(op, err)
	}
	return c.JSON(http.StatusOK, NewTax(t))
}

func (h *Handler) UpdateTax(c echo.Context) error {
	const op = errors.Op("handler.Tax.Update")
	req := TaxUpdateRequest{}
	if err := bindAndValidate(c, &req); err != nil {
		return err
	}
	t := core.TaxPartial{
		Name:        req.Name,
		Percentage:  req.Percentage,
		LocationIDs: req.LocationIDs,
	}
	ut, err := h.CatalogService.UpdateTax(c.Request().Context(), req.ID, t)
	if err != nil {
		return errors.E(op, err)
	}
	return c.JSON(http.StatusOK, NewTax(ut))
}

func (h *Handler) RetrieveTax(c echo.Context) error {
	const op = errors.Op("handler.Tax.Retrieve")
	ac, ok := c.(*AuthContext)
	if !ok {
		return errors.E(op, errors.KindUnexpected, "invalid echo.Context")
	}
	it, err := h.CatalogService.RetrieveTax(c.Request().Context(), core.TaxRetrieveRequest{
		ID:         c.Param("id"),
		MerchantID: ac.MerchantID,
	})
	if err != nil {
		return errors.E(op, err)
	}
	return c.JSON(http.StatusOK, NewTax(it))
}

func (h *Handler) ListTaxes(c echo.Context) error {
	const op = errors.Op("handler.Tax.ListAll")
	ac, ok := c.(*AuthContext)
	if !ok {
		return errors.E(op, errors.KindUnexpected, "invalid echo.Context")
	}
	req := TaxListRequest{}
	if err := bindAndValidate(c, &req); err != nil {
		return err
	}
	its, err := h.CatalogService.ListTax(c.Request().Context(), core.TaxListRequest{
		Limit:      req.Limit,
		Offset:     req.Offset,
		MerchantID: ac.MerchantID,
	})
	if err != nil {
		return errors.E(op, err)
	}
	res := make([]Tax, len(its))
	for i, it := range its {
		res[i] = NewTax(it)
	}
	return c.JSON(http.StatusOK, TaxListResponse{res})
}

func (h *Handler) DeleteTax(c echo.Context) error {
	const op = errors.Op("handler.Tax.Delete")
	ac, ok := c.(*AuthContext)
	if !ok {
		return errors.E(op, errors.KindUnexpected, "invalid echo.Context")
	}
	it, err := h.CatalogService.DeleteTax(c.Request().Context(), core.TaxDeleteRequest{
		ID:         c.Param("id"),
		MerchantID: ac.MerchantID,
	})
	if err != nil {
		return errors.E(op, err)
	}
	return c.JSON(http.StatusOK, NewTax(it))
}

type Tax struct {
	ID          string      `json:"id"`
	Name        string      `json:"name"`
	Percentage  int         `json:"percentage"`
	LocationIDs []string    `json:"location_ids"`
	MerchantID  string      `json:"merchant_id"`
	Status      core.Status `json:"status"`
}

func NewTax(t core.Tax) Tax {
	return Tax{
		ID:          t.ID,
		Name:        t.Name,
		Percentage:  t.Percentage,
		LocationIDs: t.LocationIDs,
		MerchantID:  t.MerchantID,
		Status:      t.Status,
	}
}

type TaxCreateRequest struct {
	Name        string    `json:"name" validate:"required"`
	Percentage  *int      `json:"percentage" validate:"required,min=0,max=100"`
	LocationIDs *[]string `json:"location_ids" validate:"omitempty,dive,required"`
}

type TaxUpdateRequest struct {
	ID          string    `param:"id" validate:"required"`
	Name        *string   `json:"name" validate:"omitempty,min=1"`
	Percentage  *int      `json:"percentage" validate:"omitempty,min=0,max=100"`
	LocationIDs *[]string `json:"location_ids" validate:"omitempty,dive,required"`
}

type TaxListRequest struct {
	Limit  *int64 `query:"limit" validate:"omitempty,gte=1"`
	Offset *int64 `query:"offset"`
}

type TaxListResponse struct {
	Taxes []Tax `json:"taxes"`
}
