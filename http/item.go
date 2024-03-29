package http

import (
	"fmt"
	"net/http"

	"github.com/backium/backend/core"
	"github.com/backium/backend/errors"
	"github.com/labstack/echo/v4"
)

func (h *Handler) HandleCreateItem(c echo.Context) error {
	const op = errors.Op("http/Handler.CreateItem")

	type request struct {
		Name         string     `json:"name" validate:"required"`
		Description  string     `json:"description" validate:"omitempty,max=100"`
		CategoryID   core.ID    `json:"category_id" validate:"required"`
		EnabledInPOS bool       `json:"enabled"`
		LocationIDs  *[]core.ID `json:"location_ids" validate:"omitempty,dive,required"`
	}

	ctx := c.Request().Context()

	merchant := core.MerchantFromContext(ctx)
	if merchant == nil {
		return errors.E(op, errors.KindUnexpected, "invalid echo.Context")
	}
	employee := core.EmployeeFromContext(ctx)
	if employee == nil {
		return errors.E(op, errors.KindUnexpected, "invalid echo.Context")
	}

	req := request{}
	if err := bindAndValidate(c, &req); err != nil {
		return err
	}

	item := core.NewItem(req.Name, req.CategoryID, merchant.ID)
	item.Description = req.Description
	item.EnabledInPOS = req.EnabledInPOS
	item.LastModifiedBy = employee.ID
	if req.LocationIDs != nil {
		item.LocationIDs = *req.LocationIDs
	}

	item, err := h.CatalogService.PutItem(c.Request().Context(), item)
	if err != nil {
		return errors.E(op, err)
	}
	variations, _, err := h.CatalogService.ListItemVariation(ctx, core.ItemVariationQuery{
		Filter: core.ItemVariationFilter{
			ItemIDs: []core.ID{item.ID},
		},
	})
	return c.JSON(http.StatusOK, NewItem(item, variations))
}

func (h *Handler) HandleUpdateItem(c echo.Context) error {
	const op = errors.Op("http/Handler.UpdateItem")

	type request struct {
		ID           core.ID    `param:"id" validate:"required"`
		Name         *string    `json:"name" validate:"omitempty,min=1"`
		Description  *string    `json:"description" validate:"omitempty,max=100"`
		CategoryID   *core.ID   `json:"category_id" validate:"omitempty,min=1"`
		EnabledInPOS *bool      `json:"enabled"`
		LocationIDs  *[]core.ID `json:"location_ids" validate:"omitempty,dive,required"`
	}

	ctx := c.Request().Context()

	merchant := core.MerchantFromContext(ctx)
	if merchant == nil {
		return errors.E(op, errors.KindUnexpected, "invalid echo.Context")
	}
	employee := core.EmployeeFromContext(ctx)
	if employee == nil {
		return errors.E(op, errors.KindUnexpected, "invalid echo.Context")
	}

	req := request{}
	if err := bindAndValidate(c, &req); err != nil {
		return err
	}

	item, err := h.CatalogService.GetItem(ctx, req.ID)
	if err != nil {
		return errors.E(op, err)
	}
	item.LastModifiedBy = employee.ID
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
	if req.EnabledInPOS != nil {
		item.EnabledInPOS = *req.EnabledInPOS
	}

	item, err = h.CatalogService.PutItem(ctx, item)
	if err != nil {
		return errors.E(op, err)
	}

	variations, _, err := h.CatalogService.ListItemVariation(ctx, core.ItemVariationQuery{
		Filter: core.ItemVariationFilter{
			ItemIDs: []core.ID{item.ID},
		},
	})

	return c.JSON(http.StatusOK, NewItem(item, variations))
}

func (h *Handler) HandleRetrieveItem(c echo.Context) error {
	const op = errors.Op("http/Handler.RetrieveItem")

	type request struct {
		ID core.ID `param:"id"`
	}

	ctx := c.Request().Context()

	req := request{}
	if err := bindAndValidate(c, &req); err != nil {
		return err
	}

	merchant := core.MerchantFromContext(ctx)
	if merchant == nil {
		return errors.E(op, errors.KindUnexpected, "invalid echo.Context")
	}

	item, err := h.CatalogService.GetItem(ctx, req.ID)
	if err != nil {
		return errors.E(op, err)
	}

	variations, _, err := h.CatalogService.ListItemVariation(ctx, core.ItemVariationQuery{
		Filter: core.ItemVariationFilter{
			ItemIDs: []core.ID{item.ID},
		},
	})
	resp := NewItem(item, variations)

	return c.JSON(http.StatusOK, resp)
}

func (h *Handler) HandleListItems(c echo.Context) error {
	const op = errors.Op("http/Handler.ListItems")

	type request struct {
		Limit  int64 `query:"limit" validate:"gte=0"`
		Offset int64 `query:"offset" validate:"gte=0"`
	}

	type response struct {
		Items []Item `json:"items"`
		Total int64  `json:"total_count"`
	}

	ctx := c.Request().Context()

	merchant := core.MerchantFromContext(ctx)
	if merchant == nil {
		fmt.Println("hey")
		return errors.E(op, errors.KindUnexpected, "invalid echo.Context")
	}

	req := request{}
	if err := bindAndValidate(c, &req); err != nil {
		return err
	}

	query := core.ItemQuery{
		Limit:  req.Limit,
		Offset: req.Offset,
		Filter: core.ItemFilter{MerchantID: merchant.ID},
	}

	items, count, err := h.CatalogService.ListItem(ctx, query)
	if err != nil {
		return errors.E(op, err)
	}

	itemIDs := make([]core.ID, len(items))
	for i, item := range items {
		itemIDs[i] = item.ID
	}
	variations, _, err := h.CatalogService.ListItemVariation(ctx, core.ItemVariationQuery{
		Filter: core.ItemVariationFilter{
			ItemIDs: itemIDs,
		},
	})

	resp := response{
		Items: make([]Item, len(items)),
		Total: count,
	}
	for i, item := range items {
		vars := item.ItemVariations(variations)
		resp.Items[i] = NewItem(item, vars)
	}

	return c.JSON(http.StatusOK, resp)
}

func (h *Handler) HandleSearchItem(c echo.Context) error {
	const op = errors.Op("http/Handler.ListItems")

	type filter struct {
		IDs         []core.ID `json:"ids" validate:"omitempty,dive,id"`
		LocationIDs []core.ID `json:"location_ids" validate:"omitempty,dive,id"`
		CategoryIDs []core.ID `json:"category_ids" validate:"omitempty,dive,id"`
		Name        string    `json:"name"`
	}

	type sort struct {
		Name core.SortOrder `json:"name"`
	}

	type request struct {
		Limit  int64  `json:"limit" validate:"gte=0"`
		Offset int64  `json:"offset" validate:"gte=0"`
		Filter filter `json:"filter"`
		Sort   sort   `json:"sort"`
	}

	type response struct {
		Items []Item `json:"items"`
		Total int64  `json:"total_count"`
	}

	ctx := c.Request().Context()

	merchant := core.MerchantFromContext(ctx)
	if merchant == nil {
		return errors.E(op, errors.KindUnexpected, "invalid echo.Context")
	}

	req := request{}
	if err := bindAndValidate(c, &req); err != nil {
		return err
	}

	query := core.ItemQuery{
		Limit:  req.Limit,
		Offset: req.Offset,
		Filter: core.ItemFilter{
			Name:        req.Filter.Name,
			CategoryIDs: req.Filter.CategoryIDs,
			LocationIDs: req.Filter.LocationIDs,
			MerchantID:  merchant.ID,
		},
		Sort: core.ItemSort{
			Name: req.Sort.Name,
		},
	}

	items, count, err := h.CatalogService.ListItem(ctx, query)
	if err != nil {
		return errors.E(op, err)
	}

	itemIDs := make([]core.ID, len(items))
	for i, item := range items {
		itemIDs[i] = item.ID
	}
	variations, _, err := h.CatalogService.ListItemVariation(ctx, core.ItemVariationQuery{
		Filter: core.ItemVariationFilter{ItemIDs: itemIDs},
	})

	resp := response{
		Items: make([]Item, len(items)),
		Total: count,
	}
	for i, item := range items {
		vars := item.ItemVariations(variations)
		resp.Items[i] = NewItem(item, vars)
	}

	return c.JSON(http.StatusOK, resp)
}

func (h *Handler) HandleDeleteItem(c echo.Context) error {
	const op = errors.Op("http/Handler.DeleteItem")

	type request struct {
		ID core.ID `param:"id"`
	}

	ctx := c.Request().Context()

	req := request{}
	if err := bindAndValidate(c, &req); err != nil {
		return err
	}

	merchant := core.MerchantFromContext(ctx)
	if merchant == nil {
		return errors.E(op, errors.KindUnexpected, "invalid echo.Context")
	}
	employee := core.EmployeeFromContext(ctx)
	if employee == nil {
		return errors.E(op, errors.KindUnexpected, "invalid echo.Context")
	}

	item, err := h.CatalogService.DeleteItem(ctx, req.ID)
	if err != nil {
		return errors.E(op, err)
	}
	item.LastModifiedBy = employee.ID
	variations, _, err := h.CatalogService.ListItemVariation(ctx, core.ItemVariationQuery{
		Filter: core.ItemVariationFilter{
			ItemIDs: []core.ID{item.ID},
		},
	})

	return c.JSON(http.StatusOK, NewItem(item, variations))
}

type Item struct {
	ID           core.ID         `json:"id"`
	Name         string          `json:"name"`
	Description  string          `json:"description"`
	CategoryID   core.ID         `json:"category_id"`
	Variations   []ItemVariation `json:"variations"`
	LocationIDs  []core.ID       `json:"location_ids"`
	EnabledInPOS bool            `json:"enabled"`
	MerchantID   core.ID         `json:"merchant_id"`
	Status       core.Status     `json:"status"`
}

func NewItem(item core.Item, variations []core.ItemVariation) Item {
	return Item{
		ID:           item.ID,
		Name:         item.Name,
		Description:  item.Description,
		CategoryID:   item.CategoryID,
		Variations:   NewItemVariations(variations),
		LocationIDs:  item.LocationIDs,
		EnabledInPOS: item.EnabledInPOS,
		MerchantID:   item.MerchantID,
		Status:       item.Status,
	}
}
