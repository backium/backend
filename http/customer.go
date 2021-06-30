package http

import (
	"net/http"

	"github.com/backium/backend/core"
	"github.com/backium/backend/errors"
	"github.com/labstack/echo/v4"
)

const (
	CustomerListDefaultSize = 10
	CustomerListMaxSize     = 50
)

func (h *Handler) HandleCreateCustomer(c echo.Context) error {
	const op = errors.Op("http/Handler.HandleCreateCustomer")

	type request struct {
		Name    string   `json:"name" validate:"required"`
		Email   string   `json:"email" validate:"required,email"`
		Phone   string   `json:"phone"`
		Image   string   `json:"image"`
		Address *Address `json:"address"`
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

	customer := core.NewCustomer(req.Name, req.Email, merchant.ID)
	customer.Phone = req.Phone
	customer.Image = req.Image
	if req.Address != nil {
		customer.Address = &core.Address{
			Line1:      req.Address.Line1,
			Line2:      req.Address.Line2,
			District:   req.Address.District,
			Province:   req.Address.Province,
			Department: req.Address.Department,
		}
	}

	customer, err := h.CustomerService.PutCustomer(ctx, customer)
	if err != nil {
		return errors.E(op, err)
	}

	return c.JSON(http.StatusOK, NewCustomer(customer))
}

func (h *Handler) HandleUpdateCustomer(c echo.Context) error {
	const op = errors.Op("handler.Customer.Update")

	type request struct {
		ID      string   `param:"id"`
		Name    *string  `json:"name" validate:"omitempty,min=1"`
		Email   *string  `json:"email" validate:"omitempty,email"`
		Phone   *string  `json:"phone"`
		Image   *string  `json:"image"`
		Address *Address `json:"address"`
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

	customer, err := h.CustomerService.GetCustomer(ctx, req.ID, merchant.ID)
	if req.Address != nil {
		customer.Address = &core.Address{
			Line1:      req.Address.Line1,
			Line2:      req.Address.Line2,
			Department: req.Address.Department,
			District:   req.Address.District,
			Province:   req.Address.Province,
		}
	}
	if req.Name != nil {
		customer.Name = *req.Name
	}
	if req.Email != nil {
		customer.Email = *req.Email
	}
	if req.Phone != nil {
		customer.Phone = *req.Phone
	}
	if req.Image != nil {
		customer.Image = *req.Image
	}

	customer, err = h.CustomerService.PutCustomer(ctx, customer)
	if err != nil {
		return errors.E(op, err)
	}

	return c.JSON(http.StatusOK, NewCustomer(customer))
}

func (h *Handler) HandleRetrieveCustomer(c echo.Context) error {
	const op = errors.Op("handler.Customer.Retrieve")

	ctx := c.Request().Context()

	merchant := core.MerchantFromContext(ctx)
	if merchant == nil {
		return errors.E(op, errors.KindUnexpected, "invalid echo.Context")
	}

	customer, err := h.CustomerService.GetCustomer(ctx, c.Param("id"), merchant.ID)
	if err != nil {
		return errors.E(op, err)
	}

	return c.JSON(http.StatusOK, NewCustomer(customer))
}

func (h *Handler) HandleListCustomers(c echo.Context) error {
	const op = errors.Op("handler.Customer.ListAll")

	type request struct {
		Limit  int64 `query:"limit" validate:"gte=0"`
		Offset int64 `query:"offset" validate:"gte=0"`
	}

	type response struct {
		Customers []Customer `json:"customers"`
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

	var limit, offset int64 = CustomerListDefaultSize, req.Offset
	if req.Limit <= CustomerListMaxSize {
		limit = req.Limit
	} else {
		limit = CustomerListMaxSize
	}

	customers, err := h.CustomerService.ListCustomer(ctx, core.CustomerFilter{
		Limit:      limit,
		Offset:     offset,
		MerchantID: merchant.ID,
	})
	if err != nil {
		return errors.E(op, err)
	}

	resp := response{Customers: make([]Customer, len(customers))}
	for i, customer := range customers {
		resp.Customers[i] = NewCustomer(customer)
	}

	return c.JSON(http.StatusOK, resp)
}

func (h *Handler) HandleDeleteCustomer(c echo.Context) error {
	const op = errors.Op("http/Handle.Handle.DeleteCustomer")

	ctx := c.Request().Context()

	merchant := core.MerchantFromContext(ctx)
	if merchant == nil {
		return errors.E(op, errors.KindUnexpected, "invalid echo.Context")
	}

	customer, err := h.CustomerService.DeleteCustomer(ctx, c.Param("id"), merchant.ID)
	if err != nil {
		return errors.E(op, err)
	}

	return c.JSON(http.StatusOK, NewCustomer(customer))
}

type Customer struct {
	ID         string      `json:"id"`
	Name       string      `json:"name"`
	Email      string      `json:"email"`
	Phone      string      `json:"phone"`
	Address    *Address    `json:"address,omitempty"`
	Image      string      `json:"image,omitempty"`
	MerchantID string      `json:"merchant_id"`
	CreatedAt  int64       `json:"created_at"`
	UpdatedAt  int64       `json:"updated_at"`
	Status     core.Status `json:"status"`
}

type Address struct {
	Line1      string `json:"line1,omitempty"`
	Line2      string `json:"line2,omitempty"`
	District   string `json:"district,omitempty"`
	Province   string `json:"province,omitempty"`
	Department string `json:"department,omitempty"`
}

func NewCustomer(customer core.Customer) Customer {
	c := Customer{
		ID:         customer.ID,
		Name:       customer.Name,
		Email:      customer.Email,
		Phone:      customer.Phone,
		MerchantID: customer.MerchantID,
		Image:      customer.Image,
		CreatedAt:  customer.CreatedAt,
		UpdatedAt:  customer.UpdatedAt,
		Status:     customer.Status,
	}
	if customer.Address != nil {
		c.Address = &Address{
			Line1:      customer.Address.Line1,
			Line2:      customer.Address.Line2,
			District:   customer.Address.District,
			Province:   customer.Address.Province,
			Department: customer.Address.Department,
		}
	}
	return c
}
