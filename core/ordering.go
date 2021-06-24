package core

import (
	"context"

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
	norder, err := s.OrderStorage.Get(ctx, order.ID)
	if err != nil {
		return Order{}, errors.E(op, err)
	}
	return norder, nil
}

// OrderLookup provides easy access to order elements (items, taxes, discounts) using schema uids
type OrderLookup struct {
	item map[string]ItemVariation
	tax  map[string]Tax
}

func NewOrderLookup(sch OrderSchema, items []ItemVariation, taxes []Tax) (*OrderLookup, error) {
	// Save items by UID for easy access
	itemLookup := map[string]ItemVariation{}
	for _, oit := range sch.Items {
		for _, it := range items {
			if it.ID == oit.VariationID {
				itemLookup[oit.UID] = it
			}
		}
		if _, ok := itemLookup[oit.UID]; !ok {
			return nil, errors.E("Unknow items in order")
		}
	}

	// Save taxes by UID for easy access
	taxLookup := map[string]Tax{}
	for _, ot := range sch.Taxes {
		for _, t := range taxes {
			if t.ID == ot.ID {
				taxLookup[ot.UID] = t
			}
		}
		if _, ok := taxLookup[ot.UID]; !ok {
			return nil, errors.E("Unknow taxes in order")
		}
	}

	return &OrderLookup{
		item: itemLookup,
		tax:  taxLookup,
	}, nil
}

func (l *OrderLookup) Tax(uid string) Tax {
	return l.tax[uid]
}

func (l *OrderLookup) Item(uid string) ItemVariation {
	return l.item[uid]
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
	for _, schItem := range b.schema.Items {
		item := b.lookup.Item(schItem.UID)
		orderItem := OrderItem{
			UID:         schItem.UID,
			VariationID: schItem.VariationID,
			Name:        item.Name,
			Quantity:    schItem.Quantity,
			BasePrice: Money{
				Amount:   item.Price.Amount,
				Currency: b.schema.currency,
			},
			GrossSales: Money{
				Amount:   item.Price.Amount * schItem.Quantity,
				Currency: b.schema.currency,
			},
			TotalTax: Money{
				Amount:   0,
				Currency: b.schema.currency,
			},
			Total: Money{
				Amount:   item.Price.Amount * schItem.Quantity,
				Currency: b.schema.currency,
			},
		}
		order.Items = append(order.Items, orderItem)
		itemsTotalAmount += orderItem.Total.Amount
	}
	order.Total = Money{
		Amount:   itemsTotalAmount,
		Currency: b.schema.currency,
	}
}

func (b *OrderBuilder) applyOrderLevelTaxes(order *Order) {
	// Compute and save order level taxes amount
	taxTotalAmount := map[string]int64{}
	remainderTaxTotalAmount := map[string]int64{}
	for _, ot := range b.schema.Taxes {
		t := b.lookup.Tax(ot.UID)
		ptg := d.NewFromFloat(t.Percentage).Div(hundred)
		total := d.NewFromInt(order.Total.Amount)
		amount := ptg.Mul(total).RoundBank(0).IntPart()
		taxTotalAmount[ot.UID] = amount
		remainderTaxTotalAmount[ot.UID] = amount

		ordTax := OrderTax{
			UID:   ot.UID,
			ID:    t.ID,
			Name:  t.Name,
			Scope: ot.Scope,
			Applied: Money{
				Amount:   amount,
				Currency: b.schema.currency,
			},
		}
		order.Taxes = append(order.Taxes, ordTax)
	}

	// Apply order level taxes
	for i, ordItem := range order.Items {
		var appliedTaxes []OrderItemAppliedTax
		var itemTaxTotalAmount int64
		itemAmount := ordItem.Total.Amount
		for _, t := range b.schema.Taxes {
			// Calculate item tax amount proportionally:
			//		taxItem = taxTotal * itemTotal / itemsTotal
			var itemTaxAmount int64
			if i < len(order.Items)-1 {
				total := d.NewFromInt(taxTotalAmount[t.UID])
				factor := d.NewFromInt(itemAmount).Div(d.NewFromInt(order.Total.Amount))
				itemTaxAmount = total.Mul(factor).RoundBank(0).IntPart()
			} else {
				itemTaxAmount = remainderTaxTotalAmount[t.UID]
			}

			applied := OrderItemAppliedTax{
				TaxUID: t.UID,
				Applied: Money{
					Amount:   itemTaxAmount,
					Currency: b.schema.currency,
				},
			}
			itemTaxTotalAmount += itemTaxAmount
			remainderTaxTotalAmount[t.UID] -= itemTaxAmount
			appliedTaxes = append(appliedTaxes, applied)
		}
		order.Items[i].Total.Amount += itemTaxTotalAmount
		order.Items[i].AppliedTaxes = append(ordItem.AppliedTaxes, appliedTaxes...)
		order.Items[i].TotalTax.Amount += itemTaxTotalAmount
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
	currency := "PEN"

	items, err := s.ItemVariationStorage.List(ctx, ItemVariationFilter{
		IDs: sch.itemVariationIDs(),
	})
	if err != nil {
		return nil, errors.E(op, err)
	}
	taxes, err := s.TaxStorage.List(ctx, TaxFilter{
		IDs: sch.taxIDs(),
	})

	lookup, err := NewOrderLookup(sch, items, taxes)
	if err != nil {
		return nil, errors.E(op, errors.KindValidation, err)
	}

	builder := NewOrderBuilder(sch, lookup)

	// Populate order items and set starting totals
	builder.applyItemsAndInit(&order)

	// Apply order level taxes and set tax related fields
	builder.applyOrderLevelTaxes(&order)

	// Calculate order totals
	var (
		orderTotalAmount    int64
		orderTotalTaxAmount int64
	)
	for _, it := range order.Items {
		orderTotalAmount += it.Total.Amount
	}
	for _, t := range order.Taxes {
		orderTotalTaxAmount += t.Applied.Amount
	}
	order.TotalTax = Money{
		Amount:   orderTotalTaxAmount,
		Currency: currency,
	}
	order.Total = Money{
		Amount:   orderTotalAmount,
		Currency: currency,
	}

	return &order, nil
}
