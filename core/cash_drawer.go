package core

import (
	"context"

	"github.com/backium/backend/errors"
)

type CashDrawerOp = string

const (
	CashDrawerOpAdd    = "add"
	CashDrawerOpRemove = "remove"
)

type CashDrawerAdjustment struct {
	ID           ID           `bson:"_id"`
	CashDrawerID ID           `bson:"cash_drawer_id"`
	Amount       Money        `bson:"amount"`
	Op           CashDrawerOp `bson:"operation"`
	Note         string       `bson:"note"`
	LocationID   ID           `bson:"location_id"`
	MerchantID   ID           `bson:"merchant_id"`
	CreatedAt    int64        `bson:"created_at"`
}

func NewCashDrawerAdjustment(cashDrawerID, merchantID ID) CashDrawerAdjustment {
	return CashDrawerAdjustment{
		ID:           NewID("cashadj"),
		CashDrawerID: cashDrawerID,
		MerchantID:   merchantID,
	}
}

type CashDrawer struct {
	ID           ID    `bson:"_id"`
	Amount       Money `bson:"amount"`
	CalculatedAt int64 `bson:"calculated_at"`
	LocationID   ID    `bson:"location_id"`
	MerchantID   ID    `bson:"merchant_id"`
}

func NewCashDrawer(locationID, merchantID ID) CashDrawer {
	return CashDrawer{
		ID:         NewID("cash"),
		LocationID: locationID,
		MerchantID: merchantID,
	}
}

type CashDrawerFilter struct {
	IDs         []ID
	LocationIDs []ID
	MerchantID  ID
}

type CashDrawerQuery struct {
	Limit  int64
	Offset int64
	Filter CashDrawerFilter
}

type CashDrawerStorage interface {
	Put(context.Context, CashDrawer) error
	PutAdj(context.Context, CashDrawerAdjustment) error
	Get(context.Context, ID) (CashDrawer, error)
	List(context.Context, CashDrawerQuery) ([]CashDrawer, int64, error)
	//ListAdjustment(context.Context, CashDrawerFilter) ([]CashDrawerAdjustment, int64, error)
}

func (s *LocationService) AdjustCashDrawer(ctx context.Context, adj CashDrawerAdjustment) (CashDrawer, error) {
	const op = errors.Op("core/LocationService.ApplyCashDrawerAdjustment")

	cash, err := s.CashDrawerStorage.Get(ctx, adj.CashDrawerID)
	if err != nil {
		return CashDrawer{}, errors.E(op, err)
	}

	switch adj.Op {
	case CashDrawerOpAdd:
		cash.Amount.Value += adj.Amount.Value
	case CashDrawerOpRemove:
		cash.Amount.Value -= adj.Amount.Value
	default:
		return CashDrawer{}, errors.E(op, errors.KindValidation, "Invalid cashdrawer operation")
	}

	adj.LocationID = cash.LocationID
	if err := s.CashDrawerStorage.PutAdj(ctx, adj); err != nil {
		return CashDrawer{}, errors.E(op, err)
	}

	if err := s.CashDrawerStorage.Put(ctx, cash); err != nil {
		return CashDrawer{}, errors.E(op, err)
	}

	cash, err = s.CashDrawerStorage.Get(ctx, cash.ID)
	if err != nil {
		return CashDrawer{}, errors.E(op, err)
	}

	return cash, nil
}

func (s *LocationService) ListCashDrawer(ctx context.Context, q CashDrawerQuery) ([]CashDrawer, int64, error) {
	const op = errors.Op("core/LocationService.ListCashDrawer")

	cash, count, err := s.CashDrawerStorage.List(ctx, q)
	if err != nil {
		return nil, 0, errors.E(op, err)
	}

	return cash, count, nil
}
