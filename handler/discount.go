package handler

import (
	"net/http"

	"github.com/backium/backend/base"
	
	"github.com/backium/backend/ptr"
	"github.com/backium/backend/catalog"
	"github.com/backium/backend/errors"
	"github.com/labstack/echo/v4"
)

type Discount struct {
	Controller catalog.Controller
}

func (h *Discount) Create(c echo.Context) error {
	const op = errors.Op("handler.Discount.Create")
	ac, ok := c.(*AuthContext)
	if !ok {
		return errors.E(op, errors.KindUnexpected, "invalid echo.Context")
	}
	req := DiscountCreateRequest{}
	if err := bindAndValidate(c, &req); err != nil {
		return err
	}
	disc := catalog.NewDiscount()
	if req.LocationIDs != nil {
		disc.LocationIDs = *req.LocationIDs
	}

	if req.Amount != nil {
		disc.Amount = base.Money{
			Amount:   ptr.GetInt64(req.Amount.Amount),
			Currency: req.Amount.Currency,
		}
	}

	if req.Percentage != nil {
		disc.Percentage = *req.Percentage
	}
	disc.Type = req.Type

	disc.Name = req.Name
	disc.MerchantID = ac.MerchantID

	disc, err := h.Controller.CreateDiscount(c.Request().Context(), disc)
	if err != nil {
		return errors.E(op, err)
	}
	return c.JSON(http.StatusOK, newDiscountResponse(disc))
}

func (h *Discount) Update(c echo.Context) error {
	const op = errors.Op("handler.Discount.Update")
	req := DiscountUpdateRequest{}
	if err := bindAndValidate(c, &req); err != nil {
		return err
	}
	disc := catalog.DiscountPartial{
		Name:        req.Name,
		Type:  		 req.Type,
		Percentage:  req.Percentage,
		LocationIDs: req.LocationIDs,
	}
	if req.Amount != nil {
		disc.Amount = &base.Money{
			Amount:      ptr.GetInt64(req.Amount.Amount),
			Currency:    req.Amount.Currency,
		}
	}
	ut, err := h.Controller.UpdateDiscount(c.Request().Context(), req.ID, disc)
	if err != nil {
		return errors.E(op, err)
	}
	return c.JSON(http.StatusOK, newDiscountResponse(ut))
}

func (h *Discount) Retrieve(c echo.Context) error {
	const op = errors.Op("handler.Discount.Retrieve")
	ac, ok := c.(*AuthContext)
	if !ok {
		return errors.E(op, errors.KindUnexpected, "invalid echo.Context")
	}
	it, err := h.Controller.RetrieveDiscount(c.Request().Context(), catalog.DiscountRetrieveRequest{
		ID:         c.Param("id"),
		MerchantID: ac.MerchantID,
	})
	if err != nil {
		return errors.E(op, err)
	}
	return c.JSON(http.StatusOK, newDiscountResponse(it))
}

func (h *Discount) ListAll(c echo.Context) error {
	const op = errors.Op("handler.Discount.ListAll")
	ac, ok := c.(*AuthContext)
	if !ok {
		return errors.E(op, errors.KindUnexpected, "invalid echo.Context")
	}
	req := DiscountListRequest{}
	if err := bindAndValidate(c, &req); err != nil {
		return err
	}
	its, err := h.Controller.ListDiscount(c.Request().Context(), catalog.DiscountListRequest{
		Limit:      req.Limit,
		Offset:     req.Offset,
		MerchantID: ac.MerchantID,
	})
	if err != nil {
		return errors.E(op, err)
	}
	res := make([]DiscountResponse, len(its))
	for i, it := range its {
		res[i] = newDiscountResponse(it)
	}
	return c.JSON(http.StatusOK, DiscountListResponse{res})
}

func (h *Discount) Delete(c echo.Context) error {
	const op = errors.Op("handler.Discount.Delete")
	ac, ok := c.(*AuthContext)
	if !ok {
		return errors.E(op, errors.KindUnexpected, "invalid echo.Context")
	}
	it, err := h.Controller.DeleteDiscount(c.Request().Context(), catalog.DiscountDeleteRequest{
		ID:         c.Param("id"),
		MerchantID: ac.MerchantID,
	})
	if err != nil {
		return errors.E(op, err)
	}
	return c.JSON(http.StatusOK, newDiscountResponse(it))
}

type DiscountResponse struct {
	ID          string      			`json:"id"`
	Name        string      			`json:"name"`
	Type 		catalog.DiscountType 	`bson:"discount_type"`
	Amount 		base.Money  			`bson:"amount"`
	Percentage  int         			`json:"percentage"`
	LocationIDs []string    			`json:"location_ids"`
	MerchantID  string      			`json:"merchant_id"`
	Status      base.Status 			`json:"status"`
}

func newDiscountResponse(t catalog.Discount) DiscountResponse {
	return DiscountResponse{
		ID:          t.ID,
		Name:        t.Name,
		Type: 		 t.Type,
		Amount:  	 t.Amount,
		Percentage:  t.Percentage,
		LocationIDs: t.LocationIDs,
		MerchantID:  t.MerchantID,
		Status:      t.Status,
	}
}

type DiscountCreateRequest struct {
	Name        string    				`json:"name" validate:"required"`
	Type 		catalog.DiscountType 	`bson:"type" validate:"required"`
	Amount 		*MoneyResponse 			`bson:"amount"`
	Percentage  *int      				`json:"percentage"`
	LocationIDs *[]string 				`json:"location_ids" validate:"omitempty,dive,required"`
}

type DiscountUpdateRequest struct {
	ID          string    				`param:"id" validate:"required"`
	Name        *string   				`json:"name" validate:"omitempty,min=1"`
	Type 		*catalog.DiscountType 	`bson:"type"`
	Amount 		*MoneyResponse  		`bson:"amount"`
	Percentage  *int      				`json:"percentage" validate:"omitempty,min=0,max=100"`
	LocationIDs *[]string 				`json:"location_ids" validate:"omitempty,dive,required"`
}

type DiscountListRequest struct {
	Limit  *int64 `query:"limit" validate:"omitempty,gte=1"`
	Offset *int64 `query:"offset"`
}

type DiscountListResponse struct {
	Discounts []DiscountResponse `json:"discounts"`
}
