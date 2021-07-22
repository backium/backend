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

func (h *Handler) HandleCreateEmployee(c echo.Context) error {
	const op = errors.Op("http/Handler.CreateEmployee")

	type salary struct {
		Amount *MoneyRequest `json:"amount" validate:"required"`
		Note   string        `json:"note"`
	}

	type request struct {
		FirstName   string            `json:"first_name" validate:"required"`
		LastName    string            `json:"last_name" validate:"required"`
		Email       string            `json:"email" validate:"omitempty,email"`
		Phone       string            `json:"phone" validate:"omitempty,e164"`
		Image       string            `json:"image"`
		Rate        *MoneyRequest     `json:"rate" validate:"omitempty"`
		Salary      *salary           `json:"salary" validate:"omitempty"`
		Permissions []core.Permission `json:"permissions" validate:"omitempty"`
		LocationIDs []core.ID         `json:"location_ids" validate:"omitempty,dive,required"`
	}

	ctx := c.Request().Context()

	merchant := core.MerchantFromContext(ctx)
	if merchant == nil {
		return errors.E(op, errors.KindUnexpected, "invalid echo.Context")
	}

	req := request{}
	if err := bindAndValidate(c, &req); err != nil {
		return err
	}

	employee := core.NewEmployee(req.FirstName, req.LastName, merchant.ID)
	employee.Email = req.Email
	employee.Phone = req.Phone
	employee.Image = req.Image
	employee.Permissions = req.Permissions
	if req.Rate != nil {
		rate := core.NewMoney(ptr.GetInt64(req.Rate.Value), req.Rate.Currency)
		employee.ChangeRate(rate)
	}
	if req.Salary != nil {
		amount := core.NewMoney(ptr.GetInt64(req.Salary.Amount.Value), req.Salary.Amount.Currency)
		employee.ChangeSalary(amount, req.Salary.Note)
	}
	if len(req.LocationIDs) != 0 {
		employee.LocationIDs = req.LocationIDs
	}

	employee, err := h.EmployeeService.Put(ctx, employee)
	if err != nil {
		return errors.E(op, err)
	}

	return c.JSON(http.StatusOK, NewEmployee(employee))
}

func (h *Handler) HandleUpdateEmployee(c echo.Context) error {
	const op = errors.Op("http/Handler.UpdateEmployee")

	type salary struct {
		Amount *MoneyRequest `json:"amount" validate:"required"`
		Note   string        `json:"note"`
	}

	type request struct {
		ID          core.ID            `param:"id" validate:"required"`
		FirstName   *string            `json:"first_name" validate:"omitempty,min=1"`
		LastName    *string            `json:"last_name" validate:"omitempty,min=1"`
		Email       *string            `json:"email" validate:"omitempty,email"`
		Phone       *string            `json:"phone" validate:"omitempty,e164"`
		Image       *string            `json:"image"`
		Rate        *MoneyRequest      `json:"rate" validate:"omitempty"`
		Salary      *salary            `json:"salary" validate:"omitempty"`
		Permissions *[]core.Permission `json:"permissions" validate:"omitempty"`
		LocationIDs *[]core.ID         `json:"location_ids" validate:"omitempty,dive,required"`
	}

	ctx := c.Request().Context()

	merchant := core.MerchantFromContext(ctx)
	if merchant == nil {
		return errors.E(op, errors.KindUnexpected, "invalid echo.Context")
	}

	req := request{}
	if err := bindAndValidate(c, &req); err != nil {
		return err
	}

	employee, err := h.EmployeeService.Get(ctx, req.ID)
	if err != nil {
		return errors.E(op, err)
	}
	if req.Permissions != nil {
		employee.Permissions = *req.Permissions
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
	if req.Image != nil {
		employee.Image = *req.Image
	}
	if req.Rate != nil {
		rate := core.NewMoney(ptr.GetInt64(req.Rate.Value), req.Rate.Currency)
		employee.ChangeRate(rate)
	}
	if req.Salary != nil {
		amount := core.NewMoney(ptr.GetInt64(req.Salary.Amount.Value), req.Salary.Amount.Currency)
		employee.ChangeSalary(amount, req.Salary.Note)
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

func (h *Handler) HandleRetrieveEmployee(c echo.Context) error {
	const op = errors.Op("http/Handler/RetrieveEmployee")

	type request struct {
		ID core.ID `param:"id"`
	}

	ctx := c.Request().Context()

	req := request{}
	if err := bindAndValidate(c, &req); err != nil {
		return err
	}

	merchant := core.MerchantFromContext(ctx)
	if merchant == nil {
		return errors.E(op, errors.KindUnexpected, "invalid echo.Context")
	}

	employee, err := h.EmployeeService.Get(ctx, req.ID)
	if err != nil {
		return errors.E(op, err)
	}

	return c.JSON(http.StatusOK, NewEmployee(employee))
}

func (h *Handler) HandleSearchEmployee(c echo.Context) error {
	const op = errors.Op("http/Handler.SearchEmployee")

	type filter struct {
		IDs         []core.ID `json:"ids" validate:"omitempty,dive,id"`
		LocationIDs []core.ID `json:"location_ids" validate:"omitempty,dive,id"`
		Name        string    `json:"name"`
	}

	type sort struct {
		Name core.SortOrder `json:"name"`
	}

	type request struct {
		Limit  int64  `json:"limit" validate:"gte=0"`
		Offset int64  `json:"offset" validate:"gte=0"`
		Filter filter `json:"filter"`
		Sort   sort   `json:"sort"`
	}

	type response struct {
		Employees []Employee `json:"employees"`
		Total     int64      `json:"total_count"`
	}

	ctx := c.Request().Context()

	merchant := core.MerchantFromContext(ctx)
	if merchant == nil {
		return errors.E(op, errors.KindUnexpected, "invalid echo.Context")
	}

	req := request{}
	if err := bindAndValidate(c, &req); err != nil {
		return err
	}

	var limit int64 = EmployeeListDefaultSize
	if req.Limit <= EmployeeListMaxSize {
		limit = req.Limit
	} else {
		limit = EmployeeListMaxSize
	}

	employees, count, err := h.EmployeeService.ListEmployee(ctx, core.EmployeeQuery{
		Limit:  limit,
		Offset: req.Offset,
		Filter: core.EmployeeFilter{
			Name:        req.Filter.Name,
			LocationIDs: req.Filter.LocationIDs,
			MerchantID:  merchant.ID,
		},
		Sort: core.EmployeeSort{
			Name: req.Sort.Name,
		},
	})
	if err != nil {
		return errors.E(op, err)
	}

	resp := response{
		Employees: make([]Employee, len(employees)),
		Total:     count,
	}
	for i, employee := range employees {
		resp.Employees[i] = NewEmployee(employee)
	}

	return c.JSON(http.StatusOK, resp)
}

func (h *Handler) HandleDeleteEmployee(c echo.Context) error {
	const op = errors.Op("handler.Employee.Delete")

	type request struct {
		ID core.ID `param:"id"`
	}

	ctx := c.Request().Context()

	req := request{}
	if err := bindAndValidate(c, &req); err != nil {
		return err
	}

	merchant := core.MerchantFromContext(ctx)
	if merchant == nil {
		return errors.E(op, errors.KindUnexpected, "invalid echo.Context")
	}

	employee, err := h.EmployeeService.DeleteEmployee(ctx, req.ID)
	if err != nil {
		return errors.E(op, err)
	}

	return c.JSON(http.StatusOK, NewEmployee(employee))
}

type RateEntry struct {
	Rate      MoneyRequest `json:"rate"`
	CreatedAt int64        `json:"created_at"`
}

type SalaryEntry struct {
	Salary    MoneyRequest `json:"salary"`
	Note      string       `json:"note"`
	CreatedAt int64        `json:"created_at"`
}

type Employee struct {
	ID            core.ID           `json:"id"`
	FirstName     string            `json:"first_name"`
	LastName      string            `json:"last_name"`
	Email         string            `json:"email,omitempty"`
	Phone         string            `json:"phone,omitempty"`
	Image         string            `json:"image,omitempty"`
	IsOwner       bool              `json:"is_owner"`
	Rate          *MoneyRequest     `json:"rate,omitempty"`
	RateHistory   []RateEntry       `json:"rate_history"`
	Salary        *MoneyRequest     `json:"salary,omitempty"`
	SalaryHistory []SalaryEntry     `json:"salary_history"`
	Permissions   []core.Permission `json:"permissions"`
	LocationIDs   []core.ID         `json:"location_ids"`
	MerchantID    core.ID           `json:"merchant_id"`
	CreatedAt     int64             `json:"created_at"`
	UpdatedAt     int64             `json:"updated_at"`
	Status        core.Status       `json:"status"`
}

func NewEmployee(employee core.Employee) Employee {
	rHistory := make([]RateEntry, len(employee.RateHistory))
	for i, rate := range employee.RateHistory {
		rHistory[i] = RateEntry{
			Rate: MoneyRequest{
				Value:    ptr.Int64(rate.Rate.Value),
				Currency: rate.Rate.Currency,
			},
			CreatedAt: rate.CreatedAt,
		}
	}

	sHistory := make([]SalaryEntry, len(employee.SalaryHistory))
	for i, salary := range employee.SalaryHistory {
		sHistory[i] = SalaryEntry{
			Salary: MoneyRequest{
				Value:    ptr.Int64(salary.Salary.Value),
				Currency: salary.Salary.Currency,
			},
			Note:      salary.Note,
			CreatedAt: salary.CreatedAt,
		}
	}

	var rate *MoneyRequest
	if employee.Rate != nil {
		rate = &MoneyRequest{
			Value:    &employee.Rate.Value,
			Currency: employee.Rate.Currency,
		}
	}
	var salary *MoneyRequest
	if employee.Salary != nil {
		salary = &MoneyRequest{
			Value:    &employee.Salary.Value,
			Currency: employee.Salary.Currency,
		}
	}

	return Employee{
		ID:            employee.ID,
		FirstName:     employee.FirstName,
		LastName:      employee.LastName,
		Email:         employee.Email,
		Phone:         employee.Phone,
		Image:         employee.Image,
		IsOwner:       employee.IsOwner,
		Rate:          rate,
		RateHistory:   rHistory,
		Salary:        salary,
		SalaryHistory: sHistory,
		Permissions:   employee.Permissions,
		LocationIDs:   employee.LocationIDs,
		MerchantID:    employee.MerchantID,
		CreatedAt:     employee.CreatedAt,
		UpdatedAt:     employee.UpdatedAt,
		Status:        employee.Status,
	}
}
