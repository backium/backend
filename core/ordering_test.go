package core

import (
	"context"
	"reflect"
	"testing"
)

type OrderingTestCase struct {
	Name  string
	Items []ItemVariation
	Taxes []Tax
	Req   OrderSchema
	Order Order
}

var testcases = []OrderingTestCase{
	{
		Name: "OnlyItems",
		Items: []ItemVariation{
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
					Amount:   1000,
					Currency: "PEN",
				},
			},
		},
		Req: OrderSchema{
			LocationID: "location_id",
			MerchantID: "merchant_id",
			Items: []OrderSchemaItem{
				{
					UID:         "variation1_uid",
					VariationID: "variation1_id",
					Quantity:    2,
				},
				{
					UID:         "variation2_uid",
					VariationID: "variation2_id",
					Quantity:    2,
				},
			},
		},
		Order: Order{
			LocationID: "location_id",
			MerchantID: "merchant_id",
			Items: []OrderItem{
				{
					UID:         "variation1_uid",
					VariationID: "variation1_id",
					Name:        "variation1",
					Quantity:    2,
					BasePrice: Money{
						Amount:   500,
						Currency: "PEN",
					},
					Total: Money{
						Amount:   1000,
						Currency: "PEN",
					},
				},
				{
					UID:         "variation2_uid",
					VariationID: "variation2_id",
					Name:        "variation2",
					Quantity:    2,
					BasePrice: Money{
						Amount:   1000,
						Currency: "PEN",
					},
					Total: Money{
						Amount:   2000,
						Currency: "PEN",
					},
				},
			},
			Total: Money{
				Amount:   3000,
				Currency: "PEN",
			},
		},
	},
	{
		Name: "OneItemWithTaxes",
		Items: []ItemVariation{
			{
				ID:   "variation1_id",
				Name: "variation1",
				Price: Money{
					Amount:   500,
					Currency: "PEN",
				},
			},
		},
		Taxes: []Tax{
			{
				ID:         "tax1_id",
				Name:       "IGV",
				Percentage: 20,
			},
		},
		Req: OrderSchema{
			LocationID: "location_id",
			MerchantID: "merchant_id",
			Items: []OrderSchemaItem{
				{
					UID:         "variation1_uid",
					VariationID: "variation1_id",
					Quantity:    2,
				},
			},
			Taxes: []OrderSchemaTax{
				{
					UID: "tax1_uid",
					ID:  "tax1_id",
				},
			},
		},
		Order: Order{
			LocationID: "location_id",
			MerchantID: "merchant_id",
			Items: []OrderItem{
				{
					UID:         "variation1_uid",
					VariationID: "variation1_id",
					Name:        "variation1",
					Quantity:    2,
					BasePrice: Money{
						Amount:   500,
						Currency: "PEN",
					},
					Total: Money{
						Amount:   1200,
						Currency: "PEN",
					},
					AppliedTaxes: []OrderItemAppliedTax{
						{
							TaxUID: "tax1_uid",
							Applied: Money{
								Amount:   200,
								Currency: "PEN",
							},
						},
					},
				},
			},
			Total: Money{
				Amount:   1200,
				Currency: "PEN",
			},
		},
	},
	{
		Name: "MultipleItemWithTaxes",
		Items: []ItemVariation{
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
					Amount:   1500,
					Currency: "PEN",
				},
			},
		},
		Taxes: []Tax{
			{
				ID:         "tax1_id",
				Name:       "IGV",
				Percentage: 20,
			},
		},
		Req: OrderSchema{
			LocationID: "location_id",
			MerchantID: "merchant_id",
			Items: []OrderSchemaItem{
				{
					UID:         "variation1_uid",
					VariationID: "variation1_id",
					Quantity:    2,
				},
				{
					UID:         "variation2_uid",
					VariationID: "variation2_id",
					Quantity:    3,
				},
			},
			Taxes: []OrderSchemaTax{
				{
					UID: "tax1_uid",
					ID:  "tax1_id",
				},
			},
		},
		Order: Order{
			LocationID: "location_id",
			MerchantID: "merchant_id",
			Items: []OrderItem{
				{
					UID:         "variation1_uid",
					VariationID: "variation1_id",
					Name:        "variation1",
					Quantity:    2,
					BasePrice: Money{
						Amount:   500,
						Currency: "PEN",
					},
					Total: Money{
						Amount:   1200,
						Currency: "PEN",
					},
					AppliedTaxes: []OrderItemAppliedTax{
						{
							TaxUID: "tax1_uid",
							Applied: Money{
								Amount:   200,
								Currency: "PEN",
							},
						},
					},
				},
				{
					UID:         "variation2_uid",
					VariationID: "variation2_id",
					Name:        "variation2",
					Quantity:    3,
					BasePrice: Money{
						Amount:   1500,
						Currency: "PEN",
					},
					Total: Money{
						Amount:   5400,
						Currency: "PEN",
					},
					AppliedTaxes: []OrderItemAppliedTax{
						{
							TaxUID: "tax1_uid",
							Applied: Money{
								Amount:   900,
								Currency: "PEN",
							},
						},
					},
				},
			},
			Total: Money{
				Amount:   6600,
				Currency: "PEN",
			},
		},
	},
}

func TestCreateOrder(t *testing.T) {
	for _, tc := range testcases {
		t.Run(tc.Name, func(t *testing.T) {
			ctx := context.Background()
			orderStorage := NewMockOrderStorage()
			variationStorage := NewMockItemVariationStorage()
			taxStorage := NewMockTaxStorage()

			svc := OrderingService{
				OrderStorage:         orderStorage,
				ItemVariationStorage: variationStorage,
				TaxStorage:           taxStorage,
			}

			variationStorage.ListFunc = func(ctx context.Context, fil ItemVariationFilter) ([]ItemVariation, error) {
				return tc.Items, nil
			}
			taxStorage.ListFn = func(ctx context.Context, fil TaxFilter) ([]Tax, error) {
				return tc.Taxes, nil
			}
			orderInMem := Order{}
			orderStorage.PutFn = func(ctx context.Context, order Order) error {
				orderInMem = order
				return nil
			}
			orderStorage.GetFn = func(ctx context.Context, id string) (Order, error) {
				return orderInMem, nil
			}

			order, err := svc.CreateOrder(ctx, tc.Req)
			if err != nil {
				t.Error("creating order: ", err)
			}

			if !reflect.DeepEqual(order.Total, tc.Order.Total) {
				t.Errorf("incorrent order total:\ngot: %v\nwant: %v\n", order.Total.Amount, tc.Order.Total.Amount)
			}

			for i, it := range order.Items {
				if !reflect.DeepEqual(it, tc.Order.Items[i]) {
					t.Errorf("incorrect order item[%v]:\ngot: %+v\nwant: %+v\n", i, it, tc.Order.Items[i])
				}
			}
		})
	}
}
