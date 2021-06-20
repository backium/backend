package handler

import (
	"net/http"

	"github.com/backium/backend/base"
	"github.com/backium/backend/catalog"
	"github.com/backium/backend/errors"
	"github.com/backium/backend/ptr"
	"github.com/labstack/echo/v4"
)

type ItemVariation struct {
	Controller catalog.Controller
}

func (h *ItemVariation) Create(c echo.Context) error {
	const op = errors.Op("handler.ItemVariation.Create")
	ac, ok := c.(*AuthContext)
	if !ok {
		return errors.E(op, errors.KindUnexpected, "invalid echo.Context")
	}
	req := ItemVariationCreateRequest{}
	if err := bindAndValidate(c, &req); err != nil {
		return err
	}

	itvar := catalog.NewItemVariation()
	// override defaults
	if req.LocationIDs != nil {
		itvar.LocationIDs = *req.LocationIDs
	}
	itvar.Name = req.Name
	itvar.SKU = req.SKU
	itvar.ItemID = req.ItemID
	itvar.Price = base.Money{
		Amount:   ptr.GetInt64(req.Price.Amount),
		Currency: req.Price.Currency,
	}
	itvar.MerchantID = ac.MerchantID

	itvar, err := h.Controller.CreateItemVariation(c.Request().Context(), itvar)
	if err != nil {
		return errors.E(op, err)
	}
	return c.JSON(http.StatusOK, newItemVariationResponse(itvar))
}

func (h *ItemVariation) Update(c echo.Context) error {
	const op = errors.Op("handler.ItemVariation.Update")
	req := ItemVariationUpdateRequest{}
	if err := bindAndValidate(c, &req); err != nil {
		return err
	}

	itvar := catalog.PartialItemVariation{
		Name:        req.Name,
		SKU:         req.SKU,
		LocationIDs: req.LocationIDs,
	}
	if req.Price != nil {
		itvar.Price = &base.Money{
			Amount:   ptr.GetInt64(req.Price.Amount),
			Currency: req.Price.Currency,
		}
	}
	uitvar, err := h.Controller.UpdateItemVariation(c.Request().Context(), req.ID, itvar)
	if err != nil {
		return errors.E(op, err)
	}
	return c.JSON(http.StatusOK, newItemVariationResponse(uitvar))
}

func (h *ItemVariation) Retrieve(c echo.Context) error {
	const op = errors.Op("handler.ItemVariation.Retrieve")
	ac, ok := c.(*AuthContext)
	if !ok {
		return errors.E(op, errors.KindUnexpected, "invalid echo.Context")
	}
	m, err := h.Controller.RetrieveItemVariation(c.Request().Context(), catalog.ItemVariationRetrieveRequest{
		ID:         c.Param("id"),
		MerchantID: ac.MerchantID,
	})
	if err != nil {
		return errors.E(op, err)
	}
	return c.JSON(http.StatusOK, newItemVariationResponse(m))
}

func (h *ItemVariation) ListAll(c echo.Context) error {
	const op = errors.Op("handler.ItemVariation.ListAll")
	ac, ok := c.(*AuthContext)
	if !ok {
		return errors.E(op, errors.KindUnexpected, "invalid echo.Context")
	}
	req := ItemVariationListRequest{}
	if err := bindAndValidate(c, &req); err != nil {
		return err
	}
	cuss, err := h.Controller.ListItemVariation(c.Request().Context(), catalog.ItemVariationListRequest{
		Limit:      req.Limit,
		Offset:     req.Offset,
		MerchantID: ac.MerchantID,
	})
	if err != nil {
		return errors.E(op, err)
	}
	res := make([]ItemVariationResponse, len(cuss))
	for i, cus := range cuss {
		res[i] = newItemVariationResponse(cus)
	}
	return c.JSON(http.StatusOK, ItemVariationListResponse{res})
}

func (h *ItemVariation) Delete(c echo.Context) error {
	const op = errors.Op("handler.ItemVariation.Delete")
	ac, ok := c.(*AuthContext)
	if !ok {
		return errors.E(op, errors.KindUnexpected, "invalid echo.Context")
	}
	cus, err := h.Controller.DeleteItemVariation(c.Request().Context(), catalog.ItemVariationDeleteRequest{
		ID:         c.Param("id"),
		MerchantID: ac.MerchantID,
	})
	if err != nil {
		return errors.E(op, err)
	}
	return c.JSON(http.StatusOK, newItemVariationResponse(cus))
}

type ItemVariationResponse struct {
	ID          string        `json:"id"`
	Name        string        `json:"name"`
	SKU         string        `json:"sku,omitempty"`
	ItemID      string        `json:"item_id"`
	Price       MoneyResponse `json:"price"`
	LocationIDs []string      `json:"location_ids"`
	MerchantID  string        `json:"merchant_id"`
	Status      base.Status   `json:"status"`
}

func newItemVariationResponse(itvar catalog.ItemVariation) ItemVariationResponse {
	return ItemVariationResponse{
		ID:     itvar.ID,
		Name:   itvar.Name,
		SKU:    itvar.SKU,
		ItemID: itvar.ItemID,
		Price: MoneyResponse{
			Amount:   &itvar.Price.Amount,
			Currency: itvar.Price.Currency,
		},
		LocationIDs: itvar.LocationIDs,
		MerchantID:  itvar.MerchantID,
		Status:      itvar.Status,
	}
}

type ItemVariationCreateRequest struct {
	Name        string         `json:"name" validate:"required"`
	SKU         string         `json:"sku"`
	ItemID      string         `json:"item_id" validate:"required"`
	Price       *MoneyResponse `json:"price" validate:"required"`
	LocationIDs *[]string      `json:"location_ids" validate:"omitempty,dive,required"`
}

type ItemVariationUpdateRequest struct {
	ID          string         `param:"id" validate:"required"`
	Name        *string        `json:"name" validate:"omitempty,min=1"`
	SKU         *string        `json:"sku"`
	Price       *MoneyResponse `json:"price"`
	LocationIDs *[]string      `json:"location_ids" validate:"omitempty,dive,required"`
}

type ItemVariationListRequest struct {
	Limit  *int64 `query:"limit" validate:"omitempty,gte=1"`
	Offset *int64 `query:"offset"`
}

type ItemVariationListResponse struct {
	ItemVariations []ItemVariationResponse `json:"itemVariations"`
}
