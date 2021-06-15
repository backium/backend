package handler

import (
	"fmt"
	"net/http"

	"github.com/backium/backend/controller"
	"github.com/backium/backend/entity"
	"github.com/labstack/echo/v4"
)

type Merchant struct {
	Controller controller.Merchant
}

func (h *Merchant) Create(c echo.Context) error {
	req := createMerchantRequest{}
	if err := c.Bind(&req); err != nil {
		return err
	}
	m, err := h.Controller.Create(req.merchant())
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, merchantResourceFrom(m))
}

func (h *Merchant) Update(c echo.Context) error {
	req := updateMerchantRequest{}
	if err := c.Bind(&req); err != nil {
		return err
	}
	ac := c.(*AuthContext)
	if !ac.IsSuper {
		req.ID = ac.MerchantID
	}
	m, err := h.Controller.Update(req.merchant())
	if err != nil {
		fmt.Println(err)
		return err
	}
	return c.JSON(http.StatusOK, merchantResourceFrom(m))
}

func (h *Merchant) Retrieve(c echo.Context) error {
	id := c.Param("id")
	ac := c.(*AuthContext)
	if !ac.IsSuper {
		id = ac.MerchantID
	}
	m, err := h.Controller.Retrieve(id)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, merchantResourceFrom(m))
}

func (h *Merchant) ListAll(c echo.Context) error {
	ms, err := h.Controller.ListAll()
	if err != nil {
		return err
	}
	res := make([]merchantResource, len(ms))
	for i, m := range ms {
		res[i] = merchantResourceFrom(m)
	}
	return c.JSON(http.StatusOK, listMerchantsResponse{res})
}

func (h *Merchant) Delete(c echo.Context) error {
	return nil
}

func merchantResourceFrom(m entity.Merchant) merchantResource {
	return merchantResource{
		ID:           m.ID,
		FirstName:    m.FirstName,
		LastName:     m.LastName,
		BusinessName: m.BusinessName,
	}
}

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

func (req createMerchantRequest) merchant() entity.Merchant {
	return entity.Merchant{
		FirstName:    req.FirstName,
		LastName:     req.LastName,
		BusinessName: req.BusinessName,
	}
}

type updateMerchantRequest struct {
	ID           string `param:"id"`
	FirstName    string `json:"first_name"`
	LastName     string `json:"last_name"`
	BusinessName string `json:"business_name"`
}

func (req updateMerchantRequest) merchant() entity.Merchant {
	return entity.Merchant{
		ID:           req.ID,
		FirstName:    req.FirstName,
		LastName:     req.LastName,
		BusinessName: req.BusinessName,
	}
}

type listMerchantsResponse struct {
	Merchants []merchantResource `json:"merchants"`
}
