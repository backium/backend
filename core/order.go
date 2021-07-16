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
	ID                  ID                   `bson:"_id"`
	ItemVariations      []OrderItemVariation `bson:"item_variations"`
	Taxes               []OrderTax           `bson:"taxes"`
	Discounts           []OrderDiscount      `bson:"discounts"`
	TotalDiscountAmount Money                `bson:"total_discount_amount"`
	TotalTaxAmount      Money                `bson:"total_tax_amount"`
	TotalTipAmount      Money                `bson:"total_tip_amount"`
	TotalAmount         Money                `bson:"total_amount"`
	State               OrderState           `bson:"state"`
	EmployeeID          ID                   `bson:"employee_id"`
	CustomerID          ID                   `bson:"customer_id"`
	LocationID          ID                   `bson:"location_id"`
	MerchantID          ID                   `bson:"merchant_id"`
	CreatedAt           int64                `bson:"created_at"`
	UpdatedAt           int64                `bson:"updated_at"`
	Schema              OrderSchema          `bson:"schema"`
}

func NewOrder(locationID, merchantID ID) Order {
	return Order{
		ID:             NewID("order"),
		ItemVariations: []OrderItemVariation{},
		State:          OrderStateOpen,
		LocationID:     locationID,
		MerchantID:     merchantID,
	}
}

type OrderItemVariation struct {
	UID                 string                     `bson:"uid"`
	ID                  ID                         `bson:"variation_id"`
	Name                string                     `bson:"name"`
	Quantity            int64                      `bson:"quantity"`
	GrossSales          Money                      `bson:"gross_sales"`
	TotalDiscountAmount Money                      `bson:"total_discount_amount"`
	TotalTaxAmount      Money                      `bson:"total_tax_amount"`
	TotalAmount         Money                      `bson:"total_amount"`
	BasePrice           Money                      `bson:"base_price"`
	AppliedTaxes        []OrderItemAppliedTax      `bson:"applied_taxes"`
	AppliedDiscounts    []OrderItemAppliedDiscount `bson:"applied_discounts"`

	CategoryName string `bson:"category_name"`
	ItemName     string `bson:"item_name"`
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
	ID            ID       `bson:"id"`
	Name          string   `bson:"name"`
	Scope         TaxScope `bson:"scope"`
	Percentage    float64  `bson:"percentage"`
	AppliedAmount Money    `bson:"applied_amount"`
}

type OrderDiscount struct {
	UID           string       `bson:"uid"`
	ID            ID           `bson:"id"`
	Type          DiscountType `bson:"type"`
	Name          string       `bson:"name"`
	Percentage    float64      `bson:"percentage"`
	Amount        Money        `bson:"amount"`
	AppliedAmount Money        `bson:"applied_amount"`
}

type OrderFilter struct {
	IDs         []ID
	LocationIDs []ID
	MerchantID  ID
	CreatedAt   DateFilter
}

type OrderSort struct {
	CreatedAt SortOrder
}

type OrderQuery struct {
	Limit  int64
	Offset int64
	Filter OrderFilter
	Sort   OrderSort
}

type OrderStorage interface {
	Put(context.Context, Order) error
	Get(context.Context, ID) (Order, error)
	List(context.Context, OrderQuery) ([]Order, int64, error)
}

// OrderSchema represents a potential order to be created.
type OrderSchema struct {
	ItemVariations []OrderSchemaItemVariation `bson:"item_variations"`
	Taxes          []OrderSchemaTax           `bson:"taxes"`
	Discounts      []OrderSchemaDiscount      `bson:"discounts"`
	CustomerID     ID                         `bson:"customer_id"`
	LocationID     ID                         `bson:"location_id"`
	MerchantID     ID                         `bson:"merchant_id"`
	Currency       Currency                   `bson:"currency"`
}

// Validate iterates the schema to validate the uniqueness of the uids
func (sch *OrderSchema) Validate() bool {
	usedUIDs := map[string]struct{}{}
	for _, variation := range sch.ItemVariations {
		uid := variation.UID
		if _, ok := usedUIDs[uid]; ok {
			return false
		}
		usedUIDs[uid] = struct{}{}
	}
	return true
}

func (sch *OrderSchema) itemVariationIDs() []ID {
	ids := make([]ID, len(sch.ItemVariations))
	for i, it := range sch.ItemVariations {
		ids[i] = it.ID
	}
	return ids
}

func (sch *OrderSchema) taxIDs() []ID {
	ids := make([]ID, len(sch.Taxes))
	for i, t := range sch.Taxes {
		ids[i] = t.ID
	}
	return ids
}

func (sch *OrderSchema) discountIDs() []ID {
	ids := make([]ID, len(sch.Discounts))
	for i, d := range sch.Discounts {
		ids[i] = d.ID
	}
	return ids
}

type OrderSchemaItemVariation struct {
	UID      string `bson:"uid"`
	ID       ID     `bson:"variation_id"`
	Quantity int64  `bson:"quantity"`
}

type OrderSchemaTax struct {
	UID   string   `bson:"uid"`
	ID    ID       `bson:"id"`
	Scope TaxScope `bson:"scope"`
}

type OrderSchemaDiscount struct {
	UID string `bson:"uid"`
	ID  ID     `bson:"id"`
}
