package core

import (
	"context"

	"github.com/backium/backend/errors"
)

type OrderingService struct {
	OrderStorage         OrderStorage
	ItemVariationStorage ItemVariationStorage
}

func (svc *OrderingService) CreateOrder(ctx context.Context, proto ProtoOrder) (Order, error) {
	const op = errors.Op("core/OrderingService.CreateOrder")
	preOrder := Order{
		LocationID: proto.LocationID,
		MerchantID: proto.MerchantID,
		Items:      []OrderItem{},
	}

	for _, item := range proto.Items {
		preOrder.Items = append(preOrder.Items, OrderItem{
			UID:         item.UID,
			VariationID: item.VariationID,
			Quantity:    item.Quantity,
		})
	}

	c := NewOrderCalculator(svc.ItemVariationStorage)
	order, err := c.Calculate(ctx, preOrder)
	if err == ErrInvalidOrder {
		return Order{}, errors.E(op, err, errors.KindValidation)
	}
	if err != nil {
		return Order{}, errors.E(op, err, errors.KindUnexpected)
	}

	id, err := svc.OrderStorage.Create(ctx, order)
	if err != nil {
		return Order{}, errors.E(op, err)
	}
	norder, err := svc.OrderStorage.Order(ctx, id)
	if err != nil {
		return Order{}, errors.E(op, err)
	}

	return norder, nil
}

// ProtoOrder represents a potential order to be created.
type ProtoOrder struct {
	LocationID string
	MerchantID string
	Items      []ProtoOrderItem
}

type ProtoOrderItem struct {
	UID         string
	VariationID string
	Quantity    int64
}
