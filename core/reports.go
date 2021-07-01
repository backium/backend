package core

import (
	"context"
	"strings"

	"github.com/backium/backend/errors"
)

const (
	GroupingNone          GroupingType = "none"
	GroupingItemCategory  GroupingType = "item_category"
	GroupingItem          GroupingType = "item"
	GroupingItemVariation GroupingType = "item_variation"

	//TODO: Finish remaining groupings
	GroupingDay     GroupingType = "day"
	GroupingWeekDay GroupingType = "week_day"
)

type GroupingType string

func (g *GroupingType) Validate() bool {
	switch *g {
	case GroupingNone,
		GroupingItem,
		GroupingItemCategory,
		GroupingItemVariation:
		return true
	default:
		return false
	}
}

func GroupingTypes() string {
	return strings.Join([]string{
		string(GroupingNone),
		string(GroupingItem),
		string(GroupingItemCategory),
		string(GroupingItemVariation),
	}, ",")
}

type ReportService struct {
	OrderStorage         OrderStorage
	ItemStorage          ItemStorage
	ItemVariationStorage ItemVariationStorage
	CategoryStorage      CategoryStorage
}

type ReportFilter struct {
	MerchantID  string
	LocationIDs []string
	BeginTime   int64
	EndTime     int64
}

type Aggregations struct {
	TotalSalesAmount Money
	GrossSalesAmount Money
	NetSalesAmount   Money
	TaxAmount        Money
	DiscountAmount   Money
	ItemCount        int64
	DiscountCount    int64
	TaxCount         int64
	OrderCount       int64
}

type CustomReport struct {
	GroupType    GroupingType
	GroupValue   string
	SubReport    []CustomReport
	Aggregations Aggregations
}

func (svc *ReportService) GenerateCustom(ctx context.Context, groupType []GroupingType, filter ReportFilter) ([]CustomReport, error) {
	const op = errors.Op("core/ReportService.GenerateCustom")

	orders, err := svc.OrderStorage.List(ctx, OrderFilter{
		LocationIDs: filter.LocationIDs,
		MerchantID:  filter.MerchantID,
		BeginTime:   filter.BeginTime,
		EndTime:     filter.EndTime,
	})
	if err != nil {
		return nil, errors.E(op, err)
	}

	wrappedOrders := make([]WrappedOrder, len(orders))
	for i := range orders {
		wrappedOrders[i] = NewWrappedOrder(&orders[i])
	}

	reports, err := svc.generateCustom(wrappedOrders, groupType)
	if err != nil {
		return nil, errors.E(op, err)
	}

	return reports, nil
}

// WrappedOrder is a thin layer around an order that excludes/includes item variations
type WrappedOrder struct {
	// The Order that belongs to the group
	Order *Order

	// Included order items from the group
	included map[string]bool
}

// NewWrapperOrder created a Wrapped order with all its variations included
func NewWrappedOrder(order *Order) WrappedOrder {
	included := map[string]bool{}
	for _, v := range order.ItemVariations {
		included[v.UID] = true
	}
	return WrappedOrder{
		Order:    order,
		included: included,
	}
}

func (w *WrappedOrder) CloneWith(uids []string) WrappedOrder {
	clone := WrappedOrder{
		Order:    w.Order,
		included: map[string]bool{},
	}
	for _, uid := range uids {
		clone.included[uid] = true
	}
	return clone
}

// Contains checks if an item variaton is included in the wrapped order
func (w *WrappedOrder) Contains(uid string) bool {
	included := false
	for _, v := range w.Order.ItemVariations {
		if v.UID == uid {
			included = true
		}
	}
	return w.included[uid] && included
}

// Remove excludes an item variation from the wrapper order
func (w *WrappedOrder) Remove(uid string) {
	w.included[uid] = false
}

func (svc *ReportService) generateCustom(orders []WrappedOrder, groupBy []GroupingType) ([]CustomReport, error) {
	var reports []CustomReport
	if len(groupBy) == 0 {
		return reports, nil
	}

	// First group the orders using the first groupingType
	// Iterate over the groups and calculate aggregations
	// For each group call generateCustom recursively to generate subreports for the reimaining groupingTypes
	currentGroupType := groupBy[0]
	remainingGroupTypes := groupBy[1:]

	groupOrders, err := groupOrders(orders, currentGroupType)
	if err != nil {
		return nil, err
	}

	for groupValue, orders := range groupOrders {
		subreports, err := svc.generateCustom(orders, remainingGroupTypes)
		if err != nil {
			return nil, err
		}

		report := CustomReport{
			GroupType:    currentGroupType,
			GroupValue:   groupValue,
			SubReport:    subreports,
			Aggregations: calculateAggregations(orders, "PEN"),
		}
		reports = append(reports, report)
	}
	return reports, nil
}

func calculateAggregations(orders []WrappedOrder, currency string) Aggregations {
	var (
		totalSales     int64
		grossSales     int64
		netSales       int64
		taxAmount      int64
		discountAmount int64
		itemCount      int64
		taxCount       int64
		discountCount  int64
		orderCount     = int64(len(orders))
	)

	for _, order := range orders {
		for _, variation := range order.Order.ItemVariations {
			if order.Contains(variation.UID) {
				totalSales += variation.TotalAmount.Value
				grossSales += variation.GrossSales.Value
				netSales += variation.GrossSales.Value - variation.TotalDiscountAmount.Value
				taxAmount += variation.TotalTaxAmount.Value
				discountAmount += variation.TotalDiscountAmount.Value
				itemCount += variation.Quantity
				taxCount += int64(len(variation.AppliedTaxes))
				discountCount += int64(len(variation.AppliedDiscounts))
			}
		}
	}

	return Aggregations{
		TotalSalesAmount: NewMoney(totalSales, currency),
		GrossSalesAmount: NewMoney(grossSales, currency),
		NetSalesAmount:   NewMoney(netSales, currency),
		TaxAmount:        NewMoney(taxAmount, currency),
		DiscountAmount:   NewMoney(discountAmount, currency),
		ItemCount:        itemCount,
		TaxCount:         taxCount,
		DiscountCount:    discountCount,
		OrderCount:       orderCount,
	}
}

func groupOrders(orders []WrappedOrder, groupType GroupingType) (map[string][]WrappedOrder, error) {
	groupOrders := make(map[string][]WrappedOrder)

	switch groupType {
	case GroupingNone:
		groupOrders["all"] = orders
	case GroupingItemCategory:
		groupOrders = groupOrdersByCategory(orders)
	case GroupingItem:
		groupOrders = groupOrdersByItem(orders)
	case GroupingItemVariation:
		groupOrders = groupOrdersByItemVariation(orders)
	default:
		return nil, errors.E("Unsupported groupType")
	}

	return groupOrders, nil
}

func groupOrdersByCategory(orders []WrappedOrder) map[string][]WrappedOrder {
	groupOrders := make(map[string][]WrappedOrder)

	for _, order := range orders {
		uidsByGroup := map[string][]string{}
		for _, variation := range order.Order.ItemVariations {
			name := variation.CategoryName
			if order.Contains(variation.UID) {
				uidsByGroup[name] = append(uidsByGroup[name], variation.UID)
			}
		}

		for name, uids := range uidsByGroup {
			groupOrders[name] = append(groupOrders[name], order.CloneWith(uids))
		}
	}

	return groupOrders
}

func groupOrdersByItem(orders []WrappedOrder) map[string][]WrappedOrder {
	groupOrders := make(map[string][]WrappedOrder)

	for _, order := range orders {
		groupUIDs := map[string][]string{}
		for _, variation := range order.Order.ItemVariations {
			name := variation.ItemName
			if order.Contains(variation.UID) {
				groupUIDs[name] = append(groupUIDs[name], variation.UID)
			}
		}

		for name, uids := range groupUIDs {
			groupOrders[name] = append(groupOrders[name], order.CloneWith(uids))
		}
	}

	return groupOrders
}

func groupOrdersByItemVariation(orders []WrappedOrder) map[string][]WrappedOrder {
	groupOrders := make(map[string][]WrappedOrder)

	for _, order := range orders {
		groupUIDs := map[string][]string{}
		for _, variation := range order.Order.ItemVariations {
			name := variation.Name
			if order.Contains(variation.UID) {
				groupUIDs[name] = append(groupUIDs[name], variation.UID)
			}
		}

		for name, uids := range groupUIDs {
			groupOrders[name] = append(groupOrders[name], order.CloneWith(uids))
		}
	}

	return groupOrders
}
