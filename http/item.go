package http

import (
	"net/http"

	"github.com/backium/backend/core"
	"github.com/backium/backend/errors"
	"github.com/backium/backend/ptr"
	"github.com/labstack/echo/v4"
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
	it := core.NewItem()
	if req.LocationIDs != nil {
		it.LocationIDs = *req.LocationIDs
	}
	it.Name = req.Name
	it.CategoryID = req.CategoryID
	it.Description = req.Description
	it.MerchantID = ac.MerchantID

	it, err := h.CatalogService.PutItem(c.Request().Context(), it)
	if err != nil {
		return errors.E(op, err)
	}
	return c.JSON(http.StatusOK, NewItem(it))
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
	it, err := h.CatalogService.GetItem(ctx, req.ID, ac.MerchantID, nil)
	if err != nil {
		return errors.E(op, err)
	}
	if req.Name != nil {
		it.Name = *req.Name
	}
	if req.Description != nil {
		it.Description = *req.Description
	}
	if req.LocationIDs != nil {
		it.LocationIDs = *req.LocationIDs
	}
	if req.CategoryID != nil {
		it.CategoryID = *req.CategoryID
	}

	uit, err := h.CatalogService.PutItem(ctx, it)
	if err != nil {
		return errors.E(op, err)
	}
	return c.JSON(http.StatusOK, NewItem(uit))
}

func (h *Handler) RetrieveItem(c echo.Context) error {
	const op = errors.Op("http/Handler.RetrieveItem")
	ac, ok := c.(*AuthContext)
	if !ok {
		return errors.E(op, errors.KindUnexpected, "invalid echo.Context")
	}
	ctx := c.Request().Context()
	it, err := h.CatalogService.GetItem(ctx, c.Param("id"), ac.MerchantID, nil)
	if err != nil {
		return errors.E(op, err)
	}
	itvars, err := h.CatalogService.ListItemVariation(ctx, core.ItemVariationFilter{
		ItemIDs: []string{it.ID},
	})
	resp := NewItem(it)
	for _, itvar := range itvars {
		resp.Variations = append(resp.Variations, NewItemVariation(itvar))
	}

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
	ctx := c.Request().Context()
	its, err := h.CatalogService.ListItem(ctx, core.ItemFilter{
		Limit:      ptr.GetInt64(req.Limit),
		Offset:     ptr.GetInt64(req.Offset),
		MerchantID: ac.MerchantID,
	})
	if err != nil {
		return errors.E(op, err)
	}

	ids := make([]string, len(its))
	for i, it := range its {
		ids[i] = it.ID
	}
	itvars, err := h.CatalogService.ListItemVariation(ctx, core.ItemVariationFilter{
		ItemIDs: ids,
	})
	resp := make([]Item, len(its))
	for i, it := range its {
		vars := it.ItemVariations(itvars)
		resp[i] = NewItem(it)
		resp[i].Variations = NewItemVariations(vars)
	}
	return c.JSON(http.StatusOK, ItemListResponse{resp})
}

func (h *Handler) DeleteItem(c echo.Context) error {
	const op = errors.Op("http/Handler.DeleteItem")
	ac, ok := c.(*AuthContext)
	if !ok {
		return errors.E(op, errors.KindUnexpected, "invalid echo.Context")
	}
	ctx := c.Request().Context()
	it, err := h.CatalogService.DeleteItem(ctx, c.Param("id"), ac.MerchantID, nil)
	if err != nil {
		return errors.E(op, err)
	}
	return c.JSON(http.StatusOK, NewItem(it))
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

func NewItem(it core.Item) Item {
	return Item{
		ID:          it.ID,
		Name:        it.Name,
		Description: it.Description,
		CategoryID:  it.CategoryID,
		Variations:  []ItemVariation{},
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
