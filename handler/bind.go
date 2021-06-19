package handler

import (
	"github.com/backium/backend/errors"
	"github.com/labstack/echo/v4"
)

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
