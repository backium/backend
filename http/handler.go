package http

import (
	"github.com/backium/backend/core"
	"github.com/backium/backend/errors"
	"github.com/labstack/echo/v4"
)

type Handler struct {
	UserService       core.UserService
	CatalogService    core.CatalogService
	MerchantService   core.MerchantService
	LocationService   core.LocationService
	CustomerService   core.CustomerService
	SessionRepository SessionRepository
}

func bindAndValidate(c echo.Context, req interface{}) error {
	const op = errors.Op("handler.bindAndValidate")
	if err := c.Bind(req); err != nil {
		return errors.E(op, errors.KindValidation, err)
	}

	if err := c.Validate(req); err != nil {
		return errors.E(op, errors.KindValidation, err)
	}
	return nil
}
