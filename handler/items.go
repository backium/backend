package handler

import (
	"net/http"

	"github.com/backium/backend/controller"
	"github.com/backium/backend/entity"
	"github.com/backium/backend/errors"
	"github.com/labstack/echo/v4"
)

type item struct {
	ID          string        `json:"id"`
	Name        string        `json:"name"`
	Description string        `json:"description"`
	CategoryID  string        `json:"category_id"`
	LocationIDs []string      `json:"location_ids"`
	MerchantID  string        `json:"merchant_id"`
	Status      entity.Status `json:"status"`
}

func newItem(it entity.Item) item {
	return item{
		ID:          it.ID,
		Name:        it.Name,
		Description: it.Description,
		CategoryID:  it.CategoryID,
		LocationIDs: it.LocationIDs,
		MerchantID:  it.MerchantID,
		Status:      it.Status,
	}
}

type createItemRequest struct {
	Name        string    `json:"name" validate:"required"`
	Description string    `json:"description" validate:"omitempty,max=100"`
	CategoryID  string    `json:"category_id" validate:"required"`
	LocationIDs *[]string `json:"location_ids" validate:"omitempty,dive,required"`
}

type updateItemRequest struct {
	ID          string    `param:"id" validate:"required"`
	Name        *string   `json:"name" validate:"omitempty,min=1"`
	Description *string   `json:"description" validate:"omitempty,max=100"`
	CategoryID  *string   `json:"category_id" validate:"omitempty,min=1"`
	LocationIDs *[]string `json:"location_ids" validate:"omitempty,dive,required"`
}

type listAllItemsRequest struct {
	Limit  *int64 `query:"limit" validate:"omitempty,gte=1"`
	Offset *int64 `query:"offset"`
}

type listItemsResponse struct {
	Items []item `json:"items"`
}

type Item struct {
	Controller controller.Item
}

func (h *Item) Create(c echo.Context) error {
	const op = errors.Op("handler.Item.Create")
	ac, ok := c.(*AuthContext)
	if !ok {
		return errors.E(op, errors.KindUnexpected, "invalid echo.Context")
	}
	req := createItemRequest{}
	if err := bindAndValidate(c, &req); err != nil {
		return err
	}
	it := entity.NewItem()
	if req.LocationIDs != nil {
		it.LocationIDs = *req.LocationIDs
	}
	it.Name = req.Name
	it.CategoryID = req.CategoryID
	it.Description = req.Description
	it.MerchantID = ac.MerchantID

	it, err := h.Controller.Create(c.Request().Context(), it)
	if err != nil {
		return errors.E(op, err)
	}
	return c.JSON(http.StatusOK, newItem(it))
}

func (h *Item) Update(c echo.Context) error {
	const op = errors.Op("handler.Item.Update")
	req := updateItemRequest{}
	if err := bindAndValidate(c, &req); err != nil {
		return err
	}
	it := controller.PartialItem{
		Name:        req.Name,
		Description: req.Description,
		CategoryID:  req.CategoryID,
		LocationIDs: req.LocationIDs,
	}
	uit, err := h.Controller.Update(c.Request().Context(), req.ID, it)
	if err != nil {
		return errors.E(op, err)
	}
	return c.JSON(http.StatusOK, newItem(uit))
}

func (h *Item) Retrieve(c echo.Context) error {
	const op = errors.Op("handler.Item.Retrieve")
	ac, ok := c.(*AuthContext)
	if !ok {
		return errors.E(op, errors.KindUnexpected, "invalid echo.Context")
	}
	it, err := h.Controller.Retrieve(c.Request().Context(), controller.RetrieveItemRequest{
		ID:         c.Param("id"),
		MerchantID: ac.MerchantID,
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, newItem(it))
}

func (h *Item) ListAll(c echo.Context) error {
	const op = errors.Op("handler.Item.ListAll")
	ac, ok := c.(*AuthContext)
	if !ok {
		return errors.E(op, errors.KindUnexpected, "invalid echo.Context")
	}
	req := listAllItemsRequest{}
	if err := bindAndValidate(c, &req); err != nil {
		return err
	}
	its, err := h.Controller.ListAll(c.Request().Context(), controller.ListAllItemsRequest{
		Limit:      req.Limit,
		Offset:     req.Offset,
		MerchantID: ac.MerchantID,
	})
	if err != nil {
		return errors.E(op, err)
	}
	res := make([]item, len(its))
	for i, it := range its {
		res[i] = newItem(it)
	}
	return c.JSON(http.StatusOK, listItemsResponse{res})
}

func (h *Item) Delete(c echo.Context) error {
	const op = errors.Op("handler.Item.Delete")
	ac, ok := c.(*AuthContext)
	if !ok {
		return errors.E(op, errors.KindUnexpected, "invalid echo.Context")
	}
	it, err := h.Controller.Delete(c.Request().Context(), controller.DeleteItemRequest{
		ID:         c.Param("id"),
		MerchantID: ac.MerchantID,
	})
	if err != nil {
		return errors.E(op, err)
	}
	return c.JSON(http.StatusOK, newItem(it))
}
