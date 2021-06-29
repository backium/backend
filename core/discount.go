package core

import (
	"context"

	"github.com/backium/backend/errors"
	"github.com/shopspring/decimal"
)

type DiscountType string

const (
	DiscountTypePercentage DiscountType = "percentage"
	DiscountTypeFixed      DiscountType = "fixed_amount"
)

type Discount struct {
	ID          string       `bson:"_id"`
	Name        string       `bson:"name"`
	Type        DiscountType `bson:"type"`
	Fixed       Money        `bson:"fixed"`
	Percentage  float64      `bson:"percentage"`
	LocationIDs []string     `bson:"location_ids"`
	MerchantID  string       `bson:"merchant_id"`
	CreatedAt   int64        `bson:"created_at"`
	UpdatedAt   int64        `bson:"updated_at"`
	Status      Status       `bson:"status"`
}

func NewDiscount(merchantID string) Discount {
	return Discount{
		ID:          NewID("disc"),
		LocationIDs: []string{},
		Status:      StatusActive,
		MerchantID:  merchantID,
	}
}

// calculate computes the discount over a given amount, it uses bank's rounding
func (d *Discount) calculate(amount int64) int64 {
	if d.Type == DiscountTypeFixed {
		return d.Fixed.Amount
	} else {
		ptg := decimal.NewFromFloat(d.Percentage).Div(hundred)
		total := decimal.NewFromInt(amount)
		return ptg.Mul(total).RoundBank(0).IntPart()
	}
}

type DiscountStorage interface {
	Put(context.Context, Discount) error
	PutBatch(context.Context, []Discount) error
	Get(context.Context, string, string, []string) (Discount, error)
	List(context.Context, DiscountFilter) ([]Discount, error)
}

func (s *CatalogService) PutDiscount(ctx context.Context, discount Discount) (Discount, error) {
	const op = errors.Op("core/CatalogService.PutDiscount")
	if err := s.DiscountStorage.Put(ctx, discount); err != nil {
		return Discount{}, err
	}
	discount, err := s.DiscountStorage.Get(ctx, discount.ID, discount.MerchantID, nil)
	if err != nil {
		return Discount{}, err
	}
	return discount, nil
}

func (s *CatalogService) PutDiscounts(ctx context.Context, discounts []Discount) ([]Discount, error) {
	const op = errors.Op("core/CatalogService.PutDiscounts")
	if err := s.DiscountStorage.PutBatch(ctx, discounts); err != nil {
		return nil, err
	}
	ids := make([]string, len(discounts))
	for i, d := range discounts {
		ids[i] = d.ID
	}
	discounts, err := s.DiscountStorage.List(ctx, DiscountFilter{
		Limit: int64(len(discounts)),
		IDs:   ids,
	})
	if err != nil {
		return nil, err
	}
	return discounts, nil
}

func (s *CatalogService) GetDiscount(ctx context.Context, id, merchantID string, locationIDs []string) (Discount, error) {
	const op = errors.Op("core/CatalogService.GetDiscount")
	discount, err := s.DiscountStorage.Get(ctx, id, merchantID, locationIDs)
	if err != nil {
		return Discount{}, errors.E(op, err)
	}
	return discount, nil
}

func (s *CatalogService) ListDiscount(ctx context.Context, f DiscountFilter) ([]Discount, error) {
	const op = errors.Op("core/CatalogService.ListDiscount")
	discounts, err := s.DiscountStorage.List(ctx, DiscountFilter{
		MerchantID: f.MerchantID,
		Limit:      f.Limit,
		Offset:     f.Offset,
	})
	if err != nil {
		return nil, errors.E(op, err)
	}
	return discounts, nil
}

func (s *CatalogService) DeleteDiscount(ctx context.Context, id, merchantID string, locationIDs []string) (Discount, error) {
	const op = errors.Op("core/CatalogService.DeleteDiscount")
	discount, err := s.DiscountStorage.Get(ctx, id, merchantID, locationIDs)
	if err != nil {
		return Discount{}, errors.E(op, err)
	}

	discount.Status = StatusShadowDeleted
	if err := s.DiscountStorage.Put(ctx, discount); err != nil {
		return Discount{}, errors.E(op, err)
	}
	discount, err = s.DiscountStorage.Get(ctx, id, merchantID, locationIDs)
	if err != nil {
		return Discount{}, errors.E(op, err)
	}
	return discount, nil
}

type DiscountFilter struct {
	Limit       int64
	Offset      int64
	LocationIDs []string
	MerchantID  string
	IDs         []string
}
