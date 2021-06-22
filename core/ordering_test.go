package core

import (
	"context"
	"testing"
)

func TestOrderingService(t *testing.T) {
	ctx := context.Background()
	items := []ItemVariation{
		{
			ID:   "variation1_id",
			Name: "variation1",
			Price: Money{
				Amount:   500,
				Currency: "PEN",
			},
		},
		{
			ID:   "variation2_id",
			Name: "variation2",
			Price: Money{
				Amount:   650,
				Currency: "PEN",
			},
		},
	}
	proto := ProtoOrder{
		LocationID: "location_id",
		MerchantID: "merchant_id",
		Items: []ProtoOrderItem{
			{
				UID:         "variation1_uid",
				VariationID: items[0].ID,
				Quantity:    1,
			},
			{
				UID:         "variation2_uid",
				VariationID: items[1].ID,
				Quantity:    1,
			},
		},
	}
	expectedOrder := Order{
		ID:         "order_id",
		LocationID: "location_id",
		MerchantID: "merchant_id",
		Items: []OrderItem{
			{
				UID:         "variation1_uid",
				VariationID: items[0].ID,
				Quantity:    1,
				Total: Money{
					Amount:   500,
					Currency: "PEN",
				},
			},
			{
				UID:         "variation2_uid",
				VariationID: items[1].ID,
				Quantity:    1,
				Total: Money{
					Amount:   650,
					Currency: "PEN",
				},
			},
		},
		Total: Money{
			Amount:   1150,
			Currency: "PEN",
		},
	}
	orderStorage := NewMockOrderStorage()
	variationStorage := NewMockItemVariationStorage()
	svc := OrderingService{
		OrderStorage:         orderStorage,
		ItemVariationStorage: variationStorage,
	}

	variationStorage.ListFunc = func(ctx context.Context, fil ItemVariationFilter) ([]ItemVariation, error) {
		return items, nil
	}
	orderInMem := Order{}
	orderStorage.CreateFunc = func(ctx context.Context, order Order) (string, error) {
		orderInMem = order
		orderInMem.ID = expectedOrder.ID
		return orderInMem.ID, nil
	}
	orderStorage.OrderFunc = func(ctx context.Context, id string) (Order, error) {
		return orderInMem, nil
	}

	order, err := svc.CreateOrder(ctx, proto)
	if err != nil {
		t.Error("something failed: ", err)
	}

	if order.Total.Amount != expectedOrder.Total.Amount {
		t.Errorf("incorrent order total amount:\ngot: %v\nwant: %v\n", order.Total.Amount, expectedOrder.Total.Amount)
	}
}
