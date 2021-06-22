package http

import (
	"net/http"

	"github.com/backium/backend/core"
	"github.com/backium/backend/errors"
	"github.com/labstack/echo/v4"
)

func (h *Handler) CreateItem(c echo.Context) error {
	const op = errors.Op("handler.Item.Create")
	ac, ok := c.(*AuthContext)
	if !ok {
		return errors.E(op, errors.KindUnexpected, "invalid echo.Context")
	}
	req := ItemCreateRequest{}
	if err := bindAndValidate(c, &req); err != nil {
		return err
	}
	it := core.NewItem()
	if req.LocationIDs != nil {
		it.LocationIDs = *req.LocationIDs
	}
	it.Name = req.Name
	it.CategoryID = req.CategoryID
	it.Description = req.Description
	it.MerchantID = ac.MerchantID

	it, err := h.CatalogService.CreateItem(c.Request().Context(), it)
	if err != nil {
		return errors.E(op, err)
	}
	return c.JSON(http.StatusOK, NewItem(it))
}

func (h *Handler) UpdateItem(c echo.Context) error {
	const op = errors.Op("handler.Item.Update")
	req := ItemUpdateRequest{}
	if err := bindAndValidate(c, &req); err != nil {
		return err
	}
	it := core.PartialItem{
		Name:        req.Name,
		Description: req.Description,
		CategoryID:  req.CategoryID,
		LocationIDs: req.LocationIDs,
	}
	uit, err := h.CatalogService.UpdateItem(c.Request().Context(), req.ID, it)
	if err != nil {
		return errors.E(op, err)
	}
	return c.JSON(http.StatusOK, NewItem(uit))
}

func (h *Handler) RetrieveItem(c echo.Context) error {
	const op = errors.Op("handler.Item.Retrieve")
	ac, ok := c.(*AuthContext)
	if !ok {
		return errors.E(op, errors.KindUnexpected, "invalid echo.Context")
	}
	it, err := h.CatalogService.RetrieveItem(c.Request().Context(), core.ItemRetrieveRequest{
		ID:         c.Param("id"),
		MerchantID: ac.MerchantID,
	})
	if err != nil {
		return errors.E(op, err)
	}
	return c.JSON(http.StatusOK, NewItem(it))
}

func (h *Handler) ListItems(c echo.Context) error {
	const op = errors.Op("handler.Item.ListAll")
	ac, ok := c.(*AuthContext)
	if !ok {
		return errors.E(op, errors.KindUnexpected, "invalid echo.Context")
	}
	req := ItemListRequest{}
	if err := bindAndValidate(c, &req); err != nil {
		return err
	}
	its, err := h.CatalogService.ListItem(c.Request().Context(), core.ItemListRequest{
		Limit:      req.Limit,
		Offset:     req.Offset,
		MerchantID: ac.MerchantID,
	})
	if err != nil {
		return errors.E(op, err)
	}
	res := make([]Item, len(its))
	for i, it := range its {
		res[i] = NewItem(it)
	}
	return c.JSON(http.StatusOK, ItemListResponse{res})
}

func (h *Handler) DeleteItem(c echo.Context) error {
	const op = errors.Op("handler.Item.Delete")
	ac, ok := c.(*AuthContext)
	if !ok {
		return errors.E(op, errors.KindUnexpected, "invalid echo.Context")
	}
	it, err := h.CatalogService.DeleteItem(c.Request().Context(), core.ItemDeleteRequest{
		ID:         c.Param("id"),
		MerchantID: ac.MerchantID,
	})
	if err != nil {
		return errors.E(op, err)
	}
	return c.JSON(http.StatusOK, NewItem(it))
}

type Item struct {
	ID          string      `json:"id"`
	Name        string      `json:"name"`
	Description string      `json:"description"`
	CategoryID  string      `json:"category_id"`
	LocationIDs []string    `json:"location_ids"`
	MerchantID  string      `json:"merchant_id"`
	Status      core.Status `json:"status"`
}

func NewItem(it core.Item) Item {
	return Item{
		ID:          it.ID,
		Name:        it.Name,
		Description: it.Description,
		CategoryID:  it.CategoryID,
		LocationIDs: it.LocationIDs,
		MerchantID:  it.MerchantID,
		Status:      it.Status,
	}
}

type ItemCreateRequest struct {
	Name        string    `json:"name" validate:"required"`
	Description string    `json:"description" validate:"omitempty,max=100"`
	CategoryID  string    `json:"category_id" validate:"required"`
	LocationIDs *[]string `json:"location_ids" validate:"omitempty,dive,required"`
}

type ItemUpdateRequest struct {
	ID          string    `param:"id" validate:"required"`
	Name        *string   `json:"name" validate:"omitempty,min=1"`
	Description *string   `json:"description" validate:"omitempty,max=100"`
	CategoryID  *string   `json:"category_id" validate:"omitempty,min=1"`
	LocationIDs *[]string `json:"location_ids" validate:"omitempty,dive,required"`
}

type ItemListRequest struct {
	Limit  *int64 `query:"limit" validate:"omitempty,gte=1"`
	Offset *int64 `query:"offset"`
}

type ItemListResponse struct {
	Items []Item `json:"items"`
}
