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
					GrossSales: Money{
						Amount:   1000,
						Currency: "PEN",
					},
					TotalTax: Money{
						Amount:   0,
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
					GrossSales: Money{
						Amount:   2000,
						Currency: "PEN",
					},
					TotalTax: Money{
						Amount:   0,
						Currency: "PEN",
					},
					Total: Money{
						Amount:   2000,
						Currency: "PEN",
					},
				},
			},
			TotalTax: Money{
				Amount:   0,
				Currency: "PEN",
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
				Name:       "tax1",
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
					UID:   "tax1_uid",
					ID:    "tax1_id",
					Scope: TaxScopeOrder,
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
					GrossSales: Money{
						Amount:   1000,
						Currency: "PEN",
					},
					TotalTax: Money{
						Amount:   200,
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
			Taxes: []OrderTax{
				{
					UID:   "tax1_uid",
					ID:    "tax1_id",
					Name:  "tax1",
					Scope: TaxScopeOrder,
					Applied: Money{
						Amount:   200,
						Currency: "PEN",
					},
				},
			},
			TotalTax: Money{
				Amount:   200,
				Currency: "PEN",
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
					Amount:   350,
					Currency: "PEN",
				},
			},
			{
				ID:   "variation2_id",
				Name: "variation2",
				Price: Money{
					Amount:   350,
					Currency: "PEN",
				},
			},
			{
				ID:   "variation3_id",
				Name: "variation3",
				Price: Money{
					Amount:   350,
					Currency: "PEN",
				},
			},
		},
		Taxes: []Tax{
			{
				ID:         "tax1_id",
				Name:       "tax1",
				Percentage: 9.25,
			},
		},
		Req: OrderSchema{
			LocationID: "location_id",
			MerchantID: "merchant_id",
			Items: []OrderSchemaItem{
				{
					UID:         "variation1_uid",
					VariationID: "variation1_id",
					Quantity:    1,
				},
				{
					UID:         "variation2_uid",
					VariationID: "variation2_id",
					Quantity:    1,
				},
				{
					UID:         "variation3_uid",
					VariationID: "variation3_id",
					Quantity:    1,
				},
			},
			Taxes: []OrderSchemaTax{
				{
					UID:   "tax1_uid",
					ID:    "tax1_id",
					Scope: TaxScopeOrder,
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
					Quantity:    1,
					BasePrice: Money{
						Amount:   350,
						Currency: "PEN",
					},
					GrossSales: Money{
						Amount:   350,
						Currency: "PEN",
					},
					TotalTax: Money{
						Amount:   32,
						Currency: "PEN",
					},
					Total: Money{
						Amount:   382,
						Currency: "PEN",
					},
					AppliedTaxes: []OrderItemAppliedTax{
						{
							TaxUID: "tax1_uid",
							Applied: Money{
								Amount:   32,
								Currency: "PEN",
							},
						},
					},
				},
				{
					UID:         "variation2_uid",
					VariationID: "variation2_id",
					Name:        "variation2",
					Quantity:    1,
					BasePrice: Money{
						Amount:   350,
						Currency: "PEN",
					},
					GrossSales: Money{
						Amount:   350,
						Currency: "PEN",
					},
					TotalTax: Money{
						Amount:   32,
						Currency: "PEN",
					},
					Total: Money{
						Amount:   382,
						Currency: "PEN",
					},
					AppliedTaxes: []OrderItemAppliedTax{
						{
							TaxUID: "tax1_uid",
							Applied: Money{
								Amount:   32,
								Currency: "PEN",
							},
						},
					},
				},
				{
					UID:         "variation3_uid",
					VariationID: "variation3_id",
					Name:        "variation3",
					Quantity:    1,
					BasePrice: Money{
						Amount:   350,
						Currency: "PEN",
					},
					GrossSales: Money{
						Amount:   350,
						Currency: "PEN",
					},
					TotalTax: Money{
						Amount:   33,
						Currency: "PEN",
					},
					Total: Money{
						Amount:   383,
						Currency: "PEN",
					},
					AppliedTaxes: []OrderItemAppliedTax{
						{
							TaxUID: "tax1_uid",
							Applied: Money{
								Amount:   33,
								Currency: "PEN",
							},
						},
					},
				},
			},
			Taxes: []OrderTax{
				{
					UID:   "tax1_uid",
					ID:    "tax1_id",
					Scope: TaxScopeOrder,
					Name:  "tax1",
					Applied: Money{
						Amount:   97,
						Currency: "PEN",
					},
				},
			},
			TotalTax: Money{
				Amount:   97,
				Currency: "PEN",
			},
			Total: Money{
				Amount:   1147,
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
				t.Errorf("incorrent order total:\ngot: %v\nwant: %v\n", order.Total, tc.Order.Total)
			}

			if !reflect.DeepEqual(order.TotalTax, tc.Order.TotalTax) {
				t.Errorf("incorrent order total tax:\ngot: %v\nwant: %v\n", order.TotalTax, tc.Order.TotalTax)
			}

			for i, it := range order.Items {
				if !reflect.DeepEqual(it, tc.Order.Items[i]) {
					t.Errorf("incorrect order item[%v]:\ngot: %+v\nwant: %+v\n", i, it, tc.Order.Items[i])
				}
			}

			if len(order.Taxes) != len(tc.Order.Taxes) {
				t.Errorf("incorrect number of order taxes:\ngot: %v\nwant: %v\n", len(order.Taxes), len(tc.Order.Taxes))
			}
			for i, ot := range order.Taxes {
				if !reflect.DeepEqual(ot, tc.Order.Taxes[i]) {
					t.Errorf("incorrect order tax[%v]:\ngot: %+v\nwant: %+v\n", i, ot, tc.Order.Taxes[i])
				}
			}
		})
	}
}
