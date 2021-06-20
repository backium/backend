package handler

import (
	"net/http"

	"github.com/backium/backend/base"
	"github.com/backium/backend/catalog"
	"github.com/backium/backend/errors"
	"github.com/labstack/echo/v4"
)

type Tax struct {
	Controller catalog.Controller
}

func (h *Tax) Create(c echo.Context) error {
	const op = errors.Op("handler.Tax.Create")
	ac, ok := c.(*AuthContext)
	if !ok {
		return errors.E(op, errors.KindUnexpected, "invalid echo.Context")
	}
	req := TaxCreateRequest{}
	if err := bindAndValidate(c, &req); err != nil {
		return err
	}
	t := catalog.NewTax()
	if req.LocationIDs != nil {
		t.LocationIDs = *req.LocationIDs
	}
	t.Name = req.Name
	t.Percentage = *req.Percentage
	t.MerchantID = ac.MerchantID

	t, err := h.Controller.CreateTax(c.Request().Context(), t)
	if err != nil {
		return errors.E(op, err)
	}
	return c.JSON(http.StatusOK, newTaxResponse(t))
}

func (h *Tax) Update(c echo.Context) error {
	const op = errors.Op("handler.Tax.Update")
	req := TaxUpdateRequest{}
	if err := bindAndValidate(c, &req); err != nil {
		return err
	}
	t := catalog.TaxPartial{
		Name:        req.Name,
		Percentage:  req.Percentage,
		LocationIDs: req.LocationIDs,
	}
	ut, err := h.Controller.UpdateTax(c.Request().Context(), req.ID, t)
	if err != nil {
		return errors.E(op, err)
	}
	return c.JSON(http.StatusOK, newTaxResponse(ut))
}

func (h *Tax) Retrieve(c echo.Context) error {
	const op = errors.Op("handler.Tax.Retrieve")
	ac, ok := c.(*AuthContext)
	if !ok {
		return errors.E(op, errors.KindUnexpected, "invalid echo.Context")
	}
	it, err := h.Controller.RetrieveTax(c.Request().Context(), catalog.TaxRetrieveRequest{
		ID:         c.Param("id"),
		MerchantID: ac.MerchantID,
	})
	if err != nil {
		return errors.E(op, err)
	}
	return c.JSON(http.StatusOK, newTaxResponse(it))
}

func (h *Tax) ListAll(c echo.Context) error {
	const op = errors.Op("handler.Tax.ListAll")
	ac, ok := c.(*AuthContext)
	if !ok {
		return errors.E(op, errors.KindUnexpected, "invalid echo.Context")
	}
	req := TaxListRequest{}
	if err := bindAndValidate(c, &req); err != nil {
		return err
	}
	its, err := h.Controller.ListTax(c.Request().Context(), catalog.TaxListRequest{
		Limit:      req.Limit,
		Offset:     req.Offset,
		MerchantID: ac.MerchantID,
	})
	if err != nil {
		return errors.E(op, err)
	}
	res := make([]TaxResponse, len(its))
	for i, it := range its {
		res[i] = newTaxResponse(it)
	}
	return c.JSON(http.StatusOK, TaxListResponse{res})
}

func (h *Tax) Delete(c echo.Context) error {
	const op = errors.Op("handler.Tax.Delete")
	ac, ok := c.(*AuthContext)
	if !ok {
		return errors.E(op, errors.KindUnexpected, "invalid echo.Context")
	}
	it, err := h.Controller.DeleteTax(c.Request().Context(), catalog.TaxDeleteRequest{
		ID:         c.Param("id"),
		MerchantID: ac.MerchantID,
	})
	if err != nil {
		return errors.E(op, err)
	}
	return c.JSON(http.StatusOK, newTaxResponse(it))
}

type TaxResponse struct {
	ID          string      `json:"id"`
	Name        string      `json:"name"`
	Percentage  int         `json:"percentage"`
	LocationIDs []string    `json:"location_ids"`
	MerchantID  string      `json:"merchant_id"`
	Status      base.Status `json:"status"`
}

func newTaxResponse(t catalog.Tax) TaxResponse {
	return TaxResponse{
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
	Taxes []TaxResponse `json:"taxes"`
}
