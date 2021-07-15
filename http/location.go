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

	location, err := h.LocationService.CreateLocation(ctx, location)
	if err != nil {
		return errors.E(op, err)
	}

	return c.JSON(http.StatusOK, NewLocation(location))
}

func (h *Handler) HandleUpdateLocation(c echo.Context) error {
	const op = errors.Op("http/Handler.HandleUpdateLocation")

	type request struct {
		ID           core.ID `json:"id" param:"id" validate:"required"`
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

	location, err := h.LocationService.GetLocation(ctx, req.ID)
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

	type request struct {
		ID core.ID `param:"id"`
	}

	ctx := c.Request().Context()

	req := request{}
	if err := bindAndValidate(c, &req); err != nil {
		return err
	}

	merchant := core.MerchantFromContext(ctx)
	if merchant == nil {
		return errors.E(op, errors.KindUnexpected, "invalid echo.Context")
	}

	location, err := h.LocationService.GetLocation(ctx, req.ID)
	if err != nil {
		return errors.E(op, err)
	}

	return c.JSON(http.StatusOK, NewLocation(location))
}

func (h *Handler) HandleSearchLocation(c echo.Context) error {
	const op = errors.Op("http/Handler.HandleSearchLocation")

	type filter struct {
		IDs  []core.ID `json:"ids" validate:"omitempty,dive,id"`
		Name string    `json:"name"`
	}

	type sort struct {
		Name core.SortOrder `json:"name"`
	}

	type request struct {
		Limit  int64  `json:"limit" validate:"gte=0"`
		Offset int64  `json:"offset" validate:"gte=0"`
		Filter filter `json:"filter"`
		Sort   sort   `json:"sort"`
	}

	type response struct {
		Locations []Location `json:"locations"`
		Total     int64      `json:"total_count"`
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

	var limit int64 = LocationListDefaultSize
	if req.Limit <= LocationListMaxSize {
		limit = req.Limit
	} else {
		limit = LocationListMaxSize
	}

	locations, count, err := h.LocationService.ListLocation(ctx, core.LocationQuery{
		Limit:  limit,
		Offset: req.Offset,
		Filter: core.LocationFilter{
			Name:       req.Filter.Name,
			MerchantID: merchant.ID,
		},
		Sort: core.LocationSort{
			Name: req.Sort.Name,
		},
	})
	if err != nil {
		return errors.E(op, err)
	}

	resp := response{
		Locations: make([]Location, len(locations)),
		Total:     count,
	}
	for i, loc := range locations {
		resp.Locations[i] = NewLocation(loc)
	}

	return c.JSON(http.StatusOK, resp)
}

func (h *Handler) HandleListLocations(c echo.Context) error {
	const op = errors.Op("handler.Location.ListAll")

	type request struct {
		Limit  int64 `query:"limit" validate:"gte=0"`
		Offset int64 `query:"offset" validate:"gte=0"`
	}

	type response struct {
		Locations []Location `json:"locations"`
		Total     int64      `json:"total_count"`
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

	var limit int64 = LocationListDefaultSize
	if req.Limit <= LocationListMaxSize {
		limit = req.Limit
	} else {
		limit = LocationListMaxSize
	}

	locations, count, err := h.LocationService.ListLocation(ctx, core.LocationQuery{
		Limit:  limit,
		Offset: req.Offset,
		Filter: core.LocationFilter{
			MerchantID: merchant.ID,
		},
	})
	if err != nil {
		return errors.E(op, err)
	}

	resp := response{
		Locations: make([]Location, len(locations)),
		Total:     count,
	}
	for i, loc := range locations {
		resp.Locations[i] = NewLocation(loc)
	}

	return c.JSON(http.StatusOK, resp)
}

func (h *Handler) HandleDeleteLocation(c echo.Context) error {
	const op = errors.Op("handler.Location.Delete")

	type request struct {
		ID core.ID `param:"id"`
	}

	ctx := c.Request().Context()

	req := request{}
	if err := bindAndValidate(c, &req); err != nil {
		return err
	}

	merchant := core.MerchantFromContext(ctx)
	if merchant == nil {
		return errors.E(op, errors.KindUnexpected, "invalid echo.Context")
	}

	location, err := h.LocationService.DeleteLocation(ctx, req.ID)
	if err != nil {
		return errors.E(op, err)
	}

	return c.JSON(http.StatusOK, NewLocation(location))
}

type Location struct {
	ID           core.ID     `json:"id"`
	Name         string      `json:"name"`
	BusinessName string      `json:"business_name,omitempty"`
	Image        string      `json:"image,omitempty"`
	MerchantID   core.ID     `json:"merchant_id"`
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
