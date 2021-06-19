package handler

import (
	"net/http"

	"github.com/backium/backend/controller"
	"github.com/backium/backend/entity"
	"github.com/backium/backend/errors"
	"github.com/labstack/echo/v4"
)

type category struct {
	ID          string        `json:"id"`
	Name        string        `json:"name"`
	LocationIDs []string      `json:"location_ids"`
	MerchantID  string        `json:"merchant_id"`
	Status      entity.Status `json:"status"`
}

func newCategory(cat entity.Category) category {
	return category{
		ID:          cat.ID,
		Name:        cat.Name,
		LocationIDs: cat.LocationIDs,
		MerchantID:  cat.MerchantID,
		Status:      cat.Status,
	}
}

type createCategoryRequest struct {
	Name        string    `json:"name" validate:"required"`
	LocationIDs *[]string `json:"location_ids" validate:"omitempty,dive,required"`
}

type updateCategoryRequest struct {
	ID          string    `param:"id" validate:"required"`
	Name        *string   `json:"name" validate:"omitempty,min=1"`
	LocationIDs *[]string `json:"location_ids" validate:"omitempty,dive,required"`
}

type listAllCategoriesRequest struct {
	Limit  *int64 `query:"limit" validate:"omitempty,gte=1"`
	Offset *int64 `query:"offset"`
}

type listCategoriesResponse struct {
	Categories []category `json:"categories"`
}

type Category struct {
	Controller controller.Category
}

func (h *Category) Create(c echo.Context) error {
	const op = errors.Op("handler.Category.Create")
	ac, ok := c.(*AuthContext)
	if !ok {
		return errors.E(op, errors.KindUnexpected, "invalid echo.Context")
	}
	req := createCategoryRequest{}
	if err := bindAndValidate(c, &req); err != nil {
		return err
	}
	cat := entity.NewCategory()
	if req.LocationIDs != nil {
		cat.LocationIDs = *req.LocationIDs
	}
	cat.Name = req.Name
	cat.MerchantID = ac.MerchantID

	cat, err := h.Controller.Create(c.Request().Context(), cat)
	if err != nil {
		return errors.E(op, err)
	}
	return c.JSON(http.StatusOK, newCategory(cat))
}

func (h *Category) Update(c echo.Context) error {
	const op = errors.Op("handler.Category.Update")
	req := updateCategoryRequest{}
	if err := bindAndValidate(c, &req); err != nil {
		return err
	}
	cat := controller.PartialCategory{
		Name:        req.Name,
		LocationIDs: req.LocationIDs,
	}
	ucat, err := h.Controller.Update(c.Request().Context(), req.ID, cat)
	if err != nil {
		return errors.E(op, err)
	}
	return c.JSON(http.StatusOK, newCategory(ucat))
}

func (h *Category) Retrieve(c echo.Context) error {
	const op = errors.Op("handler.Category.Retrieve")
	ac, ok := c.(*AuthContext)
	if !ok {
		return errors.E(op, errors.KindUnexpected, "invalid echo.Context")
	}
	m, err := h.Controller.Retrieve(c.Request().Context(), controller.RetrieveCategoryRequest{
		ID:         c.Param("id"),
		MerchantID: ac.MerchantID,
	})
	if err != nil {
		return errors.E(op, err)
	}
	return c.JSON(http.StatusOK, newCategory(m))
}

func (h *Category) ListAll(c echo.Context) error {
	const op = errors.Op("handler.Category.ListAll")
	ac, ok := c.(*AuthContext)
	if !ok {
		return errors.E(op, errors.KindUnexpected, "invalid echo.Context")
	}
	req := listAllCategoriesRequest{}
	if err := bindAndValidate(c, &req); err != nil {
		return err
	}
	cuss, err := h.Controller.ListAll(c.Request().Context(), controller.ListAllCategoriesRequest{
		Limit:      req.Limit,
		Offset:     req.Offset,
		MerchantID: ac.MerchantID,
	})
	if err != nil {
		return errors.E(op, err)
	}
	res := make([]category, len(cuss))
	for i, cus := range cuss {
		res[i] = newCategory(cus)
	}
	return c.JSON(http.StatusOK, listCategoriesResponse{res})
}

func (h *Category) Delete(c echo.Context) error {
	const op = errors.Op("handler.Category.Delete")
	ac, ok := c.(*AuthContext)
	if !ok {
		return errors.E(op, errors.KindUnexpected, "invalid echo.Context")
	}
	cus, err := h.Controller.Delete(c.Request().Context(), controller.DeleteCategoryRequest{
		ID:         c.Param("id"),
		MerchantID: ac.MerchantID,
	})
	if err != nil {
		return errors.E(op, err)
	}
	return c.JSON(http.StatusOK, newCategory(cus))
}
