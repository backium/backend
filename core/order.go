package core

import (
	"context"
	"errors"
)

var (
	ErrInvalidOrder = errors.New("invalid order")
)

type Order struct {
	ID            string          `bson:"_id"`
	Items         []OrderItem     `bson:"items"`
	Taxes         []OrderTax      `bson:"taxes"`
	Discounts     []OrderDiscount `bson:"discounts"`
	TotalDiscount Money           `bson:"total_discount"`
	TotalTax      Money           `bson:"total_tax"`
	Total         Money           `bson:"total"`
	LocationID    string          `bson:"location_id"`
	MerchantID    string          `bson:"merchant_id"`
	CreatedAt     int64           `bson:"created_at"`
	UpdatedAt     int64           `bson:"updated_at"`
	Schema        OrderSchema     `bson:"schema"`
}

func NewOrder(locationID, merchantID string) Order {
	return Order{
		ID:         NewID("order"),
		Items:      []OrderItem{},
		LocationID: locationID,
		MerchantID: merchantID,
	}
}

type OrderItem struct {
	UID              string                     `bson:"uid"`
	VariationID      string                     `bson:"variation_id"`
	Name             string                     `bson:"name"`
	Quantity         int64                      `bson:"quantity"`
	GrossSales       Money                      `bson:"gross_sales"`
	TotalDiscount    Money                      `bson:"total_discount"`
	TotalTax         Money                      `bson:"total_tax"`
	Total            Money                      `bson:"total"`
	BasePrice        Money                      `bson:"base_price"`
	AppliedTaxes     []OrderItemAppliedTax      `bson:"applied_taxes"`
	AppliedDiscounts []OrderItemAppliedDiscount `bson:"applied_discounts"`
}

type OrderItemAppliedTax struct {
	TaxUID  string `bson:"tax_uid"`
	Applied Money  `bson:"applied_money"`
}

type OrderItemAppliedDiscount struct {
	DiscountUID string `bson:"discount_uid"`
	Applied     Money  `bson:"applied_money"`
}

type OrderTax struct {
	UID        string   `bson:"uid"`
	ID         string   `bson:"id"`
	Name       string   `bson:"name"`
	Scope      TaxScope `bson:"scope"`
	Percentage float64  `bson:"percentage"`
	Applied    Money    `bson:"applied"`
}

type OrderDiscount struct {
	UID        string       `bson:"uid"`
	ID         string       `bson:"id"`
	Type       DiscountType `bson:"type"`
	Name       string       `bson:"name"`
	Percentage float64      `bson:"percentage"`
	Fixed      Money        `bson:"fixed"`
	Applied    Money        `bson:"applied"`
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
	Items      []OrderSchemaItem
	Taxes      []OrderSchemaTax
	Discounts  []OrderSchemaDiscount
	LocationID string
	MerchantID string
	currency   string
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

// rotazo
func (sch *OrderSchema) discountIDs() []string {
	ids := make([]string, len(sch.Discounts))
	for i, d := range sch.Discounts {
		ids[i] = d.ID
	}
	return ids
}

type OrderSchemaItem struct {
	UID         string
	VariationID string
	Quantity    int64
}

type OrderSchemaTax struct {
	UID   string
	ID    string
	Scope TaxScope
}

type OrderSchemaDiscount struct {
	UID string
	ID  string
}
