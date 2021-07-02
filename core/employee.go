package core

import (
	"context"
	"time"

	"github.com/backium/backend/errors"
)

type RateEntry struct {
	Rate      Money `bson:"rate"`
	CreatedAt int64 `bson:"created_at"`
}

type Employee struct {
	ID          ID          `bson:"_id"`
	FirstName   string      `bson:"first_name"`
	LastName    string      `bson:"last_name"`
	Email       string      `bson:"email"`
	Phone       string      `bson:"phone"`
	IsOwner     bool        `bson:"is_owner"`
	Rate        *Money      `bson:"rate"`
	RateHistory []RateEntry `bson:"rate_history"`
	LocationIDs []ID        `bson:"location_ids"`
	MerchantID  ID          `bson:"merchant_id"`
	CreatedAt   int64       `bson:"created_at"`
	UpdatedAt   int64       `bson:"updated_at"`
	Status      Status      `bson:"status"`
}

func NewEmployee(firstName, lastName string, merchantID ID) Employee {
	return Employee{
		ID:          NewID("empl"),
		FirstName:   firstName,
		LastName:    lastName,
		LocationIDs: []ID{},
		MerchantID:  merchantID,
		IsOwner:     false,
		Status:      StatusActive,
	}
}

func (e *Employee) ChangeRate(rate Money) {
	e.Rate = &Money{Value: rate.Value, Currency: rate.Currency}
	e.RateHistory = append(e.RateHistory, RateEntry{
		Rate:      rate,
		CreatedAt: time.Now().Unix(),
	})
}

type EmployeeFilter struct {
	Limit       int64
	Offset      int64
	LocationIDs []ID
	MerchantID  ID
}

type EmployeeStorage interface {
	Put(context.Context, Employee) error
	Get(context.Context, ID) (Employee, error)
	List(context.Context, EmployeeFilter) ([]Employee, error)
}

type EmployeeService struct {
	EmployeeStorage EmployeeStorage
}

func (svc *EmployeeService) Put(ctx context.Context, employee Employee) (Employee, error) {
	const op = errors.Op("core/EmployeeService.Put")

	if err := svc.EmployeeStorage.Put(ctx, employee); err != nil {
		return Employee{}, errors.E(op, err)
	}

	employee, err := svc.EmployeeStorage.Get(ctx, employee.ID)
	if err != nil {
		return Employee{}, errors.E(op, err)
	}

	return employee, nil
}

func (svc *EmployeeService) Get(ctx context.Context, id ID) (Employee, error) {
	const op = errors.Op("core/EmployeeService.Get")

	employee, err := svc.EmployeeStorage.Get(ctx, id)
	if err != nil {
		return Employee{}, errors.E(op, err)
	}

	return employee, nil

}

func (svc *EmployeeService) ListEmployee(ctx context.Context, f EmployeeFilter) ([]Employee, error) {
	const op = errors.Op("core/EmployeeService.ListEmployee")

	employees, err := svc.EmployeeStorage.List(ctx, EmployeeFilter{
		LocationIDs: f.LocationIDs,
		MerchantID:  f.MerchantID,
		Limit:       f.Limit,
		Offset:      f.Offset,
	})
	if err != nil {
		return nil, errors.E(op, err)
	}

	return employees, nil
}

func (svc *EmployeeService) DeleteEmployee(ctx context.Context, id ID) (Employee, error) {
	const op = errors.Op("core/CatalogService.DeleteItem")

	employee, err := svc.EmployeeStorage.Get(ctx, id)
	if err != nil {
		return Employee{}, errors.E(op, err)
	}

	employee.Status = StatusShadowDeleted
	if err := svc.EmployeeStorage.Put(ctx, employee); err != nil {
		return Employee{}, errors.E(op, err)
	}

	employee, err = svc.EmployeeStorage.Get(ctx, id)
	if err != nil {
		return Employee{}, errors.E(op, err)
	}

	return employee, nil
}
