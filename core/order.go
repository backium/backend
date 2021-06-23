package core

import (
	"context"
	"errors"
)

var (
	ErrInvalidOrder = errors.New("invalid order")
)

type Order struct {
	ID         string      `bson:"_id"`
	Items      []OrderItem `bson:"items"`
	Taxes      []OrderTax  `bson:"taxes"`
	Total      Money       `bson:"total"`
	LocationID string      `bson:"location_id"`
	MerchantID string      `bson:"merchant_id"`
	CreatedAt  int64       `bson:"created_at"`
	UpdatedAt  int64       `bson:"updated_at"`
	Schema     OrderSchema `bson:"schema"`
}

func NewOrder(locationID, merchantID string) Order {
	return Order{
		ID:         generateID("order"),
		Items:      []OrderItem{},
		LocationID: locationID,
		MerchantID: merchantID,
	}
}

type OrderItem struct {
	UID          string                `bson:"uid"`
	VariationID  string                `bson:"variation_id"`
	Name         string                `bson:"name"`
	Quantity     int64                 `bson:"quantity"`
	Total        Money                 `bson:"total"`
	BasePrice    Money                 `bson:"base_price"`
	AppliedTaxes []OrderItemAppliedTax `bson:"applied_taxes"`
}

type OrderItemAppliedTax struct {
	TaxUID  string `bson:"tax_uid"`
	Applied Money  `bson:"applied_money"`
}

type OrderTax struct {
	UID   string   `bson:"uid"`
	ID    string   `bson:"id"`
	Scope TaxScope `bson:"scope"`
}

type OrderStorage interface {
	Put(context.Context, Order) error
	Get(context.Context, string) (Order, error)
}

// OrderSchema represents a potential order to be created.
type OrderSchema struct {
	Items      []OrderSchemaItem
	Taxes      []OrderSchemaTax
	LocationID string
	MerchantID string
}

func (sch *OrderSchema) itemVariationIDs() []string {
	ids := make([]string, len(sch.Items))
	for i, it := range sch.Items {
		ids[i] = it.VariationID
	}
	return ids
}

func (sch *OrderSchema) taxIDs() []string {
	ids := make([]string, len(sch.Items))
	for i, it := range sch.Taxes {
		ids[i] = it.ID
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
