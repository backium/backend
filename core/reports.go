package core

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/backium/backend/errors"
)

const (
	GroupingNone          GroupingType = "none"
	GroupingCustomer      GroupingType = "customer"
	GroupingItemCategory  GroupingType = "item_category"
	GroupingItem          GroupingType = "item"
	GroupingItemVariation GroupingType = "item_variation"
	GroupingOrderState    GroupingType = "order_state"
	GroupingDay           GroupingType = "day"
	GroupingWeekday       GroupingType = "weekday"
	GroupingHourOfDay     GroupingType = "hour_of_day"
	GroupingMonth         GroupingType = "month"
)

type GroupingType string

func (g *GroupingType) Validate() bool {
	switch *g {
	case GroupingNone,
		GroupingCustomer,
		GroupingItem,
		GroupingItemCategory,
		GroupingItemVariation,
		GroupingOrderState,
		GroupingDay,
		GroupingWeekday,
		GroupingHourOfDay,
		GroupingMonth:
		return true
	default:
		return false
	}
}

func GroupingTypes() string {
	return strings.Join([]string{
		string(GroupingNone),
		string(GroupingCustomer),
		string(GroupingItem),
		string(GroupingItemCategory),
		string(GroupingItemVariation),
		string(GroupingOrderState),
		string(GroupingDay),
		string(GroupingWeekday),
		string(GroupingHourOfDay),
		string(GroupingMonth),
	}, ",")
}

type ReportService struct {
	OrderStorage         OrderStorage
	ItemStorage          ItemStorage
	ItemVariationStorage ItemVariationStorage
	CategoryStorage      CategoryStorage
}

type ReportFilter struct {
	MerchantID  ID
	LocationIDs []ID
	BeginTime   int64
	EndTime     int64
}

type Aggregations struct {
	TotalSalesAmount Money
	TotalCostAmount  Money
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

type CustomReportRequest struct {
	GroupType []GroupingType
	Timezone  string
	Filter    ReportFilter
}

func (svc *ReportService) GenerateCustom(ctx context.Context, req CustomReportRequest) ([]CustomReport, error) {
	const op = errors.Op("core/ReportService.GenerateCustom")

	orders, _, err := svc.OrderStorage.List(ctx, OrderQuery{
		Filter: OrderFilter{
			LocationIDs: req.Filter.LocationIDs,
			MerchantID:  req.Filter.MerchantID,
			CreatedAt: DateFilter{
				Gte: req.Filter.BeginTime,
				Lte: req.Filter.EndTime,
			},
		},
	})
	if err != nil {
		return nil, errors.E(op, err)
	}

	wrappedOrders := make([]WrappedOrder, len(orders))
	for i := range orders {
		wrappedOrders[i] = NewWrappedOrder(&orders[i])
	}

	reports, err := svc.generateCustom(wrappedOrders, req.GroupType, req.Timezone)
	if err != nil {
		return nil, errors.E(op, errors.KindValidation, err)
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

func (svc *ReportService) generateCustom(orders []WrappedOrder, groupBy []GroupingType, timezone string) ([]CustomReport, error) {
	var reports []CustomReport
	if len(groupBy) == 0 {
		return reports, nil
	}

	// First group the orders using the first groupingType
	// Iterate over the groups and calculate aggregations
	// For each group call generateCustom recursively to generate subreports for the reimaining groupingTypes
	currentGroupType := groupBy[0]
	remainingGroupTypes := groupBy[1:]

	orderGroups, err := groupOrders(orders, currentGroupType, timezone)
	if err != nil {
		return nil, err
	}

	for groupName, orders := range orderGroups {
		subreports, err := svc.generateCustom(orders, remainingGroupTypes, timezone)
		if err != nil {
			return nil, err
		}

		report := CustomReport{
			GroupType:    currentGroupType,
			GroupValue:   groupName,
			SubReport:    subreports,
			Aggregations: calculateAggregations(orders, "PEN"),
		}
		reports = append(reports, report)
	}
	return reports, nil
}

func calculateAggregations(orders []WrappedOrder, currency Currency) Aggregations {
	var (
		totalSales     int64
		totalCost      int64
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
				totalCost += variation.TotalCostAmount.Value
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
		TotalCostAmount:  NewMoney(totalCost, currency),
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

func groupOrders(orders []WrappedOrder, groupType GroupingType, timezone string) (map[string][]WrappedOrder, error) {
	orderGroups := make(map[string][]WrappedOrder)

	switch groupType {
	case GroupingNone:
		orderGroups["all"] = orders
	case GroupingCustomer:
		orderGroups = groupOrdersByCustomer(orders)
	case GroupingItemCategory:
		orderGroups = groupOrdersByCategory(orders)
	case GroupingItem:
		orderGroups = groupOrdersByItem(orders)
	case GroupingItemVariation:
		orderGroups = groupOrdersByItemVariation(orders)
	case GroupingOrderState:
		orderGroups = groupOrdersByState(orders)
	case GroupingDay:
		orderGroups = groupOrdersByDay(orders, timezone)
	case GroupingWeekday:
		orderGroups = groupOrdersByWeekday(orders, timezone)
	case GroupingHourOfDay:
		orderGroups = groupOrdersByHourOfDay(orders, timezone)
	case GroupingMonth:
		orderGroups = groupOrdersByMonth(orders, timezone)
	default:
		return nil, errors.E("Unsupported groupType")
	}

	return orderGroups, nil
}

func groupOrdersByCategory(orders []WrappedOrder) map[string][]WrappedOrder {
	orderGroups := make(map[string][]WrappedOrder)

	for _, order := range orders {
		uidGroups := map[string][]string{}
		for _, variation := range order.Order.ItemVariations {
			name := variation.CategoryName
			if order.Contains(variation.UID) {
				uidGroups[name] = append(uidGroups[name], variation.UID)
			}
		}

		for name, uids := range uidGroups {
			orderGroups[name] = append(orderGroups[name], order.CloneWith(uids))
		}
	}

	return orderGroups
}

func groupOrdersByItem(orders []WrappedOrder) map[string][]WrappedOrder {
	orderGroups := make(map[string][]WrappedOrder)

	for _, order := range orders {
		uidGroups := map[string][]string{}
		for _, variation := range order.Order.ItemVariations {
			name := variation.ItemName
			if order.Contains(variation.UID) {
				uidGroups[name] = append(uidGroups[name], variation.UID)
			}
		}

		for name, uids := range uidGroups {
			orderGroups[name] = append(orderGroups[name], order.CloneWith(uids))
		}
	}

	return orderGroups
}

func groupOrdersByItemVariation(orders []WrappedOrder) map[string][]WrappedOrder {
	orderGroups := make(map[string][]WrappedOrder)

	for _, order := range orders {
		uidGroups := map[string][]string{}
		for _, variation := range order.Order.ItemVariations {
			name := variation.Name
			if order.Contains(variation.UID) {
				uidGroups[name] = append(uidGroups[name], variation.UID)
			}
		}

		for name, uids := range uidGroups {
			orderGroups[name] = append(orderGroups[name], order.CloneWith(uids))
		}
	}

	return orderGroups
}

func groupOrdersByCustomer(orders []WrappedOrder) map[string][]WrappedOrder {
	orderGroups := make(map[string][]WrappedOrder)

	for _, order := range orders {
		uidGroups := map[string][]string{}

		name := string(order.Order.Customer.Name)
		for _, variation := range order.Order.ItemVariations {
			if order.Contains(variation.UID) {
				uidGroups[name] = append(uidGroups[name], variation.UID)
			}
		}

		for name, uids := range uidGroups {
			orderGroups[name] = append(orderGroups[name], order.CloneWith(uids))
		}
	}

	return orderGroups
}

func groupOrdersByState(orders []WrappedOrder) map[string][]WrappedOrder {
	orderGroups := make(map[string][]WrappedOrder)

	for _, order := range orders {
		uidGroups := map[string][]string{}

		name := string(order.Order.State)
		for _, variation := range order.Order.ItemVariations {
			if order.Contains(variation.UID) {
				uidGroups[name] = append(uidGroups[name], variation.UID)
			}
		}

		for name, uids := range uidGroups {
			orderGroups[name] = append(orderGroups[name], order.CloneWith(uids))
		}
	}

	return orderGroups
}

func groupOrdersByDay(orders []WrappedOrder, timezone string) map[string][]WrappedOrder {
	orderGroups := make(map[string][]WrappedOrder)
	location, _ := time.LoadLocation(timezone)

	for _, order := range orders {
		uidGroups := map[string][]string{}

		creationTime := time.Unix(order.Order.CreatedAt, 0).In(location)
		startOfDay := startOfDay(creationTime)
		endOfDay := endOfDay(creationTime)

		name := fmt.Sprintf("%v-%v", startOfDay.Unix(), endOfDay.Unix())
		for _, variation := range order.Order.ItemVariations {
			if order.Contains(variation.UID) {
				uidGroups[name] = append(uidGroups[name], variation.UID)
			}
		}

		for name, uids := range uidGroups {
			orderGroups[name] = append(orderGroups[name], order.CloneWith(uids))
		}
	}

	return orderGroups
}

func groupOrdersByMonth(orders []WrappedOrder, timezone string) map[string][]WrappedOrder {
	orderGroups := make(map[string][]WrappedOrder)
	location, _ := time.LoadLocation(timezone)

	for _, order := range orders {
		uidGroups := map[string][]string{}

		creationTime := time.Unix(order.Order.CreatedAt, 0).In(location)
		name := strings.ToLower(creationTime.Month().String())
		for _, variation := range order.Order.ItemVariations {
			if order.Contains(variation.UID) {
				uidGroups[name] = append(uidGroups[name], variation.UID)
			}
		}

		for name, uids := range uidGroups {
			orderGroups[name] = append(orderGroups[name], order.CloneWith(uids))
		}
	}

	return orderGroups
}

func groupOrdersByWeekday(orders []WrappedOrder, timezone string) map[string][]WrappedOrder {
	orderGroups := make(map[string][]WrappedOrder)
	location, _ := time.LoadLocation(timezone)

	for _, order := range orders {
		uidGroups := map[string][]string{}

		creationTime := time.Unix(order.Order.CreatedAt, 0).In(location)
		name := strings.ToLower(creationTime.Weekday().String())
		for _, variation := range order.Order.ItemVariations {
			if order.Contains(variation.UID) {
				uidGroups[name] = append(uidGroups[name], variation.UID)
			}
		}

		for name, uids := range uidGroups {
			orderGroups[name] = append(orderGroups[name], order.CloneWith(uids))
		}
	}

	return orderGroups
}

func groupOrdersByHourOfDay(orders []WrappedOrder, timezone string) map[string][]WrappedOrder {
	orderGroups := make(map[string][]WrappedOrder)
	location, _ := time.LoadLocation(timezone)

	for _, order := range orders {
		uidGroups := map[string][]string{}

		creationTime := time.Unix(order.Order.CreatedAt, 0).In(location)
		name := strconv.Itoa(creationTime.Hour())
		for _, variation := range order.Order.ItemVariations {
			if order.Contains(variation.UID) {
				uidGroups[name] = append(uidGroups[name], variation.UID)
			}
		}

		for name, uids := range uidGroups {
			orderGroups[name] = append(orderGroups[name], order.CloneWith(uids))
		}
	}

	return orderGroups
}

// startOfDay returns a time.Time set to the start of the day of a given time
func startOfDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

// endOfDay returns a time.Time set to the end of the day of a given time
func endOfDay(t time.Time) time.Time {
	return startOfDay(t.Add(24 * time.Hour))
}
