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
			ctx = ContextWithUser(ctx, &User{})
			orderStorage := NewMockOrderStorage()
			variationStorage := NewMockItemVariationStorage()
			taxStorage := NewMockTaxStorage()
			discountStorage := NewMockDiscountStorage()
			categoryStorage := NewMockCategoryStorage()
			itemStorage := NewMockItemStorage()
			customerStorage := NewMockCustomerStorage()
			cashDrawerStorage := NewMockCashDrawerStorage()
			inventoryStorage := NewMockInventoryStorage()

			svc := OrderingService{
				OrderStorage:         orderStorage,
				ItemVariationStorage: variationStorage,
				TaxStorage:           taxStorage,
				DiscountStorage:      discountStorage,
				CategoryStorage:      categoryStorage,
				CustomerStorage:      customerStorage,
				CashDrawerStorage:    cashDrawerStorage,
				InventoryStorage:     inventoryStorage,
				ItemStorage:          itemStorage,
			}

			categoryStorage.ListFn = func(ctx context.Context, fil CategoryQuery) ([]Category, int64, error) {
				return tc.Categories, 0, nil
			}
			itemStorage.ListFn = func(ctx context.Context, fil ItemQuery) ([]Item, int64, error) {
				return tc.Items, 0, nil
			}
			variationStorage.ListFn = func(ctx context.Context, fil ItemVariationQuery) ([]ItemVariation, int64, error) {
				return tc.ItemVariations, 0, nil
			}
			taxStorage.ListFn = func(ctx context.Context, fil TaxQuery) ([]Tax, int64, error) {
				return tc.Taxes, 0, nil
			}
			discountStorage.ListFn = func(ctx context.Context, fil DiscountQuery) ([]Discount, int64, error) {
				return tc.Discounts, 0, nil
			}
			customerStorage.GetFn = func(ctx context.Context, id ID) (Customer, error) {
				return Customer{}, nil
			}
			cashDrawerStorage.ListFn = func(ctx context.Context, q CashDrawerQuery) ([]CashDrawer, int64, error) {
				return []CashDrawer{}, 0, nil
			}
			inventoryStorage.ListCountFn = func(ctx context.Context, q InventoryFilter) ([]InventoryCount, int64, error) {
				return []InventoryCount{}, 0, nil
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
