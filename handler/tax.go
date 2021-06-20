package handler

import (
	"net/http"

	"github.com/backium/backend/controller"
	"github.com/backium/backend/entity"
	"github.com/backium/backend/errors"
	"github.com/labstack/echo/v4"
)

type tax struct {
	ID          string        `json:"id"`
	Name        string        `json:"name"`
	Percentage  int           `json:"percentage"`
	LocationIDs []string      `json:"location_ids"`
	MerchantID  string        `json:"merchant_id"`
	Status      entity.Status `json:"status"`
}

func newTax(t entity.Tax) tax {
	return tax{
		ID:          t.ID,
		Name:        t.Name,
		Percentage:  t.Percentage,
		LocationIDs: t.LocationIDs,
		MerchantID:  t.MerchantID,
		Status:      t.Status,
	}
}

type createTaxRequest struct {
	Name        string    `json:"name" validate:"required"`
	Percentage  *int      `json:"percentage" validate:"required,min=0,max=100"`
	LocationIDs *[]string `json:"location_ids" validate:"omitempty,dive,required"`
}

type updateTaxRequest struct {
	ID          string    `param:"id" validate:"required"`
	Name        *string   `json:"name" validate:"omitempty,min=1"`
	Percentage  *int      `json:"percentage" validate:"omitempty,min=0,max=100"`
	LocationIDs *[]string `json:"location_ids" validate:"omitempty,dive,required"`
}

type listAllTaxesRequest struct {
	Limit  *int64 `query:"limit" validate:"omitempty,gte=1"`
	Offset *int64 `query:"offset"`
}

type listTaxesResponse struct {
	Taxes []tax `json:"taxes"`
}

type Tax struct {
	Controller controller.Tax
}

func (h *Tax) Create(c echo.Context) error {
	const op = errors.Op("handler.Tax.Create")
	ac, ok := c.(*AuthContext)
	if !ok {
		return errors.E(op, errors.KindUnexpected, "invalid echo.Context")
	}
	req := createTaxRequest{}
	if err := bindAndValidate(c, &req); err != nil {
		return err
	}
	t := entity.NewTax()
	if req.LocationIDs != nil {
		t.LocationIDs = *req.LocationIDs
	}
	t.Name = req.Name
	t.Percentage = *req.Percentage
	t.MerchantID = ac.MerchantID

	t, err := h.Controller.Create(c.Request().Context(), t)
	if err != nil {
		return errors.E(op, err)
	}
	return c.JSON(http.StatusOK, newTax(t))
}

func (h *Tax) Update(c echo.Context) error {
	const op = errors.Op("handler.Tax.Update")
	req := updateTaxRequest{}
	if err := bindAndValidate(c, &req); err != nil {
		return err
	}
	t := controller.PartialTax{
		Name:        req.Name,
		Percentage:  req.Percentage,
		LocationIDs: req.LocationIDs,
	}
	ut, err := h.Controller.Update(c.Request().Context(), req.ID, t)
	if err != nil {
		return errors.E(op, err)
	}
	return c.JSON(http.StatusOK, newTax(ut))
}

func (h *Tax) Retrieve(c echo.Context) error {
	const op = errors.Op("handler.Tax.Retrieve")
	ac, ok := c.(*AuthContext)
	if !ok {
		return errors.E(op, errors.KindUnexpected, "invalid echo.Context")
	}
	it, err := h.Controller.Retrieve(c.Request().Context(), controller.RetrieveTaxRequest{
		ID:         c.Param("id"),
		MerchantID: ac.MerchantID,
	})
	if err != nil {
		return errors.E(op, err)
	}
	return c.JSON(http.StatusOK, newTax(it))
}

func (h *Tax) ListAll(c echo.Context) error {
	const op = errors.Op("handler.Tax.ListAll")
	ac, ok := c.(*AuthContext)
	if !ok {
		return errors.E(op, errors.KindUnexpected, "invalid echo.Context")
	}
	req := listAllTaxesRequest{}
	if err := bindAndValidate(c, &req); err != nil {
		return err
	}
	its, err := h.Controller.ListAll(c.Request().Context(), controller.ListAllTaxesRequest{
		Limit:      req.Limit,
		Offset:     req.Offset,
		MerchantID: ac.MerchantID,
	})
	if err != nil {
		return errors.E(op, err)
	}
	res := make([]tax, len(its))
	for i, it := range its {
		res[i] = newTax(it)
	}
	return c.JSON(http.StatusOK, listTaxesResponse{res})
}

func (h *Tax) Delete(c echo.Context) error {
	const op = errors.Op("handler.Tax.Delete")
	ac, ok := c.(*AuthContext)
	if !ok {
		return errors.E(op, errors.KindUnexpected, "invalid echo.Context")
	}
	it, err := h.Controller.Delete(c.Request().Context(), controller.DeleteTaxRequest{
		ID:         c.Param("id"),
		MerchantID: ac.MerchantID,
	})
	if err != nil {
		return errors.E(op, err)
	}
	return c.JSON(http.StatusOK, newTax(it))
}
