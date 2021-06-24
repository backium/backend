package http

import (
	"net/http"

	"github.com/backium/backend/core"
	"github.com/backium/backend/errors"
	"github.com/backium/backend/ptr"
	"github.com/labstack/echo/v4"
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
	cat := core.NewCategory()
	if req.LocationIDs != nil {
		cat.LocationIDs = *req.LocationIDs
	}
	cat.Name = req.Name
	cat.MerchantID = ac.MerchantID

	ctx := c.Request().Context()
	cat, err := h.CatalogService.PutCategory(ctx, cat)
	if err != nil {
		return errors.E(op, err)
	}
	return c.JSON(http.StatusOK, NewCategory(cat))
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
	cat, err := h.CatalogService.GetCategory(ctx, req.ID, ac.MerchantID, nil)
	if err != nil {
		return errors.E(op, err)
	}
	if req.Name != nil {
		cat.Name = *req.Name
	}
	if req.LocationIDs != nil {
		cat.LocationIDs = *req.LocationIDs
	}
	ucat, err := h.CatalogService.PutCategory(ctx, cat)
	if err != nil {
		return errors.E(op, err)
	}
	return c.JSON(http.StatusOK, NewCategory(ucat))
}

func (h *Handler) RetrieveCategory(c echo.Context) error {
	const op = errors.Op("handler.Category.Retrieve")
	ac, ok := c.(*AuthContext)
	if !ok {
		return errors.E(op, errors.KindUnexpected, "invalid echo.Context")
	}
	ctx := c.Request().Context()
	m, err := h.CatalogService.GetCategory(ctx, c.Param("id"), ac.MerchantID, nil)
	if err != nil {
		return errors.E(op, err)
	}
	return c.JSON(http.StatusOK, NewCategory(m))
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
	cuss, err := h.CatalogService.ListCategory(c.Request().Context(), core.CategoryFilter{
		Limit:      ptr.GetInt64(req.Limit),
		Offset:     ptr.GetInt64(req.Offset),
		MerchantID: ac.MerchantID,
	})
	if err != nil {
		return errors.E(op, err)
	}
	res := make([]Category, len(cuss))
	for i, cus := range cuss {
		res[i] = NewCategory(cus)
	}
	return c.JSON(http.StatusOK, CategoryListResponse{res})
}

func (h *Handler) DeleteCategory(c echo.Context) error {
	const op = errors.Op("handler.Category.Delete")
	ac, ok := c.(*AuthContext)
	if !ok {
		return errors.E(op, errors.KindUnexpected, "invalid echo.Context")
	}
	ctx := c.Request().Context()
	cus, err := h.CatalogService.DeleteCategory(ctx, c.Param("id"), ac.MerchantID, nil)
	if err != nil {
		return errors.E(op, err)
	}
	return c.JSON(http.StatusOK, NewCategory(cus))
}

type Category struct {
	ID          string      `json:"id"`
	Name        string      `json:"name"`
	LocationIDs []string    `json:"location_ids"`
	MerchantID  string      `json:"merchant_id"`
	CreatedAt   int64       `json:"created_at"`
	UpdatedAt   int64       `json:"updated_at"`
	Status      core.Status `json:"status"`
}

func NewCategory(cat core.Category) Category {
	return Category{
		ID:          cat.ID,
		Name:        cat.Name,
		LocationIDs: cat.LocationIDs,
		MerchantID:  cat.MerchantID,
		CreatedAt:   cat.CreatedAt,
		UpdatedAt:   cat.UpdatedAt,
		Status:      cat.Status,
	}
}

type CategoryCreateRequest struct {
	Name        string    `json:"name" validate:"required"`
	LocationIDs *[]string `json:"location_ids" validate:"omitempty,dive,required"`
}

type CategoryUpdateRequest struct {
	ID          string    `param:"id" validate:"required"`
	Name        *string   `json:"name" validate:"omitempty,min=1"`
	LocationIDs *[]string `json:"location_ids" validate:"omitempty,dive,required"`
}

type CategoryListRequest struct {
	Limit  *int64 `query:"limit" validate:"omitempty,gte=1"`
	Offset *int64 `query:"offset"`
}

type CategoryListResponse struct {
	Categories []Category `json:"categories"`
}
