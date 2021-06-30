package http

import (
	"net/http"

	"github.com/backium/backend/core"
	"github.com/backium/backend/errors"
	"github.com/labstack/echo/v4"
)

const (
	LocationListDefaultSize = 10
	LocationListMaxSize     = 50
)

func (h *Handler) HandleCreateLocation(c echo.Context) error {
	const op = errors.Op("http/Handler.HandleCreateLocation")

	type request struct {
		Name         string `json:"name" validate:"required"`
		BusinessName string `json:"business_name"`
		Image        string `json:"image"`
	}

	ctx := c.Request().Context()

	merchant := core.MerchantFromContext(ctx)
	if merchant == nil {
		return errors.E(op, errors.KindUnexpected, "invalid echo.Context")
	}

	req := request{}
	if err := bindAndValidate(c, &req); err != nil {
		return err
	}

	location := core.NewLocation(req.Name, merchant.ID)
	location.BusinessName = req.BusinessName
	location.Image = req.Image

	location, err := h.LocationService.PutLocation(ctx, location)
	if err != nil {
		return errors.E(op, err)
	}

	return c.JSON(http.StatusOK, NewLocation(location))
}

func (h *Handler) HandleUpdateLocation(c echo.Context) error {
	const op = errors.Op("http/Handler.HandleUpdateLocation")

	type request struct {
		ID           string  `json:"id" param:"id" validate:"required"`
		Name         *string `json:"name" validate:"omitempty,min=1"`
		BusinessName *string `json:"business_name" validate:"omitempty"`
		Image        *string `json:"image"`
	}

	ctx := c.Request().Context()

	merchant := core.MerchantFromContext(ctx)
	if merchant == nil {
		return errors.E(op, errors.KindUnexpected, "invalid echo.Context")
	}

	req := request{}
	if err := bindAndValidate(c, &req); err != nil {
		return err
	}

	location, err := h.LocationService.GetLocation(ctx, req.ID, merchant.ID)
	if err != nil {
		return err
	}
	if req.Name != nil {
		location.Name = *req.Name
	}
	if req.BusinessName != nil {
		location.BusinessName = *req.BusinessName
	}
	if req.Image != nil {
		location.Image = *req.Image
	}

	location, err = h.LocationService.PutLocation(ctx, location)
	if err != nil {
		return errors.E(op, err)
	}

	return c.JSON(http.StatusOK, NewLocation(location))
}

func (h *Handler) HandleRetrieveLocation(c echo.Context) error {
	const op = errors.Op("http/Handler.HandleRetrieveLocation")

	ctx := c.Request().Context()

	merchant := core.MerchantFromContext(ctx)
	if merchant == nil {
		return errors.E(op, errors.KindUnexpected, "invalid echo.Context")
	}

	location, err := h.LocationService.GetLocation(ctx, c.Param("id"), merchant.ID)
	if err != nil {
		return errors.E(op, err)
	}

	return c.JSON(http.StatusOK, NewLocation(location))
}

func (h *Handler) HandleListLocations(c echo.Context) error {
	const op = errors.Op("handler.Location.ListAll")

	type request struct {
		Limit  int64 `query:"limit" validate:"gte=0"`
		Offset int64 `query:"offset" validate:"gte=0"`
	}

	type response struct {
		Locations []Location `json:"locations"`
	}

	ctx := c.Request().Context()

	merchant := core.MerchantFromContext(ctx)
	if merchant == nil {
		return errors.E(op, errors.KindUnexpected, "invalid echo.Context")
	}

	req := request{}
	if err := bindAndValidate(c, &req); err != nil {
		return err
	}

	var limit, offset int64 = LocationListDefaultSize, req.Offset
	if req.Limit <= LocationListMaxSize {
		limit = req.Limit
	} else {
		limit = LocationListMaxSize
	}

	locations, err := h.LocationService.ListLocation(ctx, core.LocationFilter{
		Limit:      limit,
		Offset:     offset,
		MerchantID: merchant.ID,
	})
	if err != nil {
		return errors.E(op, err)
	}

	resp := response{Locations: make([]Location, len(locations))}
	for i, loc := range locations {
		resp.Locations[i] = NewLocation(loc)
	}

	return c.JSON(http.StatusOK, resp)
}

func (h *Handler) HandleDeleteLocation(c echo.Context) error {
	const op = errors.Op("handler.Location.Delete")

	ctx := c.Request().Context()

	merchant := core.MerchantFromContext(ctx)
	if merchant == nil {
		return errors.E(op, errors.KindUnexpected, "invalid echo.Context")
	}

	location, err := h.LocationService.DeleteLocation(ctx, c.Param("id"), merchant.ID)
	if err != nil {
		return errors.E(op, err)
	}

	return c.JSON(http.StatusOK, NewLocation(location))
}

type Location struct {
	ID           string      `json:"id"`
	Name         string      `json:"name"`
	BusinessName string      `json:"business_name,omitempty"`
	Image        string      `json:"image,omitempty"`
	MerchantID   string      `json:"merchant_id"`
	CreatedAt    int64       `json:"created_at"`
	UpdatedAt    int64       `json:"updated_at"`
	Status       core.Status `json:"status"`
}

func NewLocation(location core.Location) Location {
	return Location{
		ID:           location.ID,
		Name:         location.Name,
		BusinessName: location.BusinessName,
		Image:        location.Image,
		MerchantID:   location.MerchantID,
		CreatedAt:    location.CreatedAt,
		UpdatedAt:    location.UpdatedAt,
		Status:       location.Status,
	}
}
