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
	ID         ID       `bson:"_id"`
	Name       string   `bson:"name"`
	Email      string   `bson:"email"`
	Phone      string   `bson:"phone"`
	Address    *Address `bson:"address"`
	Image      string   `bson:"image"`
	MerchantID ID       `bson:"merchant_id"`
	CreatedAt  int64    `bson:"created_at"`
	UpdatedAt  int64    `bson:"updated_at"`
	Status     Status   `bson:"status"`
}

// Creates a Customer with default values
func NewCustomer(name, email string, merchantID ID) Customer {
	return Customer{
		ID:         NewID("cust"),
		Name:       name,
		Email:      email,
		Status:     StatusActive,
		MerchantID: merchantID,
	}
}

type CustomerStorage interface {
	Put(context.Context, Customer) error
	PutBatch(context.Context, []Customer) error
	Get(context.Context, ID) (Customer, error)
	List(context.Context, CustomerFilter) ([]Customer, int64, error)
}

type CustomerService struct {
	CustomerStorage CustomerStorage
}

func (svc *CustomerService) PutCustomer(ctx context.Context, customer Customer) (Customer, error) {
	const op = errors.Op("controller.Customer.Create")

	if err := svc.CustomerStorage.Put(ctx, customer); err != nil {
		return Customer{}, err
	}

	customer, err := svc.CustomerStorage.Get(ctx, customer.ID)
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

	ids := make([]ID, len(customers))
	for i, t := range customers {
		ids[i] = t.ID
	}
	customers, _, err := svc.CustomerStorage.List(ctx, CustomerFilter{
		Limit: int64(len(customers)),
		IDs:   ids,
	})
	if err != nil {
		return nil, err
	}

	return customers, nil
}

func (svc *CustomerService) GetCustomer(ctx context.Context, id ID) (Customer, error) {
	const op = errors.Op("core/CustomerService.GetCustomer")

	customer, err := svc.CustomerStorage.Get(ctx, id)
	if err != nil {
		return Customer{}, errors.E(op, err)
	}

	return customer, nil
}

func (svc *CustomerService) ListCustomer(ctx context.Context, f CustomerFilter) ([]Customer, int64, error) {
	const op = errors.Op("core/CustomerService.ListCustomer")

	customers, count, err := svc.CustomerStorage.List(ctx, CustomerFilter{
		MerchantID: f.MerchantID,
		Limit:      f.Limit,
		Offset:     f.Offset,
	})
	if err != nil {
		return nil, 0, errors.E(op, err)
	}

	return customers, count, nil
}

func (svc *CustomerService) DeleteCustomer(ctx context.Context, id ID) (Customer, error) {
	const op = errors.Op("core/CustomerService.DeleteCustomer")

	customer, err := svc.CustomerStorage.Get(ctx, id)
	if err != nil {
		return Customer{}, errors.E(op, err)
	}

	customer.Status = StatusShadowDeleted
	if err := svc.CustomerStorage.Put(ctx, customer); err != nil {
		return Customer{}, errors.E(op, err)
	}

	customer, err = svc.CustomerStorage.Get(ctx, id)
	if err != nil {
		return Customer{}, errors.E(op, err)
	}

	return customer, nil
}

type CustomerFilter struct {
	Limit      int64
	Offset     int64
	MerchantID ID
	IDs        []ID
}
