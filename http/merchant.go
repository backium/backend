package http

import (
	"net/http"

	"github.com/backium/backend/core"
	"github.com/backium/backend/errors"
	"github.com/labstack/echo/v4"
)

func (h *Handler) CreateMerchant(c echo.Context) error {
	req := MerchantCreateRequest{}
	if err := c.Bind(&req); err != nil {
		return err
	}
	merchant := core.NewMerchant()
	merchant.FirstName = req.FirstName
	merchant.LastName = req.LastName
	merchant.BusinessName = req.BusinessName
	ctx := c.Request().Context()
	merchant, err := h.MerchantService.PutMerchant(ctx, merchant)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, NewMerchant(merchant))
}

func (h *Handler) CreateAPIKey(c echo.Context) error {
	const op = errors.Op("http/Handler.CreateAPIKey")
	ac, ok := c.(*AuthContext)
	if !ok {
		return errors.E(op, errors.KindUnexpected, "invalid echo.Context")
	}
	req := APIKeyCreateRequest{}
	if err := c.Bind(&req); err != nil {
		return err
	}
	ctx := c.Request().Context()
	key, err := h.MerchantService.CreateKey(ctx, req.Name, ac.MerchantID)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, APIKey{
		Name:  key.Name,
		Token: key.Token,
	})
}

func (h *Handler) RetrieveMerchant(c echo.Context) error {
	const op = errors.Op("http/Handler.RetrieveMerchant")
	ac, ok := c.(*AuthContext)
	if !ok {
		return errors.E(op, errors.KindUnexpected, "invalid echo.Context")
	}
	ctx := c.Request().Context()
	merch, err := h.MerchantService.GetMerchant(ctx, ac.MerchantID)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, NewMerchant(merch))
}

type Merchant struct {
	ID           string `json:"id"`
	FirstName    string `json:"first_name"`
	LastName     string `json:"last_name"`
	BusinessName string `json:"business_name"`
}

func NewMerchant(m core.Merchant) Merchant {
	return Merchant{
		ID:           m.ID,
		FirstName:    m.FirstName,
		LastName:     m.LastName,
		BusinessName: m.BusinessName,
	}
}

type APIKey struct {
	Name  string `json:"name"`
	Token string `json:"token"`
}

type MerchantCreateRequest struct {
	FirstName    string `json:"first_name"`
	LastName     string `json:"last_name"`
	BusinessName string `json:"business_name"`
}

type APIKeyCreateRequest struct {
	Name string `json:"name"`
}

type MerchantUpdateRequest struct {
	ID           string `param:"id"`
	FirstName    string `json:"first_name"`
	LastName     string `json:"last_name"`
	BusinessName string `json:"business_name"`
}
