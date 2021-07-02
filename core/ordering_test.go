package core

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateOrder(t *testing.T) {
	type testcase struct {
		Name           string
		Categories     []Category
		Items          []Item
		ItemVariations []ItemVariation
		Taxes          []Tax
		Discounts      []Discount
		Schema         OrderSchema
		Order          Order
	}

	var testcases []testcase
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
			categoryStorage := NewMockCategoryStorage()
			itemStorage := NewMockItemStorage()

			svc := OrderingService{
				OrderStorage:         orderStorage,
				ItemVariationStorage: variationStorage,
				TaxStorage:           taxStorage,
				DiscountStorage:      discountStorage,
				CategoryStorage:      categoryStorage,
				ItemStorage:          itemStorage,
			}

			categoryStorage.ListFn = func(ctx context.Context, fil CategoryFilter) ([]Category, error) {
				return tc.Categories, nil
			}
			itemStorage.ListFn = func(ctx context.Context, fil ItemFilter) ([]Item, error) {
				return tc.Items, nil
			}
			variationStorage.ListFn = func(ctx context.Context, fil ItemVariationFilter) ([]ItemVariation, error) {
				return tc.ItemVariations, nil
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
			orderStorage.GetFn = func(ctx context.Context, id ID) (Order, error) {
				return orderInMem, nil
			}

			order, err := svc.CreateOrder(ctx, tc.Schema)
			if err != nil {
				t.Error("creating order: ", err)
			}

			assert.Equal(t, tc.Order.TotalAmount, order.TotalAmount, "incorrect order total")
			assert.Equal(t, tc.Order.TotalTaxAmount, order.TotalTaxAmount, "incorrect order total tax")
			assert.Equal(t, tc.Order.ItemVariations, order.ItemVariations, "incorrect order items")
			assert.Equal(t, tc.Order.Taxes, order.Taxes, "incorrect order taxes")
			assert.Equal(t, tc.Order.Discounts, order.Discounts, "incorrect order discounts")
		})
	}
}
