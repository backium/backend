package controller

import (
	"context"

	"github.com/backium/backend/entity"
	"github.com/backium/backend/errors"
)

const (
	maxReturnedCustomers = 50
)

type RetrieveCustomerRequest struct {
	ID         string
	MerchantID string
}

type DeleteCustomerRequest struct {
	ID         string
	MerchantID string
}

type ListAllCustomersRequest struct {
	Limit      *int64
	Offset     *int64
	MerchantID string
}

type ListCustomersFilter struct {
	Limit      int64
	Offset     int64
	MerchantID string
	IDs        []string
}

type CustomerRepository interface {
	Create(context.Context, entity.Customer) (entity.Customer, error)
	Update(context.Context, entity.Customer) (entity.Customer, error)
	Retrieve(context.Context, string) (entity.Customer, error)
	List(context.Context, ListCustomersFilter) ([]entity.Customer, error)
	Delete(context.Context, string) (entity.Customer, error)
}

type Customer struct {
	Repository CustomerRepository
}

func (c *Customer) Create(ctx context.Context, cus entity.Customer) (entity.Customer, error) {
	const op = errors.Op("controller.Customer.Create")
	cus, err := c.Repository.Create(ctx, cus)
	if err != nil {
		return cus, errors.E(op, err)
	}
	return cus, nil
}

func (c *Customer) Update(ctx context.Context, cus entity.Customer) (entity.Customer, error) {
	const op = errors.Op("controller.Customer.Update")
	loc, err := c.Repository.Update(ctx, cus)
	if err != nil {
		return loc, errors.E(op, err)
	}
	return loc, nil
}

func (c *Customer) Retrieve(ctx context.Context, req RetrieveCustomerRequest) (entity.Customer, error) {
	const op = errors.Op("controller.Customer.Retrieve")
	cus, err := c.Repository.Retrieve(ctx, req.ID)
	if err != nil {
		return entity.Customer{}, errors.E(op, err)
	}

	if cus.MerchantID != req.MerchantID {
		return entity.Customer{}, errors.E(op, errors.KindNotFound, "trying to retrieve an external customer")
	}
	return cus, nil
}

func (c *Customer) ListAll(ctx context.Context, req ListAllCustomersRequest) ([]entity.Customer, error) {
	const op = errors.Op("controller.Customer.ListAll")
	limit := int64(maxReturnedCustomers)
	offset := int64(0)
	if req.Limit != nil {
		limit = *req.Limit
	}
	if req.Offset != nil {
		offset = *req.Offset
	}

	cuss, err := c.Repository.List(ctx, ListCustomersFilter{
		MerchantID: req.MerchantID,
		Limit:      limit,
		Offset:     offset,
	})
	if err != nil {
		return nil, errors.E(op, err)
	}
	return cuss, nil
}

func (c *Customer) Delete(ctx context.Context, req DeleteCustomerRequest) (entity.Customer, error) {
	const op = errors.Op("controller.Customer.Delete")
	cus, err := c.Repository.Retrieve(ctx, req.ID)
	if err != nil {
		return entity.Customer{}, errors.E(op, err)
	}

	if cus.MerchantID != req.MerchantID {
		return entity.Customer{}, errors.E(op, errors.KindNotFound, "trying to retrieve an external customer")
	}

	loc, err := c.Repository.Update(ctx, entity.Customer{
		ID:     req.ID,
		Status: entity.StatusShadowDeleted,
	})
	if err != nil {
		return loc, errors.E(op, err)
	}
	return loc, nil
}
