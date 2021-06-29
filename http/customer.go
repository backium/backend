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

func (h *Handler) CreateCustomer(c echo.Context) error {
	const op = errors.Op("handler.Customer.Create")
	ac, ok := c.(*AuthContext)
	if !ok {
		return errors.E(op, errors.KindUnexpected, "invalid echo.Context")
	}
	req := CustomerCreateRequest{}
	if err := bindAndValidate(c, &req); err != nil {
		return err
	}
	customer := core.NewCustomer(ac.MerchantID)
	if req.Address != nil {
		customer.Address = &core.Address{
			Line1:      req.Address.Line1,
			Line2:      req.Address.Line2,
			District:   req.Address.District,
			Province:   req.Address.Province,
			Department: req.Address.Department,
		}
	}
	customer.Name = req.Name
	customer.Email = req.Email
	customer.Phone = req.Phone
	customer.Image = req.Image

	ctx := c.Request().Context()
	customer, err := h.CustomerService.PutCustomer(ctx, customer)
	if err != nil {
		return errors.E(op, err)
	}
	return c.JSON(http.StatusOK, NewCustomer(customer))
}

func (h *Handler) UpdateCustomer(c echo.Context) error {
	const op = errors.Op("handler.Customer.Update")
	ac, ok := c.(*AuthContext)
	if !ok {
		return errors.E(op, errors.KindUnexpected, "invalid echo.Context")
	}
	req := CustomerUpdateRequest{}
	if err := bindAndValidate(c, &req); err != nil {
		return err
	}
	ctx := c.Request().Context()
	customer, err := h.CustomerService.GetCustomer(ctx, c.Param("id"), ac.MerchantID)
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

func (h *Handler) RetrieveCustomer(c echo.Context) error {
	const op = errors.Op("handler.Customer.Retrieve")
	ac, ok := c.(*AuthContext)
	if !ok {
		return errors.E(op, errors.KindUnexpected, "invalid echo.Context")
	}
	ctx := c.Request().Context()
	customer, err := h.CustomerService.GetCustomer(ctx, c.Param("id"), ac.MerchantID)
	if err != nil {
		return errors.E(op, err)
	}
	return c.JSON(http.StatusOK, NewCustomer(customer))
}

func (h *Handler) ListCustomers(c echo.Context) error {
	const op = errors.Op("handler.Customer.ListAll")
	ac, ok := c.(*AuthContext)
	if !ok {
		return errors.E(op, errors.KindUnexpected, "invalid echo.Context")
	}
	req := CustomerListRequest{}
	if err := bindAndValidate(c, &req); err != nil {
		return err
	}

	var limit, offset int64 = CustomerListDefaultSize, 0
	if req.Limit <= CustomerListMaxSize {
		limit = req.Limit
	}
	if req.Limit > CustomerListMaxSize {
		limit = CustomerListMaxSize
	}
	if req.Offset != 0 {
		offset = req.Offset
	}

	ctx := c.Request().Context()
	customers, err := h.CustomerService.ListCustomer(ctx, core.CustomerFilter{
		Limit:      limit,
		Offset:     offset,
		MerchantID: ac.MerchantID,
	})
	if err != nil {
		return errors.E(op, err)
	}
	resp := CustomerListResponse{Customers: make([]Customer, len(customers))}
	for i, customer := range customers {
		resp.Customers[i] = NewCustomer(customer)
	}
	return c.JSON(http.StatusOK, resp)
}

func (h *Handler) DeleteCustomer(c echo.Context) error {
	const op = errors.Op("handler.Customer.Delete")
	ac, ok := c.(*AuthContext)
	if !ok {
		return errors.E(op, errors.KindUnexpected, "invalid echo.Context")
	}
	ctx := c.Request().Context()
	customer, err := h.CustomerService.DeleteCustomer(ctx, c.Param("id"), ac.MerchantID)
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
	resp := Customer{
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
		resp.Address = &Address{
			Line1:      customer.Address.Line1,
			Line2:      customer.Address.Line2,
			District:   customer.Address.District,
			Province:   customer.Address.Province,
			Department: customer.Address.Department,
		}
	}
	return resp
}

type CustomerCreateRequest struct {
	Name    string   `json:"name" validate:"required"`
	Email   string   `json:"email" validate:"required,email"`
	Phone   string   `json:"phone"`
	Image   string   `json:"image"`
	Address *Address `json:"address"`
}

type CustomerUpdateRequest struct {
	ID      string   `param:"id"`
	Name    *string  `json:"name" validate:"omitempty,min=1"`
	Email   *string  `json:"email" validate:"omitempty,email"`
	Phone   *string  `json:"phone"`
	Image   *string  `json:"image"`
	Address *Address `json:"address"`
}

type CustomerListRequest struct {
	Limit  int64 `query:"limit" validate:"gte=0"`
	Offset int64 `query:"offset" validate:"gte=0"`
}

type CustomerListResponse struct {
	Customers []Customer `json:"customers"`
}
