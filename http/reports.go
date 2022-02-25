package http

import (
	"fmt"
	"net/http"

	"github.com/backium/backend/core"
	"github.com/backium/backend/errors"
	"github.com/labstack/echo/v4"
)

func (h *Handler) HandleGenerateStockReport(c echo.Context) error {
	const op = errors.Op("http/Handler.HandleGenerateStockReport")

	type request struct {
		LocationIDs      []core.ID `json:"location_ids" validate:"omitempty,dive,required"`
		ItemVariationIDs []core.ID `json:"item_variation_ids" validate:"omitempty,dive,required"`
	}

	type response struct {
		Report StockReport `json:"report"`
	}

	ctx := c.Request().Context()

	merchant := core.MerchantFromContext(ctx)
	if merchant == nil {
		return errors.E(op, errors.KindUnexpected, "invalid echo.Context")
	}

	req := request{}
	if err := bindAndValidate(c, &req); err != nil {
		return errors.E(op, err)
	}

	report, err := h.ReportService.GenerateStockReport(ctx, core.StockReportRequest{
		Filter: core.StockFilter{
			MerchantID:       merchant.ID,
			LocationIDs:      req.LocationIDs,
			ItemVariationIDs: req.ItemVariationIDs,
		},
	})
	if err != nil {
		return errors.E(op, err)
	}

	resp := response{
		Report: NewStockReport(report),
	}

	return c.JSON(http.StatusOK, resp)
}

func (h *Handler) HandleGenerateCustomReport(c echo.Context) error {
	const op = errors.Op("http/Handler.HandleGenerateCustomReport")

	type request struct {
		LocationIDs  []core.ID           `json:"location_ids" validate:"omitempty,dive,required"`
		EmployeeIDs  []core.ID           `json:"employee_ids" validate:"omitempty,dive,required"`
		CustomerIDs  []core.ID           `json:"customer_ids" validate:"omitempty,dive,required"`
		PaymentTypes []core.PaymentType  `json:"payment_types" validate:"omitempty,dive,required"`
		OrderStates  []core.OrderState   `json:"order_states" validate:"omitempty,dive,required"`
		GroupByType  []core.GroupingType `json:"group_by_type" validate:"required"`
		Timezone     string              `json:"timezone" validate:"required"`
		BeginTime    int64               `json:"begin_time" validate:"required"`
		EndTime      int64               `json:"end_time" validate:"required"`
	}

	type response struct {
		Report []CustomReport `json:"reports"`
	}

	ctx := c.Request().Context()

	merchant := core.MerchantFromContext(ctx)
	if merchant == nil {
		return errors.E(op, errors.KindUnexpected, "invalid echo.Context")
	}

	req := request{}
	if err := bindAndValidate(c, &req); err != nil {
		return errors.E(op, err)
	}

	for i, group := range req.GroupByType {
		if ok := group.Validate(); !ok {
			msg := fmt.Sprintf("request field 'group_by_type[%v]' is not valid, it should be one of: %v",
				i,
				core.GroupingTypes())
			return errors.E(op, errors.KindValidation, msg)
		}
	}

	reports, err := h.ReportService.GenerateCustom(ctx, core.CustomReportRequest{
		GroupType: req.GroupByType,
		Timezone:  req.Timezone,
		Filter: core.ReportFilter{
			MerchantID:   merchant.ID,
			LocationIDs:  req.LocationIDs,
			EmployeeIDs:  req.EmployeeIDs,
			CustomerIDs:  req.CustomerIDs,
			OrderStates:  req.OrderStates,
			PaymentTypes: req.PaymentTypes,
			BeginTime:    req.BeginTime,
			EndTime:      req.EndTime,
		},
	})
	if err != nil {
		return errors.E(op, err)
	}

	resp := response{}
	for _, report := range reports {
		resp.Report = append(resp.Report, NewCustomReport(report))
	}

	return c.JSON(http.StatusOK, resp)
}

type Aggregations struct {
	TotalSalesAmount Money `json:"total_sales_amount"`
	TotalCostAmount  Money `json:"total_cost_amount"`
	GrossSalesAmount Money `json:"gross_sales_amount"`
	NetSalesAmount   Money `json:"net_sales_amount"`
	TaxAmount        Money `json:"tax_amount"`
	DiscountAmount   Money `json:"discount_amount"`
	ItemCount        int64 `json:"item_count"`
	DiscountCount    int64 `json:"discount_count"`
	TaxCount         int64 `json:"tax_count"`
	OrderCount       int64 `json:"order_count"`
}

type StockReport struct {
	TotalStock  int64 `json:"total_stock"`
	TotalCost   Money `json:"total_cost"`
	TotalPrice  Money `json:"total_price"`
	TotalProfit Money `json:"total_profit"`
}

func NewStockReport(report core.StockReport) StockReport {
	return StockReport{
		TotalStock:  report.TotalStock,
		TotalCost:   NewMoney(report.TotalCost),
		TotalPrice:  NewMoney(report.TotalPrice),
		TotalProfit: NewMoney(report.TotalProfit),
	}
}

type CustomReport struct {
	GroupType    core.GroupingType `json:"group_type"`
	GroupValue   string            `json:"group_value"`
	SubReport    []CustomReport    `json:"subreport"`
	Aggregations Aggregations      `json:"aggregations"`
}

func NewCustomReport(report core.CustomReport) CustomReport {
	subreport := []CustomReport{}
	for _, sub := range report.SubReport {
		subreport = append(subreport, NewCustomReport(sub))
	}

	return CustomReport{
		GroupType:  report.GroupType,
		GroupValue: report.GroupValue,
		SubReport:  subreport,
		Aggregations: Aggregations{
			TotalSalesAmount: NewMoney(report.Aggregations.TotalSalesAmount),
			TotalCostAmount:  NewMoney(report.Aggregations.TotalCostAmount),
			GrossSalesAmount: NewMoney(report.Aggregations.GrossSalesAmount),
			NetSalesAmount:   NewMoney(report.Aggregations.NetSalesAmount),
			TaxAmount:        NewMoney(report.Aggregations.TaxAmount),
			DiscountAmount:   NewMoney(report.Aggregations.DiscountAmount),
			ItemCount:        report.Aggregations.ItemCount,
			TaxCount:         report.Aggregations.TaxCount,
			DiscountCount:    report.Aggregations.DiscountCount,
			OrderCount:       report.Aggregations.OrderCount,
		},
	}
}
