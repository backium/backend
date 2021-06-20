package handler

import (
	"net/http"

	"github.com/backium/backend/base"
	"github.com/backium/backend/catalog"
	"github.com/backium/backend/errors"
	"github.com/labstack/echo/v4"
)

type Category struct {
	Controller catalog.Controller
}

func (h *Category) Create(c echo.Context) error {
	const op = errors.Op("handler.Category.Create")
	ac, ok := c.(*AuthContext)
	if !ok {
		return errors.E(op, errors.KindUnexpected, "invalid echo.Context")
	}
	req := CategoryCreateRequest{}
	if err := bindAndValidate(c, &req); err != nil {
		return err
	}
	cat := catalog.NewCategory()
	if req.LocationIDs != nil {
		cat.LocationIDs = *req.LocationIDs
	}
	cat.Name = req.Name
	cat.MerchantID = ac.MerchantID

	cat, err := h.Controller.CreateCategory(c.Request().Context(), cat)
	if err != nil {
		return errors.E(op, err)
	}
	return c.JSON(http.StatusOK, newCategoryResponse(cat))
}

func (h *Category) Update(c echo.Context) error {
	const op = errors.Op("handler.Category.Update")
	req := CategoryUpdateRequest{}
	if err := bindAndValidate(c, &req); err != nil {
		return err
	}
	cat := catalog.CategoryPartial{
		Name:        req.Name,
		LocationIDs: req.LocationIDs,
	}
	ucat, err := h.Controller.UpdateCategory(c.Request().Context(), req.ID, cat)
	if err != nil {
		return errors.E(op, err)
	}
	return c.JSON(http.StatusOK, newCategoryResponse(ucat))
}

func (h *Category) Retrieve(c echo.Context) error {
	const op = errors.Op("handler.Category.Retrieve")
	ac, ok := c.(*AuthContext)
	if !ok {
		return errors.E(op, errors.KindUnexpected, "invalid echo.Context")
	}
	m, err := h.Controller.RetrieveCategory(c.Request().Context(), catalog.CategoryRetrieveRequest{
		ID:         c.Param("id"),
		MerchantID: ac.MerchantID,
	})
	if err != nil {
		return errors.E(op, err)
	}
	return c.JSON(http.StatusOK, newCategoryResponse(m))
}

func (h *Category) ListAll(c echo.Context) error {
	const op = errors.Op("handler.Category.ListAll")
	ac, ok := c.(*AuthContext)
	if !ok {
		return errors.E(op, errors.KindUnexpected, "invalid echo.Context")
	}
	req := CategoryListRequest{}
	if err := bindAndValidate(c, &req); err != nil {
		return err
	}
	cuss, err := h.Controller.ListCategory(c.Request().Context(), catalog.CategoryListRequest{
		Limit:      req.Limit,
		Offset:     req.Offset,
		MerchantID: ac.MerchantID,
	})
	if err != nil {
		return errors.E(op, err)
	}
	res := make([]CategoryResponse, len(cuss))
	for i, cus := range cuss {
		res[i] = newCategoryResponse(cus)
	}
	return c.JSON(http.StatusOK, CategoryListResponse{res})
}

func (h *Category) Delete(c echo.Context) error {
	const op = errors.Op("handler.Category.Delete")
	ac, ok := c.(*AuthContext)
	if !ok {
		return errors.E(op, errors.KindUnexpected, "invalid echo.Context")
	}
	cus, err := h.Controller.DeleteCategory(c.Request().Context(), catalog.CategoryDeleteRequest{
		ID:         c.Param("id"),
		MerchantID: ac.MerchantID,
	})
	if err != nil {
		return errors.E(op, err)
	}
	return c.JSON(http.StatusOK, newCategoryResponse(cus))
}

type CategoryResponse struct {
	ID          string      `json:"id"`
	Name        string      `json:"name"`
	LocationIDs []string    `json:"location_ids"`
	MerchantID  string      `json:"merchant_id"`
	Status      base.Status `json:"status"`
}

func newCategoryResponse(cat catalog.Category) CategoryResponse {
	return CategoryResponse{
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
	Categories []CategoryResponse `json:"categories"`
}
