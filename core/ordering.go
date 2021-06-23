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
	order, err := s.build(ctx, sch)
	if err != nil {
		return Order{}, errors.E(op, err)
	}

	if err := s.OrderStorage.Put(ctx, order); err != nil {
		return Order{}, errors.E(op, err)
	}
	norder, err := s.OrderStorage.Get(ctx, order.ID)
	if err != nil {
		return Order{}, errors.E(op, err)
	}
	return norder, nil
}

// Build creates a new order from an schema, will all monetary fields set
// TODO: Add item level taxes and order level discounts
func (s *OrderingService) build(ctx context.Context, sch OrderSchema) (Order, error) {
	const op = errors.Op("core/OrderingService.build")
	if sch.LocationID == "" || sch.MerchantID == "" {
		return Order{}, errors.E(op, errors.KindValidation, "Invalid order schema")
	}
	order := NewOrder(sch.LocationID, sch.MerchantID)
	order.Schema = sch
	currency := "PEN"

	items, err := s.ItemVariationStorage.List(ctx, ItemVariationFilter{
		IDs: sch.itemVariationIDs(),
	})
	if err != nil {
		return Order{}, errors.E(op, err)
	}
	taxes, err := s.TaxStorage.List(ctx, TaxFilter{
		IDs: sch.taxIDs(),
	})

	// Save items by UID for easy access
	itemLookup := map[string]ItemVariation{}
	for _, oit := range sch.Items {
		for _, it := range items {
			if it.ID == oit.VariationID {
				itemLookup[oit.UID] = it
			}
		}
		if _, ok := itemLookup[oit.UID]; !ok {
			return Order{}, errors.E(op, errors.KindValidation, "Unknow items in order")
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
			return Order{}, errors.E(op, errors.KindValidation, "Unknow taxes in order")
		}
	}

	// Set order items and calculate initial totals
	var itemsTotalAmount int64
	for _, schItem := range sch.Items {
		item := itemLookup[schItem.UID]
		orderItem := OrderItem{
			UID:         schItem.UID,
			VariationID: schItem.VariationID,
			Name:        item.Name,
			Quantity:    schItem.Quantity,
			BasePrice: Money{
				Amount:   item.Price.Amount,
				Currency: currency,
			},
			Total: Money{
				Amount:   item.Price.Amount * schItem.Quantity,
				Currency: currency,
			},
		}
		order.Items = append(order.Items, orderItem)
		itemsTotalAmount += orderItem.Total.Amount
	}

	// Compute and save order level taxes amount
	taxTotalAmount := map[string]int64{}
	for _, ot := range sch.Taxes {
		t := taxLookup[ot.UID]
		ptg := d.NewFromInt(t.Percentage).Div(hundred)
		total := d.NewFromInt(itemsTotalAmount)
		amount := ptg.Mul(total).RoundBank(0).IntPart()
		taxTotalAmount[ot.UID] = amount

		ordTax := OrderTax{
			UID:   ot.UID,
			ID:    t.ID,
			Name:  t.Name,
			Scope: ot.Scope,
			Applied: Money{
				Amount:   amount,
				Currency: currency,
			},
		}
		order.Taxes = append(order.Taxes, ordTax)
	}

	// Apply order level taxes
	for i, ordItem := range order.Items {
		var appliedTaxes []OrderItemAppliedTax
		var itemTaxTotalAmount int64
		itemAmount := ordItem.Total.Amount
		for _, schTax := range sch.Taxes {
			total := d.NewFromInt(taxTotalAmount[schTax.UID])
			factor := d.NewFromInt(itemAmount).Div(d.NewFromInt(itemsTotalAmount))
			itemTaxAmount := total.Mul(factor).RoundBank(0).IntPart()

			applied := OrderItemAppliedTax{
				TaxUID: schTax.UID,
				Applied: Money{
					Amount:   itemTaxAmount,
					Currency: currency,
				},
			}
			itemTaxTotalAmount += itemTaxAmount
			appliedTaxes = append(appliedTaxes, applied)
		}
		order.Items[i].Total.Amount += itemTaxTotalAmount
		order.Items[i].AppliedTaxes = append(ordItem.AppliedTaxes, appliedTaxes...)
	}

	// Calculate order totals
	var (
		orderTotalAmount    int64
		orderTotalTaxAmount int64
	)
	for _, ordItem := range order.Items {
		orderTotalAmount += ordItem.Total.Amount
	}
	for _, v := range taxTotalAmount {
		orderTotalTaxAmount += v
	}
	order.TotalTax = Money{
		Amount:   orderTotalTaxAmount,
		Currency: currency,
	}
	order.Total = Money{
		Amount:   orderTotalAmount,
		Currency: currency,
	}

	return order, nil
}
