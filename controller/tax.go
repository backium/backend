package controller

import (
	"context"

	"github.com/backium/backend/entity"
	"github.com/backium/backend/errors"
)

const (
	maxReturnedTaxes = 50
)

type RetrieveTaxRequest struct {
	ID         string
	MerchantID string
}

type DeleteTaxRequest struct {
	ID         string
	MerchantID string
}

type ListAllTaxesRequest struct {
	Limit       *int64
	Offset      *int64
	LocationIDs []string
	MerchantID  string
}

type ListTaxesFilter struct {
	Limit       int64
	Offset      int64
	LocationIDs []string
	MerchantID  string
	IDs         []string
}

type PartialTax struct {
	Name        *string        `bson:"name,omitempty"`
	Percentage  *int           `bson:"percentage,omitempty"`
	LocationIDs *[]string      `bson:"location_ids,omitempty"`
	Status      *entity.Status `bson:"status,omitempty"`
}

type TaxRepository interface {
	Create(context.Context, entity.Tax) (string, error)
	Update(context.Context, entity.Tax) error
	UpdatePartial(context.Context, string, PartialTax) error
	Retrieve(context.Context, string) (entity.Tax, error)
	List(context.Context, ListTaxesFilter) ([]entity.Tax, error)
}

type Tax struct {
	Repository TaxRepository
}

func (c *Tax) Create(ctx context.Context, it entity.Tax) (entity.Tax, error) {
	const op = errors.Op("controller.Tax.Create")
	id, err := c.Repository.Create(ctx, it)
	if err != nil {
		return entity.Tax{}, err
	}
	it, err = c.Repository.Retrieve(ctx, id)
	if err != nil {
		return entity.Tax{}, err
	}
	return it, nil
}

func (c *Tax) Update(ctx context.Context, id string, it PartialTax) (entity.Tax, error) {
	const op = errors.Op("controller.Tax.Update")
	if err := c.Repository.UpdatePartial(ctx, id, it); err != nil {
		return entity.Tax{}, errors.E(op, err)
	}
	uit, err := c.Repository.Retrieve(ctx, id)
	if err != nil {
		return entity.Tax{}, err
	}
	return uit, nil
}

func (c *Tax) Retrieve(ctx context.Context, req RetrieveTaxRequest) (entity.Tax, error) {
	const op = errors.Op("controller.Tax.Retrieve")
	it, err := c.Repository.Retrieve(ctx, req.ID)
	if err != nil {
		return entity.Tax{}, errors.E(op, err)
	}
	if it.MerchantID != req.MerchantID {
		return entity.Tax{}, errors.E(op, errors.KindNotFound, "trying to retrieve an external tax")
	}
	return it, nil
}

func (c *Tax) ListAll(ctx context.Context, req ListAllTaxesRequest) ([]entity.Tax, error) {
	const op = errors.Op("controller.Tax.ListAll")
	limit := int64(maxReturnedTaxes)
	offset := int64(0)
	if req.Limit != nil {
		limit = *req.Limit
	}
	if req.Offset != nil {
		offset = *req.Offset
	}

	its, err := c.Repository.List(ctx, ListTaxesFilter{
		MerchantID: req.MerchantID,
		Limit:      limit,
		Offset:     offset,
	})
	if err != nil {
		return nil, errors.E(op, err)
	}
	return its, nil
}

func (c *Tax) Delete(ctx context.Context, req DeleteTaxRequest) (entity.Tax, error) {
	const op = errors.Op("controller.Tax.Delete")
	it, err := c.Repository.Retrieve(ctx, req.ID)
	if err != nil {
		return entity.Tax{}, errors.E(op, err)
	}

	if it.MerchantID != req.MerchantID {
		return entity.Tax{}, errors.E(op, errors.KindNotFound, "trying to retrieve an external tax")
	}

	status := entity.StatusShadowDeleted
	update := PartialTax{Status: &status}
	if err := c.Repository.UpdatePartial(ctx, req.ID, update); err != nil {
		return entity.Tax{}, errors.E(op, err)
	}
	dit, err := c.Repository.Retrieve(ctx, req.ID)
	if err != nil {
		return entity.Tax{}, errors.E(op, err)
	}
	return dit, nil
}
