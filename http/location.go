package http

import (
	"net/http"

	"github.com/backium/backend/core"
	"github.com/backium/backend/errors"
	"github.com/labstack/echo/v4"
)

func (h *Handler) CreateLocation(c echo.Context) error {
	const op = errors.Op("handler.Location.Create")
	ac, ok := c.(*AuthContext)
	if !ok {
		return errors.E(op, errors.KindUnexpected, "invalid echo.Context")
	}
	req := LocationCreateRequest{}
	if err := bindAndValidate(c, &req); err != nil {
		return err
	}

	loc := core.NewLocation()
	loc.Name = req.Name
	loc.BusinessName = req.BusinessName
	loc.MerchantID = ac.MerchantID

	loc, err := h.LocationService.Create(c.Request().Context(), loc)
	if err != nil {
		return errors.E(op, err)
	}
	return c.JSON(http.StatusOK, NewLocation(loc))
}

func (h *Handler) UpdateLocation(c echo.Context) error {
	const op = errors.Op("handler.Location.Update")
	req := LocationUpdateRequest{}
	if err := bindAndValidate(c, &req); err != nil {
		return err
	}
	loc := core.LocationPartial{
		Name:         req.Name,
		BusinessName: req.BusinessName,
	}

	uloc, err := h.LocationService.Update(c.Request().Context(), req.ID, loc)
	if err != nil {
		return errors.E(op, err)
	}
	return c.JSON(http.StatusOK, NewLocation(uloc))
}

func (h *Handler) RetrieveLocation(c echo.Context) error {
	const op = errors.Op("handler.Location.Retrieve")
	ac, ok := c.(*AuthContext)
	if !ok {
		return errors.E(op, errors.KindUnexpected, "invalid echo.Context")
	}
	id := c.Param("id")
	loc, err := h.LocationService.Retrieve(c.Request().Context(), core.RetrieveLocationRequest{
		ID:         id,
		MerchantID: ac.MerchantID,
	})
	if err != nil {
		return errors.E(op, err)
	}
	return c.JSON(http.StatusOK, NewLocation(loc))
}

func (h *Handler) ListLocations(c echo.Context) error {
	const op = errors.Op("handler.Location.ListAll")
	ac, ok := c.(*AuthContext)
	if !ok {
		return errors.E(op, errors.KindUnexpected, "invalid echo.Context")
	}
	req := LocationListRequest{}
	if err := bindAndValidate(c, &req); err != nil {
		return err
	}
	locs, err := h.LocationService.ListAll(c.Request().Context(), core.ListAllLocationsRequest{
		Limit:      req.Limit,
		Offset:     req.Offset,
		MerchantID: ac.MerchantID,
	})
	if err != nil {
		return errors.E(op, err)
	}
	res := make([]Location, len(locs))
	for i, loc := range locs {
		res[i] = NewLocation(loc)
	}
	return c.JSON(http.StatusOK, LocationListResponse{res})
}

func (h *Handler) DeleteLocation(c echo.Context) error {
	const op = errors.Op("handler.Location.Delete")
	ac, ok := c.(*AuthContext)
	if !ok {
		return errors.E(op, errors.KindUnexpected, "invalid echo.Context")
	}
	id := c.Param("id")
	loc, err := h.LocationService.Delete(c.Request().Context(), core.DeleteLocationRequest{
		ID:         id,
		MerchantID: ac.MerchantID,
	})
	if err != nil {
		return errors.E(op, err)
	}
	return c.JSON(http.StatusOK, NewLocation(loc))
}

type Location struct {
	ID           string      `json:"id"`
	Name         string      `json:"name"`
	BusinessName string      `json:"business_name,omitempty"`
	MerchantID   string      `json:"merchant_id"`
	Status       core.Status `json:"status"`
}

func NewLocation(loc core.Location) Location {
	return Location{
		ID:           loc.ID,
		Name:         loc.Name,
		BusinessName: loc.BusinessName,
		MerchantID:   loc.MerchantID,
		Status:       loc.Status,
	}
}

type LocationCreateRequest struct {
	Name         string `json:"name" validate:"required"`
	BusinessName string `json:"business_name"`
}

type LocationUpdateRequest struct {
	ID           string  `json:"id" param:"id" validate:"required"`
	Name         *string `json:"name" validate:"omitempty,min=1"`
	BusinessName *string `json:"business_name" validate:"omitempty"`
}

type LocationListRequest struct {
	Limit  *int64 `query:"limit" validate:"omitempty,gte=1"`
	Offset *int64 `query:"offset"`
}

type LocationListResponse struct {
	Locations []Location `json:"locations"`
}
