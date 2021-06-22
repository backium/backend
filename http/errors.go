package http

import (
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
	c.Logger().Error(err)

	serr := Error{}
	switch true {
	case errors.Is(err, errors.KindNotFound):
		code = http.StatusNotFound
		serr.Type = ErrTypeInvalidRequest
		serr.Message = "No such resource"
	case errors.Is(err, errors.KindValidation):
		code = http.StatusBadRequest
		serr.Type = ErrTypeInvalidRequest
		serr.Message = "Validation error"
	case errors.Is(err, errors.KindInvalidCredentials):
		code = http.StatusBadRequest
		serr.Type = ErrTypeAuthentication
		serr.Message = "Invalid credentials"
	case errors.Is(err, errors.KindInvalidSession):
		code = http.StatusUnauthorized
		serr.Type = ErrTypeAuthentication
		serr.Message = "Authentication required"
	default:
		if he, ok := err.(*echo.HTTPError); ok {
			code = he.Code
			serr.Type = "invalid_request_error"
			serr.Message = he.Message.(string)
			break
		}
		serr.Type = ErrTypeApi
		serr.Message = "Something went wrong. Please contact support"
	}

	c.JSON(code, map[string]interface{}{
		"error": serr,
	})
}
