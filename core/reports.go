package core

import (
	"context"

	"github.com/backium/backend/errors"
)

type GroupingType string

const (
	GroupingNone          GroupingType = "none"
	GroupingItemCategory  GroupingType = "item_category"
	GroupingItem          GroupingType = "item"
	GroupingItemVariation GroupingType = "item_variation"
	GroupingDay           GroupingType = "day"
)

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
	GroupBy      GroupingType
	GroupValue   string
	SubReport    []CustomReport
	Aggregations Aggregations
}

func (svc *ReportService) GenerateCustom(ctx context.Context, groupBy []GroupingType, generalFilter ReportFilter) ([]CustomReport, error) {
	const op = errors.Op("core/ReportService.GenerateCustom")
	orders, err := svc.OrderStorage.List(ctx, OrderFilter{
		LocationIDs: generalFilter.LocationIDs,
		MerchantID:  generalFilter.MerchantID,
		BeginTime:   generalFilter.BeginTime,
		EndTime:     generalFilter.EndTime,
	})
	if err != nil {
		return nil, errors.E(op, err)
	}

	report, err := svc.generateCustom(ctx, orders, groupBy, generalFilter)
	if err != nil {
		return nil, errors.E(op, err)
	}

	return report, nil
}

func (svc *ReportService) generateCustom(ctx context.Context, orders []Order, groupBy []GroupingType, generalFilter ReportFilter) ([]CustomReport, error) {
	var reports []CustomReport
	firstGroupType := groupBy[0]

	switch firstGroupType {
	case GroupingNone:
		reports = append(reports, CustomReport{
			GroupBy:    firstGroupType,
			GroupValue: "",
		})
	}
	return reports, nil
}

func (svc *ReportService) groupOrders(ctx context.Context, orders []Order, groupBy GroupingType) (map[string][]Order, error) {
	groupedOrders := make(map[string][]Order)
	switch groupBy {
	case GroupingNone:
		groupedOrders["all"] = orders
	}

	return groupedOrders, nil
}
