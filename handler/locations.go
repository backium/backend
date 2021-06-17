package handler

import (
	"net/http"

	"github.com/backium/backend/controller"
	"github.com/backium/backend/entity"
	"github.com/backium/backend/errors"
	"github.com/labstack/echo/v4"
)

type locationResource struct {
	ID           string        `json:"id"`
	Name         string        `json:"name"`
	BusinessName string        `json:"business_name"`
	MerchantID   string        `json:"merchant_id"`
	Status       entity.Status `json:"status"`
}

func newLocationResource(loc entity.Location) locationResource {
	return locationResource{
		ID:           loc.ID,
		Name:         loc.Name,
		BusinessName: loc.BusinessName,
		MerchantID:   loc.MerchantID,
		Status:       loc.Status,
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

type listAllLocationsRequest struct {
	Limit  *int64 `query:"limit" validate:"omitempty,gte=1"`
	Offset *int64 `query:"offset"`
}

type listLocationsResponse struct {
	Locations []locationResource `json:"locations"`
}

type Location struct {
	Controller controller.Location
}

func (h *Location) Create(c echo.Context) error {
	const op = errors.Op("handler.Location.Create")
	ac, ok := c.(*AuthContext)
	if !ok {
		return errors.E(op, errors.KindUnexpected, "invalid echo.Context")
	}
	req := createLocationRequest{}
	if err := bindAndValidate(c, &req); err != nil {
		return err
	}
	loc, err := h.Controller.Create(c.Request().Context(), entity.Location{
		Name:         req.Name,
		BusinessName: req.BusinessName,
		MerchantID:   ac.MerchantID,
	})
	if err != nil {
		return errors.E(op, err)
	}
	return c.JSON(http.StatusOK, newLocationResource(loc))
}

func (h *Location) Update(c echo.Context) error {
	const op = errors.Op("handler.Location.Update")
	ac, ok := c.(*AuthContext)
	if !ok {
		return errors.E(op, errors.KindUnexpected, "invalid echo.Context")
	}
	req := updateLocationRequest{}
	if err := bindAndValidate(c, &req); err != nil {
		return err
	}
	loc, err := h.Controller.Update(c.Request().Context(), entity.Location{
		ID:           c.Param("id"),
		Name:         req.Name,
		BusinessName: req.BusinessName,
		MerchantID:   ac.MerchantID,
	})
	if err != nil {
		return errors.E(op, err)
	}
	return c.JSON(http.StatusOK, newLocationResource(loc))
}

func (h *Location) Retrieve(c echo.Context) error {
	const op = errors.Op("handler.Location.Retrieve")
	ac, ok := c.(*AuthContext)
	if !ok {
		return errors.E(op, errors.KindUnexpected, "invalid echo.Context")
	}
	id := c.Param("id")
	loc, err := h.Controller.Retrieve(c.Request().Context(), controller.RetrieveLocationRequest{
		ID:         id,
		MerchantID: ac.MerchantID,
	})
	if err != nil {
		return errors.E(op, err)
	}
	return c.JSON(http.StatusOK, newLocationResource(loc))
}

func (h *Location) ListAll(c echo.Context) error {
	const op = errors.Op("handler.Location.ListAll")
	ac, ok := c.(*AuthContext)
	if !ok {
		return errors.E(op, errors.KindUnexpected, "invalid echo.Context")
	}
	req := listAllLocationsRequest{}
	if err := bindAndValidate(c, &req); err != nil {
		return err
	}
	locs, err := h.Controller.ListAll(c.Request().Context(), controller.ListAllLocationsRequest{
		Limit:      req.Limit,
		Offset:     req.Offset,
		MerchantID: ac.MerchantID,
	})
	if err != nil {
		return errors.E(op, err)
	}
	res := make([]locationResource, len(locs))
	for i, loc := range locs {
		res[i] = newLocationResource(loc)
	}
	return c.JSON(http.StatusOK, listLocationsResponse{res})
}

func (h *Location) Delete(c echo.Context) error {
	const op = errors.Op("handler.Location.Delete")
	ac, ok := c.(*AuthContext)
	if !ok {
		return errors.E(op, errors.KindUnexpected, "invalid echo.Context")
	}
	id := c.Param("id")
	loc, err := h.Controller.Delete(c.Request().Context(), controller.DeleteLocationRequest{
		ID:         id,
		MerchantID: ac.MerchantID,
	})
	if err != nil {
		return errors.E(op, err)
	}
	return c.JSON(http.StatusOK, newLocationResource(loc))
}
