package core

import (
	"context"

	"github.com/backium/backend/errors"
)

type OrderingService struct {
	OrderStorage         OrderStorage
	ItemVariationStorage ItemVariationStorage
	TaxStorage           TaxStorage
}

func (svc *OrderingService) CreateOrder(ctx context.Context, proto ProtoOrder) (Order, error) {
	const op = errors.Op("core/OrderingService.CreateOrder")
	preOrder := NewOrder()
	preOrder.LocationID = proto.LocationID
	preOrder.MerchantID = proto.MerchantID

	for _, it := range proto.Items {
		preOrder.Items = append(preOrder.Items, OrderItem{
			UID:         it.UID,
			VariationID: it.VariationID,
			Quantity:    it.Quantity,
		})
	}
	for _, t := range proto.Taxes {
		preOrder.Taxes = append(preOrder.Taxes, OrderTax{
			UID: t.UID,
			ID:  t.ID,
		})
	}

	c := NewOrderBuilder(svc.ItemVariationStorage, svc.TaxStorage)
	order, err := c.Build(ctx, preOrder)
	if err == ErrInvalidOrder {
		return Order{}, errors.E(op, err, errors.KindValidation)
	}
	if err != nil {
		return Order{}, errors.E(op, err, errors.KindUnexpected)
	}

	if err := svc.OrderStorage.Put(ctx, order); err != nil {
		return Order{}, errors.E(op, err)
	}
	norder, err := svc.OrderStorage.Get(ctx, order.ID)
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
	Taxes      []ProtoOrderTax
}

type ProtoOrderItem struct {
	UID         string
	VariationID string
	Quantity    int64
}

type ProtoOrderTax struct {
	UID   string
	ID    string
	Scope string
}
