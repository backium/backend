package http

import (
	"net/http"

	"github.com/backium/backend/core"
	"github.com/backium/backend/errors"
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

	cat, err := h.CatalogService.CreateCategory(c.Request().Context(), cat)
	if err != nil {
		return errors.E(op, err)
	}
	return c.JSON(http.StatusOK, NewCategory(cat))
}

func (h *Handler) UpdateCategory(c echo.Context) error {
	const op = errors.Op("handler.Category.Update")
	req := CategoryUpdateRequest{}
	if err := bindAndValidate(c, &req); err != nil {
		return err
	}
	cat := core.CategoryPartial{
		Name:        req.Name,
		LocationIDs: req.LocationIDs,
	}
	ucat, err := h.CatalogService.UpdateCategory(c.Request().Context(), req.ID, cat)
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
	m, err := h.CatalogService.RetrieveCategory(c.Request().Context(), core.CategoryRetrieveRequest{
		ID:         c.Param("id"),
		MerchantID: ac.MerchantID,
	})
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
	cuss, err := h.CatalogService.ListCategory(c.Request().Context(), core.CategoryListRequest{
		Limit:      req.Limit,
		Offset:     req.Offset,
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
	cus, err := h.CatalogService.DeleteCategory(c.Request().Context(), core.CategoryDeleteRequest{
		ID:         c.Param("id"),
		MerchantID: ac.MerchantID,
	})
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
	Status      core.Status `json:"status"`
}

func NewCategory(cat core.Category) Category {
	return Category{
		ID:          cat.ID,
		Name:        cat.Name,
		LocationIDs: cat.LocationIDs,
		MerchantID:  cat.MerchantID,
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
