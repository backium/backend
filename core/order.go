package core

import (
	"context"
	"errors"

	d "github.com/shopspring/decimal"
)

var (
	ErrInvalidOrder = errors.New("invalid order")
	hundred         = d.NewFromInt(100)
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
}

type OrderItem struct {
	UID          string                `bson:"uid"`
	VariationID  string                `bson:"variation_id"`
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
	UID string `bson:"uid"`
	ID  string `bson:"id"`
}

func NewOrder() Order {
	return Order{
		ID:    generateID("order"),
		Items: []OrderItem{},
	}
}

type OrderStorage interface {
	Put(context.Context, Order) error
	Get(context.Context, string) (Order, error)
}

// OrderBuilder is used to calculate and set the monetary fields of an order.
type OrderBuilder struct {
	ItemVariationStorage ItemVariationStorage
	TaxStorage           TaxStorage
}

func NewOrderBuilder(ivs ItemVariationStorage, ts TaxStorage) *OrderBuilder {
	return &OrderBuilder{
		ItemVariationStorage: ivs,
		TaxStorage:           ts,
	}
}

// Build will take a snapshot of the order catalog object prices (items, taxes,
// discounts, etc) and return an order with item base prices set, discounts applied,
// taxes applied, and all gross-related fields set.
func (c *OrderBuilder) Build(ctx context.Context, order Order) (Order, error) {
	vids := make([]string, len(order.Items))
	for i, item := range order.Items {
		vids[i] = item.VariationID
	}
	tids := make([]string, len(order.Taxes))
	for i, t := range order.Taxes {
		tids[i] = t.ID
	}

	taxes, err := c.TaxStorage.List(ctx, TaxFilter{
		IDs: tids,
	})
	itemVars, err := c.ItemVariationStorage.List(ctx, ItemVariationFilter{
		IDs: vids,
	})
	if err != nil {
		return Order{}, err
	}
	if len(itemVars) == 0 {
		return Order{}, ErrInvalidOrder
	}
	for _, v := range itemVars {
		if v.Price.Currency != "PEN" {
			return Order{}, ErrInvalidOrder
		}
	}

	// Calculate base price and starting total for each item
	for i, it := range order.Items {
		itvar := ItemVariation{}
		for _, v := range itemVars {
			if v.ID == it.VariationID {
				itvar = v
				break
			}
		}
		if itvar.ID == "" {
			return Order{}, ErrInvalidOrder
		}
		order.Items[i].BasePrice = Money{
			Amount:   itvar.Price.Amount,
			Currency: "PEN",
		}
		order.Items[i].Total = Money{
			Amount:   itvar.Price.Amount * it.Quantity,
			Currency: "PEN",
		}
	}

	// Calculate current items total for tax calculation
	var itemsTotal int64
	for _, it := range order.Items {
		itemsTotal += it.Total.Amount
	}

	// Populate map from uid to Tax
	orderTax := map[string]Tax{}
	for _, ot := range order.Taxes {
		for _, t := range taxes {
			if t.ID == ot.ID {
				orderTax[ot.UID] = t
				break
			}
		}
		if _, ok := orderTax[ot.UID]; !ok {
			return Order{}, ErrInvalidOrder
		}
	}

	// Precompute order level tax to be applied
	orderTaxAmount := map[string]int64{}
	orderTaxRemainder := map[string]int64{}
	for _, ot := range order.Taxes {
		t := orderTax[ot.UID]
		ptg := d.NewFromInt(t.Percentage).Div(hundred)
		amount := d.NewFromInt(itemsTotal).Mul(ptg).RoundBank(0).IntPart()
		orderTaxAmount[ot.UID] = amount
		orderTaxRemainder[ot.UID] = amount
	}

	// Apply order level tax to each item proportionally
	for i, it := range order.Items {
		itemTotal := it.Total.Amount
		for _, ot := range order.Taxes {
			var amount int64
			if i < len(order.Items)-1 {
				// Apply tax proportionally : taxItem = taxTotal * itemTotal / itemsTotal
				factor := d.NewFromInt(itemTotal).Div(d.NewFromInt(itemsTotal))
				amount = d.NewFromInt(orderTaxAmount[ot.UID]).Mul(factor).RoundBank(0).IntPart()
			} else {
				amount = orderTaxRemainder[ot.UID]
			}

			applied := OrderItemAppliedTax{
				TaxUID: ot.UID,
				Applied: Money{
					Amount:   amount,
					Currency: "PEN",
				},
			}
			order.Items[i].AppliedTaxes = append(it.AppliedTaxes, applied)
			order.Items[i].Total.Amount += amount
			orderTaxRemainder[ot.UID] -= amount
		}
	}

	// Calculate order totals
	var orderTotal int64
	for _, it := range order.Items {
		orderTotal += it.Total.Amount
	}

	order.Total = Money{
		Amount:   orderTotal,
		Currency: "PEN",
	}

	return order, nil
}
