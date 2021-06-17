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
	Name        string   `json:"name"`
	Description string   `json:"description"`
	CategoryID  string   `json:"category_id"`
	LocationIDs []string `json:"location_ids"`
}

type updateItemRequest struct {
	ID          string   `param:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	CategoryID  string   `json:"category_id"`
	LocationIDs []string `json:"location_ids"`
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
	cat := entity.Item{
		Name:        req.Name,
		Description: req.Description,
		CategoryID:  req.CategoryID,
		LocationIDs: req.LocationIDs,
		MerchantID:  ac.MerchantID,
	}
	cat, err := h.Controller.Create(c.Request().Context(), cat)
	if err != nil {
		return errors.E(op, err)
	}
	return c.JSON(http.StatusOK, newItem(cat))
}

func (h *Item) Update(c echo.Context) error {
	const op = errors.Op("handler.Item.Update")
	ac, ok := c.(*AuthContext)
	if !ok {
		return errors.E(op, errors.KindUnexpected, "invalid echo.Context")
	}
	req := updateItemRequest{}
	if err := bindAndValidate(c, &req); err != nil {
		return err
	}
	it := entity.Item{
		ID:          req.ID,
		Name:        req.Name,
		Description: req.Description,
		CategoryID:  req.CategoryID,
		LocationIDs: req.LocationIDs,
		MerchantID:  ac.MerchantID,
	}
	it, err := h.Controller.Update(c.Request().Context(), it)
	if err != nil {
		return errors.E(op, err)
	}
	return c.JSON(http.StatusOK, newItem(it))
}

func (h *Item) Retrieve(c echo.Context) error {
	const op = errors.Op("handler.Item.Retrieve")
	ac, ok := c.(*AuthContext)
	if !ok {
		return errors.E(op, errors.KindUnexpected, "invalid echo.Context")
	}
	m, err := h.Controller.Retrieve(c.Request().Context(), controller.RetrieveItemRequest{
		ID:         c.Param("id"),
		MerchantID: ac.MerchantID,
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, newItem(m))
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
	cuss, err := h.Controller.ListAll(c.Request().Context(), controller.ListAllItemsRequest{
		Limit:      req.Limit,
		Offset:     req.Offset,
		MerchantID: ac.MerchantID,
	})
	if err != nil {
		return errors.E(op, err)
	}
	res := make([]item, len(cuss))
	for i, cus := range cuss {
		res[i] = newItem(cus)
	}
	return c.JSON(http.StatusOK, listItemsResponse{res})
}

func (h *Item) Delete(c echo.Context) error {
	const op = errors.Op("handler.Item.Delete")
	ac, ok := c.(*AuthContext)
	if !ok {
		return errors.E(op, errors.KindUnexpected, "invalid echo.Context")
	}
	cus, err := h.Controller.Delete(c.Request().Context(), controller.DeleteItemRequest{
		ID:         c.Param("id"),
		MerchantID: ac.MerchantID,
	})
	if err != nil {
		return errors.E(op, err)
	}
	return c.JSON(http.StatusOK, newItem(cus))
}
