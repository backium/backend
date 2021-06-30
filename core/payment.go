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
	ID      string      `bson:"_id"`
	OrderID string      `bson:"order_id"`
	Type    PaymentType `bson:"type"`
	// The Payment amount without tips
	Amount     Money  `bson:"amount"`
	TipAmount  Money  `bson:"tip_amount"`
	LocationID string `bson:"location_id"`
	MerchantID string `bson:"merchant_id"`
	CreatedAt  int64  `bson:"created_at"`
	UpdatedAt  int64  `bson:"updated_at"`
}

func NewPayment(ptype PaymentType, orderID, merchantID, locationID string) Payment {
	return Payment{
		ID:         NewID("payment"),
		OrderID:    orderID,
		Type:       ptype,
		LocationID: locationID,
		MerchantID: merchantID,
	}
}

type PaymentFilter struct {
	Limit       int64
	Offset      int64
	IDs         []string
	OrderIDs    []string
	LocationIDs []string
	MerchantID  string
}

type PaymentStorage interface {
	Put(context.Context, Payment) error
	Get(context.Context, string) (Payment, error)
	List(context.Context, PaymentFilter) ([]Payment, error)
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
