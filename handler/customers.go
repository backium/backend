package handler

import (
	"net/http"

	"github.com/backium/backend/controller"
	"github.com/backium/backend/entity"
	"github.com/backium/backend/errors"
	"github.com/labstack/echo/v4"
)

type customer struct {
	ID         string        `json:"id"`
	Name       string        `json:"name"`
	Email      string        `json:"email"`
	Phone      string        `json:"phone"`
	Address    *address      `json:"address,omitempty"`
	MerchantID string        `json:"merchant_id"`
	Status     entity.Status `json:"status"`
}

type address struct {
	Line1      string `json:"line1,omitempty"`
	Line2      string `json:"line2,omitempty"`
	District   string `json:"district,omitempty"`
	Province   string `json:"province,omitempty"`
	Department string `json:"department,omitempty"`
}

func newCustomer(cus entity.Customer) customer {
	cusr := customer{
		ID:         cus.ID,
		Name:       cus.Name,
		Email:      cus.Email,
		Phone:      cus.Phone,
		MerchantID: cus.MerchantID,
		Status:     cus.Status,
	}
	if cus.Address != nil {
		cusr.Address = &address{
			Line1:      cus.Address.Line1,
			Line2:      cus.Address.Line2,
			District:   cus.Address.District,
			Province:   cus.Address.Province,
			Department: cus.Address.Department,
		}
	}
	return cusr
}

type createCustomerRequest struct {
	Name    string   `json:"name" validate:"required"`
	Email   string   `json:"email" validate:"required,email"`
	Phone   string   `json:"phone"`
	Address *address `json:"address"`
}

type updateCustomerRequest struct {
	ID      string   `param:"id"`
	Name    *string  `json:"name" validate:"omitempty,min=1"`
	Email   *string  `json:"email" validate:"omitempty,email"`
	Phone   *string  `json:"phone"`
	Address *address `json:"address"`
}

type listAllCustomersRequest struct {
	Limit  *int64 `query:"limit" validate:"omitempty,gte=1"`
	Offset *int64 `query:"offset"`
}

type listCustomersResponse struct {
	Customers []customer `json:"customers"`
}

type Customer struct {
	Controller controller.Customer
}

func (h *Customer) Create(c echo.Context) error {
	const op = errors.Op("handler.Customer.Create")
	ac, ok := c.(*AuthContext)
	if !ok {
		return errors.E(op, errors.KindUnexpected, "invalid echo.Context")
	}
	req := createCustomerRequest{}
	if err := bindAndValidate(c, &req); err != nil {
		return err
	}
	cus := entity.NewCustomer()
	if req.Address != nil {
		cus.Address = &entity.Address{
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
	cus, err := h.Controller.Create(c.Request().Context(), cus)
	if err != nil {
		return errors.E(op, err)
	}
	return c.JSON(http.StatusOK, newCustomer(cus))
}

func (h *Customer) Update(c echo.Context) error {
	const op = errors.Op("handler.Customer.Update")
	req := updateCustomerRequest{}
	if err := bindAndValidate(c, &req); err != nil {
		return err
	}
	cus := controller.PartialCustomer{
		Name:  req.Name,
		Email: req.Email,
		Phone: req.Phone,
	}
	if req.Address != nil {
		cus.Address = &entity.Address{
			Line1:      req.Address.Line1,
			Line2:      req.Address.Line2,
			Department: req.Address.Department,
			District:   req.Address.District,
			Province:   req.Address.Province,
		}
	}
	ucus, err := h.Controller.Update(c.Request().Context(), req.ID, cus)
	if err != nil {
		return errors.E(op, err)
	}
	return c.JSON(http.StatusOK, newCustomer(ucus))
}

func (h *Customer) Retrieve(c echo.Context) error {
	const op = errors.Op("handler.Customer.Retrieve")
	ac, ok := c.(*AuthContext)
	if !ok {
		return errors.E(op, errors.KindUnexpected, "invalid echo.Context")
	}
	m, err := h.Controller.Retrieve(c.Request().Context(), controller.RetrieveCustomerRequest{
		ID:         c.Param("id"),
		MerchantID: ac.MerchantID,
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, newCustomer(m))
}

func (h *Customer) ListAll(c echo.Context) error {
	const op = errors.Op("handler.Customer.ListAll")
	ac, ok := c.(*AuthContext)
	if !ok {
		return errors.E(op, errors.KindUnexpected, "invalid echo.Context")
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
		return errors.E(op, err)
	}
	res := make([]customer, len(cuss))
	for i, cus := range cuss {
		res[i] = newCustomer(cus)
	}
	return c.JSON(http.StatusOK, listCustomersResponse{res})
}

func (h *Customer) Delete(c echo.Context) error {
	const op = errors.Op("handler.Customer.Delete")
	ac, ok := c.(*AuthContext)
	if !ok {
		return errors.E(op, errors.KindUnexpected, "invalid echo.Context")
	}
	cus, err := h.Controller.Delete(c.Request().Context(), controller.DeleteCustomerRequest{
		ID:         c.Param("id"),
		MerchantID: ac.MerchantID,
	})
	if err != nil {
		return errors.E(op, err)
	}
	return c.JSON(http.StatusOK, newCustomer(cus))
}
