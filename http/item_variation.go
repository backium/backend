package http

import (
	"net/http"

	"github.com/backium/backend/core"
	"github.com/backium/backend/errors"
	"github.com/backium/backend/ptr"
	"github.com/labstack/echo/v4"
)

type Money struct {
	Amount   *int64 `json:"amount" validate:"required"`
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

	itvar := core.NewItemVariation()
	// override defaults
	if req.LocationIDs != nil {
		itvar.LocationIDs = *req.LocationIDs
	}
	itvar.Name = req.Name
	itvar.SKU = req.SKU
	itvar.ItemID = req.ItemID
	itvar.Price = core.Money{
		Amount:   ptr.GetInt64(req.Price.Amount),
		Currency: req.Price.Currency,
	}
	itvar.MerchantID = ac.MerchantID

	itvar, err := h.CatalogService.CreateItemVariation(c.Request().Context(), itvar)
	if err != nil {
		return errors.E(op, err)
	}
	return c.JSON(http.StatusOK, NewItemVariation(itvar))
}

func (h *Handler) UpdateItemVariation(c echo.Context) error {
	const op = errors.Op("handler.ItemVariation.Update")
	req := ItemVariationUpdateRequest{}
	if err := bindAndValidate(c, &req); err != nil {
		return err
	}

	itvar := core.ItemVariationPartial{
		Name:        req.Name,
		SKU:         req.SKU,
		LocationIDs: req.LocationIDs,
	}
	if req.Price != nil {
		itvar.Price = &core.Money{
			Amount:   ptr.GetInt64(req.Price.Amount),
			Currency: req.Price.Currency,
		}
	}
	uitvar, err := h.CatalogService.UpdateItemVariation(c.Request().Context(), req.ID, itvar)
	if err != nil {
		return errors.E(op, err)
	}
	return c.JSON(http.StatusOK, NewItemVariation(uitvar))
}

func (h *Handler) RetrieveItemVariation(c echo.Context) error {
	const op = errors.Op("handler.ItemVariation.Retrieve")
	ac, ok := c.(*AuthContext)
	if !ok {
		return errors.E(op, errors.KindUnexpected, "invalid echo.Context")
	}
	m, err := h.CatalogService.RetrieveItemVariation(c.Request().Context(), core.ItemVariationRetrieveRequest{
		ID:         c.Param("id"),
		MerchantID: ac.MerchantID,
	})
	if err != nil {
		return errors.E(op, err)
	}
	return c.JSON(http.StatusOK, NewItemVariation(m))
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
	cuss, err := h.CatalogService.ListItemVariation(c.Request().Context(), core.ItemVariationListRequest{
		Limit:      req.Limit,
		Offset:     req.Offset,
		MerchantID: ac.MerchantID,
	})
	if err != nil {
		return errors.E(op, err)
	}
	res := make([]ItemVariation, len(cuss))
	for i, cus := range cuss {
		res[i] = NewItemVariation(cus)
	}
	return c.JSON(http.StatusOK, ItemVariationListResponse{res})
}

func (h *Handler) DeleteItemVariation(c echo.Context) error {
	const op = errors.Op("handler.ItemVariation.Delete")
	ac, ok := c.(*AuthContext)
	if !ok {
		return errors.E(op, errors.KindUnexpected, "invalid echo.Context")
	}
	cus, err := h.CatalogService.DeleteItemVariation(c.Request().Context(), core.ItemVariationDeleteRequest{
		ID:         c.Param("id"),
		MerchantID: ac.MerchantID,
	})
	if err != nil {
		return errors.E(op, err)
	}
	return c.JSON(http.StatusOK, NewItemVariation(cus))
}

type ItemVariation struct {
	ID          string      `json:"id"`
	Name        string      `json:"name"`
	SKU         string      `json:"sku,omitempty"`
	ItemID      string      `json:"item_id"`
	Price       Money       `json:"price"`
	LocationIDs []string    `json:"location_ids"`
	MerchantID  string      `json:"merchant_id"`
	Status      core.Status `json:"status"`
}

func NewItemVariation(itvar core.ItemVariation) ItemVariation {
	return ItemVariation{
		ID:     itvar.ID,
		Name:   itvar.Name,
		SKU:    itvar.SKU,
		ItemID: itvar.ItemID,
		Price: Money{
			Amount:   &itvar.Price.Amount,
			Currency: itvar.Price.Currency,
		},
		LocationIDs: itvar.LocationIDs,
		MerchantID:  itvar.MerchantID,
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
	LocationIDs *[]string `json:"location_ids" validate:"omitempty,dive,required"`
}

type ItemVariationUpdateRequest struct {
	ID          string    `param:"id" validate:"required"`
	Name        *string   `json:"name" validate:"omitempty,min=1"`
	SKU         *string   `json:"sku"`
	Price       *Money    `json:"price"`
	LocationIDs *[]string `json:"location_ids" validate:"omitempty,dive,required"`
}

type ItemVariationListRequest struct {
	Limit  *int64 `query:"limit" validate:"omitempty,gte=1"`
	Offset *int64 `query:"offset"`
}

type ItemVariationListResponse struct {
	ItemVariations []ItemVariation `json:"itemVariations"`
}
