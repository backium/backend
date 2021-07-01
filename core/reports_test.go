package core

import (
	"context"
	"testing"
)

func TestGenerateCustom(t *testing.T) {
	currency := "PEN"
	orders := []Order{
		{
			ID: NewID("order"),
			ItemVariations: []OrderItemVariation{
				{
					CategoryName: "Food",
					ItemName:     "Burguer",
					UID:          "item1",
					Quantity:     2,
					BasePrice:    NewMoney(2000, currency),
					GrossSales:   NewMoney(4000, currency),
					AppliedDiscounts: []OrderItemAppliedDiscount{
						{
							AppliedAmount: NewMoney(500, currency),
						},
					},
					AppliedTaxes: []OrderItemAppliedTax{
						{
							AppliedAmount: NewMoney(300, currency),
						},
					},
					TotalDiscountAmount: NewMoney(500, currency),
					TotalTaxAmount:      NewMoney(300, currency),
					TotalAmount:         NewMoney(3800, currency),
				},
				{
					CategoryName: "Food",
					ItemName:     "Cookies",
					UID:          "item2",
					Quantity:     2,
					BasePrice:    NewMoney(1000, currency),
					GrossSales:   NewMoney(2000, currency),
					AppliedDiscounts: []OrderItemAppliedDiscount{
						{
							AppliedAmount: NewMoney(500, currency),
						},
					},
					AppliedTaxes: []OrderItemAppliedTax{
						{
							AppliedAmount: NewMoney(300, currency),
						},
					},
					TotalDiscountAmount: NewMoney(500, currency),
					TotalTaxAmount:      NewMoney(300, currency),
					TotalAmount:         NewMoney(2800, currency),
				},
				{
					CategoryName: "Drinks",
					ItemName:     "Soda",
					UID:          "item3",
					Quantity:     2,
					BasePrice:    NewMoney(1000, currency),
					GrossSales:   NewMoney(2000, currency),
					AppliedDiscounts: []OrderItemAppliedDiscount{
						{
							AppliedAmount: NewMoney(500, currency),
						},
					},
					AppliedTaxes: []OrderItemAppliedTax{
						{
							AppliedAmount: NewMoney(300, currency),
						},
					},
					TotalDiscountAmount: NewMoney(500, currency),
					TotalTaxAmount:      NewMoney(300, currency),
					TotalAmount:         NewMoney(2800, currency),
				},
			},
		},
		{
			ID: NewID("order"),
			ItemVariations: []OrderItemVariation{
				{
					CategoryName: "Drinks",
					ItemName:     "Soda",
					UID:          "item1",
					Quantity:     5,
					BasePrice:    NewMoney(1000, currency),
					GrossSales:   NewMoney(5000, currency),
					AppliedDiscounts: []OrderItemAppliedDiscount{
						{
							AppliedAmount: NewMoney(500, currency),
						},
					},
					AppliedTaxes: []OrderItemAppliedTax{
						{
							AppliedAmount: NewMoney(300, currency),
						},
					},
					TotalDiscountAmount: NewMoney(500, currency),
					TotalTaxAmount:      NewMoney(300, currency),
					TotalAmount:         NewMoney(4800, currency),
				},
			},
		},
	}

	service := ReportService{}

	wrappedOrders := make([]WrappedOrder, len(orders))
	for i := range orders {
		wrappedOrders[i] = NewWrappedOrder(&orders[i])
	}

	reports, _ := service.generateCustom(context.TODO(), wrappedOrders, []GroupingType{GroupingItemCategory, GroupingItem})

	if categories := len(reports); categories != 2 {
		t.Errorf("Wrong number of report groups:\ngot: %v\nwant: %v", categories, 2)
	}

	for _, report := range reports {
		switch report.GroupValue {
		case "Food":
		case "Drinks":
		default:
			t.Errorf("Invalid report group value:\ngot: %v\nwant: %v", report.GroupValue, "Food or Drinks")
		}
	}
}

func TestGroupOrdersByCategory(t *testing.T) {
	orders := []Order{
		{
			ID: NewID("order"),
			ItemVariations: []OrderItemVariation{
				{
					CategoryName: "Food",
				},
				{
					CategoryName: "Food",
				},
				{
					CategoryName: "Drinks",
				},
			},
		},
		{
			ID: NewID("order"),
			ItemVariations: []OrderItemVariation{
				{
					CategoryName: "Drinks",
				},
			},
		},
	}

	service := ReportService{}

	wrapped := make([]WrappedOrder, len(orders))
	for i := range orders {
		wrapped[i] = NewWrappedOrder(&orders[i])
	}

	groupedOrders := service.groupOrdersByCategory(wrapped)

	if groups := len(groupedOrders); groups != 2 {
		t.Errorf("Wrong number of groups:\ngot :%v\nwant: %v", groups, 2)
	}

	if foodOrders := len(groupedOrders["Food"]); foodOrders != 1 {
		t.Errorf("Wrong number of food orders:\ngot :%v\nwant: %v", foodOrders, 1)
	}

	if drinksOrders := len(groupedOrders["Food"]); drinksOrders != 1 {
		t.Errorf("Wrong number of drinks orders:\ngot :%v\nwant: %v", drinksOrders, 2)
	}
}
