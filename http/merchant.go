package http

import (
	"net/http"

	"github.com/backium/backend/core"
	"github.com/backium/backend/errors"
	"github.com/labstack/echo/v4"
)

func (h *Handler) HandleCreateAPIKey(c echo.Context) error {
	const op = errors.Op("http/Handler.CreateAPIKey")

	type request struct {
		Name string `json:"name" validate:"required"`
	}

	ctx := c.Request().Context()

	merchant := core.MerchantFromContext(ctx)
	if merchant == nil {
		return errors.E(op, errors.KindUnexpected, "invalid echo.Context")
	}

	req := request{}
	if err := c.Bind(&req); err != nil {
		return err
	}

	key, err := h.MerchantService.CreateKey(ctx, req.Name, merchant.ID)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, APIKey{
		Name:  key.Name,
		Token: key.Token,
	})
}

func (h *Handler) HandleRetrieveMerchant(c echo.Context) error {
	const op = errors.Op("http/Handler.RetrieveMerchant")

	ctx := c.Request().Context()

	merchant := core.MerchantFromContext(ctx)
	if merchant == nil {
		return errors.E(op, errors.KindUnexpected, "invalid echo.Context")
	}

	return c.JSON(http.StatusOK, NewMerchant(*merchant))
}

type Merchant struct {
	ID           core.ID `json:"id"`
	FirstName    string  `json:"first_name"`
	LastName     string  `json:"last_name"`
	BusinessName string  `json:"business_name"`
	Currency     string  `json:"currency"`
}

func NewMerchant(m core.Merchant) Merchant {
	return Merchant{
		ID:           m.ID,
		FirstName:    m.FirstName,
		LastName:     m.LastName,
		BusinessName: m.BusinessName,
		Currency:     m.Currency,
	}
}

type APIKey struct {
	Name  string  `json:"name"`
	Token core.ID `json:"token"`
}
