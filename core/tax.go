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
	ID          string   `bson:"_id"`
	Name        string   `bson:"name,omitempty"`
	Percentage  float64  `bson:"percentage"`
	LocationIDs []string `bson:"location_ids"`
	MerchantID  string   `bson:"merchant_id,omitempty"`
	CreatedAt   int64    `bson:"created_at"`
	UpdatedAt   int64    `bson:"updated_at"`
	Status      Status   `bson:"status,omitempty"`
}

func NewTax() Tax {
	return Tax{
		ID:          generateID("tax"),
		LocationIDs: []string{},
		Status:      StatusActive,
	}
}

type TaxStorage interface {
	Put(context.Context, Tax) error
	PutBatch(context.Context, []Tax) error
	Get(context.Context, string) (Tax, error)
	List(context.Context, TaxFilter) ([]Tax, error)
}

func (svc *CatalogService) PutTax(ctx context.Context, t Tax) (Tax, error) {
	const op = errors.Op("controller.Tax.Create")
	if err := svc.TaxStorage.Put(ctx, t); err != nil {
		return Tax{}, err
	}
	t, err := svc.TaxStorage.Get(ctx, t.ID)
	if err != nil {
		return Tax{}, err
	}
	return t, nil
}

func (svc *CatalogService) PutTaxes(ctx context.Context, tt []Tax) ([]Tax, error) {
	const op = errors.Op("controller.Tax.Create")
	if err := svc.TaxStorage.PutBatch(ctx, tt); err != nil {
		return nil, err
	}
	ids := make([]string, len(tt))
	for i, t := range tt {
		ids[i] = t.ID
	}
	tt, err := svc.TaxStorage.List(ctx, TaxFilter{
		Limit: int64(len(tt)),
		IDs:   ids,
	})
	if err != nil {
		return nil, err
	}
	return tt, nil
}

func (svc *CatalogService) GetTax(ctx context.Context, req TaxRetrieveRequest) (Tax, error) {
	const op = errors.Op("controller.Tax.Retrieve")
	it, err := svc.TaxStorage.Get(ctx, req.ID)
	if err != nil {
		return Tax{}, errors.E(op, err)
	}
	if it.MerchantID != req.MerchantID {
		return Tax{}, errors.E(op, errors.KindNotFound, "trying to retrieve an external tax")
	}
	return it, nil
}

func (svc *CatalogService) ListTax(ctx context.Context, req TaxListRequest) ([]Tax, error) {
	const op = errors.Op("controller.Tax.ListAll")
	limit := int64(maxReturnedTaxes)
	offset := int64(0)
	if req.Limit != nil {
		limit = *req.Limit
	}
	if req.Offset != nil {
		offset = *req.Offset
	}

	its, err := svc.TaxStorage.List(ctx, TaxFilter{
		MerchantID: req.MerchantID,
		Limit:      limit,
		Offset:     offset,
	})
	if err != nil {
		return nil, errors.E(op, err)
	}
	return its, nil
}

func (svc *CatalogService) DeleteTax(ctx context.Context, req TaxDeleteRequest) (Tax, error) {
	const op = errors.Op("controller.Tax.Delete")
	tax, err := svc.TaxStorage.Get(ctx, req.ID)
	if err != nil {
		return Tax{}, errors.E(op, err)
	}

	if tax.MerchantID != req.MerchantID {
		return Tax{}, errors.E(op, errors.KindNotFound, "trying to retrieve an external tax")
	}

	tax.Status = StatusShadowDeleted
	if err := svc.TaxStorage.Put(ctx, tax); err != nil {
		return Tax{}, errors.E(op, err)
	}
	resp, err := svc.TaxStorage.Get(ctx, req.ID)
	if err != nil {
		return Tax{}, errors.E(op, err)
	}
	return resp, nil
}

type TaxRetrieveRequest struct {
	ID         string
	MerchantID string
}

type TaxDeleteRequest struct {
	ID         string
	MerchantID string
}

type TaxListRequest struct {
	Limit       *int64
	Offset      *int64
	LocationIDs []string
	MerchantID  string
}

type TaxFilter struct {
	Limit       int64
	Offset      int64
	LocationIDs []string
	MerchantID  string
	IDs         []string
}
