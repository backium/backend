package core

import (
	"context"

	"github.com/backium/backend/errors"
)

const (
	maxReturnedCustomers = 50
)

type Customer struct {
	ID         string   `bson:"_id"`
	Name       string   `bson:"name,omitempty"`
	Email      string   `bson:"email,omitempty"`
	Phone      string   `bson:"phone,omitempty"`
	Address    *Address `bson:"address,omitempty"`
	MerchantID string   `bson:"merchant_id,omitempty"`
	Status     Status   `bson:"status,omitempty"`
}

// Creates a Customer with default values
func NewCustomer() Customer {
	return Customer{
		Status: StatusActive,
	}
}

type PartialCustomer struct {
	Name    *string  `bson:"name,omitempty"`
	Email   *string  `bson:"email,omitempty"`
	Phone   *string  `bson:"phone,omitempty"`
	Address *Address `bson:"address,omitempty"`
	Status  *Status  `bson:"status,omitempty"`
}

type CustomerRepository interface {
	Create(context.Context, Customer) (string, error)
	Update(context.Context, Customer) error
	UpdatePartial(context.Context, string, PartialCustomer) error
	Retrieve(context.Context, string) (Customer, error)
	List(context.Context, ListCustomersFilter) ([]Customer, error)
}

type CustomerService struct {
	CustomerRepository CustomerRepository
}

func (svc *CustomerService) Create(ctx context.Context, cus Customer) (Customer, error) {
	const op = errors.Op("controller.Customer.Create")
	id, err := svc.CustomerRepository.Create(ctx, cus)
	if err != nil {
		return Customer{}, err
	}
	ccus, err := svc.CustomerRepository.Retrieve(ctx, id)
	if err != nil {
		return Customer{}, err
	}
	return ccus, nil
}

func (svc *CustomerService) Update(ctx context.Context, id string, cus PartialCustomer) (Customer, error) {
	const op = errors.Op("controller.Customer.Update")
	if err := svc.CustomerRepository.UpdatePartial(ctx, id, cus); err != nil {
		return Customer{}, errors.E(op, err)
	}
	ucus, err := svc.CustomerRepository.Retrieve(ctx, id)
	if err != nil {
		return Customer{}, err
	}
	return ucus, nil
}

func (svc *CustomerService) Retrieve(ctx context.Context, req RetrieveCustomerRequest) (Customer, error) {
	const op = errors.Op("controller.Customer.Retrieve")
	cus, err := svc.CustomerRepository.Retrieve(ctx, req.ID)
	if err != nil {
		return Customer{}, errors.E(op, err)
	}
	if cus.MerchantID != req.MerchantID {
		return Customer{}, errors.E(op, errors.KindNotFound, "trying to retrieve an external customer")
	}
	return cus, nil
}

func (svc *CustomerService) ListAll(ctx context.Context, req ListAllCustomersRequest) ([]Customer, error) {
	const op = errors.Op("controller.Customer.ListAll")
	limit := int64(maxReturnedCustomers)
	offset := int64(0)
	if req.Limit != nil {
		limit = *req.Limit
	}
	if req.Offset != nil {
		offset = *req.Offset
	}

	cuss, err := svc.CustomerRepository.List(ctx, ListCustomersFilter{
		MerchantID: req.MerchantID,
		Limit:      limit,
		Offset:     offset,
	})
	if err != nil {
		return nil, errors.E(op, err)
	}
	return cuss, nil
}

func (svc *CustomerService) Delete(ctx context.Context, req DeleteCustomerRequest) (Customer, error) {
	const op = errors.Op("controller.Customer.Delete")
	it, err := svc.CustomerRepository.Retrieve(ctx, req.ID)
	if err != nil {
		return Customer{}, errors.E(op, err)
	}

	if it.MerchantID != req.MerchantID {
		return Customer{}, errors.E(op, errors.KindNotFound, "trying to retrieve an external customer")
	}

	status := StatusShadowDeleted
	update := PartialCustomer{Status: &status}
	if err := svc.CustomerRepository.UpdatePartial(ctx, req.ID, update); err != nil {
		return Customer{}, errors.E(op, err)
	}
	dcus, err := svc.CustomerRepository.Retrieve(ctx, req.ID)
	if err != nil {
		return Customer{}, errors.E(op, err)
	}
	return dcus, nil
}

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
