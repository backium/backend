package core

import (
	"context"
	"fmt"
	"html/template"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/backium/backend/errors"
)

type receiptContent struct {
	LocationName string
	CustomerName string
	Date         string
	Hour         string
	ReceiptID    string
	Items        []item
	Subtotal     string
	Tips         string
	Total        string
}

type item struct {
	Name     string
	Quantity int64
	Price    string
}

func (s *OrderingService) GenerateOrderReceipt(ctx context.Context, orderID ID) (string, error) {
	const op = errors.Op("core/OrderingService.GenerateOrderReceipt")

	order, err := s.OrderStorage.Get(ctx, orderID)
	if err != nil {
		return "", errors.E(op, err)
	}
	location, err := s.LocationStorage.Get(ctx, order.LocationID)
	if err != nil {
		return "", errors.E(op, err)
	}
	customer, err := s.CustomerStorage.Get(ctx, order.CustomerID)
	if err != nil {
		customer.Name = "Guest"
		log.Printf("generate receipt: %v", err)
	}

	htmlFilename, err := compileReceiptHtml(location, order, customer)
	if err != nil {
		return "", errors.E(op, errors.KindUnexpected, errors.Errorf("compiling receipt: %v", err))
	}
	defer os.Remove(htmlFilename)

	pdfFilename, err := htmlToPdf(htmlFilename)
	if err != nil {
		return "", errors.E(op, errors.KindUnexpected, errors.Errorf("html to pdf: %v", err))
	}
	defer os.Remove(pdfFilename)

	url, err := s.Uploader.Upload(ctx, pdfFilename)
	if err != nil {
		return "", errors.E(op, errors.KindUnexpected, errors.Errorf("uploading receipt: %v", err))
	}

	return url, nil
}

func htmlToPdf(filename string) (string, error) {
	pdfFilename := fmt.Sprint(time.Now().Unix())
	cmd := exec.Command("wkhtmltopdf", "--page-width", "200px", "--page-height", "600px", filename, pdfFilename)
	if err := cmd.Run(); err != nil {
		return "", err
	}
	return pdfFilename, nil
}

func compileReceiptHtml(location Location, order Order, customer Customer) (string, error) {
	t, err := template.ParseFiles("core/receipt/order.html")
	if err != nil {
		return "", err
	}

	id := time.Now().UnixNano()
	filename := fmt.Sprintf("%v.html", id)

	f, err := os.Create(filename)
	if err != nil {
		return "", err
	}

	items := make([]item, len(order.ItemVariations))
	for i, v := range order.ItemVariations {
		items[i] = item{
			Name:     strings.ToUpper(v.Name),
			Quantity: v.Quantity,
			Price:    centsToString(v.TotalAmount.Value),
		}
	}

	now := time.Now()
	receiptID := string(order.ID[len(order.ID)-7:])
	receipt := receiptContent{
		LocationName: location.Name,
		CustomerName: customer.Name,
		ReceiptID:    receiptID,
		Date:         now.Format("January 2, 2006"),
		Hour:         now.Format("15:04:02 AM"),
		Items:        items,
		Subtotal:     centsToString(order.TotalAmount.Value - order.TotalTipAmount.Value),
		Tips:         centsToString(order.TotalTipAmount.Value),
		Total:        centsToString(order.TotalAmount.Value),
	}

	err = t.Execute(f, receipt)
	if err != nil {
		return "", err
	}

	return filename, nil
}

func centsToString(c int64) string {
	i := c / 100
	return fmt.Sprintf("%d.%d", i, c-i*100)
}
