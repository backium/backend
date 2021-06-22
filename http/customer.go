package http

import (
	"net/http"

	"github.com/backium/backend/core"
	"github.com/backium/backend/errors"
	"github.com/labstack/echo/v4"
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
	cus := core.NewCustomer()
	if req.Address != nil {
		cus.Address = &core.Address{
			Line1:      req.Address.Line1,
			Line2:      req.Address.Line2,
			District:   req.Address.District,
			Province:   req.Address.Province,
			Department: req.Address.Department,
		}
	}
	cus.Name = req.Name
	cus.Email = req.Email
	cus.Phone = req.Phone
	cus.MerchantID = ac.MerchantID
	cus, err := h.CustomerService.Create(c.Request().Context(), cus)
	if err != nil {
		return errors.E(op, err)
	}
	return c.JSON(http.StatusOK, NewCustomer(cus))
}

func (h *Handler) UpdateCustomer(c echo.Context) error {
	const op = errors.Op("handler.Customer.Update")
	req := CustomerUpdateRequest{}
	if err := bindAndValidate(c, &req); err != nil {
		return err
	}
	cus := core.PartialCustomer{
		Name:  req.Name,
		Email: req.Email,
		Phone: req.Phone,
	}
	if req.Address != nil {
		cus.Address = &core.Address{
			Line1:      req.Address.Line1,
			Line2:      req.Address.Line2,
			Department: req.Address.Department,
			District:   req.Address.District,
			Province:   req.Address.Province,
		}
	}
	ucus, err := h.CustomerService.Update(c.Request().Context(), req.ID, cus)
	if err != nil {
		return errors.E(op, err)
	}
	return c.JSON(http.StatusOK, NewCustomer(ucus))
}

func (h *Handler) RetrieveCustomer(c echo.Context) error {
	const op = errors.Op("handler.Customer.Retrieve")
	ac, ok := c.(*AuthContext)
	if !ok {
		return errors.E(op, errors.KindUnexpected, "invalid echo.Context")
	}
	m, err := h.CustomerService.Retrieve(c.Request().Context(), core.RetrieveCustomerRequest{
		ID:         c.Param("id"),
		MerchantID: ac.MerchantID,
	})
	if err != nil {
		return errors.E(op, err)
	}
	return c.JSON(http.StatusOK, NewCustomer(m))
}

func (h *Handler) ListCustomers(c echo.Context) error {
	const op = errors.Op("handler.Customer.ListAll")
	ac, ok := c.(*AuthContext)
	if !ok {
		return errors.E(op, errors.KindUnexpected, "invalid echo.Context")
	}
	req := listAllCustomersRequest{}
	if err := bindAndValidate(c, &req); err != nil {
		return err
	}
	cuss, err := h.CustomerService.ListAll(c.Request().Context(), core.ListAllCustomersRequest{
		Limit:      req.Limit,
		Offset:     req.Offset,
		MerchantID: ac.MerchantID,
	})
	if err != nil {
		return errors.E(op, err)
	}
	res := make([]Customer, len(cuss))
	for i, cus := range cuss {
		res[i] = NewCustomer(cus)
	}
	return c.JSON(http.StatusOK, listCustomersResponse{res})
}

func (h *Handler) DeleteCustomer(c echo.Context) error {
	const op = errors.Op("handler.Customer.Delete")
	ac, ok := c.(*AuthContext)
	if !ok {
		return errors.E(op, errors.KindUnexpected, "invalid echo.Context")
	}
	cus, err := h.CustomerService.Delete(c.Request().Context(), core.DeleteCustomerRequest{
		ID:         c.Param("id"),
		MerchantID: ac.MerchantID,
	})
	if err != nil {
		return errors.E(op, err)
	}
	return c.JSON(http.StatusOK, NewCustomer(cus))
}

type Customer struct {
	ID         string      `json:"id"`
	Name       string      `json:"name"`
	Email      string      `json:"email"`
	Phone      string      `json:"phone"`
	Address    *Address    `json:"address,omitempty"`
	MerchantID string      `json:"merchant_id"`
	Status     core.Status `json:"status"`
}

type Address struct {
	Line1      string `json:"line1,omitempty"`
	Line2      string `json:"line2,omitempty"`
	District   string `json:"district,omitempty"`
	Province   string `json:"province,omitempty"`
	Department string `json:"department,omitempty"`
}

func NewCustomer(cus core.Customer) Customer {
	cusr := Customer{
		ID:         cus.ID,
		Name:       cus.Name,
		Email:      cus.Email,
		Phone:      cus.Phone,
		MerchantID: cus.MerchantID,
		Status:     cus.Status,
	}
	if cus.Address != nil {
		cusr.Address = &Address{
			Line1:      cus.Address.Line1,
			Line2:      cus.Address.Line2,
			District:   cus.Address.District,
			Province:   cus.Address.Province,
			Department: cus.Address.Department,
		}
	}
	return cusr
}

type CustomerCreateRequest struct {
	Name    string   `json:"name" validate:"required"`
	Email   string   `json:"email" validate:"required,email"`
	Phone   string   `json:"phone"`
	Address *Address `json:"address"`
}

type CustomerUpdateRequest struct {
	ID      string   `param:"id"`
	Name    *string  `json:"name" validate:"omitempty,min=1"`
	Email   *string  `json:"email" validate:"omitempty,email"`
	Phone   *string  `json:"phone"`
	Address *Address `json:"address"`
}

type listAllCustomersRequest struct {
	Limit  *int64 `query:"limit" validate:"omitempty,gte=1"`
	Offset *int64 `query:"offset"`
}

type listCustomersResponse struct {
	Customers []Customer `json:"customers"`
}
