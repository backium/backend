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
	LocationID string      `bson:"location_id"`
	MerchantID string      `bson:"merchant_id"`
	Items      []OrderItem `bson:"items"`
	Total      Money       `bson:"total"`
}

type OrderItem struct {
	UID         string `bson:"_uid"`
	VariationID string `bson:"variation_id"`
	Quantity    int64  `bson:"quantity"`
	Total       Money  `bson:"total"`
}

type OrderStorage interface {
	Create(context.Context, Order) (string, error)
	Order(context.Context, string) (Order, error)
}

// OrderCalculator is used to calculate and set the monetary fields of an order.
type OrderCalculator struct {
	ItemVariationStorage ItemVariationStorage
}

func NewOrderCalculator(ivs ItemVariationStorage) *OrderCalculator {
	return &OrderCalculator{
		ItemVariationStorage: ivs,
	}
}

// Calculate will take a snapshot of the order's item prices and return an order with
// all the monetary fields set to the corresponding calculated amounts.
func (c *OrderCalculator) Calculate(ctx context.Context, order Order) (Order, error) {
	ids := make([]string, len(order.Items))
	for i, item := range order.Items {
		ids[i] = item.VariationID
	}

	vars, err := c.ItemVariationStorage.List(ctx, ItemVariationFilter{
		IDs: ids,
	})
	if err != nil {
		return Order{}, err
	}

	if len(vars) == 0 {
		return Order{}, ErrInvalidOrder
	}

	var total int64
	for i, item := range order.Items {
		for _, v := range vars {
			if v.ID == item.VariationID {
				itemTotal := v.Price.Amount * item.Quantity
				order.Items[i].Total = Money{
					Amount:   itemTotal,
					Currency: v.Price.Currency,
				}
				total += itemTotal
			}
		}
	}
	order.Total = Money{
		Amount:   total,
		Currency: "PEN",
	}

	return order, nil
}
