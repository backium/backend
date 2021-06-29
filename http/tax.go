package http

import (
	"net/http"

	"github.com/backium/backend/core"
	"github.com/backium/backend/errors"
	"github.com/labstack/echo/v4"
)

const (
	TaxListDefaultSize = 10
	TaxListMaxSize     = 50
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
	tax := core.NewTax(ac.MerchantID)
	if req.LocationIDs != nil {
		tax.LocationIDs = *req.LocationIDs
	}
	tax.Name = req.Name
	tax.Percentage = *req.Percentage

	ctx := c.Request().Context()
	tax, err := h.CatalogService.PutTax(ctx, tax)
	if err != nil {
		return errors.E(op, err)
	}
	return c.JSON(http.StatusOK, NewTax(tax))
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
	tax, err := h.CatalogService.GetTax(ctx, req.ID, ac.MerchantID, nil)
	if err != nil {
		return errors.E(op, err)
	}

	if req.Name != nil {
		tax.Name = *req.Name
	}
	if req.Percentage != nil {
		tax.Percentage = *req.Percentage
	}
	if req.LocationIDs != nil {
		tax.LocationIDs = *req.LocationIDs
	}

	tax, err = h.CatalogService.PutTax(ctx, tax)
	if err != nil {
		return errors.E(op, err)
	}
	return c.JSON(http.StatusOK, NewTax(tax))
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
	taxes := make([]core.Tax, len(req.Taxes))
	for i, tax := range req.Taxes {
		taxes[i] = core.NewTax(ac.MerchantID)
		if tax.LocationIDs != nil {
			taxes[i].LocationIDs = *tax.LocationIDs
		}
		taxes[i].Name = tax.Name
		taxes[i].Percentage = *tax.Percentage
	}

	ctx := c.Request().Context()
	taxes, err := h.CatalogService.PutTaxes(ctx, taxes)
	if err != nil {
		return errors.E(op, err)
	}
	resp := TaxListResponse{Taxes: make([]Tax, len(taxes))}
	for i, tax := range taxes {
		resp.Taxes[i] = NewTax(tax)
	}
	return c.JSON(http.StatusOK, resp)
}

func (h *Handler) RetrieveTax(c echo.Context) error {
	const op = errors.Op("handler.Tax.Retrieve")
	ac, ok := c.(*AuthContext)
	if !ok {
		return errors.E(op, errors.KindUnexpected, "invalid echo.Context")
	}
	ctx := c.Request().Context()
	tax, err := h.CatalogService.GetTax(ctx, c.Param("id"), ac.MerchantID, nil)
	if err != nil {
		return errors.E(op, err)
	}
	return c.JSON(http.StatusOK, NewTax(tax))
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

	var limit, offset int64 = TaxListDefaultSize, 0
	if req.Limit <= TaxListMaxSize {
		limit = req.Limit
	}
	if req.Limit > TaxListMaxSize {
		limit = TaxListMaxSize
	}
	if req.Offset != 0 {
		offset = req.Offset
	}

	ctx := c.Request().Context()
	taxes, err := h.CatalogService.ListTax(ctx, core.TaxFilter{
		Limit:      limit,
		Offset:     offset,
		MerchantID: ac.MerchantID,
	})
	if err != nil {
		return errors.E(op, err)
	}
	resp := TaxListResponse{Taxes: make([]Tax, len(taxes))}
	for i, tax := range taxes {
		resp.Taxes[i] = NewTax(tax)
	}
	return c.JSON(http.StatusOK, resp)
}

func (h *Handler) DeleteTax(c echo.Context) error {
	const op = errors.Op("handler.Tax.Delete")
	ac, ok := c.(*AuthContext)
	if !ok {
		return errors.E(op, errors.KindUnexpected, "invalid echo.Context")
	}
	ctx := c.Request().Context()
	it, err := h.CatalogService.DeleteTax(ctx, c.Param("id"), ac.MerchantID, nil)
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

func NewTax(tax core.Tax) Tax {
	return Tax{
		ID:          tax.ID,
		Name:        tax.Name,
		Percentage:  tax.Percentage,
		LocationIDs: tax.LocationIDs,
		MerchantID:  tax.MerchantID,
		CreatedAt:   tax.CreatedAt,
		UpdatedAt:   tax.UpdatedAt,
		Status:      tax.Status,
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
	Limit  int64 `query:"limit"`
	Offset int64 `query:"offset"`
}

type TaxListResponse struct {
	Taxes []Tax `json:"taxes"`
}
