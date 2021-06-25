package core

import (
	"context"
	"fmt"

	"github.com/backium/backend/errors"
	d "github.com/shopspring/decimal"
)

var (
	hundred = d.NewFromInt(100)
)

type OrderingService struct {
	OrderStorage         OrderStorage
	ItemVariationStorage ItemVariationStorage
	TaxStorage           TaxStorage
	DiscountStorage      DiscountStorage
}

func (s *OrderingService) CreateOrder(ctx context.Context, sch OrderSchema) (Order, error) {
	const op = errors.Op("core/OrderingService.CreateOrder")
	sch.currency = "PEN"
	order, err := s.build(ctx, sch)
	if err != nil {
		return Order{}, errors.E(op, err)
	}

	if err := s.OrderStorage.Put(ctx, *order); err != nil {
		return Order{}, errors.E(op, err)
	}
	norder, err := s.OrderStorage.Get(ctx, order.ID, order.MerchantID, nil)
	if err != nil {
		return Order{}, errors.E(op, err)
	}
	return norder, nil
}

// OrderLookup provides easy access to order elements (items, taxes, discounts) using schema uids
type OrderLookup struct {
	item     map[string]ItemVariation
	tax      map[string]Tax
	discount map[string]Discount
}

func NewOrderLookup(sch OrderSchema, items []ItemVariation, taxes []Tax, discounts []Discount) (*OrderLookup, error) {
	// Save items by UID for easy access
	itemLookup := map[string]ItemVariation{}
	for _, oit := range sch.Items {
		for _, it := range items {
			if it.ID == oit.VariationID {
				itemLookup[oit.UID] = it
			}
		}
		if _, ok := itemLookup[oit.UID]; !ok {
			return nil, errors.E(fmt.Sprintf("Item variation '%v' doesn't exist or is not available.", oit.UID))
		}
	}

	taxLookup := map[string]Tax{}
	for _, ot := range sch.Taxes {
		for _, t := range taxes {
			if t.ID == ot.ID {
				taxLookup[ot.UID] = t
			}
		}
		if _, ok := taxLookup[ot.UID]; !ok {
			return nil, errors.E(fmt.Sprintf("Tax '%v' doesn't exist or is not available.", ot.UID))
		}
	}

	discountLookup := map[string]Discount{}
	for _, ot := range sch.Discounts {
		for _, d := range discounts {
			if d.ID == ot.ID {
				discountLookup[ot.UID] = d
			}
		}
		if _, ok := discountLookup[ot.UID]; !ok {
			return nil, errors.E("Unknow discounts in order")
		}
	}

	return &OrderLookup{
		item:     itemLookup,
		tax:      taxLookup,
		discount: discountLookup,
	}, nil
}

func (l *OrderLookup) Tax(uid string) Tax {
	return l.tax[uid]
}

func (l *OrderLookup) Item(uid string) ItemVariation {
	return l.item[uid]
}

func (l *OrderLookup) Discount(uid string) Discount {
	return l.discount[uid]
}

// OrderBuilder helps to build an order from a schema
type OrderBuilder struct {
	lookup *OrderLookup
	schema OrderSchema
}

func NewOrderBuilder(sch OrderSchema, lookup *OrderLookup) *OrderBuilder {
	return &OrderBuilder{
		lookup: lookup,
		schema: sch,
	}
}

// applyItemsAndInit will populate order items from the schema and calculate initial totals
func (b *OrderBuilder) applyItemsAndInit(order *Order) {
	var itemsTotalAmount int64
	currency := b.schema.currency
	for _, schItem := range b.schema.Items {
		item := b.lookup.Item(schItem.UID)
		orderItem := OrderItem{
			UID:           schItem.UID,
			VariationID:   schItem.VariationID,
			Name:          item.Name,
			Quantity:      schItem.Quantity,
			BasePrice:     NewMoney(item.Price.Amount, currency),
			GrossSales:    NewMoney(item.Price.Amount*schItem.Quantity, currency),
			TotalDiscount: NewMoney(0, currency),
			TotalTax:      NewMoney(0, currency),
			Total:         NewMoney(item.Price.Amount*schItem.Quantity, currency),
		}
		order.Items = append(order.Items, orderItem)
		itemsTotalAmount += orderItem.Total.Amount
	}
	order.TotalDiscount = NewMoney(0, currency)
	order.TotalTax = NewMoney(0, currency)
	order.Total = NewMoney(itemsTotalAmount, currency)
}

func (b *OrderBuilder) applyOrderLevelFixedDiscounts(order *Order) {
	var schemaDiscounts []OrderSchemaDiscount
	currency := b.schema.currency
	for _, d := range b.schema.Discounts {
		if b.lookup.Discount(d.UID).Type == DiscountTypeFixed {
			schemaDiscounts = append(schemaDiscounts, d)
		}
	}
	// Compute and save order level taxes amount
	discountTotalAmount := map[string]int64{}
	remainDiscountTotalAmount := map[string]int64{}
	for _, schemaDiscount := range schemaDiscounts {
		discount := b.lookup.Discount(schemaDiscount.UID)
		amount := discount.Fixed.Amount
		discountTotalAmount[schemaDiscount.UID] = amount
		remainDiscountTotalAmount[schemaDiscount.UID] = amount

		orderDiscount := OrderDiscount{
			UID:     schemaDiscount.UID,
			ID:      discount.ID,
			Name:    discount.Name,
			Fixed:   NewMoney(discount.Fixed.Amount, currency),
			Type:    DiscountTypeFixed,
			Applied: NewMoney(amount, currency),
		}
		order.Discounts = append(order.Discounts, orderDiscount)
	}

	// Apply order level discounts
	for i, orderItem := range order.Items {
		var appliedDiscounts []OrderItemAppliedDiscount
		var itemDiscountTotalAmount int64
		itemAmount := orderItem.Total.Amount
		for _, schemaDiscount := range schemaDiscounts {
			var itemDiscountAmount int64
			if i < len(order.Items)-1 {
				total := d.NewFromInt(discountTotalAmount[schemaDiscount.UID])
				factor := d.NewFromInt(itemAmount).Div(d.NewFromInt(order.Total.Amount))
				itemDiscountAmount = total.Mul(factor).RoundBank(0).IntPart()
			} else {
				itemDiscountAmount = remainDiscountTotalAmount[schemaDiscount.UID]
			}

			applied := OrderItemAppliedDiscount{
				DiscountUID: schemaDiscount.UID,
				Applied:     NewMoney(itemDiscountAmount, currency),
			}
			itemDiscountTotalAmount += itemDiscountAmount
			remainDiscountTotalAmount[schemaDiscount.UID] -= itemDiscountAmount
			appliedDiscounts = append(appliedDiscounts, applied)
		}
		order.Items[i].Total.Amount -= itemDiscountTotalAmount
		order.Items[i].AppliedDiscounts = append(orderItem.AppliedDiscounts, appliedDiscounts...)
		order.Items[i].TotalDiscount.Amount += itemDiscountTotalAmount
	}
	for _, v := range discountTotalAmount {
		order.Total.Amount -= v
		order.TotalDiscount.Amount += v
	}

}

func (b *OrderBuilder) applyOrderLevelPercentageDiscounts(order *Order) {
	var schemaDiscounts []OrderSchemaDiscount
	currency := b.schema.currency
	for _, d := range b.schema.Discounts {
		if b.lookup.Discount(d.UID).Type == DiscountTypePercentage {
			schemaDiscounts = append(schemaDiscounts, d)
		}
	}
	// Compute and save order total discounts
	discountTotalAmount := map[string]int64{}
	remainDiscountTotalAmount := map[string]int64{}
	for _, schemaDiscount := range schemaDiscounts {
		discount := b.lookup.Discount(schemaDiscount.UID)
		amount := discount.calculate(order.Total.Amount)
		discountTotalAmount[schemaDiscount.UID] = amount
		remainDiscountTotalAmount[schemaDiscount.UID] = amount

		orderDiscount := OrderDiscount{
			UID:        schemaDiscount.UID,
			ID:         discount.ID,
			Name:       discount.Name,
			Percentage: discount.Percentage,
			Type:       DiscountTypePercentage,
			Applied:    NewMoney(amount, currency),
		}
		order.Discounts = append(order.Discounts, orderDiscount)
	}

	// Apply order level discounts
	for i, orderItem := range order.Items {
		var appliedDiscounts []OrderItemAppliedDiscount
		var itemDiscountTotalAmount int64
		itemAmount := orderItem.Total.Amount
		for _, schemaDiscount := range schemaDiscounts {
			var itemDiscountAmount int64
			if i < len(order.Items)-1 {
				// Calculate item discount amount proportionally:
				//		discountItem = discountTotal * itemTotal / itemsTotal
				total := d.NewFromInt(discountTotalAmount[schemaDiscount.UID])
				factor := d.NewFromInt(itemAmount).Div(d.NewFromInt(order.Total.Amount))
				itemDiscountAmount = total.Mul(factor).RoundBank(0).IntPart()
			} else {
				itemDiscountAmount = remainDiscountTotalAmount[schemaDiscount.UID]
			}

			applied := OrderItemAppliedDiscount{
				DiscountUID: schemaDiscount.UID,
				Applied:     NewMoney(itemDiscountAmount, currency),
			}
			itemDiscountTotalAmount += itemDiscountAmount
			remainDiscountTotalAmount[schemaDiscount.UID] -= itemDiscountAmount
			appliedDiscounts = append(appliedDiscounts, applied)
		}
		order.Items[i].Total.Amount -= itemDiscountTotalAmount
		order.Items[i].AppliedDiscounts = append(orderItem.AppliedDiscounts, appliedDiscounts...)
		order.Items[i].TotalDiscount.Amount += itemDiscountTotalAmount
	}
	for _, v := range discountTotalAmount {
		order.Total.Amount -= v
		order.TotalDiscount.Amount += v
	}
}

func (b *OrderBuilder) applyOrderLevelTaxes(order *Order) {
	// Compute and save order level taxes amount
	taxTotalAmount := map[string]int64{}
	remainderTaxTotalAmount := map[string]int64{}
	currency := b.schema.currency
	for _, schemaTax := range b.schema.Taxes {
		tax := b.lookup.Tax(schemaTax.UID)
		ptg := d.NewFromFloat(tax.Percentage).Div(hundred)
		total := d.NewFromInt(order.Total.Amount)
		amount := ptg.Mul(total).RoundBank(0).IntPart()
		taxTotalAmount[schemaTax.UID] = amount
		remainderTaxTotalAmount[schemaTax.UID] = amount

		orderTax := OrderTax{
			UID:        schemaTax.UID,
			ID:         tax.ID,
			Name:       tax.Name,
			Percentage: tax.Percentage,
			Scope:      schemaTax.Scope,
			Applied:    NewMoney(amount, currency),
		}
		order.Taxes = append(order.Taxes, orderTax)
	}

	// Apply order level taxes
	for i, orderItem := range order.Items {
		var appliedTaxes []OrderItemAppliedTax
		var itemTaxTotalAmount int64
		itemAmount := orderItem.Total.Amount
		for _, schemaTax := range b.schema.Taxes {
			var itemTaxAmount int64
			if i < len(order.Items)-1 {
				// Calculate item tax amount proportionally:
				//		taxItem = taxTotal * itemTotal / itemsTotal
				total := d.NewFromInt(taxTotalAmount[schemaTax.UID])
				factor := d.NewFromInt(itemAmount).Div(d.NewFromInt(order.Total.Amount))
				itemTaxAmount = total.Mul(factor).RoundBank(0).IntPart()
			} else {
				itemTaxAmount = remainderTaxTotalAmount[schemaTax.UID]
			}

			applied := OrderItemAppliedTax{
				TaxUID:  schemaTax.UID,
				Applied: NewMoney(itemTaxAmount, currency),
			}
			itemTaxTotalAmount += itemTaxAmount
			remainderTaxTotalAmount[schemaTax.UID] -= itemTaxAmount
			appliedTaxes = append(appliedTaxes, applied)
		}
		order.Items[i].Total.Amount += itemTaxTotalAmount
		order.Items[i].AppliedTaxes = append(orderItem.AppliedTaxes, appliedTaxes...)
		order.Items[i].TotalTax.Amount += itemTaxTotalAmount
	}
	for _, v := range taxTotalAmount {
		order.Total.Amount += v
		order.TotalTax.Amount += v
	}
}

// Build creates a new order from an schema, will all monetary fields set
// TODO: Add item level taxes and order level discounts
func (s *OrderingService) build(ctx context.Context, sch OrderSchema) (*Order, error) {
	const op = errors.Op("core/OrderingService.build")
	if sch.LocationID == "" || sch.MerchantID == "" {
		return nil, errors.E(op, errors.KindValidation, "Invalid order schema")
	}
	order := NewOrder(sch.LocationID, sch.MerchantID)
	order.Schema = sch

	items, err := s.ItemVariationStorage.List(ctx, ItemVariationFilter{
		IDs: sch.itemVariationIDs(),
	})
	if err != nil {
		return nil, errors.E(op, err)
	}
	taxes, err := s.TaxStorage.List(ctx, TaxFilter{
		IDs: sch.taxIDs(),
	})
	if err != nil {
		return nil, errors.E(op, err)
	}
	discounts, err := s.DiscountStorage.List(ctx, DiscountFilter{
		IDs: sch.discountIDs(),
	})

	lookup, err := NewOrderLookup(sch, items, taxes, discounts)
	if err != nil {
		return nil, errors.E(op, errors.KindValidation, err)
	}

	builder := NewOrderBuilder(sch, lookup)

	// Populate order items and set starting totals
	builder.applyItemsAndInit(&order)

	// Apply order level discounts and update totals
	builder.applyOrderLevelPercentageDiscounts(&order)
	builder.applyOrderLevelFixedDiscounts(&order)

	// Apply order level taxes and set tax related fields
	builder.applyOrderLevelTaxes(&order)

	return &order, nil
}
