package http

import (
	"net/http"

	"github.com/backium/backend/core"
	"github.com/backium/backend/errors"
	"github.com/backium/backend/ptr"
	"github.com/labstack/echo/v4"
)

func (h *Handler) CreateOrder(c echo.Context) error {
	const op = errors.Op("http/Handler.CreateOrder")
	ac, ok := c.(*AuthContext)
	if !ok {
		return errors.E(op, errors.KindUnexpected, "invalid echo.Context")
	}
	req := OrderCreateRequest{}
	if err := bindAndValidate(c, &req); err != nil {
		return err
	}
	proto := core.OrderSchema{
		LocationID: req.LocationID,
		MerchantID: ac.MerchantID,
	}
	for _, it := range req.Items {
		proto.Items = append(proto.Items, core.OrderSchemaItem{
			UID:         it.UID,
			VariationID: it.VariationID,
			Quantity:    it.Quantity,
		})
	}
	for _, ot := range req.Taxes {
		proto.Taxes = append(proto.Taxes, core.OrderSchemaTax{
			UID: ot.UID,
			ID:  ot.ID,
		})
	}

	ctx := c.Request().Context()
	order, err := h.OrderingService.CreateOrder(ctx, proto)
	if err != nil {
		return errors.E(op, err)
	}
	return c.JSON(http.StatusOK, NewOrder(order))
}

type Order struct {
	ID         string      `json:"id"`
	Items      []OrderItem `json:"items"`
	Total      Money       `json:"total"`
	LocationID string      `json:"location_id"`
	MerchantID string      `json:"merchant_id"`
	CreatedAt  int64       `json:"created_at"`
	UpdatedAt  int64       `json:"updated_at"`
}

type OrderItem struct {
	UID          string                `json:"uid"`
	VariationID  string                `json:"variation_id"`
	Name         string                `json:"name"`
	Quantity     int64                 `json:"quantity"`
	AppliedTaxes []OrderItemAppliedTax `json:"applied_taxes"`
	Total        Money                 `json:"total"`
}

func NewOrderItem(it core.OrderItem) OrderItem {
	taxes := make([]OrderItemAppliedTax, len(it.AppliedTaxes))
	for i, t := range it.AppliedTaxes {
		taxes[i] = OrderItemAppliedTax{
			TaxUID: t.TaxUID,
			Applied: Money{
				Amount:   ptr.Int64(t.Applied.Amount),
				Currency: t.Applied.Currency,
			},
		}
	}
	return OrderItem{
		UID:         it.UID,
		VariationID: it.VariationID,
		Name:        it.Name,
		Quantity:    it.Quantity,
		Total: Money{
			Amount:   ptr.Int64(it.Total.Amount),
			Currency: it.Total.Currency,
		},
		AppliedTaxes: taxes,
	}
}

type OrderItemAppliedTax struct {
	TaxUID  string `json:"tax_uid"`
	Applied Money  `json:"applied"`
}

type OrderTax struct {
	UID string `json:"uid"`
	ID  string `json:"id"`
}

func NewOrder(o core.Order) Order {
	items := make([]OrderItem, len(o.Items))
	for i, oi := range o.Items {
		items[i] = NewOrderItem(oi)
	}
	return Order{
		ID:    o.ID,
		Items: items,
		Total: Money{
			Amount:   ptr.Int64(o.Total.Amount),
			Currency: o.Total.Currency,
		},
		LocationID: o.LocationID,
		MerchantID: o.MerchantID,
		CreatedAt:  o.CreatedAt,
		UpdatedAt:  o.UpdatedAt,
	}
}

type OrderItemRequest struct {
	UID         string `json:"uid" validate:"required"`
	VariationID string `json:"variation_id" validate:"required"`
	Quantity    int64  `json:"quantity" validate:"required"`
}

type OrderTaxRequest struct {
	UID string `json:"uid" validate:"required"`
	ID  string `json:"id" validate:"required"`
}

type OrderCreateRequest struct {
	Items      []OrderItemRequest `json:"items" validate:"required,dive"`
	LocationID string             `json:"location_id" validate:"required"`
	Taxes      []OrderTaxRequest  `json:"taxes" validate:"omitempty,dive"`
}
