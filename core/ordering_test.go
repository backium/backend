package core

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

type OrderingTestCase struct {
	Name      string
	Items     []ItemVariation
	Taxes     []Tax
	Discounts []Discount
	Schema    OrderSchema
	Order     Order
}

func TestCreateOrder(t *testing.T) {
	var testcases []OrderingTestCase
	f, err := os.ReadFile("../testdata/orders.json")
	if err != nil {
		t.Errorf("reading orders test file: %v", err)
	}
	if err := json.Unmarshal(f, &testcases); err != nil {
		t.Errorf("unmarshaling ordering testcase file: %v", err)
	}

	for _, tc := range testcases {
		t.Run(tc.Name, func(t *testing.T) {
			ctx := context.Background()
			orderStorage := NewMockOrderStorage()
			variationStorage := NewMockItemVariationStorage()
			taxStorage := NewMockTaxStorage()
			discountStorage := NewMockDiscountStorage()

			svc := OrderingService{
				OrderStorage:         orderStorage,
				ItemVariationStorage: variationStorage,
				TaxStorage:           taxStorage,
				DiscountStorage:      discountStorage,
			}

			variationStorage.ListFn = func(ctx context.Context, fil ItemVariationFilter) ([]ItemVariation, error) {
				return tc.Items, nil
			}
			taxStorage.ListFn = func(ctx context.Context, fil TaxFilter) ([]Tax, error) {
				return tc.Taxes, nil
			}
			discountStorage.ListFn = func(ctx context.Context, fil DiscountFilter) ([]Discount, error) {
				return tc.Discounts, nil
			}
			orderInMem := Order{}
			orderStorage.PutFn = func(ctx context.Context, order Order) error {
				orderInMem = order
				return nil
			}
			orderStorage.GetFn = func(ctx context.Context, id, merchantID string, locationIDs []string) (Order, error) {
				return orderInMem, nil
			}

			order, err := svc.CreateOrder(ctx, tc.Schema)
			if err != nil {
				t.Error("creating order: ", err)
			}

			assert.Equal(t, tc.Order.Total, order.Total, "incorrect order total")
			assert.Equal(t, tc.Order.TotalTax, order.TotalTax, "incorrect order total tax")
			assert.Equal(t, tc.Order.Items, order.Items, "incorrect order items")
			assert.Equal(t, tc.Order.Taxes, order.Taxes, "incorrect order taxes")
			assert.Equal(t, tc.Order.Discounts, order.Discounts, "incorrect order discounts")
		})
	}
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
		Schema: OrderSchema{
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
					TotalDiscount: Money{
						Amount:   0,
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
					TotalDiscount: Money{
						Amount:   0,
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
			TotalDiscount: Money{
				Amount:   0,
				Currency: "PEN",
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
		Name: "OneItemWithDiscounts",
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
		Discounts: []Discount{
			{
				ID:         "discount1_id",
				Name:       "discount1",
				Type:       DiscountTypePercentage,
				Percentage: 20,
			},
		},
		Schema: OrderSchema{
			LocationID: "location_id",
			MerchantID: "merchant_id",
			Items: []OrderSchemaItem{
				{
					UID:         "variation1_uid",
					VariationID: "variation1_id",
					Quantity:    2,
				},
			},
			Discounts: []OrderSchemaDiscount{
				{
					UID: "discount1_uid",
					ID:  "discount1_id",
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
					AppliedDiscounts: []OrderItemAppliedDiscount{
						{
							DiscountUID: "discount1_uid",
							Applied: Money{
								Amount:   200,
								Currency: "PEN",
							},
						},
					},
					BasePrice: Money{
						Amount:   500,
						Currency: "PEN",
					},
					GrossSales: Money{
						Amount:   1000,
						Currency: "PEN",
					},
					TotalDiscount: Money{
						Amount:   200,
						Currency: "PEN",
					},
					TotalTax: Money{
						Amount:   0,
						Currency: "PEN",
					},
					Total: Money{
						Amount:   800,
						Currency: "PEN",
					},
				},
			},
			Discounts: []OrderDiscount{
				{
					UID:  "discount1_uid",
					ID:   "discount1_id",
					Name: "discount1",
					Applied: Money{
						Amount:   200,
						Currency: "PEN",
					},
				},
			},
			TotalDiscount: Money{
				Amount:   200,
				Currency: "PEN",
			},
			TotalTax: Money{
				Amount:   0,
				Currency: "PEN",
			},
			Total: Money{
				Amount:   800,
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
		Schema: OrderSchema{
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
					TotalDiscount: Money{
						Amount:   0,
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
			TotalDiscount: Money{
				Amount:   0,
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
		Schema: OrderSchema{
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
					TotalDiscount: Money{
						Amount:   0,
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
					TotalDiscount: Money{
						Amount:   0,
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
					TotalDiscount: Money{
						Amount:   0,
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
			TotalDiscount: Money{
				Amount:   0,
				Currency: "PEN",
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
