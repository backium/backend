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

func (h *Handler) CreateCategory(c echo.Context) error {
	const op = errors.Op("handler.Category.Create")
	ac, ok := c.(*AuthContext)
	if !ok {
		return errors.E(op, errors.KindUnexpected, "invalid echo.Context")
	}
	req := CategoryCreateRequest{}
	if err := bindAndValidate(c, &req); err != nil {
		return err
	}
	category := core.NewCategory()
	if req.LocationIDs != nil {
		category.LocationIDs = *req.LocationIDs
	}
	category.Name = req.Name
	category.Image = req.Image
	category.MerchantID = ac.MerchantID

	ctx := c.Request().Context()
	category, err := h.CatalogService.PutCategory(ctx, category)
	if err != nil {
		return errors.E(op, err)
	}
	return c.JSON(http.StatusOK, NewCategory(category))
}

func (h *Handler) UpdateCategory(c echo.Context) error {
	const op = errors.Op("handler.Category.Update")
	ac, ok := c.(*AuthContext)
	if !ok {
		return errors.E(op, errors.KindUnexpected, "invalid echo.Context")
	}
	req := CategoryUpdateRequest{}
	if err := bindAndValidate(c, &req); err != nil {
		return err
	}
	ctx := c.Request().Context()
	category, err := h.CatalogService.GetCategory(ctx, req.ID, ac.MerchantID, nil)
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

func (h *Handler) RetrieveCategory(c echo.Context) error {
	const op = errors.Op("handler.Category.Retrieve")
	ac, ok := c.(*AuthContext)
	if !ok {
		return errors.E(op, errors.KindUnexpected, "invalid echo.Context")
	}
	ctx := c.Request().Context()
	cat, err := h.CatalogService.GetCategory(ctx, c.Param("id"), ac.MerchantID, nil)
	if err != nil {
		return errors.E(op, err)
	}
	return c.JSON(http.StatusOK, NewCategory(cat))
}

func (h *Handler) ListCategories(c echo.Context) error {
	const op = errors.Op("handler.Category.ListAll")
	ac, ok := c.(*AuthContext)
	if !ok {
		return errors.E(op, errors.KindUnexpected, "invalid echo.Context")
	}
	req := CategoryListRequest{}
	if err := bindAndValidate(c, &req); err != nil {
		return err
	}

	var limit, offset int64 = CategoryListDefaultSize, 0
	if req.Limit <= CategoryListMaxSize {
		limit = req.Limit
	}
	if req.Limit > CategoryListMaxSize {
		limit = CategoryListMaxSize
	}
	if req.Offset != 0 {
		offset = req.Offset
	}

	categories, err := h.CatalogService.ListCategory(c.Request().Context(), core.CategoryFilter{
		Limit:      limit,
		Offset:     offset,
		MerchantID: ac.MerchantID,
	})
	if err != nil {
		return errors.E(op, err)
	}
	resp := CategoryListResponse{Categories: make([]Category, len(categories))}
	for i, category := range categories {
		resp.Categories[i] = NewCategory(category)
	}
	return c.JSON(http.StatusOK, resp)
}

func (h *Handler) DeleteCategory(c echo.Context) error {
	const op = errors.Op("handler.Category.Delete")
	ac, ok := c.(*AuthContext)
	if !ok {
		return errors.E(op, errors.KindUnexpected, "invalid echo.Context")
	}
	ctx := c.Request().Context()
	category, err := h.CatalogService.DeleteCategory(ctx, c.Param("id"), ac.MerchantID, nil)
	if err != nil {
		return errors.E(op, err)
	}
	return c.JSON(http.StatusOK, NewCategory(category))
}

type Category struct {
	ID          string      `json:"id"`
	Name        string      `json:"name"`
	Image       string      `json:"image"`
	LocationIDs []string    `json:"location_ids"`
	MerchantID  string      `json:"merchant_id"`
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

type CategoryCreateRequest struct {
	Name        string    `json:"name" validate:"required"`
	Image       string    `json:"image"`
	LocationIDs *[]string `json:"location_ids" validate:"omitempty,dive,required"`
}

type CategoryUpdateRequest struct {
	ID          string    `param:"id" validate:"required"`
	Name        *string   `json:"name" validate:"omitempty,min=1"`
	Image       *string   `json:"image"`
	LocationIDs *[]string `json:"location_ids" validate:"omitempty,dive,required"`
}

type CategoryListRequest struct {
	Limit  int64 `query:"limit" validate:"gte=0"`
	Offset int64 `query:"offset" validate:"gte=0"`
}

type CategoryListResponse struct {
	Categories []Category `json:"categories"`
}
