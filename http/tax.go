package http

import (
	"net/http"

	"github.com/backium/backend/core"
	"github.com/backium/backend/errors"
	"github.com/labstack/echo/v4"
)

func (h *Handler) HandleCreateTax(c echo.Context) error {
	const op = errors.Op("http/Handler.HandleCreateTax")

	type request struct {
		Name        string     `json:"name" validate:"required"`
		Percentage  *float64   `json:"percentage" validate:"required,min=0,max=100"`
		Enabled     bool       `json:"enabled"`
		LocationIDs *[]core.ID `json:"location_ids" validate:"omitempty,dive,required,id"`
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

	tax := core.NewTax(req.Name, merchant.ID)
	tax.Percentage = *req.Percentage
	tax.EnabledInPOS = req.Enabled
	if req.LocationIDs != nil {
		tax.LocationIDs = *req.LocationIDs
	}

	tax, err := h.CatalogService.PutTax(ctx, tax)
	if err != nil {
		return errors.E(op, err)
	}

	return c.JSON(http.StatusOK, NewTax(tax))
}

func (h *Handler) HandleUpdateTax(c echo.Context) error {
	const op = errors.Op("http/Handler.HandleUpdateTax")

	type request struct {
		ID          core.ID    `param:"id" validate:"required"`
		Name        *string    `json:"name" validate:"omitempty,min=1"`
		Percentage  *float64   `json:"percentage" validate:"omitempty,min=0,max=100"`
		Enabled     *bool      `json:"enabled"`
		LocationIDs *[]core.ID `json:"location_ids" validate:"omitempty,dive,required"`
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

	tax, err := h.CatalogService.GetTax(ctx, req.ID)
	if err != nil {
		return errors.E(op, err)
	}
	if req.Name != nil {
		tax.Name = *req.Name
	}
	if req.Percentage != nil {
		tax.Percentage = *req.Percentage
	}
	if req.Enabled != nil {
		tax.EnabledInPOS = *req.Enabled
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

func (h *Handler) HandleBatchCreateTax(c echo.Context) error {
	const op = errors.Op("http/Handler.BatchCreateTax")

	type tax struct {
		Name        string     `json:"name" validate:"required"`
		Percentage  *float64   `json:"percentage" validate:"required,min=0,max=100"`
		LocationIDs *[]core.ID `json:"location_ids" validate:"omitempty,dive,required"`
	}

	type request struct {
		Taxes []tax
	}

	type response struct {
		Taxes []Tax `json:"taxes"`
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

	taxes := make([]core.Tax, len(req.Taxes))
	for i, tax := range req.Taxes {
		taxes[i] = core.NewTax(tax.Name, merchant.ID)
		taxes[i].Percentage = *tax.Percentage
		if tax.LocationIDs != nil {
			taxes[i].LocationIDs = *tax.LocationIDs
		}
	}

	taxes, err := h.CatalogService.PutTaxes(ctx, taxes)
	if err != nil {
		return errors.E(op, err)
	}

	resp := response{Taxes: make([]Tax, len(taxes))}
	for i, tax := range taxes {
		resp.Taxes[i] = NewTax(tax)
	}

	return c.JSON(http.StatusOK, resp)
}

func (h *Handler) HandleRetrieveTax(c echo.Context) error {
	const op = errors.Op("http/Handler.HandleRetrieveTax")

	type request struct {
		ID core.ID `param:"id" validate:"required,id"`
	}

	ctx := c.Request().Context()

	req := request{}
	if err := bindAndValidate(c, &req); err != nil {
		return err
	}

	merchant := core.MerchantFromContext(ctx)
	if merchant == nil {
		return errors.E(op, errors.KindUnexpected, "invalid echo.Context")
	}

	tax, err := h.CatalogService.GetTax(ctx, req.ID)
	if err != nil {
		return errors.E(op, err)
	}

	return c.JSON(http.StatusOK, NewTax(tax))
}

func (h *Handler) HandleListTaxes(c echo.Context) error {
	const op = errors.Op("http/Handler.HandleListTaxes")

	type request struct {
		Limit  int64 `query:"limit"`
		Offset int64 `query:"offset"`
	}

	type response struct {
		Taxes []Tax `json:"taxes"`
		Total int64 `json:"total_count"`
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

	taxes, count, err := h.CatalogService.ListTax(ctx, core.TaxQuery{
		Limit:  req.Limit,
		Offset: req.Offset,
		Filter: core.TaxFilter{MerchantID: merchant.ID},
	})
	if err != nil {
		return errors.E(op, err)
	}

	resp := response{
		Taxes: make([]Tax, len(taxes)),
		Total: count,
	}
	for i, tax := range taxes {
		resp.Taxes[i] = NewTax(tax)
	}

	return c.JSON(http.StatusOK, resp)
}

func (h *Handler) HandleSearchTax(c echo.Context) error {
	const op = errors.Op("http/Handler.HandleSearchTax")

	type filter struct {
		IDs         []core.ID `json:"ids" validate:"omitempty,dive,id"`
		LocationIDs []core.ID `json:"location_ids" validate:"omitempty,dive,id"`
		Name        string    `json:"name"`
	}

	type sort struct {
		Name core.SortOrder `json:"name"`
	}

	type request struct {
		Limit  int64  `json:"limit" validate:"gte=0"`
		Offset int64  `json:"offset" validate:"gte=0"`
		Filter filter `json:"filter"`
		Sort   sort   `json:"sort"`
	}

	type response struct {
		Taxes []Tax `json:"taxes"`
		Total int64 `json:"total_count"`
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

	taxes, count, err := h.CatalogService.ListTax(ctx, core.TaxQuery{
		Limit:  req.Limit,
		Offset: req.Offset,
		Filter: core.TaxFilter{
			Name:        req.Filter.Name,
			LocationIDs: req.Filter.LocationIDs,
			MerchantID:  merchant.ID,
		},
		Sort: core.TaxSort{
			Name: req.Sort.Name,
		},
	})
	if err != nil {
		return errors.E(op, err)
	}

	resp := response{
		Taxes: make([]Tax, len(taxes)),
		Total: count,
	}
	for i, tax := range taxes {
		resp.Taxes[i] = NewTax(tax)
	}

	return c.JSON(http.StatusOK, resp)
}

func (h *Handler) HandleDeleteTax(c echo.Context) error {
	const op = errors.Op("http/Handler.HandleDeleteTax")

	type request struct {
		ID core.ID `param:"id"`
	}

	ctx := c.Request().Context()

	req := request{}
	if err := bindAndValidate(c, &req); err != nil {
		return err
	}

	merchant := core.MerchantFromContext(ctx)
	if merchant == nil {
		return errors.E(op, errors.KindUnexpected, "invalid echo.Context")
	}

	tax, err := h.CatalogService.DeleteTax(ctx, req.ID)
	if err != nil {
		return errors.E(op, err)
	}

	return c.JSON(http.StatusOK, NewTax(tax))
}

type Tax struct {
	ID           core.ID     `json:"id"`
	Name         string      `json:"name"`
	Percentage   float64     `json:"percentage"`
	EnabledInPOS bool        `json:"enabled"`
	LocationIDs  []core.ID   `json:"location_ids"`
	MerchantID   core.ID     `json:"merchant_id"`
	CreatedAt    int64       `json:"created_at"`
	UpdatedAt    int64       `json:"updated_at"`
	Status       core.Status `json:"status"`
}

func NewTax(tax core.Tax) Tax {
	return Tax{
		ID:           tax.ID,
		Name:         tax.Name,
		Percentage:   tax.Percentage,
		EnabledInPOS: tax.EnabledInPOS,
		LocationIDs:  tax.LocationIDs,
		MerchantID:   tax.MerchantID,
		CreatedAt:    tax.CreatedAt,
		UpdatedAt:    tax.UpdatedAt,
		Status:       tax.Status,
	}
}
