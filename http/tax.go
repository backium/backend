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

	ctx := c.Request().Context()
	t, err := h.CatalogService.PutTax(ctx, t)
	if err != nil {
		return errors.E(op, err)
	}
	return c.JSON(http.StatusOK, NewTax(t))
}

func (h *Handler) UpdateTax(c echo.Context) error {
	const op = errors.Op("handler.Tax.Update")
	ac, ok := c.(*AuthContext)
	if !ok {
		return errors.E(op, errors.KindUnexpected, "invalid echo.Context")
	}
	req := TaxUpdateRequest{}
	if err := bindAndValidate(c, &req); err != nil {
		return err
	}
	ctx := c.Request().Context()
	t, err := h.CatalogService.GetTax(ctx, core.TaxRetrieveRequest{
		ID:         req.ID,
		MerchantID: ac.MerchantID,
	})
	if err != nil {
		return errors.E(op, err)
	}

	if req.Name != nil {
		t.Name = *req.Name
	}
	if req.Percentage != nil {
		t.Percentage = *req.Percentage
	}
	if req.LocationIDs != nil {
		t.LocationIDs = *req.LocationIDs
	}

	ut, err := h.CatalogService.PutTax(ctx, t)
	if err != nil {
		return errors.E(op, err)
	}
	return c.JSON(http.StatusOK, NewTax(ut))
}

func (h *Handler) BatchCreateTax(c echo.Context) error {
	const op = errors.Op("http/Handler.BatchCreateTax")
	ac, ok := c.(*AuthContext)
	if !ok {
		return errors.E(op, errors.KindUnexpected, "invalid echo.Context")
	}
	req := TaxBatchCreateRequest{}
	if err := bindAndValidate(c, &req); err != nil {
		return err
	}
	tt := make([]core.Tax, len(req.Taxes))
	for i, tr := range req.Taxes {
		tt[i] = core.NewTax()
		if tr.LocationIDs != nil {
			tt[i].LocationIDs = *tr.LocationIDs
		}
		tt[i].Name = tr.Name
		tt[i].Percentage = *tr.Percentage
		tt[i].MerchantID = ac.MerchantID
	}

	ctx := c.Request().Context()
	tt, err := h.CatalogService.PutTaxes(ctx, tt)
	if err != nil {
		return errors.E(op, err)
	}
	res := make([]Tax, len(tt))
	for i, t := range tt {
		res[i] = NewTax(t)
	}
	return c.JSON(http.StatusOK, TaxListResponse{Taxes: res})
}

func (h *Handler) RetrieveTax(c echo.Context) error {
	const op = errors.Op("handler.Tax.Retrieve")
	ac, ok := c.(*AuthContext)
	if !ok {
		return errors.E(op, errors.KindUnexpected, "invalid echo.Context")
	}
	ctx := c.Request().Context()
	it, err := h.CatalogService.GetTax(ctx, core.TaxRetrieveRequest{
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
	ctx := c.Request().Context()
	its, err := h.CatalogService.ListTax(ctx, core.TaxListRequest{
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
	ctx := c.Request().Context()
	it, err := h.CatalogService.DeleteTax(ctx, core.TaxDeleteRequest{
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
	Percentage  float64     `json:"percentage"`
	LocationIDs []string    `json:"location_ids"`
	MerchantID  string      `json:"merchant_id"`
	CreatedAt   int64       `json:"created_at"`
	UpdatedAt   int64       `json:"updated_at"`
	Status      core.Status `json:"status"`
}

func NewTax(t core.Tax) Tax {
	return Tax{
		ID:          t.ID,
		Name:        t.Name,
		Percentage:  t.Percentage,
		LocationIDs: t.LocationIDs,
		MerchantID:  t.MerchantID,
		CreatedAt:   t.CreatedAt,
		UpdatedAt:   t.UpdatedAt,
		Status:      t.Status,
	}
}

type TaxCreateRequest struct {
	Name        string    `json:"name" validate:"required"`
	Percentage  *float64  `json:"percentage" validate:"required,min=0,max=100"`
	LocationIDs *[]string `json:"location_ids" validate:"omitempty,dive,required"`
}

type TaxUpdateRequest struct {
	ID          string    `param:"id" validate:"required"`
	Name        *string   `json:"name" validate:"omitempty,min=1"`
	Percentage  *float64  `json:"percentage" validate:"omitempty,min=0,max=100"`
	LocationIDs *[]string `json:"location_ids" validate:"omitempty,dive,required"`
}

type TaxBatchCreateRequest struct {
	Taxes []TaxCreateRequest
}

type TaxListRequest struct {
	Limit  *int64 `query:"limit" validate:"omitempty,gte=1"`
	Offset *int64 `query:"offset"`
}

type TaxListResponse struct {
	Taxes []Tax `json:"taxes"`
}
