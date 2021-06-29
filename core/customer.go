package core

import (
	"context"

	"github.com/backium/backend/errors"
)

const (
	maxReturnedCustomers     = 50
	defaultReturnedCustomers = 10
)

type Customer struct {
	ID         string   `bson:"_id"`
	Name       string   `bson:"name"`
	Email      string   `bson:"email"`
	Phone      string   `bson:"phone"`
	Address    *Address `bson:"address"`
	Image      string   `bson:"image"`
	MerchantID string   `bson:"merchant_id"`
	CreatedAt  int64    `bson:"created_at"`
	UpdatedAt  int64    `bson:"updated_at"`
	Status     Status   `bson:"status"`
}

// Creates a Customer with default values
func NewCustomer(merchantID string) Customer {
	return Customer{
		ID:         NewID("cust"),
		Status:     StatusActive,
		MerchantID: merchantID,
	}
}

type CustomerStorage interface {
	Put(context.Context, Customer) error
	PutBatch(context.Context, []Customer) error
	Get(context.Context, string, string) (Customer, error)
	List(context.Context, CustomerFilter) ([]Customer, error)
}

type CustomerService struct {
	CustomerStorage CustomerStorage
}

func (svc *CustomerService) PutCustomer(ctx context.Context, customer Customer) (Customer, error) {
	const op = errors.Op("controller.Customer.Create")
	if err := svc.CustomerStorage.Put(ctx, customer); err != nil {
		return Customer{}, err
	}
	customer, err := svc.CustomerStorage.Get(ctx, customer.ID, customer.MerchantID)
	if err != nil {
		return Customer{}, err
	}
	return customer, nil
}

func (svc *CustomerService) PutCustomers(ctx context.Context, customers []Customer) ([]Customer, error) {
	const op = errors.Op("core/CustomerService.PutCustomers")
	if err := svc.CustomerStorage.PutBatch(ctx, customers); err != nil {
		return nil, err
	}
	ids := make([]string, len(customers))
	for i, t := range customers {
		ids[i] = t.ID
	}
	customers, err := svc.CustomerStorage.List(ctx, CustomerFilter{
		Limit: int64(len(customers)),
		IDs:   ids,
	})
	if err != nil {
		return nil, err
	}
	return customers, nil
}

func (svc *CustomerService) GetCustomer(ctx context.Context, id, merchantID string) (Customer, error) {
	const op = errors.Op("core/CustomerService.GetCustomer")
	customer, err := svc.CustomerStorage.Get(ctx, id, merchantID)
	if err != nil {
		return Customer{}, errors.E(op, err)
	}
	return customer, nil
}

func (svc *CustomerService) ListCustomer(ctx context.Context, f CustomerFilter) ([]Customer, error) {
	const op = errors.Op("core/CustomerService.ListCustomer")
	limit, offset := int64(defaultReturnedCustomers), int64(0)
	if f.Limit != 0 && f.Limit < maxReturnedCustomers {
		limit = f.Limit
	}
	if f.Offset != 0 {
		offset = f.Offset
	}

	customers, err := svc.CustomerStorage.List(ctx, CustomerFilter{
		MerchantID: f.MerchantID,
		Limit:      limit,
		Offset:     offset,
	})
	if err != nil {
		return nil, errors.E(op, err)
	}
	return customers, nil
}

func (svc *CustomerService) DeleteCustomer(ctx context.Context, id, merchantID string) (Customer, error) {
	const op = errors.Op("core/CustomerService.DeleteCustomer")
	customer, err := svc.CustomerStorage.Get(ctx, id, merchantID)
	if err != nil {
		return Customer{}, errors.E(op, err)
	}

	customer.Status = StatusShadowDeleted
	if err := svc.CustomerStorage.Put(ctx, customer); err != nil {
		return Customer{}, errors.E(op, err)
	}
	customer, err = svc.CustomerStorage.Get(ctx, id, merchantID)
	if err != nil {
		return Customer{}, errors.E(op, err)
	}
	return customer, nil
}

type CustomerFilter struct {
	Limit      int64
	Offset     int64
	MerchantID string
	IDs        []string
}
