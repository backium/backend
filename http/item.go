package http

import (
	"net/http"

	"github.com/backium/backend/core"
	"github.com/backium/backend/errors"
	"github.com/labstack/echo/v4"
)

const (
	ItemListDefaultSize = 10
	ItemListMaxSize     = 50
)

func (h *Handler) CreateItem(c echo.Context) error {
	const op = errors.Op("http/Handler.CreateItem")
	ac, ok := c.(*AuthContext)
	if !ok {
		return errors.E(op, errors.KindUnexpected, "invalid echo.Context")
	}
	req := ItemCreateRequest{}
	if err := bindAndValidate(c, &req); err != nil {
		return err
	}
	item := core.NewItem(ac.MerchantID)
	if req.LocationIDs != nil {
		item.LocationIDs = *req.LocationIDs
	}
	item.Name = req.Name
	item.CategoryID = req.CategoryID
	item.Description = req.Description

	ctx := c.Request().Context()
	item, err := h.CatalogService.PutItem(c.Request().Context(), item)
	if err != nil {
		return errors.E(op, err)
	}
	variations, err := h.CatalogService.ListItemVariation(ctx, core.ItemVariationFilter{
		ItemIDs: []string{item.ID},
	})
	return c.JSON(http.StatusOK, NewItem(item, variations))
}

func (h *Handler) UpdateItem(c echo.Context) error {
	const op = errors.Op("http/Handler.UpdateItem")
	ac, ok := c.(*AuthContext)
	if !ok {
		return errors.E(op, errors.KindUnexpected, "invalid echo.Context")
	}
	req := ItemUpdateRequest{}
	if err := bindAndValidate(c, &req); err != nil {
		return err
	}
	ctx := c.Request().Context()
	item, err := h.CatalogService.GetItem(ctx, req.ID, ac.MerchantID, nil)
	if err != nil {
		return errors.E(op, err)
	}
	if req.Name != nil {
		item.Name = *req.Name
	}
	if req.Description != nil {
		item.Description = *req.Description
	}
	if req.LocationIDs != nil {
		item.LocationIDs = *req.LocationIDs
	}
	if req.CategoryID != nil {
		item.CategoryID = *req.CategoryID
	}

	item, err = h.CatalogService.PutItem(ctx, item)
	if err != nil {
		return errors.E(op, err)
	}
	variations, err := h.CatalogService.ListItemVariation(ctx, core.ItemVariationFilter{
		ItemIDs: []string{item.ID},
	})
	return c.JSON(http.StatusOK, NewItem(item, variations))
}

func (h *Handler) RetrieveItem(c echo.Context) error {
	const op = errors.Op("http/Handler.RetrieveItem")
	ac, ok := c.(*AuthContext)
	if !ok {
		return errors.E(op, errors.KindUnexpected, "invalid echo.Context")
	}
	ctx := c.Request().Context()
	item, err := h.CatalogService.GetItem(ctx, c.Param("id"), ac.MerchantID, nil)
	if err != nil {
		return errors.E(op, err)
	}
	variations, err := h.CatalogService.ListItemVariation(ctx, core.ItemVariationFilter{
		ItemIDs: []string{item.ID},
	})
	resp := NewItem(item, variations)
	return c.JSON(http.StatusOK, resp)
}

func (h *Handler) ListItems(c echo.Context) error {
	const op = errors.Op("http/Handler.ListItems")
	ac, ok := c.(*AuthContext)
	if !ok {
		return errors.E(op, errors.KindUnexpected, "invalid echo.Context")
	}
	req := ItemListRequest{}
	if err := bindAndValidate(c, &req); err != nil {
		return err
	}
	var limit, offset int64 = ItemListDefaultSize, 0
	if req.Limit <= ItemListMaxSize {
		limit = req.Limit
	}
	if req.Limit > ItemListMaxSize {
		limit = ItemListMaxSize
	}
	if req.Offset != 0 {
		offset = req.Offset
	}

	ctx := c.Request().Context()
	items, err := h.CatalogService.ListItem(ctx, core.ItemFilter{
		Limit:      limit,
		Offset:     offset,
		MerchantID: ac.MerchantID,
	})
	if err != nil {
		return errors.E(op, err)
	}

	itemIDs := make([]string, len(items))
	for i, item := range items {
		itemIDs[i] = item.ID
	}
	variations, err := h.CatalogService.ListItemVariation(ctx, core.ItemVariationFilter{
		ItemIDs: itemIDs,
	})
	resp := ItemListResponse{Items: make([]Item, len(items))}
	for i, item := range items {
		vars := item.ItemVariations(variations)
		resp.Items[i] = NewItem(item, vars)
	}
	return c.JSON(http.StatusOK, resp)
}

func (h *Handler) DeleteItem(c echo.Context) error {
	const op = errors.Op("http/Handler.DeleteItem")
	ac, ok := c.(*AuthContext)
	if !ok {
		return errors.E(op, errors.KindUnexpected, "invalid echo.Context")
	}
	ctx := c.Request().Context()
	item, err := h.CatalogService.DeleteItem(ctx, c.Param("id"), ac.MerchantID, nil)
	if err != nil {
		return errors.E(op, err)
	}
	variations, err := h.CatalogService.ListItemVariation(ctx, core.ItemVariationFilter{
		ItemIDs: []string{item.ID},
	})
	return c.JSON(http.StatusOK, NewItem(item, variations))
}

type Item struct {
	ID          string          `json:"id"`
	Name        string          `json:"name"`
	Description string          `json:"description"`
	CategoryID  string          `json:"category_id"`
	Variations  []ItemVariation `json:"variations"`
	LocationIDs []string        `json:"location_ids"`
	MerchantID  string          `json:"merchant_id"`
	Status      core.Status     `json:"status"`
}

func NewItem(item core.Item, variations []core.ItemVariation) Item {
	return Item{
		ID:          item.ID,
		Name:        item.Name,
		Description: item.Description,
		CategoryID:  item.CategoryID,
		Variations:  NewItemVariations(variations),
		LocationIDs: item.LocationIDs,
		MerchantID:  item.MerchantID,
		Status:      item.Status,
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
	Limit  int64 `query:"limit" validate:"gte=0"`
	Offset int64 `query:"offset" validate:"gte=0"`
}

type ItemListResponse struct {
	Items []Item `json:"items"`
}
