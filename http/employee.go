package http

import (
	"net/http"

	"github.com/backium/backend/core"
	"github.com/backium/backend/errors"
	"github.com/backium/backend/ptr"
	"github.com/labstack/echo/v4"
)

const (
	EmployeeListDefaultSize = 10
	EmployeeListMaxSize     = 50
)

func (h *Handler) CreateEmployee(c echo.Context) error {
	const op = errors.Op("http/Handler.CreateEmployee")
	ac, ok := c.(*AuthContext)
	if !ok {
		return errors.E(op, errors.KindUnexpected, "invalid echo.Context")
	}
	req := EmployeeCreateRequest{}
	if err := bindAndValidate(c, &req); err != nil {
		return err
	}
	employee := core.NewEmployee(ac.MerchantID)
	employee.FirstName = req.FirstName
	employee.LastName = req.LastName
	employee.Email = req.Email
	employee.Phone = req.Phone
	if req.Rate != nil {
		rate := core.NewMoney(ptr.GetInt64(req.Rate.Value), req.Rate.Currency)
		employee.ChangeRate(rate)
	}
	if len(req.LocationIDs) != 0 {
		employee.LocationIDs = req.LocationIDs
	}

	ctx := c.Request().Context()
	employee, err := h.EmployeeService.Put(ctx, employee)
	if err != nil {
		return errors.E(op, err)
	}
	return c.JSON(http.StatusOK, NewEmployee(employee))
}

func (h *Handler) UpdateEmployee(c echo.Context) error {
	const op = errors.Op("http/Handler.UpdateEmployee")
	ac, ok := c.(*AuthContext)
	if !ok {
		return errors.E(op, errors.KindUnexpected, "invalid echo.Context")
	}
	req := EmployeeUpdateRequest{}
	if err := bindAndValidate(c, &req); err != nil {
		return err
	}
	ctx := c.Request().Context()
	employee, err := h.EmployeeService.Get(ctx, req.ID, ac.MerchantID)
	if err != nil {
		return errors.E(op, err)
	}

	if req.FirstName != nil {
		employee.FirstName = *req.FirstName
	}
	if req.LastName != nil {
		employee.LastName = *req.LastName
	}
	if req.Email != nil {
		employee.Email = *req.Email
	}
	if req.Phone != nil {
		employee.Phone = *req.Phone
	}
	if req.Rate != nil {
		rate := core.NewMoney(ptr.GetInt64(req.Rate.Value), req.Rate.Currency)
		employee.ChangeRate(rate)
	}
	if req.LocationIDs != nil {
		employee.LocationIDs = *req.LocationIDs
	}

	employee, err = h.EmployeeService.Put(ctx, employee)
	if err != nil {
		return errors.E(op, err)
	}
	return c.JSON(http.StatusOK, NewEmployee(employee))
}

func (h *Handler) RetrieveEmployee(c echo.Context) error {
	const op = errors.Op("http/Handler/RetrieveEmployee")
	ac, ok := c.(*AuthContext)
	if !ok {
		return errors.E(op, errors.KindUnexpected, "invalid echo.Context")
	}
	ctx := c.Request().Context()
	employee, err := h.EmployeeService.Get(ctx, c.Param("id"), ac.MerchantID)
	if err != nil {
		return errors.E(op, err)
	}
	return c.JSON(http.StatusOK, NewEmployee(employee))
}

func (h *Handler) SearchEmployee(c echo.Context) error {
	const op = errors.Op("http/Handler.SearchEmployee")
	ac, ok := c.(*AuthContext)
	if !ok {
		return errors.E(op, errors.KindUnexpected, "invalid echo.Context")
	}
	req := EmployeeSearchRequest{}
	if err := bindAndValidate(c, &req); err != nil {
		return err
	}

	var limit, offset int64 = EmployeeListDefaultSize, 0
	if req.Limit <= EmployeeListMaxSize {
		limit = req.Limit
	}
	if req.Limit > EmployeeListMaxSize {
		limit = EmployeeListMaxSize
	}
	if req.Offset != 0 {
		offset = req.Offset
	}

	ctx := c.Request().Context()
	employees, err := h.EmployeeService.ListEmployee(ctx, core.EmployeeFilter{
		Limit:       limit,
		Offset:      offset,
		LocationIDs: req.LocationIDs,
		MerchantID:  ac.MerchantID,
	})
	if err != nil {
		return errors.E(op, err)
	}
	resp := EmployeeSearchResponse{Employees: make([]Employee, len(employees))}
	for i, employee := range employees {
		resp.Employees[i] = NewEmployee(employee)
	}
	return c.JSON(http.StatusOK, resp)
}

func (h *Handler) DeleteEmployee(c echo.Context) error {
	const op = errors.Op("handler.Employee.Delete")
	ac, ok := c.(*AuthContext)
	if !ok {
		return errors.E(op, errors.KindUnexpected, "invalid echo.Context")
	}
	ctx := c.Request().Context()
	employee, err := h.EmployeeService.DeleteEmployee(ctx, c.Param("id"), ac.MerchantID)
	if err != nil {
		return errors.E(op, err)
	}
	return c.JSON(http.StatusOK, NewEmployee(employee))
}

type RateEntry struct {
	Rate      Money `json:"rate"`
	CreatedAt int64 `json:"created_at"`
}

type Employee struct {
	ID          string      `json:"id"`
	FirstName   string      `json:"first_name"`
	LastName    string      `json:"last_name"`
	Email       string      `json:"email,omitempty"`
	Phone       string      `json:"phone,omitempty"`
	IsOwner     bool        `json:"is_owner"`
	Rate        *Money      `json:"rate,omitempty"`
	RateHistory []RateEntry `json:"rate_history"`
	LocationIDs []string    `json:"location_ids"`
	MerchantID  string      `json:"merchant_id"`
	CreatedAt   int64       `json:"created_at"`
	UpdatedAt   int64       `json:"updated_at"`
	Status      core.Status `json:"status"`
}

func NewEmployee(employee core.Employee) Employee {
	history := make([]RateEntry, len(employee.RateHistory))
	for i, rate := range employee.RateHistory {
		history[i] = RateEntry{
			Rate: Money{
				Value:    ptr.Int64(rate.Rate.Value),
				Currency: rate.Rate.Currency,
			},
			CreatedAt: rate.CreatedAt,
		}
	}
	var rate *Money
	if employee.Rate != nil {
		rate = &Money{
			Value:    &employee.Rate.Value,
			Currency: employee.Rate.Currency,
		}
	}
	return Employee{
		ID:          employee.ID,
		FirstName:   employee.FirstName,
		LastName:    employee.LastName,
		Email:       employee.Email,
		Phone:       employee.Phone,
		IsOwner:     employee.IsOwner,
		Rate:        rate,
		RateHistory: history,
		LocationIDs: employee.LocationIDs,
		MerchantID:  employee.MerchantID,
		CreatedAt:   employee.CreatedAt,
		UpdatedAt:   employee.UpdatedAt,
		Status:      employee.Status,
	}
}

type EmployeeCreateRequest struct {
	FirstName   string   `json:"first_name" validate:"required"`
	LastName    string   `json:"last_name" validate:"required"`
	Email       string   `json:"email" validate:"omitempty,email"`
	Phone       string   `json:"phone" validate:"omitempty,e164"`
	Rate        *Money   `json:"rate" validate:"omitempty"`
	LocationIDs []string `json:"location_ids" validate:"omitempty,dive,required"`
}

type EmployeeUpdateRequest struct {
	ID          string    `param:"id" validate:"required"`
	FirstName   *string   `json:"first_name" validate:"omitempty,min=1"`
	LastName    *string   `json:"last_name" validate:"omitempty,min=1"`
	Email       *string   `json:"email" validate:"omitempty,email"`
	Phone       *string   `json:"phone" validate:"omitempty,e164"`
	Rate        *Money    `json:"rate" validate:"omitempty"`
	LocationIDs *[]string `json:"location_ids" validate:"omitempty,dive,required"`
}

type EmployeeSearchRequest struct {
	Limit       int64    `json:"limit"`
	Offset      int64    `json:"offset"`
	LocationIDs []string `json:"location_ids"`
}

type EmployeeSearchResponse struct {
	Employees []Employee `json:"employees"`
}
