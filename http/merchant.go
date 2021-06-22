package http

import (
	"fmt"
	"net/http"

	"github.com/backium/backend/core"
	"github.com/labstack/echo/v4"
)

func (h *Handler) CreateMerchant(c echo.Context) error {
	req := MerchantCreateRequest{}
	if err := c.Bind(&req); err != nil {
		return err
	}
	m, err := h.MerchantService.Create(req.merchant())
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, NewMerchant(m))
}

func (h *Handler) UpdateMerchant(c echo.Context) error {
	req := MerchantUpdateRequest{}
	if err := c.Bind(&req); err != nil {
		return err
	}
	ac := c.(*AuthContext)
	if ac.Kind != core.UserKindSuper {
		req.ID = ac.MerchantID
	}
	m, err := h.MerchantService.Update(req.merchant())
	if err != nil {
		fmt.Println(err)
		return err
	}
	return c.JSON(http.StatusOK, NewMerchant(m))
}

func (h *Handler) RetrieveMerchant(c echo.Context) error {
	id := c.Param("id")
	ac := c.(*AuthContext)
	if ac.Kind != core.UserKindSuper {
		id = ac.MerchantID
	}
	m, err := h.MerchantService.Retrieve(id)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, NewMerchant(m))
}

func (h *Handler) ListMerchants(c echo.Context) error {
	ms, err := h.MerchantService.ListAll()
	if err != nil {
		return err
	}
	res := make([]Merchant, len(ms))
	for i, m := range ms {
		res[i] = NewMerchant(m)
	}
	return c.JSON(http.StatusOK, MerchantListResponse{res})
}

func (h *Handler) DeleteMerchant(c echo.Context) error {
	return nil
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

type MerchantCreateRequest struct {
	FirstName    string `json:"first_name"`
	LastName     string `json:"last_name"`
	BusinessName string `json:"business_name"`
}

func (req MerchantCreateRequest) merchant() core.Merchant {
	return core.Merchant{
		FirstName:    req.FirstName,
		LastName:     req.LastName,
		BusinessName: req.BusinessName,
	}
}

type MerchantUpdateRequest struct {
	ID           string `param:"id"`
	FirstName    string `json:"first_name"`
	LastName     string `json:"last_name"`
	BusinessName string `json:"business_name"`
}

func (req MerchantUpdateRequest) merchant() core.Merchant {
	return core.Merchant{
		ID:           req.ID,
		FirstName:    req.FirstName,
		LastName:     req.LastName,
		BusinessName: req.BusinessName,
	}
}

type MerchantListResponse struct {
	Merchants []Merchant `json:"merchants"`
}
