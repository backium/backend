package http

import (
	"net/http"

	"github.com/backium/backend/core"
	"github.com/backium/backend/errors"
	"github.com/backium/backend/ptr"
	"github.com/labstack/echo/v4"
)

const (
	OrderListDefaultSize = 10
	OrderListMaxSize     = 50
)

func (h *Handler) SearchOrders(c echo.Context) error {
	const op = errors.Op("http/Handler.SearchOrders")
	ac, ok := c.(*AuthContext)
	if !ok {
		return errors.E(op, errors.KindUnexpected, "invalid echo.Context")
	}
	req := OrderSearchRequest{}
	if err := bindAndValidate(c, &req); err != nil {
		return err
	}

	var limit, offset int64 = OrderListDefaultSize, 0
	if req.Limit <= OrderListMaxSize {
		limit = req.Limit
	}
	if req.Limit > OrderListMaxSize {
		limit = OrderListMaxSize
	}
	if req.Offset != 0 {
		offset = req.Offset
	}

	ctx := c.Request().Context()
	orders, err := h.OrderingService.ListOrder(ctx, core.OrderFilter{
		Limit:       limit,
		Offset:      offset,
		LocationIDs: req.LocationIDs,
		MerchantID:  ac.MerchantID,
	})
	if err != nil {
		return errors.E(op, err)
	}

	resp := OrderSearchResponse{Orders: make([]Order, len(orders))}
	for i, order := range orders {
		resp.Orders[i] = NewOrder(order)
	}

	return c.JSON(http.StatusOK, resp)
}

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
	schema := core.OrderSchema{
		LocationID: req.LocationID,
		MerchantID: ac.MerchantID,
	}
	for _, item := range req.Items {
		schema.Items = append(schema.Items, core.OrderSchemaItem{
			UID:         item.UID,
			VariationID: item.VariationID,
			Quantity:    item.Quantity,
		})
	}
	for _, tax := range req.Taxes {
		schema.Taxes = append(schema.Taxes, core.OrderSchemaTax{
			UID:   tax.UID,
			ID:    tax.ID,
			Scope: tax.Scope,
		})
	}
	for _, discount := range req.Discounts {
		schema.Discounts = append(schema.Discounts, core.OrderSchemaDiscount{
			UID: discount.UID,
			ID:  discount.ID,
		})
	}

	ctx := c.Request().Context()
	order, err := h.OrderingService.CreateOrder(ctx, schema)
	if err != nil {
		return errors.E(op, err)
	}
	return c.JSON(http.StatusOK, NewOrder(order))
}

type Order struct {
	ID            string          `json:"id"`
	Items         []OrderItem     `json:"items"`
	Total         Money           `json:"total"`
	TotalDiscount Money           `json:"total_discount"`
	TotalTax      Money           `json:"total_tax"`
	Taxes         []OrderTax      `json:"taxes"`
	Discounts     []OrderDiscount `json:"discounts"`
	LocationID    string          `json:"location_id"`
	MerchantID    string          `json:"merchant_id"`
	CreatedAt     int64           `json:"created_at"`
	UpdatedAt     int64           `json:"updated_at"`
}

func NewOrder(order core.Order) Order {
	items := make([]OrderItem, len(order.Items))
	for i, orderItem := range order.Items {
		items[i] = NewOrderItem(orderItem)
	}
	taxes := make([]OrderTax, len(order.Taxes))
	for i, orderTax := range order.Taxes {
		taxes[i] = NewOrderTax(orderTax)
	}
	discounts := make([]OrderDiscount, len(order.Discounts))
	for i, orderDiscount := range order.Discounts {
		discounts[i] = NewOrderDiscount(orderDiscount)
	}
	return Order{
		ID:        order.ID,
		Items:     items,
		Taxes:     taxes,
		Discounts: discounts,
		TotalDiscount: Money{
			Amount:   ptr.Int64(order.TotalDiscount.Amount),
			Currency: order.TotalTax.Currency,
		},
		TotalTax: Money{
			Amount:   ptr.Int64(order.TotalTax.Amount),
			Currency: order.TotalTax.Currency,
		},
		Total: Money{
			Amount:   ptr.Int64(order.Total.Amount),
			Currency: order.Total.Currency,
		},
		LocationID: order.LocationID,
		MerchantID: order.MerchantID,
		CreatedAt:  order.CreatedAt,
		UpdatedAt:  order.UpdatedAt,
	}
}

type OrderItem struct {
	UID              string                     `json:"uid"`
	VariationID      string                     `json:"variation_id"`
	Name             string                     `json:"name"`
	Quantity         int64                      `json:"quantity"`
	AppliedTaxes     []OrderItemAppliedTax      `json:"applied_taxes"`
	AppliedDiscounts []OrderItemAppliedDiscount `json:"applied_discounts"`
	BasePrice        Money                      `json:"base_price"`
	GrossSales       Money                      `json:"gross_sales"`
	TotalDiscount    Money                      `json:"total_discount"`
	TotalTax         Money                      `json:"total_tax"`
	Total            Money                      `json:"total"`
}

func NewOrderItem(item core.OrderItem) OrderItem {
	taxes := make([]OrderItemAppliedTax, len(item.AppliedTaxes))
	for i, tax := range item.AppliedTaxes {
		taxes[i] = OrderItemAppliedTax{
			TaxUID: tax.TaxUID,
			Applied: Money{
				Amount:   ptr.Int64(tax.Applied.Amount),
				Currency: tax.Applied.Currency,
			},
		}
	}
	discounts := make([]OrderItemAppliedDiscount, len(item.AppliedDiscounts))
	for i, discount := range item.AppliedDiscounts {
		discounts[i] = OrderItemAppliedDiscount{
			DiscountUID: discount.DiscountUID,
			Applied: Money{
				Amount:   ptr.Int64(discount.Applied.Amount),
				Currency: discount.Applied.Currency,
			},
		}
	}
	return OrderItem{
		UID:         item.UID,
		VariationID: item.VariationID,
		Name:        item.Name,
		Quantity:    item.Quantity,
		BasePrice: Money{
			Amount:   ptr.Int64(item.BasePrice.Amount),
			Currency: item.BasePrice.Currency,
		},
		GrossSales: Money{
			Amount:   ptr.Int64(item.GrossSales.Amount),
			Currency: item.GrossSales.Currency,
		},
		TotalDiscount: Money{
			Amount:   ptr.Int64(item.TotalDiscount.Amount),
			Currency: item.TotalDiscount.Currency,
		},
		TotalTax: Money{
			Amount:   ptr.Int64(item.TotalTax.Amount),
			Currency: item.TotalTax.Currency,
		},
		Total: Money{
			Amount:   ptr.Int64(item.Total.Amount),
			Currency: item.Total.Currency,
		},
		AppliedTaxes:     taxes,
		AppliedDiscounts: discounts,
	}
}

type OrderItemAppliedTax struct {
	TaxUID  string `json:"tax_uid"`
	Applied Money  `json:"applied"`
}

type OrderItemAppliedDiscount struct {
	DiscountUID string `json:"discount_uid"`
	Applied     Money  `json:"applied"`
}

type OrderTax struct {
	UID        string        `json:"uid"`
	ID         string        `json:"id"`
	Scope      core.TaxScope `json:"scope"`
	Name       string        `json:"name"`
	Percentage float64       `json:"percentage"`
	Applied    Money         `json:"applied"`
}

func NewOrderTax(tax core.OrderTax) OrderTax {
	return OrderTax{
		UID:        tax.UID,
		ID:         tax.ID,
		Scope:      tax.Scope,
		Name:       tax.Name,
		Percentage: tax.Percentage,
		Applied: Money{
			Amount:   ptr.Int64(tax.Applied.Amount),
			Currency: tax.Applied.Currency,
		},
	}
}

type OrderDiscount struct {
	UID        string            `json:"uid"`
	ID         string            `json:"id"`
	Name       string            `json:"name"`
	Type       core.DiscountType `json:"type"`
	Fixed      *Money            `json:"fixed,omitempty"`
	Percentage *float64          `json:"percentage,omitempty"`
	Applied    Money             `json:"applied"`
}

func NewOrderDiscount(discount core.OrderDiscount) OrderDiscount {
	orderDiscount := OrderDiscount{
		UID:  discount.UID,
		ID:   discount.ID,
		Name: discount.Name,
		Type: discount.Type,
		Applied: Money{
			Amount:   ptr.Int64(discount.Applied.Amount),
			Currency: discount.Applied.Currency,
		},
	}
	if discount.Type == core.DiscountTypeFixed {
		orderDiscount.Fixed = &Money{
			Amount:   ptr.Int64(discount.Fixed.Amount),
			Currency: discount.Fixed.Currency,
		}
	} else {
		orderDiscount.Percentage = ptr.Float64(discount.Percentage)
	}
	return orderDiscount
}

type OrderItemRequest struct {
	UID         string `json:"uid" validate:"required"`
	VariationID string `json:"variation_id" validate:"required"`
	Quantity    int64  `json:"quantity" validate:"required"`
}

type OrderTaxRequest struct {
	UID   string        `json:"uid" validate:"required"`
	ID    string        `json:"id" validate:"required"`
	Scope core.TaxScope `json:"scope" validate:"required"`
}

type OrderDiscountRequest struct {
	UID string `json:"uid" validate:"required"`
	ID  string `json:"id" validate:"required"`
}

type OrderCreateRequest struct {
	Items      []OrderItemRequest     `json:"items" validate:"required,dive"`
	LocationID string                 `json:"location_id" validate:"required"`
	Taxes      []OrderTaxRequest      `json:"taxes" validate:"omitempty,dive"`
	Discounts  []OrderDiscountRequest `json:"discounts" validate:"omitempty,dive"`
}

type OrderSearchRequest struct {
	LocationIDs []string `json:"location_ids" validate:"omitempty,dive,required"`
	Limit       int64    `json:"limit" validate:"gte=0"`
	Offset      int64    `json:"offset" validate:"gte=0"`
}

type OrderSearchResponse struct {
	Orders []Order `json:"orders"`
}
