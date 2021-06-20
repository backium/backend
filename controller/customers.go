package controller

import (
	"context"

	"github.com/backium/backend/base"
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

type PartialCustomer struct {
	Name    *string       `bson:"name,omitempty"`
	Email   *string       `bson:"email,omitempty"`
	Phone   *string       `bson:"phone,omitempty"`
	Address *base.Address `bson:"address,omitempty"`
	Status  *base.Status  `bson:"status,omitempty"`
}

type CustomerRepository interface {
	Create(context.Context, entity.Customer) (string, error)
	Update(context.Context, entity.Customer) error
	UpdatePartial(context.Context, string, PartialCustomer) error
	Retrieve(context.Context, string) (entity.Customer, error)
	List(context.Context, ListCustomersFilter) ([]entity.Customer, error)
}

type Customer struct {
	Repository CustomerRepository
}

func (c *Customer) Create(ctx context.Context, cus entity.Customer) (entity.Customer, error) {
	const op = errors.Op("controller.Customer.Create")
	id, err := c.Repository.Create(ctx, cus)
	if err != nil {
		return entity.Customer{}, err
	}
	ccus, err := c.Repository.Retrieve(ctx, id)
	if err != nil {
		return entity.Customer{}, err
	}
	return ccus, nil
}

func (c *Customer) Update(ctx context.Context, id string, cus PartialCustomer) (entity.Customer, error) {
	const op = errors.Op("controller.Customer.Update")
	if err := c.Repository.UpdatePartial(ctx, id, cus); err != nil {
		return entity.Customer{}, errors.E(op, err)
	}
	ucus, err := c.Repository.Retrieve(ctx, id)
	if err != nil {
		return entity.Customer{}, err
	}
	return ucus, nil
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
	it, err := c.Repository.Retrieve(ctx, req.ID)
	if err != nil {
		return entity.Customer{}, errors.E(op, err)
	}

	if it.MerchantID != req.MerchantID {
		return entity.Customer{}, errors.E(op, errors.KindNotFound, "trying to retrieve an external customer")
	}

	status := base.StatusShadowDeleted
	update := PartialCustomer{Status: &status}
	if err := c.Repository.UpdatePartial(ctx, req.ID, update); err != nil {
		return entity.Customer{}, errors.E(op, err)
	}
	dcus, err := c.Repository.Retrieve(ctx, req.ID)
	if err != nil {
		return entity.Customer{}, errors.E(op, err)
	}
	return dcus, nil
}
