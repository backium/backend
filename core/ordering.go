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
	PaymentStorage       PaymentStorage
}

func (s *OrderingService) ListOrder(ctx context.Context, f OrderFilter) ([]Order, error) {
	const op = errors.Op("core/OrderingService.ListOrder")
	orders, err := s.OrderStorage.List(ctx, OrderFilter{
		LocationIDs: f.LocationIDs,
		MerchantID:  f.MerchantID,
		Limit:       f.Limit,
		Offset:      f.Offset,
	})
	if err != nil {
		return nil, errors.E(op, err)
	}
	return orders, nil
}

func (s *OrderingService) CreateOrder(ctx context.Context, schema OrderSchema) (Order, error) {
	const op = errors.Op("core/OrderingService.CreateOrder")
	schema.Currency = "PEN"
	order, err := s.build(ctx, schema)
	if err != nil {
		return Order{}, errors.E(op, err)
	}

	if err := s.OrderStorage.Put(ctx, *order); err != nil {
		return Order{}, errors.E(op, err)
	}
	newOrder, err := s.OrderStorage.Get(ctx, order.ID, order.MerchantID, nil)
	if err != nil {
		return Order{}, errors.E(op, err)
	}
	return newOrder, nil
}

func (s *OrderingService) PayOrder(ctx context.Context, orderID, merchantID string,
	paymentIDs []string) (Order, error) {
	const op = errors.Op("core/OrderingService.CreateOrder")
	order, err := s.OrderStorage.Get(ctx, orderID, merchantID, nil)
	if err != nil {
		return Order{}, errors.E(op, err)
	}
	payments, err := s.PaymentStorage.List(ctx, PaymentFilter{
		IDs: paymentIDs,
	})
	if err != nil {
		return Order{}, errors.E(op, errors.KindUnexpected, err)
	}
	if len(payments) == 0 {
		return Order{}, errors.E(op, errors.KindValidation, "Payments not found")
	}
	var payAmount int64
	var tipAmount int64
	for _, payment := range payments {
		if payment.OrderID != order.ID {
			return Order{}, errors.E(op, errors.KindValidation,
				fmt.Sprintf("Payment '%v' is not attached to the order", payment.ID))
		}
		payAmount += payment.Amount.Value
		tipAmount += payment.TipAmount.Value
	}

	order.TotalTipAmount = NewMoney(tipAmount, order.Schema.Currency)
	order.TotalAmount.Value += tipAmount
	if order.TotalAmount.Value == payAmount {
		order.State = OrderStateCompleted
	}

	if err := s.OrderStorage.Put(ctx, order); err != nil {
		return Order{}, errors.E(op, errors.KindUnexpected, err)
	}
	order, err = s.OrderStorage.Get(ctx, order.ID, merchantID, nil)
	if err != nil {
		return Order{}, errors.E(op, errors.KindUnexpected, err)
	}

	return order, nil
}

// OrderLookup provides easy access to order elements (items, taxes, discounts) using schema uids
type OrderLookup struct {
	item     map[string]ItemVariation
	tax      map[string]Tax
	discount map[string]Discount
}

func NewOrderLookup(schema OrderSchema, items []ItemVariation, taxes []Tax, discounts []Discount) (*OrderLookup, error) {
	// Save items by UID for easy access
	itemLookup := map[string]ItemVariation{}
	for _, schemaItem := range schema.Items {
		for _, item := range items {
			if item.ID == schemaItem.VariationID {
				itemLookup[schemaItem.UID] = item
			}
		}
		if _, ok := itemLookup[schemaItem.UID]; !ok {
			return nil, errors.E(fmt.Sprintf("Item variation '%v' doesn't exist or is not available.", schemaItem.UID))
		}
	}

	taxLookup := map[string]Tax{}
	for _, schemaTax := range schema.Taxes {
		for _, tax := range taxes {
			if tax.ID == schemaTax.ID {
				taxLookup[schemaTax.UID] = tax
			}
		}
		if _, ok := taxLookup[schemaTax.UID]; !ok {
			return nil, errors.E(fmt.Sprintf("Tax '%v' doesn't exist or is not available.", schemaTax.UID))
		}
	}

	discountLookup := map[string]Discount{}
	for _, schemaDiscount := range schema.Discounts {
		for _, discount := range discounts {
			if discount.ID == schemaDiscount.ID {
				discountLookup[schemaDiscount.UID] = discount
			}
		}
		if _, ok := discountLookup[schemaDiscount.UID]; !ok {
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
	currency := b.schema.Currency
	for _, schemaItem := range b.schema.Items {
		item := b.lookup.Item(schemaItem.UID)
		orderItem := OrderItem{
			UID:                 schemaItem.UID,
			VariationID:         schemaItem.VariationID,
			Name:                item.Name,
			Quantity:            schemaItem.Quantity,
			BasePrice:           NewMoney(item.Price.Value, currency),
			GrossSales:          NewMoney(item.Price.Value*schemaItem.Quantity, currency),
			TotalDiscountAmount: NewMoney(0, currency),
			TotalTaxAmount:      NewMoney(0, currency),
			TotalAmount:         NewMoney(item.Price.Value*schemaItem.Quantity, currency),
		}
		order.Items = append(order.Items, orderItem)
		itemsTotalAmount += orderItem.TotalAmount.Value
	}
	order.TotalDiscountAmount = NewMoney(0, currency)
	order.TotalTaxAmount = NewMoney(0, currency)
	order.TotalAmount = NewMoney(itemsTotalAmount, currency)
}

func (b *OrderBuilder) applyOrderLevelFixedDiscounts(order *Order) {
	var schemaDiscounts []OrderSchemaDiscount
	currency := b.schema.Currency
	for _, schemaDiscount := range b.schema.Discounts {
		if b.lookup.Discount(schemaDiscount.UID).Type == DiscountFixed {
			schemaDiscounts = append(schemaDiscounts, schemaDiscount)
		}
	}
	// Compute and save order level taxes amount
	discountTotalAmount := map[string]int64{}
	remainDiscountTotalAmount := map[string]int64{}
	for _, schemaDiscount := range schemaDiscounts {
		discount := b.lookup.Discount(schemaDiscount.UID)
		amount := discount.Amount.Value
		discountTotalAmount[schemaDiscount.UID] = amount
		remainDiscountTotalAmount[schemaDiscount.UID] = amount

		orderDiscount := OrderDiscount{
			UID:           schemaDiscount.UID,
			ID:            discount.ID,
			Name:          discount.Name,
			Amount:        NewMoney(discount.Amount.Value, currency),
			Type:          DiscountFixed,
			AppliedAmount: NewMoney(amount, currency),
		}
		order.Discounts = append(order.Discounts, orderDiscount)
	}

	// Apply order level discounts
	for i, orderItem := range order.Items {
		var appliedDiscounts []OrderItemAppliedDiscount
		var itemDiscountTotalAmount int64
		itemAmount := orderItem.TotalAmount.Value
		for _, schemaDiscount := range schemaDiscounts {
			var itemDiscountAmount int64
			if i < len(order.Items)-1 {
				total := d.NewFromInt(discountTotalAmount[schemaDiscount.UID])
				factor := d.NewFromInt(itemAmount).Div(d.NewFromInt(order.TotalAmount.Value))
				itemDiscountAmount = total.Mul(factor).RoundBank(0).IntPart()
			} else {
				itemDiscountAmount = remainDiscountTotalAmount[schemaDiscount.UID]
			}

			applied := OrderItemAppliedDiscount{
				DiscountUID:   schemaDiscount.UID,
				AppliedAmount: NewMoney(itemDiscountAmount, currency),
			}
			itemDiscountTotalAmount += itemDiscountAmount
			remainDiscountTotalAmount[schemaDiscount.UID] -= itemDiscountAmount
			appliedDiscounts = append(appliedDiscounts, applied)
		}
		order.Items[i].TotalAmount.Value -= itemDiscountTotalAmount
		order.Items[i].AppliedDiscounts = append(orderItem.AppliedDiscounts, appliedDiscounts...)
		order.Items[i].TotalDiscountAmount.Value += itemDiscountTotalAmount
	}
	for _, amount := range discountTotalAmount {
		order.TotalAmount.Value -= amount
		order.TotalDiscountAmount.Value += amount
	}
}

func (b *OrderBuilder) applyOrderLevelPercentageDiscounts(order *Order) {
	var schemaDiscounts []OrderSchemaDiscount
	currency := b.schema.Currency
	for _, d := range b.schema.Discounts {
		if b.lookup.Discount(d.UID).Type == DiscountPercentage {
			schemaDiscounts = append(schemaDiscounts, d)
		}
	}
	// Compute and save order total discounts
	discountTotalAmount := map[string]int64{}
	remainDiscountTotalAmount := map[string]int64{}
	for _, schemaDiscount := range schemaDiscounts {
		discount := b.lookup.Discount(schemaDiscount.UID)
		amount := discount.calculate(order.TotalAmount.Value)
		discountTotalAmount[schemaDiscount.UID] = amount
		remainDiscountTotalAmount[schemaDiscount.UID] = amount

		orderDiscount := OrderDiscount{
			UID:           schemaDiscount.UID,
			ID:            discount.ID,
			Name:          discount.Name,
			Percentage:    discount.Percentage,
			Type:          DiscountPercentage,
			AppliedAmount: NewMoney(amount, currency),
		}
		order.Discounts = append(order.Discounts, orderDiscount)
	}

	// Apply order level discounts
	for i, orderItem := range order.Items {
		var appliedDiscounts []OrderItemAppliedDiscount
		var itemDiscountTotalAmount int64
		itemAmount := orderItem.TotalAmount.Value
		for _, schemaDiscount := range schemaDiscounts {
			var itemDiscountAmount int64
			if i < len(order.Items)-1 {
				// Calculate item discount amount proportionally:
				//		discountItem = discountTotal * itemTotal / itemsTotal
				total := d.NewFromInt(discountTotalAmount[schemaDiscount.UID])
				factor := d.NewFromInt(itemAmount).Div(d.NewFromInt(order.TotalAmount.Value))
				itemDiscountAmount = total.Mul(factor).RoundBank(0).IntPart()
			} else {
				itemDiscountAmount = remainDiscountTotalAmount[schemaDiscount.UID]
			}

			applied := OrderItemAppliedDiscount{
				DiscountUID:   schemaDiscount.UID,
				AppliedAmount: NewMoney(itemDiscountAmount, currency),
			}
			itemDiscountTotalAmount += itemDiscountAmount
			remainDiscountTotalAmount[schemaDiscount.UID] -= itemDiscountAmount
			appliedDiscounts = append(appliedDiscounts, applied)
		}
		order.Items[i].TotalAmount.Value -= itemDiscountTotalAmount
		order.Items[i].AppliedDiscounts = append(orderItem.AppliedDiscounts, appliedDiscounts...)
		order.Items[i].TotalDiscountAmount.Value += itemDiscountTotalAmount
	}
	for _, amount := range discountTotalAmount {
		order.TotalAmount.Value -= amount
		order.TotalDiscountAmount.Value += amount
	}
}

func (b *OrderBuilder) applyOrderLevelTaxes(order *Order) {
	// Compute and save order level taxes amount
	taxTotalAmount := map[string]int64{}
	remainderTaxTotalAmount := map[string]int64{}
	currency := b.schema.Currency
	for _, schemaTax := range b.schema.Taxes {
		tax := b.lookup.Tax(schemaTax.UID)
		ptg := d.NewFromFloat(tax.Percentage).Div(hundred)
		total := d.NewFromInt(order.TotalAmount.Value)
		amount := ptg.Mul(total).RoundBank(0).IntPart()
		taxTotalAmount[schemaTax.UID] = amount
		remainderTaxTotalAmount[schemaTax.UID] = amount

		orderTax := OrderTax{
			UID:           schemaTax.UID,
			ID:            tax.ID,
			Name:          tax.Name,
			Percentage:    tax.Percentage,
			Scope:         schemaTax.Scope,
			AppliedAmount: NewMoney(amount, currency),
		}
		order.Taxes = append(order.Taxes, orderTax)
	}

	// Apply order level taxes
	for i, orderItem := range order.Items {
		var appliedTaxes []OrderItemAppliedTax
		var itemTaxTotalAmount int64
		itemAmount := orderItem.TotalAmount.Value
		for _, schemaTax := range b.schema.Taxes {
			var itemTaxAmount int64
			if i < len(order.Items)-1 {
				// Calculate item tax amount proportionally:
				//		taxItem = taxTotal * itemTotal / itemsTotal
				total := d.NewFromInt(taxTotalAmount[schemaTax.UID])
				factor := d.NewFromInt(itemAmount).Div(d.NewFromInt(order.TotalAmount.Value))
				itemTaxAmount = total.Mul(factor).RoundBank(0).IntPart()
			} else {
				itemTaxAmount = remainderTaxTotalAmount[schemaTax.UID]
			}

			applied := OrderItemAppliedTax{
				TaxUID:        schemaTax.UID,
				AppliedAmount: NewMoney(itemTaxAmount, currency),
			}
			itemTaxTotalAmount += itemTaxAmount
			remainderTaxTotalAmount[schemaTax.UID] -= itemTaxAmount
			appliedTaxes = append(appliedTaxes, applied)
		}
		order.Items[i].TotalAmount.Value += itemTaxTotalAmount
		order.Items[i].AppliedTaxes = append(orderItem.AppliedTaxes, appliedTaxes...)
		order.Items[i].TotalTaxAmount.Value += itemTaxTotalAmount
	}
	for _, amount := range taxTotalAmount {
		order.TotalAmount.Value += amount
		order.TotalTaxAmount.Value += amount
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
