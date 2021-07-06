package http

import (
	"net/http"

	"github.com/backium/backend/core"
	"github.com/backium/backend/errors"
	"github.com/backium/backend/ptr"
	"github.com/labstack/echo/v4"
)

const (
	ItemVariationListDefaultSize = 10
	ItemVariationListMaxSize     = 50
)

func (h *Handler) HandleCreateItemVariation(c echo.Context) error {
	const op = errors.Op("handler.ItemVariation.Create")

	type request struct {
		Name        string        `json:"name" validate:"required"`
		SKU         string        `json:"sku"`
		ItemID      core.ID       `json:"item_id" validate:"required"`
		Price       *MoneyRequest `json:"price" validate:"required"`
		Cost        *MoneyRequest `json:"cost" validate:"omitempty"`
		Image       string        `json:"image"`
		LocationIDs *[]core.ID    `json:"location_ids" validate:"omitempty,dive,required"`
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

	variation := core.NewItemVariation(req.Name, req.ItemID, merchant.ID)
	variation.SKU = req.SKU
	variation.Image = req.Image
	variation.Price = core.Money{
		Value:    ptr.GetInt64(req.Price.Value),
		Currency: req.Price.Currency,
	}
	if req.Cost != nil {
		variation.Cost = &core.Money{
			Value:    ptr.GetInt64(req.Cost.Value),
			Currency: req.Cost.Currency,
		}
	}
	if req.LocationIDs != nil {
		variation.LocationIDs = *req.LocationIDs
	}

	variation, err := h.CatalogService.PutItemVariation(ctx, variation)
	if err != nil {
		return errors.E(op, err)
	}

	return c.JSON(http.StatusOK, NewItemVariation(variation))
}

func (h *Handler) HandleUpdateItemVariation(c echo.Context) error {
	const op = errors.Op("handler.ItemVariation.Update")

	type request struct {
		ID          core.ID       `param:"id" validate:"required"`
		Name        *string       `json:"name" validate:"omitempty,min=1"`
		SKU         *string       `json:"sku"`
		Price       *MoneyRequest `json:"price"`
		Cost        *MoneyRequest `json:"cost"`
		Image       *string       `json:"image"`
		LocationIDs *[]core.ID    `json:"location_ids" validate:"omitempty,dive,required"`
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

	variation, err := h.CatalogService.GetItemVariation(ctx, req.ID)
	if req.Price != nil {
		variation.Price = core.Money{
			Value:    ptr.GetInt64(req.Price.Value),
			Currency: req.Price.Currency,
		}
	}
	if req.Cost != nil {
		variation.Cost = &core.Money{
			Value:    ptr.GetInt64(req.Cost.Value),
			Currency: req.Cost.Currency,
		}
	}
	if req.Name != nil {
		variation.Name = *req.Name
	}
	if req.SKU != nil {
		variation.SKU = *req.SKU
	}
	if req.Image != nil {
		variation.Image = *req.Image
	}
	if req.LocationIDs != nil {
		variation.LocationIDs = *req.LocationIDs
	}

	variation, err = h.CatalogService.PutItemVariation(ctx, variation)
	if err != nil {
		return errors.E(op, err)
	}

	return c.JSON(http.StatusOK, NewItemVariation(variation))
}

func (h *Handler) HandleRetrieveItemVariation(c echo.Context) error {
	const op = errors.Op("handler.ItemVariation.Retrieve")

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

	variation, err := h.CatalogService.GetItemVariation(ctx, req.ID)
	if err != nil {
		return errors.E(op, err)
	}

	return c.JSON(http.StatusOK, NewItemVariation(variation))
}

func (h *Handler) HandleListItemVariations(c echo.Context) error {
	const op = errors.Op("handler.ItemVariation.ListAll")

	type request struct {
		Limit  int64 `query:"limit" validate:"gte=0"`
		Offset int64 `query:"offset" validate:"gte=0"`
	}

	type response struct {
		ItemVariations []ItemVariation `json:"item_variations"`
		Total          int64           `json:"total_count"`
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

	var limit int64 = ItemVariationListDefaultSize
	if req.Limit <= ItemVariationListMaxSize {
		limit = req.Limit
	} else {
		limit = ItemVariationListMaxSize
	}

	variations, count, err := h.CatalogService.ListItemVariation(ctx, core.ItemVariationQuery{
		Limit:  limit,
		Offset: req.Offset,
		Filter: core.ItemVariationFilter{MerchantID: merchant.ID},
	})
	if err != nil {
		return errors.E(op, err)
	}

	resp := response{
		ItemVariations: make([]ItemVariation, len(variations)),
		Total:          count,
	}
	for i, v := range variations {
		resp.ItemVariations[i] = NewItemVariation(v)
	}

	return c.JSON(http.StatusOK, resp)
}

func (h *Handler) HandleSearchItemVariation(c echo.Context) error {
	const op = errors.Op("http/Handler.HandleSearchItemVariation")

	type filter struct {
		IDs         []core.ID `json:"ids" validate:"omitempty,dive,id"`
		LocationIDs []core.ID `json:"location_ids" validate:"omitempty,dive,id"`
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
		ItemVariations []ItemVariation `json:"item_variations"`
		Total          int64           `json:"total_count"`
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

	var limit int64 = ItemVariationListDefaultSize
	if req.Limit <= ItemVariationListMaxSize {
		limit = req.Limit
	} else {
		limit = ItemVariationListMaxSize
	}

	variations, count, err := h.CatalogService.ListItemVariation(ctx, core.ItemVariationQuery{
		Limit:  limit,
		Offset: req.Offset,
		Filter: core.ItemVariationFilter{
			Name:        req.Filter.Name,
			LocationIDs: req.Filter.LocationIDs,
			MerchantID:  merchant.ID,
		},
		Sort: core.ItemVariationSort{
			Name: req.Sort.Name,
		},
	})
	if err != nil {
		return errors.E(op, err)
	}

	resp := response{
		ItemVariations: make([]ItemVariation, len(variations)),
		Total:          count,
	}
	for i, v := range variations {
		resp.ItemVariations[i] = NewItemVariation(v)
	}

	return c.JSON(http.StatusOK, resp)
}

func (h *Handler) HandleDeleteItemVariation(c echo.Context) error {
	const op = errors.Op("handler.ItemVariation.Delete")

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

	variation, err := h.CatalogService.DeleteItemVariation(ctx, req.ID)
	if err != nil {
		return errors.E(op, err)
	}

	return c.JSON(http.StatusOK, NewItemVariation(variation))
}

type ItemVariation struct {
	ID          core.ID     `json:"id"`
	Name        string      `json:"name"`
	SKU         string      `json:"sku,omitempty"`
	ItemID      core.ID     `json:"item_id"`
	Price       Money       `json:"price"`
	Cost        *Money      `json:"cost,omitempty"`
	Image       string      `json:"image,omitempty"`
	LocationIDs []core.ID   `json:"location_ids"`
	MerchantID  core.ID     `json:"merchant_id"`
	CreatedAt   int64       `json:"created_at"`
	UpdatedAt   int64       `json:"updated_at"`
	Status      core.Status `json:"status"`
}

func NewItemVariation(variation core.ItemVariation) ItemVariation {
	var cost *Money
	if variation.Cost != nil {
		cost = &Money{variation.Cost.Value, variation.Cost.Currency}
	}
	return ItemVariation{
		ID:          variation.ID,
		Name:        variation.Name,
		SKU:         variation.SKU,
		ItemID:      variation.ItemID,
		Price:       NewMoney(variation.Price),
		Cost:        cost,
		Image:       variation.Image,
		LocationIDs: variation.LocationIDs,
		MerchantID:  variation.MerchantID,
		CreatedAt:   variation.CreatedAt,
		UpdatedAt:   variation.UpdatedAt,
		Status:      variation.Status,
	}
}

func NewItemVariations(variations []core.ItemVariation) []ItemVariation {
	resp := make([]ItemVariation, len(variations))
	for i, v := range variations {
		resp[i] = NewItemVariation(v)
	}
	return resp
}
