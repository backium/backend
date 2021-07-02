package core

import (
	"testing"

	"github.com/stretchr/testify/assert"
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

	reports, _ := service.generateCustom(wrappedOrders, []GroupingType{GroupingItemCategory, GroupingItem}, "America/Bogota")

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
	wrappedOrders := []WrappedOrder{
		NewWrappedOrder(&Order{
			ID: NewID("order"),
			ItemVariations: []OrderItemVariation{
				{UID: "item1", Name: "burguer", CategoryName: "Food"},
				{UID: "item2", Name: "soda", CategoryName: "Drinks"},
			},
		}),
		NewWrappedOrder(&Order{
			ID: NewID("order"),
			ItemVariations: []OrderItemVariation{
				{UID: "item1", Name: "soda", CategoryName: "Drinks"},
				{UID: "item2", Name: "cookies", CategoryName: "Food"},
			},
		}),
	}
	expectedOrderGroups := map[string][]WrappedOrder{
		"Food": {
			{Order: wrappedOrders[0].Order, included: map[string]bool{"item1": true}},
			{Order: wrappedOrders[1].Order, included: map[string]bool{"item2": true}},
		},
		"Drinks": {
			{Order: wrappedOrders[0].Order, included: map[string]bool{"item2": true}},
			{Order: wrappedOrders[1].Order, included: map[string]bool{"item1": true}},
		},
	}

	orderGroups := groupOrdersByCategory(wrappedOrders)

	for category, groups := range expectedOrderGroups {
		if !assert.Equal(t, groups, orderGroups[category]) {
			t.Errorf("bad grouping for category %v", category)
		}
	}
}

func TestGroupOrdersByDay(t *testing.T) {
	timezone := "America/Bogota"
	wrappedOrders := []WrappedOrder{
		NewWrappedOrder(&Order{
			ID: NewID("order"),
			ItemVariations: []OrderItemVariation{
				{UID: "item1", Name: "burguer", CategoryName: "Food"},
				{UID: "item2", Name: "soda", CategoryName: "Drinks"},
			},
			// July 01
			CreatedAt: 1625179364,
			UpdatedAt: 1625179364,
		}),
		NewWrappedOrder(&Order{
			ID: NewID("order"),
			ItemVariations: []OrderItemVariation{
				{UID: "item1", Name: "soda", CategoryName: "Drinks"},
				{UID: "item2", Name: "cookies", CategoryName: "Food"},
			},
			// July 03
			CreatedAt: 1625299364,
			UpdatedAt: 1625299364,
		}),
	}

	expectedOrderGroups := map[string][]WrappedOrder{
		"1625115600-1625202000": {wrappedOrders[0]},
		"1625288400-1625374800": {wrappedOrders[1]},
	}

	orderGroups := groupOrdersByDay(wrappedOrders, timezone)

	for day, groups := range expectedOrderGroups {
		if !assert.Equal(t, groups, orderGroups[day]) {
			t.Errorf("bad grouping for day %v", day)
		}
	}
}

func TestGroupOrdersByWeekday(t *testing.T) {
	timezone := "America/Bogota"
	wrappedOrders := []WrappedOrder{
		NewWrappedOrder(&Order{
			ID: NewID("order"),
			ItemVariations: []OrderItemVariation{
				{UID: "item1", Name: "burguer", CategoryName: "Food"},
				{UID: "item2", Name: "soda", CategoryName: "Drinks"},
			},
			// Friday, July 2, 2021
			CreatedAt: 1625242000,
			UpdatedAt: 1625242000,
		}),
		NewWrappedOrder(&Order{
			ID: NewID("order"),
			ItemVariations: []OrderItemVariation{
				{UID: "item1", Name: "soda", CategoryName: "Drinks"},
				{UID: "item2", Name: "cookies", CategoryName: "Food"},
			},
			// Wednesday, July 14, 2021
			CreatedAt: 1626242000,
			UpdatedAt: 1626242000,
		}),
		NewWrappedOrder(&Order{
			ID: NewID("order"),
			ItemVariations: []OrderItemVariation{
				{UID: "item1", Name: "cookies", CategoryName: "Food"},
				{UID: "item2", Name: "burguer", CategoryName: "Food"},
			},
			// Sunday, July 25, 2021
			CreatedAt: 1627242000,
			UpdatedAt: 1627242000,
		}),
	}
	expectedOrderGroups := map[string][]WrappedOrder{
		"friday":    {NewWrappedOrder(wrappedOrders[0].Order)},
		"wednesday": {NewWrappedOrder(wrappedOrders[1].Order)},
		"sunday":    {NewWrappedOrder(wrappedOrders[2].Order)},
	}

	orderGroups := groupOrdersByWeekday(wrappedOrders, timezone)

	for weekday, groups := range expectedOrderGroups {
		if !assert.Equal(t, groups, orderGroups[weekday]) {
			t.Errorf("bad grouping for weekday %v", weekday)
		}
	}
}

func TestGroupOrdersByHourOfDay(t *testing.T) {
	timezone := "America/Bogota"
	wrappedOrders := []WrappedOrder{
		NewWrappedOrder(&Order{
			ID: NewID("order"),
			ItemVariations: []OrderItemVariation{
				{UID: "item1", Name: "burguer", CategoryName: "Food"},
				{UID: "item2", Name: "soda", CategoryName: "Drinks"},
			},
			// Friday, July 2, 2021 11:06:40 AM
			CreatedAt: 1625242000,
			UpdatedAt: 1625242000,
		}),
		NewWrappedOrder(&Order{
			ID: NewID("order"),
			ItemVariations: []OrderItemVariation{
				{UID: "item1", Name: "soda", CategoryName: "Drinks"},
				{UID: "item2", Name: "cookies", CategoryName: "Food"},
			},
			// Wednesday, July 14, 2021 12:53:20 AM
			CreatedAt: 1626242000,
			UpdatedAt: 1626242000,
		}),
		NewWrappedOrder(&Order{
			ID: NewID("order"),
			ItemVariations: []OrderItemVariation{
				{UID: "item1", Name: "cookies", CategoryName: "Food"},
				{UID: "item2", Name: "burguer", CategoryName: "Food"},
			},
			// Sunday, July 25, 2021 2:40:00 PM
			CreatedAt: 1627242000,
			UpdatedAt: 1627242000,
		}),
	}
	expectedOrderGroups := map[string][]WrappedOrder{
		"11": {NewWrappedOrder(wrappedOrders[0].Order)},
		"0":  {NewWrappedOrder(wrappedOrders[1].Order)},
		"14": {NewWrappedOrder(wrappedOrders[2].Order)},
	}

	orderGroups := groupOrdersByHourOfDay(wrappedOrders, timezone)

	for hourOfDay, groups := range expectedOrderGroups {
		if !assert.Equal(t, groups, orderGroups[hourOfDay]) {
			t.Errorf("bad grouping for hourOfDay %v", hourOfDay)
		}
	}
}
