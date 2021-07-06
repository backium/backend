package core

import (
	"context"

	"github.com/backium/backend/errors"
)

type PaymentType string

const (
	PaymentCash        PaymentType = "cash"
	PaymentCard        PaymentType = "card"
	PaymentBankAccount PaymentType = "bank_account"
	PaymentYape        PaymentType = "yape"
)

type Payment struct {
	ID      ID          `bson:"_id"`
	OrderID ID          `bson:"order_id"`
	Type    PaymentType `bson:"type"`
	// The Payment amount without tips
	Amount     Money `bson:"amount"`
	TipAmount  Money `bson:"tip_amount"`
	LocationID ID    `bson:"location_id"`
	MerchantID ID    `bson:"merchant_id"`
	CreatedAt  int64 `bson:"created_at"`
	UpdatedAt  int64 `bson:"updated_at"`
}

func NewPayment(ptype PaymentType, orderID, merchantID, locationID ID) Payment {
	return Payment{
		ID:         NewID("payment"),
		OrderID:    orderID,
		Type:       ptype,
		LocationID: locationID,
		MerchantID: merchantID,
	}
}

type PaymentFilter struct {
	IDs         []ID
	OrderIDs    []ID
	LocationIDs []ID
	MerchantID  ID
	CreatedAt   DateFilter
}

type PaymentSort struct {
	CreatedAt SortOrder
}

type PaymentQuery struct {
	Limit  int64
	Offset int64
	Filter PaymentFilter
	Sort   PaymentSort
}

type PaymentStorage interface {
	Put(context.Context, Payment) error
	Get(context.Context, ID) (Payment, error)
	List(context.Context, PaymentQuery) ([]Payment, int64, error)
}

type PaymentService struct {
	PaymentStorage PaymentStorage
}

func (svc *PaymentService) CreatePayment(ctx context.Context, payment Payment) (Payment, error) {
	const op = errors.Op("core/PaymentService.PutPayment")

	if err := svc.PaymentStorage.Put(ctx, payment); err != nil {
		return Payment{}, err
	}

	payment, err := svc.PaymentStorage.Get(ctx, payment.ID)
	if err != nil {
		return Payment{}, err
	}

	return payment, nil
}

func (svc *PaymentService) ListPayment(ctx context.Context, q PaymentQuery) ([]Payment, int64, error) {
	const op = errors.Op("core/PaymentService.ListPayment")

	payments, count, err := svc.PaymentStorage.List(ctx, q)
	if err != nil {
		return nil, 0, errors.E(op, err)
	}

	return payments, count, nil
}
