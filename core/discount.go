package core

import (
	"context"

	"github.com/backium/backend/errors"
	"github.com/shopspring/decimal"
)

type DiscountType string

const (
	DiscountPercentage DiscountType = "percentage"
	DiscountFixed      DiscountType = "fixed_amount"
)

type Discount struct {
	ID          ID           `bson:"_id"`
	Name        string       `bson:"name"`
	Type        DiscountType `bson:"type"`
	Amount      Money        `bson:"amount"`
	Percentage  float64      `bson:"percentage"`
	LocationIDs []ID         `bson:"location_ids"`
	MerchantID  ID           `bson:"merchant_id"`
	CreatedAt   int64        `bson:"created_at"`
	UpdatedAt   int64        `bson:"updated_at"`
	Status      Status       `bson:"status"`
}

func NewDiscount(name string, typ DiscountType, merchantID ID) Discount {
	return Discount{
		ID:          NewID("disc"),
		Name:        name,
		Type:        typ,
		LocationIDs: []ID{},
		Status:      StatusActive,
		MerchantID:  merchantID,
	}
}

// calculate computes the discount over a given amount, it uses bank's rounding
func (d *Discount) calculate(amount int64) int64 {
	if d.Type == DiscountFixed {
		return d.Amount.Value
	} else {
		ptg := decimal.NewFromFloat(d.Percentage).Div(hundred)
		total := decimal.NewFromInt(amount)
		return ptg.Mul(total).RoundBank(0).IntPart()
	}
}

type DiscountStorage interface {
	Put(context.Context, Discount) error
	PutBatch(context.Context, []Discount) error
	Get(context.Context, ID) (Discount, error)
	List(context.Context, DiscountQuery) ([]Discount, int64, error)
}

func (s *CatalogService) PutDiscount(ctx context.Context, discount Discount) (Discount, error) {
	const op = errors.Op("core/CatalogService.PutDiscount")

	if err := s.DiscountStorage.Put(ctx, discount); err != nil {
		return Discount{}, err
	}

	discount, err := s.DiscountStorage.Get(ctx, discount.ID)
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

	ids := make([]ID, len(discounts))
	for i, d := range discounts {
		ids[i] = d.ID
	}
	discounts, _, err := s.DiscountStorage.List(ctx, DiscountQuery{
		Limit:  int64(len(discounts)),
		Filter: DiscountFilter{IDs: ids},
	})
	if err != nil {
		return nil, err
	}

	return discounts, nil
}

func (s *CatalogService) GetDiscount(ctx context.Context, id ID) (Discount, error) {
	const op = errors.Op("core/CatalogService.GetDiscount")

	discount, err := s.DiscountStorage.Get(ctx, id)
	if err != nil {
		return Discount{}, errors.E(op, err)
	}

	return discount, nil
}

func (s *CatalogService) ListDiscount(ctx context.Context, q DiscountQuery) ([]Discount, int64, error) {
	const op = errors.Op("core/CatalogService.ListDiscount")

	discounts, count, err := s.DiscountStorage.List(ctx, q)
	if err != nil {
		return nil, 0, errors.E(op, err)
	}

	return discounts, count, nil
}

func (s *CatalogService) DeleteDiscount(ctx context.Context, id ID) (Discount, error) {
	const op = errors.Op("core/CatalogService.DeleteDiscount")

	discount, err := s.DiscountStorage.Get(ctx, id)
	if err != nil {
		return Discount{}, errors.E(op, err)
	}

	discount.Status = StatusShadowDeleted
	if err := s.DiscountStorage.Put(ctx, discount); err != nil {
		return Discount{}, errors.E(op, err)
	}

	discount, err = s.DiscountStorage.Get(ctx, id)
	if err != nil {
		return Discount{}, errors.E(op, err)
	}

	return discount, nil
}

type DiscountFilter struct {
	Name        string
	IDs         []ID
	LocationIDs []ID
	MerchantID  ID
}

type DiscountSort struct {
	Name SortOrder
}

type DiscountQuery struct {
	Limit  int64
	Offset int64
	Filter DiscountFilter
	Sort   DiscountSort
}
