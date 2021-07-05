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

func (h *Handler) HandleSearchOrders(c echo.Context) error {
	const op = errors.Op("http/Handler.SearchOrders")

	type dateFilter struct {
		Gte int64 `json:"gte" validate:"gte=0"`
		Lte int64 `json:"lte" validate:"gte=0"`
	}

	type filter struct {
		IDs         []core.ID  `json:"ids" validate:"omitempty,dive,id"`
		LocationIDs []core.ID  `json:"location_ids" validate:"omitempty,dive,id"`
		CreatedAt   dateFilter `json:"created_at"`
	}

	type sort struct {
		CreatedAt core.SortOrder `json:"created_at"`
	}

	type request struct {
		Limit  int64  `json:"limit" validate:"gte=0"`
		Offset int64  `json:"offset" validate:"gte=0"`
		Filter filter `json:"filter"`
		Sort   sort   `json:"sort"`
	}

	type response struct {
		Orders []Order `json:"orders"`
		Total  int64   `json:"total_count"`
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

	var limit, offset int64 = OrderListDefaultSize, req.Offset
	if req.Limit <= OrderListMaxSize {
		limit = req.Limit
	} else {
		limit = OrderListMaxSize
	}

	orders, count, err := h.OrderingService.ListOrder(ctx, core.OrderQuery{
		Limit:  limit,
		Offset: offset,
		Filter: core.OrderFilter{
			LocationIDs: req.Filter.LocationIDs,
			MerchantID:  merchant.ID,
			CreatedAt: core.DateFilter{
				Gte: req.Filter.CreatedAt.Gte,
				Lte: req.Filter.CreatedAt.Lte,
			},
		},
		Sort: core.OrderSort{
			CreatedAt: req.Sort.CreatedAt,
		},
	})
	if err != nil {
		return errors.E(op, err)
	}

	resp := response{
		Orders: make([]Order, len(orders)),
		Total:  count,
	}
	for i, order := range orders {
		resp.Orders[i] = NewOrder(order)
	}

	return c.JSON(http.StatusOK, resp)
}

func (h *Handler) HandleCreateOrder(c echo.Context) error {
	const op = errors.Op("http/Handler.CreateOrder")

	type item struct {
		UID         string  `json:"uid" validate:"required"`
		VariationID core.ID `json:"variation_id" validate:"required"`
		Quantity    int64   `json:"quantity" validate:"required"`
	}

	type tax struct {
		UID   string        `json:"uid" validate:"required"`
		ID    core.ID       `json:"id" validate:"required"`
		Scope core.TaxScope `json:"scope" validate:"required"`
	}

	type discount struct {
		UID string  `json:"uid" validate:"required"`
		ID  core.ID `json:"id" validate:"required"`
	}

	type request struct {
		Items      []item     `json:"items" validate:"required,dive"`
		LocationID core.ID    `json:"location_id" validate:"required"`
		Taxes      []tax      `json:"taxes" validate:"omitempty,dive"`
		Discounts  []discount `json:"discounts" validate:"omitempty,dive"`
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

	schema := core.OrderSchema{
		LocationID: req.LocationID,
		MerchantID: merchant.ID,
	}
	for _, item := range req.Items {
		schema.ItemVariations = append(schema.ItemVariations, core.OrderSchemaItemVariation{
			UID:      item.UID,
			ID:       item.VariationID,
			Quantity: item.Quantity,
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

	order, err := h.OrderingService.CreateOrder(ctx, schema)
	if err != nil {
		return errors.E(op, err)
	}

	return c.JSON(http.StatusOK, NewOrder(order))
}

func (h *Handler) HandlePayOrder(c echo.Context) error {
	const op = errors.Op("http/Handler.PayOrder")

	type request struct {
		PaymentIDs []core.ID `json:"payment_ids" validate:"omitempty,dive,required"`
		OrderID    core.ID   `param:"order_id" validate:"required"`
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

	order, err := h.OrderingService.PayOrder(ctx, req.OrderID, req.PaymentIDs)
	if err != nil {
		return errors.E(op, err)
	}

	return c.JSON(http.StatusOK, NewOrder(order))
}

type Order struct {
	ID                  core.ID         `json:"id"`
	Items               []OrderItem     `json:"items"`
	TotalAmount         MoneyRequest    `json:"total_amount"`
	TotalDiscountAmount MoneyRequest    `json:"total_discount_amount"`
	TotalTaxAmount      MoneyRequest    `json:"total_tax_amount"`
	Taxes               []OrderTax      `json:"taxes"`
	Discounts           []OrderDiscount `json:"discounts"`
	State               core.OrderState `json:"state"`
	LocationID          core.ID         `json:"location_id"`
	MerchantID          core.ID         `json:"merchant_id"`
	CreatedAt           int64           `json:"created_at"`
	UpdatedAt           int64           `json:"updated_at"`
}

func NewOrder(order core.Order) Order {
	items := make([]OrderItem, len(order.ItemVariations))
	for i, orderItem := range order.ItemVariations {
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
		State:     order.State,
		TotalDiscountAmount: MoneyRequest{
			Value:    ptr.Int64(order.TotalDiscountAmount.Value),
			Currency: order.TotalTaxAmount.Currency,
		},
		TotalTaxAmount: MoneyRequest{
			Value:    ptr.Int64(order.TotalTaxAmount.Value),
			Currency: order.TotalTaxAmount.Currency,
		},
		TotalAmount: MoneyRequest{
			Value:    ptr.Int64(order.TotalAmount.Value),
			Currency: order.TotalAmount.Currency,
		},
		LocationID: order.LocationID,
		MerchantID: order.MerchantID,
		CreatedAt:  order.CreatedAt,
		UpdatedAt:  order.UpdatedAt,
	}
}

type OrderItem struct {
	UID                 string                     `json:"uid"`
	VariationID         core.ID                    `json:"variation_id"`
	Name                string                     `json:"name"`
	Quantity            int64                      `json:"quantity"`
	AppliedTaxes        []OrderItemAppliedTax      `json:"applied_taxes"`
	AppliedDiscounts    []OrderItemAppliedDiscount `json:"applied_discounts"`
	BasePrice           MoneyRequest               `json:"base_price"`
	GrossSales          MoneyRequest               `json:"gross_sales"`
	TotalDiscountAmount MoneyRequest               `json:"total_discount_amount"`
	TotalTaxAmount      MoneyRequest               `json:"total_tax_amount"`
	TotalAmount         MoneyRequest               `json:"total_amount"`
}

func NewOrderItem(item core.OrderItemVariation) OrderItem {
	taxes := make([]OrderItemAppliedTax, len(item.AppliedTaxes))
	for i, tax := range item.AppliedTaxes {
		taxes[i] = OrderItemAppliedTax{
			TaxUID: tax.TaxUID,
			AppliedAmount: MoneyRequest{
				Value:    ptr.Int64(tax.AppliedAmount.Value),
				Currency: tax.AppliedAmount.Currency,
			},
		}
	}
	discounts := make([]OrderItemAppliedDiscount, len(item.AppliedDiscounts))
	for i, discount := range item.AppliedDiscounts {
		discounts[i] = OrderItemAppliedDiscount{
			DiscountUID: discount.DiscountUID,
			AppliedAmount: MoneyRequest{
				Value:    ptr.Int64(discount.AppliedAmount.Value),
				Currency: discount.AppliedAmount.Currency,
			},
		}
	}
	return OrderItem{
		UID:         item.UID,
		VariationID: item.ID,
		Name:        item.Name,
		Quantity:    item.Quantity,
		BasePrice: MoneyRequest{
			Value:    ptr.Int64(item.BasePrice.Value),
			Currency: item.BasePrice.Currency,
		},
		GrossSales: MoneyRequest{
			Value:    ptr.Int64(item.GrossSales.Value),
			Currency: item.GrossSales.Currency,
		},
		TotalDiscountAmount: MoneyRequest{
			Value:    ptr.Int64(item.TotalDiscountAmount.Value),
			Currency: item.TotalDiscountAmount.Currency,
		},
		TotalTaxAmount: MoneyRequest{
			Value:    ptr.Int64(item.TotalTaxAmount.Value),
			Currency: item.TotalTaxAmount.Currency,
		},
		TotalAmount: MoneyRequest{
			Value:    ptr.Int64(item.TotalAmount.Value),
			Currency: item.TotalAmount.Currency,
		},
		AppliedTaxes:     taxes,
		AppliedDiscounts: discounts,
	}
}

type OrderItemAppliedTax struct {
	TaxUID        string       `json:"tax_uid"`
	AppliedAmount MoneyRequest `json:"applied_amount"`
}

type OrderItemAppliedDiscount struct {
	DiscountUID   string       `json:"discount_uid"`
	AppliedAmount MoneyRequest `json:"applied_amount"`
}

type OrderTax struct {
	UID           string        `json:"uid"`
	ID            core.ID       `json:"id"`
	Scope         core.TaxScope `json:"scope"`
	Name          string        `json:"name"`
	Percentage    float64       `json:"percentage"`
	AppliedAmount MoneyRequest  `json:"applied_amount"`
}

func NewOrderTax(tax core.OrderTax) OrderTax {
	return OrderTax{
		UID:        tax.UID,
		ID:         tax.ID,
		Scope:      tax.Scope,
		Name:       tax.Name,
		Percentage: tax.Percentage,
		AppliedAmount: MoneyRequest{
			Value:    ptr.Int64(tax.AppliedAmount.Value),
			Currency: tax.AppliedAmount.Currency,
		},
	}
}

type OrderDiscount struct {
	UID           string            `json:"uid"`
	ID            core.ID           `json:"id"`
	Name          string            `json:"name"`
	Type          core.DiscountType `json:"type"`
	Amount        *MoneyRequest     `json:"amount,omitempty"`
	Percentage    *float64          `json:"percentage,omitempty"`
	AppliedAmount MoneyRequest      `json:"applied_amount"`
}

func NewOrderDiscount(discount core.OrderDiscount) OrderDiscount {
	orderDiscount := OrderDiscount{
		UID:  discount.UID,
		ID:   discount.ID,
		Name: discount.Name,
		Type: discount.Type,
		AppliedAmount: MoneyRequest{
			Value:    ptr.Int64(discount.AppliedAmount.Value),
			Currency: discount.AppliedAmount.Currency,
		},
	}
	if discount.Type == core.DiscountFixed {
		orderDiscount.Amount = &MoneyRequest{
			Value:    ptr.Int64(discount.Amount.Value),
			Currency: discount.Amount.Currency,
		}
	} else {
		orderDiscount.Percentage = ptr.Float64(discount.Percentage)
	}
	return orderDiscount
}
