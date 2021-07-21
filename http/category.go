package http

import (
	"net/http"

	"github.com/backium/backend/core"
	"github.com/backium/backend/errors"
	"github.com/labstack/echo/v4"
)

const (
	CategoryListDefaultSize = 10
	CategoryListMaxSize     = 50
)

func (h *Handler) HandleCreateCategory(c echo.Context) error {
	const op = errors.Op("handler.Category.Create")

	type request struct {
		Name        string     `json:"name" validate:"required"`
		Image       string     `json:"image"`
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

	category := core.NewCategory(req.Name, merchant.ID)
	category.Image = req.Image
	if req.LocationIDs != nil {
		category.LocationIDs = *req.LocationIDs
	}

	category, err := h.CatalogService.PutCategory(ctx, category)
	if err != nil {
		return errors.E(op, err)
	}

	return c.JSON(http.StatusOK, NewCategory(category))
}

func (h *Handler) HandleUpdateCategory(c echo.Context) error {
	const op = errors.Op("handler.Category.Update")

	type request struct {
		ID          core.ID    `param:"id" validate:"required"`
		Name        *string    `json:"name" validate:"omitempty,min=1"`
		Image       *string    `json:"image"`
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

	category, err := h.CatalogService.GetCategory(ctx, req.ID)
	if err != nil {
		return errors.E(op, err)
	}
	if req.Name != nil {
		category.Name = *req.Name
	}
	if req.Image != nil {
		category.Image = *req.Image
	}
	if req.LocationIDs != nil {
		category.LocationIDs = *req.LocationIDs
	}

	category, err = h.CatalogService.PutCategory(ctx, category)
	if err != nil {
		return errors.E(op, err)
	}

	return c.JSON(http.StatusOK, NewCategory(category))
}

func (h *Handler) HandleRetrieveCategory(c echo.Context) error {
	const op = errors.Op("handler.Category.Retrieve")

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

	category, err := h.CatalogService.GetCategory(ctx, req.ID)
	if err != nil {
		return errors.E(op, err)
	}

	return c.JSON(http.StatusOK, NewCategory(category))
}

func (h *Handler) HandleSearchCategory(c echo.Context) error {
	const op = errors.Op("http/Handler.HandleSearchCategory")

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
		Categories []Category `json:"categories"`
		Total      int64      `json:"total_count"`
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

	categories, count, err := h.CatalogService.ListCategory(ctx, core.CategoryQuery{
		Limit:  req.Limit,
		Offset: req.Offset,
		Filter: core.CategoryFilter{
			Name:        req.Filter.Name,
			LocationIDs: req.Filter.LocationIDs,
			MerchantID:  merchant.ID,
		},
		Sort: core.CategorySort{
			Name: req.Sort.Name,
		},
	})
	if err != nil {
		return errors.E(op, err)
	}

	resp := response{
		Categories: make([]Category, len(categories)),
		Total:      count,
	}
	for i, category := range categories {
		resp.Categories[i] = NewCategory(category)
	}

	return c.JSON(http.StatusOK, resp)
}

func (h *Handler) HandleListCategories(c echo.Context) error {
	const op = errors.Op("handler.Category.ListAll")

	type request struct {
		Limit  int64 `query:"limit" validate:"gte=0"`
		Offset int64 `query:"offset" validate:"gte=0"`
	}

	type response struct {
		Categories []Category `json:"categories"`
		Total      int64      `json:"total_count"`
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

	categories, count, err := h.CatalogService.ListCategory(ctx, core.CategoryQuery{
		Limit:  req.Limit,
		Offset: req.Offset,
		Filter: core.CategoryFilter{MerchantID: merchant.ID},
	})
	if err != nil {
		return errors.E(op, err)
	}

	resp := response{
		Categories: make([]Category, len(categories)),
		Total:      count,
	}
	for i, category := range categories {
		resp.Categories[i] = NewCategory(category)
	}

	return c.JSON(http.StatusOK, resp)
}

func (h *Handler) HandleDeleteCategory(c echo.Context) error {
	const op = errors.Op("handler.Category.Delete")

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

	category, err := h.CatalogService.DeleteCategory(ctx, req.ID)
	if err != nil {
		return errors.E(op, err)
	}

	return c.JSON(http.StatusOK, NewCategory(category))
}

type Category struct {
	ID          core.ID     `json:"id"`
	Name        string      `json:"name"`
	Image       string      `json:"image"`
	LocationIDs []core.ID   `json:"location_ids"`
	MerchantID  core.ID     `json:"merchant_id"`
	CreatedAt   int64       `json:"created_at"`
	UpdatedAt   int64       `json:"updated_at"`
	Status      core.Status `json:"status"`
}

func NewCategory(category core.Category) Category {
	return Category{
		ID:          category.ID,
		Name:        category.Name,
		Image:       category.Image,
		LocationIDs: category.LocationIDs,
		MerchantID:  category.MerchantID,
		CreatedAt:   category.CreatedAt,
		UpdatedAt:   category.UpdatedAt,
		Status:      category.Status,
	}
}
