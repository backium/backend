package app

import (
	"fmt"
	"net/http"

	"github.com/backium/backend/errors"
	"github.com/labstack/echo/v4"
)

const (
	ErrTypeInvalidRequest ErrorType = "invalid_request_error"
	ErrTypeAuthentication ErrorType = "authentication_error"
	ErrTypeApi            ErrorType = "api_error"
)

type ErrorType string

type Error struct {
	Type    ErrorType `json:"type"`
	Message string    `json:"message"`
}

func errorHandler(err error, c echo.Context) {
	code := http.StatusInternalServerError
	if he, ok := err.(*echo.HTTPError); ok {
		code = he.Code
	}
	c.Logger().Error(err)

	serr := Error{}
	fmt.Println(err)
	switch true {
	case errors.Is(err, errors.KindNotFound):
		code = http.StatusNotFound
		serr.Type = "invalid_request_error"
		serr.Message = "No such resource"
	case errors.Is(err, errors.KindValidation):
		code = http.StatusBadRequest
		serr.Type = "invalid_request_error"
		serr.Message = "Validation error"
	case errors.Is(err, errors.KindInvalidCredentials):
		code = http.StatusBadRequest
		serr.Type = "authentication_error"
		serr.Message = "Invalid credentials"
	case errors.Is(err, errors.KindInvalidSession):
		code = http.StatusUnauthorized
		serr.Type = "authentication_error"
		serr.Message = "Authentication required"
	default:
		serr.Type = "api_error"
		serr.Message = "Something went wrong"
	}

	c.JSON(code, map[string]interface{}{
		"error": serr,
	})
}
