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

type Money struct {
	Value    *int64 `json:"value" validate:"required"`
	Currency string `json:"currency" validate:"required"`
}

func (h *Handler) CreateItemVariation(c echo.Context) error {
	const op = errors.Op("handler.ItemVariation.Create")
	ac, ok := c.(*AuthContext)
	if !ok {
		return errors.E(op, errors.KindUnexpected, "invalid echo.Context")
	}
	req := ItemVariationCreateRequest{}
	if err := bindAndValidate(c, &req); err != nil {
		return err
	}

	variation := core.NewItemVariation(ac.MerchantID)
	// override defaults
	if req.LocationIDs != nil {
		variation.LocationIDs = *req.LocationIDs
	}
	variation.Name = req.Name
	variation.SKU = req.SKU
	variation.ItemID = req.ItemID
	variation.Image = req.Image
	variation.Price = core.Money{
		Value:    ptr.GetInt64(req.Price.Value),
		Currency: req.Price.Currency,
	}

	ctx := c.Request().Context()
	variation, err := h.CatalogService.PutItemVariation(ctx, variation)
	if err != nil {
		return errors.E(op, err)
	}
	return c.JSON(http.StatusOK, NewItemVariation(variation))
}

func (h *Handler) UpdateItemVariation(c echo.Context) error {
	const op = errors.Op("handler.ItemVariation.Update")
	ac, ok := c.(*AuthContext)
	if !ok {
		return errors.E(op, errors.KindUnexpected, "invalid echo.Context")
	}
	req := ItemVariationUpdateRequest{}
	if err := bindAndValidate(c, &req); err != nil {
		return err
	}

	ctx := c.Request().Context()
	variation, err := h.CatalogService.GetItemVariation(ctx, c.Param("id"), ac.MerchantID, nil)
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

func (h *Handler) RetrieveItemVariation(c echo.Context) error {
	const op = errors.Op("handler.ItemVariation.Retrieve")
	ac, ok := c.(*AuthContext)
	if !ok {
		return errors.E(op, errors.KindUnexpected, "invalid echo.Context")
	}
	ctx := c.Request().Context()
	variation, err := h.CatalogService.GetItemVariation(ctx, c.Param("id"), ac.MerchantID, nil)
	if err != nil {
		return errors.E(op, err)
	}
	return c.JSON(http.StatusOK, NewItemVariation(variation))
}

func (h *Handler) ListItemVariations(c echo.Context) error {
	const op = errors.Op("handler.ItemVariation.ListAll")
	ac, ok := c.(*AuthContext)
	if !ok {
		return errors.E(op, errors.KindUnexpected, "invalid echo.Context")
	}
	req := ItemVariationListRequest{}
	if err := bindAndValidate(c, &req); err != nil {
		return err
	}

	var limit, offset int64 = ItemVariationListDefaultSize, 0
	if req.Limit <= ItemVariationListMaxSize {
		limit = req.Limit
	}
	if req.Limit > ItemVariationListMaxSize {
		limit = ItemVariationListMaxSize
	}
	if req.Offset != 0 {
		offset = req.Offset
	}

	ctx := c.Request().Context()
	variations, err := h.CatalogService.ListItemVariation(ctx, core.ItemVariationFilter{
		Limit:      limit,
		Offset:     offset,
		MerchantID: ac.MerchantID,
	})
	if err != nil {
		return errors.E(op, err)
	}
	resp := ItemVariationListResponse{ItemVariations: make([]ItemVariation, len(variations))}
	for i, cus := range variations {
		resp.ItemVariations[i] = NewItemVariation(cus)
	}
	return c.JSON(http.StatusOK, resp)
}

func (h *Handler) DeleteItemVariation(c echo.Context) error {
	const op = errors.Op("handler.ItemVariation.Delete")
	ac, ok := c.(*AuthContext)
	if !ok {
		return errors.E(op, errors.KindUnexpected, "invalid echo.Context")
	}
	ctx := c.Request().Context()
	variation, err := h.CatalogService.DeleteItemVariation(ctx, c.Param("id"), ac.MerchantID, nil)
	if err != nil {
		return errors.E(op, err)
	}
	return c.JSON(http.StatusOK, NewItemVariation(variation))
}

type ItemVariation struct {
	ID          string      `json:"id"`
	Name        string      `json:"name"`
	SKU         string      `json:"sku,omitempty"`
	ItemID      string      `json:"item_id"`
	Price       Money       `json:"price"`
	Image       string      `json:"image,omitempty"`
	LocationIDs []string    `json:"location_ids"`
	MerchantID  string      `json:"merchant_id"`
	CreatedAt   int64       `json:"created_at"`
	UpdatedAt   int64       `json:"updated_at"`
	Status      core.Status `json:"status"`
}

func NewItemVariation(itvar core.ItemVariation) ItemVariation {
	return ItemVariation{
		ID:     itvar.ID,
		Name:   itvar.Name,
		SKU:    itvar.SKU,
		ItemID: itvar.ItemID,
		Price: Money{
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

type ItemVariationCreateRequest struct {
	Name        string    `json:"name" validate:"required"`
	SKU         string    `json:"sku"`
	ItemID      string    `json:"item_id" validate:"required"`
	Price       *Money    `json:"price" validate:"required"`
	Image       string    `json:"image"`
	LocationIDs *[]string `json:"location_ids" validate:"omitempty,dive,required"`
}

type ItemVariationUpdateRequest struct {
	ID          string    `param:"id" validate:"required"`
	Name        *string   `json:"name" validate:"omitempty,min=1"`
	SKU         *string   `json:"sku"`
	Price       *Money    `json:"price"`
	Image       *string   `json:"image"`
	LocationIDs *[]string `json:"location_ids" validate:"omitempty,dive,required"`
}

type ItemVariationListRequest struct {
	Limit  int64 `query:"limit" validate:"gte=0"`
	Offset int64 `query:"offset" validate:"gte=0"`
}

type ItemVariationListResponse struct {
	ItemVariations []ItemVariation `json:"item_variations"`
}
