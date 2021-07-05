package core

import (
	"context"

	"github.com/backium/backend/errors"
)

type TaxScope string

const (
	TaxScopeOrder TaxScope = "order"
	TaxScopeItem  TaxScope = "item"
)

type Tax struct {
	ID           ID      `bson:"_id"`
	Name         string  `bson:"name,omitempty"`
	Percentage   float64 `bson:"percentage"`
	LocationIDs  []ID    `bson:"location_ids"`
	MerchantID   ID      `bson:"merchant_id,omitempty"`
	EnabledInPOS bool    `bson:"enabled_in_pos"`
	CreatedAt    int64   `bson:"created_at"`
	UpdatedAt    int64   `bson:"updated_at"`
	Status       Status  `bson:"status,omitempty"`
}

func NewTax(name string, merchantID ID) Tax {
	return Tax{
		ID:          NewID("tax"),
		Name:        name,
		LocationIDs: []ID{},
		Status:      StatusActive,
		MerchantID:  merchantID,
	}
}

type TaxStorage interface {
	Put(context.Context, Tax) error
	PutBatch(context.Context, []Tax) error
	Get(context.Context, ID) (Tax, error)
	List(context.Context, TaxQuery) ([]Tax, error)
}

func (svc *CatalogService) PutTax(ctx context.Context, tax Tax) (Tax, error) {
	const op = errors.Op("core/CatalogService.PutTax")

	if err := svc.TaxStorage.Put(ctx, tax); err != nil {
		return Tax{}, err
	}

	tax, err := svc.TaxStorage.Get(ctx, tax.ID)
	if err != nil {
		return Tax{}, err
	}

	return tax, nil
}

func (svc *CatalogService) PutTaxes(ctx context.Context, taxes []Tax) ([]Tax, error) {
	const op = errors.Op("core/CatalogService.PutTaxes")

	if err := svc.TaxStorage.PutBatch(ctx, taxes); err != nil {
		return nil, err
	}

	ids := make([]ID, len(taxes))
	for i, t := range taxes {
		ids[i] = t.ID
	}
	taxes, err := svc.TaxStorage.List(ctx, TaxQuery{
		Filter: TaxFilter{IDs: ids},
	})
	if err != nil {
		return nil, err
	}

	return taxes, nil
}

func (svc *CatalogService) GetTax(ctx context.Context, id ID) (Tax, error) {
	const op = errors.Op("core/CatalogService.GetTax")

	tax, err := svc.TaxStorage.Get(ctx, id)
	if err != nil {
		return Tax{}, errors.E(op, err)
	}

	return tax, nil
}

func (svc *CatalogService) ListTax(ctx context.Context, q TaxQuery) ([]Tax, error) {
	const op = errors.Op("core/CatalogService.ListTax")

	taxes, err := svc.TaxStorage.List(ctx, q)
	if err != nil {
		return nil, errors.E(op, err)
	}

	return taxes, nil
}

func (svc *CatalogService) DeleteTax(ctx context.Context, id ID) (Tax, error) {
	const op = errors.Op("core/CatalogService.DeleteTax")

	tax, err := svc.TaxStorage.Get(ctx, id)
	if err != nil {
		return Tax{}, errors.E(op, err)
	}

	tax.Status = StatusShadowDeleted
	if err := svc.TaxStorage.Put(ctx, tax); err != nil {
		return Tax{}, errors.E(op, err)
	}

	tax, err = svc.TaxStorage.Get(ctx, id)
	if err != nil {
		return Tax{}, errors.E(op, err)
	}

	return tax, nil
}

type TaxFilter struct {
	Name        string
	IDs         []ID
	LocationIDs []ID
	MerchantID  ID
}

type TaxSort struct {
	Name SortOrder
}

type TaxQuery struct {
	Limit  int64
	Offset int64
	Filter TaxFilter
	Sort   TaxSort
}
