package controller

import (
	"context"
	"errors"

	"github.com/backium/backend/entity"
)

const (
	maxReturnedCustomers = 50
)

type RetrieveCustomerRequest struct {
	ID         string
	MerchantID string
}

type ListAllCustomersRequest struct {
	Limit      *int64
	Offset     *int64
	MerchantID string
}

type SearchCustomersFilter struct {
	Limit      int64
	Offset     int64
	MerchantID string
	IDs        []string
}

type CustomerRepository interface {
	Create(context.Context, entity.Customer) (entity.Customer, error)
	Update(context.Context, entity.Customer) (entity.Customer, error)
	Retrieve(context.Context, string) (entity.Customer, error)
	ListAll(context.Context, string) ([]entity.Customer, error)
	Search(context.Context, SearchCustomersFilter) ([]entity.Customer, error)
	Delete(context.Context, string) (entity.Customer, error)
}

type Customer struct {
	Repository CustomerRepository
}

func (c *Customer) Create(ctx context.Context, cus entity.Customer) (entity.Customer, error) {
	return c.Repository.Create(ctx, cus)
}

func (c *Customer) Update(ctx context.Context, cus entity.Customer) (entity.Customer, error) {
	cuss, err := c.Repository.Search(ctx, SearchCustomersFilter{
		IDs:        []string{cus.ID},
		MerchantID: cus.MerchantID,
		Limit:      1,
	})
	if err != nil {
		return entity.Customer{}, err
	}
	if len(cuss) == 0 {
		return entity.Customer{}, errors.New("customer not found")
	}
	return c.Repository.Update(ctx, cus)
}

func (c *Customer) Retrieve(ctx context.Context, req RetrieveCustomerRequest) (entity.Customer, error) {
	cus, err := c.Repository.Retrieve(ctx, req.ID)
	if err != nil {
		return entity.Customer{}, err
	}

	if cus.MerchantID != req.MerchantID {
		return entity.Customer{}, errors.New("customer not found")
	}
	return cus, nil
}

func (c *Customer) ListAll(ctx context.Context, req ListAllCustomersRequest) ([]entity.Customer, error) {
	limit := int64(maxReturnedCustomers)
	offset := int64(0)
	if req.Limit != nil {
		limit = *req.Limit
	}
	if req.Offset != nil {
		offset = *req.Offset
	}
	return c.Repository.Search(ctx, SearchCustomersFilter{
		MerchantID: req.MerchantID,
		Limit:      limit,
		Offset:     offset,
	})
}
