package core

import (
	"context"
	"errors"
)

var (
	ErrInvalidOrder = errors.New("invalid order")
)

type OrderState string

const (
	OrderStateOpen      OrderState = "open"
	OrderStateCompleted OrderState = "completed"
	OrderStateCanceled  OrderState = "canceled"
)

type Order struct {
	ID                  string          `bson:"_id"`
	Items               []OrderItem     `bson:"items"`
	Taxes               []OrderTax      `bson:"taxes"`
	Discounts           []OrderDiscount `bson:"discounts"`
	TotalDiscountAmount Money           `bson:"total_discount_amount"`
	TotalTaxAmount      Money           `bson:"total_tax_amount"`
	TotalTipAmount      Money           `bson:"total_tip_amount"`
	TotalAmount         Money           `bson:"total_amount"`
	State               OrderState      `bson:"state"`
	LocationID          string          `bson:"location_id"`
	MerchantID          string          `bson:"merchant_id"`
	CreatedAt           int64           `bson:"created_at"`
	UpdatedAt           int64           `bson:"updated_at"`
	Schema              OrderSchema     `bson:"schema"`
}

func NewOrder(locationID, merchantID string) Order {
	return Order{
		ID:         NewID("order"),
		Items:      []OrderItem{},
		State:      OrderStateOpen,
		LocationID: locationID,
		MerchantID: merchantID,
	}
}

type OrderItem struct {
	UID                 string                     `bson:"uid"`
	VariationID         string                     `bson:"variation_id"`
	Name                string                     `bson:"name"`
	Quantity            int64                      `bson:"quantity"`
	GrossSales          Money                      `bson:"gross_sales"`
	TotalDiscountAmount Money                      `bson:"total_discount_amount"`
	TotalTaxAmount      Money                      `bson:"total_tax_amount"`
	TotalAmount         Money                      `bson:"total_amount"`
	BasePrice           Money                      `bson:"base_price"`
	AppliedTaxes        []OrderItemAppliedTax      `bson:"applied_taxes"`
	AppliedDiscounts    []OrderItemAppliedDiscount `bson:"applied_discounts"`
}

type OrderItemAppliedTax struct {
	TaxUID        string `bson:"tax_uid"`
	AppliedAmount Money  `bson:"applied_amount"`
}

type OrderItemAppliedDiscount struct {
	DiscountUID   string `bson:"discount_uid"`
	AppliedAmount Money  `bson:"applied_amount"`
}

type OrderTax struct {
	UID           string   `bson:"uid"`
	ID            string   `bson:"id"`
	Name          string   `bson:"name"`
	Scope         TaxScope `bson:"scope"`
	Percentage    float64  `bson:"percentage"`
	AppliedAmount Money    `bson:"applied_amount"`
}

type OrderDiscount struct {
	UID           string       `bson:"uid"`
	ID            string       `bson:"id"`
	Type          DiscountType `bson:"type"`
	Name          string       `bson:"name"`
	Percentage    float64      `bson:"percentage"`
	Amount        Money        `bson:"amount"`
	AppliedAmount Money        `bson:"applied_amount"`
}

type OrderFilter struct {
	Limit       int64
	Offset      int64
	LocationIDs []string
	MerchantID  string
}

type OrderStorage interface {
	Put(context.Context, Order) error
	Get(context.Context, string, string, []string) (Order, error)
	List(context.Context, OrderFilter) ([]Order, error)
}

// OrderSchema represents a potential order to be created.
type OrderSchema struct {
	Items      []OrderSchemaItem     `bson:"items"`
	Taxes      []OrderSchemaTax      `bson:"taxes"`
	Discounts  []OrderSchemaDiscount `bson:"discounts"`
	LocationID string                `bson:"location_id"`
	MerchantID string                `bson:"merchant_id"`
	Currency   string                `bson:"currency"`
}

func (sch *OrderSchema) itemVariationIDs() []string {
	ids := make([]string, len(sch.Items))
	for i, it := range sch.Items {
		ids[i] = it.VariationID
	}
	return ids
}

func (sch *OrderSchema) taxIDs() []string {
	ids := make([]string, len(sch.Taxes))
	for i, t := range sch.Taxes {
		ids[i] = t.ID
	}
	return ids
}

func (sch *OrderSchema) discountIDs() []string {
	ids := make([]string, len(sch.Discounts))
	for i, d := range sch.Discounts {
		ids[i] = d.ID
	}
	return ids
}

type OrderSchemaItem struct {
	UID         string `bson:"uid"`
	VariationID string `bson:"variation_id"`
	Quantity    int64  `bson:"quantity"`
}

type OrderSchemaTax struct {
	UID   string   `bson:"uid"`
	ID    string   `bson:"id"`
	Scope TaxScope `bson:"scope"`
}

type OrderSchemaDiscount struct {
	UID string `bson:"uid"`
	ID  string `bson:"id"`
}
