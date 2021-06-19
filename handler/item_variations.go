package handler

import (
	"net/http"

	"github.com/backium/backend/controller"
	"github.com/backium/backend/entity"
	"github.com/backium/backend/errors"
	"github.com/backium/backend/ptr"
	"github.com/labstack/echo/v4"
)

type money struct {
	Amount   *int64 `json:"amount" validate:"required"`
	Currency string `json:"currency" validate:"required"`
}

type itemVariation struct {
	ID          string        `json:"id"`
	Name        string        `json:"name"`
	SKU         string        `json:"sku,omitempty"`
	ItemID      string        `json:"item_id"`
	Price       money         `json:"price"`
	LocationIDs []string      `json:"location_ids"`
	MerchantID  string        `json:"merchant_id"`
	Status      entity.Status `json:"status"`
}

func newItemVariation(itvar entity.ItemVariation) itemVariation {
	return itemVariation{
		ID:     itvar.ID,
		Name:   itvar.Name,
		SKU:    itvar.SKU,
		ItemID: itvar.ItemID,
		Price: money{
			Amount:   &itvar.Price.Amount,
			Currency: itvar.Price.Currency,
		},
		LocationIDs: itvar.LocationIDs,
		MerchantID:  itvar.MerchantID,
		Status:      itvar.Status,
	}
}

type createItemVariationRequest struct {
	Name        string    `json:"name" validate:"required"`
	SKU         string    `json:"sku"`
	ItemID      string    `json:"item_id" validate:"required"`
	Price       *money    `json:"price" validate:"required"`
	LocationIDs *[]string `json:"location_ids" validate:"omitempty,dive,required"`
}

type updateItemVariationRequest struct {
	ID          string    `param:"id" validate:"required"`
	Name        *string   `json:"name" validate:"omitempty,min=1"`
	SKU         *string   `json:"sku"`
	Price       *money    `json:"price"`
	LocationIDs *[]string `json:"location_ids" validate:"omitempty,dive,required"`
}

type listAllItemVariationsRequest struct {
	Limit  *int64 `query:"limit" validate:"omitempty,gte=1"`
	Offset *int64 `query:"offset"`
}

type listItemVariationsResponse struct {
	ItemVariations []itemVariation `json:"itemVariations"`
}

type ItemVariation struct {
	Controller controller.ItemVariation
}

func (h *ItemVariation) Create(c echo.Context) error {
	const op = errors.Op("handler.ItemVariation.Create")
	ac, ok := c.(*AuthContext)
	if !ok {
		return errors.E(op, errors.KindUnexpected, "invalid echo.Context")
	}
	req := createItemVariationRequest{}
	if err := bindAndValidate(c, &req); err != nil {
		return err
	}

	itvar := entity.NewItemVariation()
	// override defaults
	if req.LocationIDs != nil {
		itvar.LocationIDs = *req.LocationIDs
	}
	itvar.Name = req.Name
	itvar.SKU = req.SKU
	itvar.ItemID = req.ItemID
	itvar.Price = entity.Money{
		Amount:   ptr.GetInt64(req.Price.Amount),
		Currency: req.Price.Currency,
	}
	itvar.MerchantID = ac.MerchantID

	itvar, err := h.Controller.Create(c.Request().Context(), itvar)
	if err != nil {
		return errors.E(op, err)
	}
	return c.JSON(http.StatusOK, newItemVariation(itvar))
}

func (h *ItemVariation) Update(c echo.Context) error {
	const op = errors.Op("handler.ItemVariation.Update")
	req := updateItemVariationRequest{}
	if err := bindAndValidate(c, &req); err != nil {
		return err
	}

	itvar := controller.PartialItemVariation{
		Name:        req.Name,
		SKU:         req.SKU,
		LocationIDs: req.LocationIDs,
	}
	if req.Price != nil {
		itvar.Price = &entity.Money{
			Amount:   ptr.GetInt64(req.Price.Amount),
			Currency: req.Price.Currency,
		}
	}
	uitvar, err := h.Controller.Update(c.Request().Context(), req.ID, itvar)
	if err != nil {
		return errors.E(op, err)
	}
	return c.JSON(http.StatusOK, newItemVariation(uitvar))
}

func (h *ItemVariation) Retrieve(c echo.Context) error {
	const op = errors.Op("handler.ItemVariation.Retrieve")
	ac, ok := c.(*AuthContext)
	if !ok {
		return errors.E(op, errors.KindUnexpected, "invalid echo.Context")
	}
	m, err := h.Controller.Retrieve(c.Request().Context(), controller.RetrieveItemVariationRequest{
		ID:         c.Param("id"),
		MerchantID: ac.MerchantID,
	})
	if err != nil {
		return errors.E(op, err)
	}
	return c.JSON(http.StatusOK, newItemVariation(m))
}

func (h *ItemVariation) ListAll(c echo.Context) error {
	const op = errors.Op("handler.ItemVariation.ListAll")
	ac, ok := c.(*AuthContext)
	if !ok {
		return errors.E(op, errors.KindUnexpected, "invalid echo.Context")
	}
	req := listAllItemVariationsRequest{}
	if err := bindAndValidate(c, &req); err != nil {
		return err
	}
	cuss, err := h.Controller.ListAll(c.Request().Context(), controller.ListAllItemVariationsRequest{
		Limit:      req.Limit,
		Offset:     req.Offset,
		MerchantID: ac.MerchantID,
	})
	if err != nil {
		return errors.E(op, err)
	}
	res := make([]itemVariation, len(cuss))
	for i, cus := range cuss {
		res[i] = newItemVariation(cus)
	}
	return c.JSON(http.StatusOK, listItemVariationsResponse{res})
}

func (h *ItemVariation) Delete(c echo.Context) error {
	const op = errors.Op("handler.ItemVariation.Delete")
	ac, ok := c.(*AuthContext)
	if !ok {
		return errors.E(op, errors.KindUnexpected, "invalid echo.Context")
	}
	cus, err := h.Controller.Delete(c.Request().Context(), controller.DeleteItemVariationRequest{
		ID:         c.Param("id"),
		MerchantID: ac.MerchantID,
	})
	if err != nil {
		return errors.E(op, err)
	}
	return c.JSON(http.StatusOK, newItemVariation(cus))
}
