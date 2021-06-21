package core

import (
	"context"

	"github.com/backium/backend/errors"
)

const maxReturnedTaxes = 50

type TaxPartial struct {
	Name        *string   `bson:"name,omitempty"`
	Percentage  *int      `bson:"percentage,omitempty"`
	LocationIDs *[]string `bson:"location_ids,omitempty"`
	Status      *Status   `bson:"status,omitempty"`
}

type Tax struct {
	ID          string   `bson:"_id"`
	Name        string   `bson:"name,omitempty"`
	Percentage  int      `bson:"percentage"`
	LocationIDs []string `bson:"location_ids"`
	MerchantID  string   `bson:"merchant_id,omitempty"`
	Status      Status   `bson:"status,omitempty"`
}

func NewTax() Tax {
	return Tax{
		LocationIDs: []string{},
		Status:      StatusActive,
	}
}

type TaxRepository interface {
	Create(context.Context, Tax) (string, error)
	Update(context.Context, Tax) error
	UpdatePartial(context.Context, string, TaxPartial) error
	Retrieve(context.Context, string) (Tax, error)
	List(context.Context, TaxFilter) ([]Tax, error)
}

type TaxService struct {
	TaxRepository TaxRepository
}

func (c *TaxService) CreateTax(ctx context.Context, it Tax) (Tax, error) {
	const op = errors.Op("controller.Tax.Create")
	id, err := c.TaxRepository.Create(ctx, it)
	if err != nil {
		return Tax{}, err
	}
	it, err = c.TaxRepository.Retrieve(ctx, id)
	if err != nil {
		return Tax{}, err
	}
	return it, nil
}

func (c *TaxService) UpdateTax(ctx context.Context, id string, it TaxPartial) (Tax, error) {
	const op = errors.Op("controller.Tax.Update")
	if err := c.TaxRepository.UpdatePartial(ctx, id, it); err != nil {
		return Tax{}, errors.E(op, err)
	}
	uit, err := c.TaxRepository.Retrieve(ctx, id)
	if err != nil {
		return Tax{}, err
	}
	return uit, nil
}

func (c *TaxService) RetrieveTax(ctx context.Context, req TaxRetrieveRequest) (Tax, error) {
	const op = errors.Op("controller.Tax.Retrieve")
	it, err := c.TaxRepository.Retrieve(ctx, req.ID)
	if err != nil {
		return Tax{}, errors.E(op, err)
	}
	if it.MerchantID != req.MerchantID {
		return Tax{}, errors.E(op, errors.KindNotFound, "trying to retrieve an external tax")
	}
	return it, nil
}

func (c *TaxService) ListTax(ctx context.Context, req TaxListRequest) ([]Tax, error) {
	const op = errors.Op("controller.Tax.ListAll")
	limit := int64(maxReturnedTaxes)
	offset := int64(0)
	if req.Limit != nil {
		limit = *req.Limit
	}
	if req.Offset != nil {
		offset = *req.Offset
	}

	its, err := c.TaxRepository.List(ctx, TaxFilter{
		MerchantID: req.MerchantID,
		Limit:      limit,
		Offset:     offset,
	})
	if err != nil {
		return nil, errors.E(op, err)
	}
	return its, nil
}

func (c *TaxService) DeleteTax(ctx context.Context, req TaxDeleteRequest) (Tax, error) {
	const op = errors.Op("controller.Tax.Delete")
	it, err := c.TaxRepository.Retrieve(ctx, req.ID)
	if err != nil {
		return Tax{}, errors.E(op, err)
	}

	if it.MerchantID != req.MerchantID {
		return Tax{}, errors.E(op, errors.KindNotFound, "trying to retrieve an external tax")
	}

	status := StatusShadowDeleted
	update := TaxPartial{Status: &status}
	if err := c.TaxRepository.UpdatePartial(ctx, req.ID, update); err != nil {
		return Tax{}, errors.E(op, err)
	}
	dit, err := c.TaxRepository.Retrieve(ctx, req.ID)
	if err != nil {
		return Tax{}, errors.E(op, err)
	}
	return dit, nil
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
