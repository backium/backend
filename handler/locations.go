package handler

import (
	"errors"
	"net/http"

	"github.com/backium/backend/controller"
	"github.com/backium/backend/entity"
	"github.com/labstack/echo/v4"
)

type locationResource struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	BusinessName string `json:"business_name"`
	MerchantID   string `json:"merchant_id"`
}

func locationResourceFrom(l entity.Location) locationResource {
	return locationResource{
		ID:           l.ID,
		Name:         l.Name,
		BusinessName: l.BusinessName,
		MerchantID:   l.MerchantID,
	}
}

type createLocationRequest struct {
	Name         string `json:"name"`
	BusinessName string `json:"business_name"`
}

type updateLocationRequest struct {
	Name         string `json:"name"`
	BusinessName string `json:"business_name"`
}

type listLocationsResponse struct {
	Locations []locationResource `json:"locations"`
}

type Location struct {
	Controller controller.Location
}

func (h *Location) Create(c echo.Context) error {
	ac, ok := c.(*AuthContext)
	if !ok {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid context")
	}
	req := createLocationRequest{}
	if err := bindAndValidate(c, &req); err != nil {
		return err
	}
	m, err := h.Controller.Create(c.Request().Context(), entity.Location{
		Name:         req.Name,
		BusinessName: req.BusinessName,
		MerchantID:   ac.MerchantID,
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, locationResourceFrom(m))
}

func (h *Location) Update(c echo.Context) error {
	ac, ok := c.(*AuthContext)
	if !ok {
		return errors.New("Invalid context")
	}
	req := updateLocationRequest{}
	if err := bindAndValidate(c, &req); err != nil {
		return err
	}
	m, err := h.Controller.Update(c.Request().Context(), entity.Location{
		ID:           c.Param("id"),
		Name:         req.Name,
		BusinessName: req.BusinessName,
		MerchantID:   ac.MerchantID,
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, locationResourceFrom(m))
}

func (h *Location) Retrieve(c echo.Context) error {
	ac, ok := c.(*AuthContext)
	if !ok {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid context")
	}
	id := c.Param("id")
	m, err := h.Controller.Retrieve(c.Request().Context(), controller.RetrieveLocationRequest{
		ID:         id,
		MerchantID: ac.MerchantID,
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, locationResourceFrom(m))
}

func (h *Location) ListAll(c echo.Context) error {
	ac, ok := c.(*AuthContext)
	if !ok {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid context")
	}
	ms, err := h.Controller.ListAll(c.Request().Context(), ac.MerchantID)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	res := make([]locationResource, len(ms))
	for i, m := range ms {
		res[i] = locationResourceFrom(m)
	}
	return c.JSON(http.StatusOK, listLocationsResponse{res})
}

func (h *Location) Delete(c echo.Context) error {
	return nil
}
