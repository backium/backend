package http

import (
	"fmt"
	"strings"

	"github.com/backium/backend/core"
	"github.com/backium/backend/errors"
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
)

type Handler struct {
	UserService       core.UserService
	EmployeeService   core.EmployeeService
	CatalogService    core.CatalogService
	MerchantService   core.MerchantService
	LocationService   core.LocationService
	CustomerService   core.CustomerService
	OrderingService   core.OrderingService
	PaymentService    core.PaymentService
	ReportService     core.ReportService
	SessionRepository core.SessionStorage
}

func bindAndValidate(c echo.Context, req interface{}) error {
	const op = errors.Op("handler.bindAndValidate")

	if err := c.Bind(req); err != nil {
		return errors.E(op, errors.KindValidation, err)
	}

	if err := c.Validate(req); err != nil {
		e, ok := err.(validator.ValidationErrors)
		if !ok {
			return errors.E(op, errors.KindValidation, "unknow failed validation")
		}
		ferr := e[0]
		path := strings.SplitN(ferr.Namespace(), ".", 2)[1]
		msg := fmt.Sprintf("request field '%v' is not valid, it should satisfy: %v", path, ferr.Tag())
		return errors.E(op, errors.KindValidation, msg)
	}

	return nil
}
