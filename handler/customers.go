package handler

import (
	"errors"
	"net/http"

	"github.com/backium/backend/controller"
	"github.com/backium/backend/entity"
	"github.com/labstack/echo/v4"
)

type customerResource struct {
	ID         string   `json:"id"`
	Name       string   `json:"name"`
	Email      string   `json:"email"`
	Phone      string   `json:"phone"`
	Address    *address `json:"address"`
	MerchantID string   `json:"merchant_id"`
}

type address struct {
	Line1      string `json:"line1"`
	Line2      string `json:"line2"`
	District   string `json:"district"`
	Province   string `json:"province"`
	Department string `json:"department"`
}

func customerResourceFrom(cus entity.Customer) customerResource {
	var addr *address
	if cus.Address != nil {
		addr = &address{
			Line1:      cus.Address.Line1,
			Line2:      cus.Address.Line2,
			District:   cus.Address.District,
			Province:   cus.Address.Province,
			Department: cus.Address.Department,
		}
	}
	return customerResource{
		ID:         cus.ID,
		Name:       cus.Name,
		Email:      cus.Email,
		Phone:      cus.Phone,
		Address:    addr,
		MerchantID: cus.MerchantID,
	}
}

func (cus *customerResource) customer() entity.Customer {
	var addr *entity.Address
	if cus.Address != nil {
		addr = &entity.Address{
			Line1:      cus.Address.Line1,
			Line2:      cus.Address.Line2,
			District:   cus.Address.District,
			Province:   cus.Address.Province,
			Department: cus.Address.Department,
		}
	}
	return entity.Customer{
		ID:         cus.ID,
		Name:       cus.Name,
		Email:      cus.Email,
		Phone:      cus.Phone,
		Address:    addr,
		MerchantID: cus.MerchantID,
	}
}

type createCustomerRequest struct {
	Name    string   `json:"name" validate:"required"`
	Email   string   `json:"email" validate:"required,email"`
	Phone   string   `json:"phone"`
	Address *address `json:"address"`
}

func (req *createCustomerRequest) customer() entity.Customer {
	var addr *entity.Address
	if req.Address != nil {
		addr = &entity.Address{
			Line1:      req.Address.Line1,
			Line2:      req.Address.Line2,
			District:   req.Address.District,
			Province:   req.Address.Province,
			Department: req.Address.Department,
		}
	}
	return entity.Customer{
		Name:    req.Name,
		Email:   req.Email,
		Phone:   req.Phone,
		Address: addr,
	}
}

type updateCustomerRequest struct {
	ID      string   `param:"id"`
	Name    string   `json:"name"`
	Email   string   `json:"email" validate:"omitempty,email"`
	Phone   string   `json:"phone"`
	Address *address `json:"address"`
}

func (req *updateCustomerRequest) customer() entity.Customer {
	var addr *entity.Address
	if req.Address != nil {
		addr = &entity.Address{
			Line1:      req.Address.Line1,
			Line2:      req.Address.Line2,
			District:   req.Address.District,
			Province:   req.Address.Province,
			Department: req.Address.Department,
		}
	}
	return entity.Customer{
		ID:      req.ID,
		Name:    req.Name,
		Email:   req.Email,
		Phone:   req.Phone,
		Address: addr,
	}
}

type listAllCustomersRequest struct {
	Limit  *int64 `query:"limit" validate:"gte=1"`
	Offset *int64 `query:"offset"`
}

type listCustomersResponse struct {
	Customers []customerResource `json:"customers"`
}

type Customer struct {
	Controller controller.Customer
}

func (h *Customer) Create(c echo.Context) error {
	ac, ok := c.(*AuthContext)
	if !ok {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid context")
	}
	req := createCustomerRequest{}
	if err := bindAndValidate(c, &req); err != nil {
		return err
	}
	cus := req.customer()
	cus.MerchantID = ac.MerchantID
	m, err := h.Controller.Create(c.Request().Context(), cus)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, customerResourceFrom(m))
}

func (h *Customer) Update(c echo.Context) error {
	ac, ok := c.(*AuthContext)
	if !ok {
		return errors.New("Invalid context")
	}
	req := updateCustomerRequest{}
	if err := bindAndValidate(c, &req); err != nil {
		return err
	}
	cus := req.customer()
	cus.MerchantID = ac.MerchantID
	m, err := h.Controller.Update(c.Request().Context(), cus)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, customerResourceFrom(m))
}

func (h *Customer) Retrieve(c echo.Context) error {
	ac, ok := c.(*AuthContext)
	if !ok {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid context")
	}
	m, err := h.Controller.Retrieve(c.Request().Context(), controller.RetrieveCustomerRequest{
		ID:         c.Param("id"),
		MerchantID: ac.MerchantID,
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, customerResourceFrom(m))
}

func (h *Customer) ListAll(c echo.Context) error {
	ac, ok := c.(*AuthContext)
	if !ok {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid context")
	}
	req := listAllCustomersRequest{}
	if err := bindAndValidate(c, &req); err != nil {
		return err
	}
	cuss, err := h.Controller.ListAll(c.Request().Context(), controller.ListAllCustomersRequest{
		Limit:      req.Limit,
		Offset:     req.Offset,
		MerchantID: ac.MerchantID,
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	res := make([]customerResource, len(cuss))
	for i, m := range cuss {
		res[i] = customerResourceFrom(m)
	}
	return c.JSON(http.StatusOK, listCustomersResponse{res})
}

func (h *Customer) Delete(c echo.Context) error {
	return nil
}
