package api

import (
	"fmt"
	"net/http"

	"github.com/backium/backend/merchants"
	"github.com/labstack/echo/v4"
)

type merchantResource struct {
	ID           string `json:"id"`
	FirstName    string `json:"first_name"`
	LastName     string `json:"last_name"`
	BusinessName string `json:"business_name"`
}

type createMerchantRequest struct {
	FirstName    string `json:"first_name"`
	LastName     string `json:"last_name"`
	BusinessName string `json:"business_name"`
}

type updateMerchantRequest struct {
	ID           string `param:"id"`
	FirstName    string `json:"first_name"`
	LastName     string `json:"last_name"`
	BusinessName string `json:"business_name"`
}

type listMerchantsResponse struct {
	Merchants []merchantResource `json:"merchants"`
}

type merchantHandler struct {
	controller merchants.MerchantController
}

func (h *merchantHandler) Create(c echo.Context) error {
	req := createMerchantRequest{}
	if err := c.Bind(&req); err != nil {
		return err
	}
	m, err := h.controller.Create(req.merchant())
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, merchantResourceFrom(m))
}

func (h *merchantHandler) Update(c echo.Context) error {
	req := updateMerchantRequest{}
	if err := c.Bind(&req); err != nil {
		return err
	}
	fmt.Println("binding done")
	m, err := h.controller.Update(req.merchant())
	if err != nil {
		fmt.Println(err)
		return err
	}
	return c.JSON(http.StatusOK, merchantResourceFrom(m))
}

func (h *merchantHandler) Retrieve(c echo.Context) error {
	id := c.Param("id")
	m, err := h.controller.Retrieve(id)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, merchantResourceFrom(m))
}

func (h *merchantHandler) ListAll(c echo.Context) error {
	ms, err := h.controller.ListAll()
	if err != nil {
		return err
	}
	res := make([]merchantResource, len(ms))
	for i, m := range ms {
		res[i] = merchantResourceFrom(m)
	}
	return c.JSON(http.StatusOK, listMerchantsResponse{res})
}

func (h *merchantHandler) Delete(c echo.Context) error {
	return nil
}

func merchantResourceFrom(m merchants.Merchant) merchantResource {
	return merchantResource{
		ID:           m.ID,
		FirstName:    m.FirstName,
		LastName:     m.LastName,
		BusinessName: m.BusinessName,
	}
}

func (req createMerchantRequest) merchant() merchants.Merchant {
	return merchants.Merchant{
		FirstName:    req.FirstName,
		LastName:     req.LastName,
		BusinessName: req.BusinessName,
	}
}

func (req updateMerchantRequest) merchant() merchants.Merchant {
	return merchants.Merchant{
		ID:           req.ID,
		FirstName:    req.FirstName,
		LastName:     req.LastName,
		BusinessName: req.BusinessName,
	}
}
