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
		ItemID      string        `json:"item_id" validate:"required"`
		Price       *MoneyRequest `json:"price" validate:"required"`
		Image       string        `json:"image"`
		LocationIDs *[]string     `json:"location_ids" validate:"omitempty,dive,required"`
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
		ID          string        `param:"id" validate:"required"`
		Name        *string       `json:"name" validate:"omitempty,min=1"`
		SKU         *string       `json:"sku"`
		Price       *MoneyRequest `json:"price"`
		Image       *string       `json:"image"`
		LocationIDs *[]string     `json:"location_ids" validate:"omitempty,dive,required"`
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

	variation, err := h.CatalogService.GetItemVariation(ctx, req.ID, merchant.ID, nil)
	if req.Price != nil {
		variation.Price = core.Money{
			Value:    ptr.GetInt64(req.Price.Value),
			Currency: req.Price.Currency,
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

	ctx := c.Request().Context()

	merchant := core.MerchantFromContext(ctx)
	if merchant == nil {
		return errors.E(op, errors.KindUnexpected, "invalid echo.Context")
	}

	variation, err := h.CatalogService.GetItemVariation(ctx, c.Param("id"), merchant.ID, nil)
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

	var limit, offset int64 = ItemVariationListDefaultSize, req.Offset
	if req.Limit <= ItemVariationListMaxSize {
		limit = req.Limit
	} else {
		limit = ItemVariationListMaxSize
	}

	variations, err := h.CatalogService.ListItemVariation(ctx, core.ItemVariationFilter{
		Limit:      limit,
		Offset:     offset,
		MerchantID: merchant.ID,
	})
	if err != nil {
		return errors.E(op, err)
	}

	resp := response{ItemVariations: make([]ItemVariation, len(variations))}
	for i, cus := range variations {
		resp.ItemVariations[i] = NewItemVariation(cus)
	}

	return c.JSON(http.StatusOK, resp)
}

func (h *Handler) HandleDeleteItemVariation(c echo.Context) error {
	const op = errors.Op("handler.ItemVariation.Delete")

	ctx := c.Request().Context()

	merchant := core.MerchantFromContext(ctx)
	if merchant == nil {
		return errors.E(op, errors.KindUnexpected, "invalid echo.Context")
	}

	variation, err := h.CatalogService.DeleteItemVariation(ctx, c.Param("id"), merchant.ID, nil)
	if err != nil {
		return errors.E(op, err)
	}

	return c.JSON(http.StatusOK, NewItemVariation(variation))
}

type ItemVariation struct {
	ID          string       `json:"id"`
	Name        string       `json:"name"`
	SKU         string       `json:"sku,omitempty"`
	ItemID      string       `json:"item_id"`
	Price       MoneyRequest `json:"price"`
	Image       string       `json:"image,omitempty"`
	LocationIDs []string     `json:"location_ids"`
	MerchantID  string       `json:"merchant_id"`
	CreatedAt   int64        `json:"created_at"`
	UpdatedAt   int64        `json:"updated_at"`
	Status      core.Status  `json:"status"`
}

func NewItemVariation(itvar core.ItemVariation) ItemVariation {
	return ItemVariation{
		ID:     itvar.ID,
		Name:   itvar.Name,
		SKU:    itvar.SKU,
		ItemID: itvar.ItemID,
		Price: MoneyRequest{
			Value:    &itvar.Price.Value,
			Currency: itvar.Price.Currency,
		},
		Image:       itvar.Image,
		LocationIDs: itvar.LocationIDs,
		MerchantID:  itvar.MerchantID,
		CreatedAt:   itvar.CreatedAt,
		UpdatedAt:   itvar.UpdatedAt,
		Status:      itvar.Status,
	}
}

func NewItemVariations(itvars []core.ItemVariation) []ItemVariation {
	vars := make([]ItemVariation, len(itvars))
	for i, itvar := range itvars {
		vars[i] = NewItemVariation(itvar)
	}
	return vars
}
