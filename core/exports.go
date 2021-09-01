package core

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/backium/backend/errors"
	"github.com/xuri/excelize/v2"
)

type ExportService struct {
	OrderStorage    OrderStorage
	LocationStorage LocationStorage
	EmployeeStorage EmployeeStorage
	Uploader        Uploader
}

func (svc *ExportService) ExportOrders(ctx context.Context, q OrderQuery, timezone string) (string, error) {
	const op = errors.Op("core/ExportService.ExportOrders")

	locations, _, err := svc.LocationStorage.List(ctx, LocationQuery{
		Filter: LocationFilter{MerchantID: q.Filter.MerchantID},
	})
	if err != nil {
		return "", errors.E(op, err)
	}

	employees, _, err := svc.EmployeeStorage.List(ctx, EmployeeQuery{
		Filter: EmployeeFilter{MerchantID: q.Filter.MerchantID},
	})

	q.Sort = OrderSort{CreatedAt: SortAscending}
	orders, _, err := svc.OrderStorage.List(ctx, q)
	if err != nil {
		return "", errors.E(op, err)
	}

	sheetName := "Orders"
	filename := fmt.Sprintf("export_%v.xlsx", time.Now().UnixNano())
	title := &[]interface{}{
		"Date",
		"Hour",
		"Items",
		"Customer",
		"Employee",
		"Location",
		"Cost",
		"Price",
	}

	f := excelize.NewFile()
	f.NewSheet(sheetName)
	f.DeleteSheet("Sheet1")
	f.SetSheetRow(sheetName, "A1", title)
	f.SetColWidth(sheetName, "C", "C", 50)

	for i, order := range orders {
		var items []string
		var location string
		var employee string

		for _, item := range order.ItemVariations {
			items = append(items, item.ItemName)
		}
		for _, loc := range locations {
			if loc.ID == order.LocationID {
				location = loc.Name
			}
		}
		for _, emp := range employees {
			if emp.ID == order.EmployeeID {
				employee = fmt.Sprintf("%s %s", emp.FirstName, emp.LastName)
			}
		}

		locale, _ := time.LoadLocation(timezone)
		date := time.Unix(order.CreatedAt, 0).In(locale)

		f.SetSheetRow(sheetName, fmt.Sprintf("A%v", i+2), &[]interface{}{
			date.Format("02/01/2006"),
			date.Format("15:04 AM"),
			strings.Join(items, ","),
			order.Customer.Name,
			employee,
			location,
			centsToString(order.TotalCostAmount.Value),
			centsToString(order.TotalAmount.Value),
		})
	}

	if err := f.SaveAs(filename); err != nil {
		return "", errors.E(op, err)
	}
	defer os.Remove(filename)

	url, err := svc.Uploader.Upload(ctx, filename)
	if err != nil {
		return "", errors.E(op, err)
	}

	return url, nil
}
