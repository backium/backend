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

	location := core.NewLocation(ac.MerchantID)
	location.Name = req.Name
	location.BusinessName = req.BusinessName
	location.Image = req.Image

	ctx := c.Request().Context()
	location, err := h.LocationService.PutLocation(ctx, location)
	if err != nil {
		return errors.E(op, err)
	}
	return c.JSON(http.StatusOK, NewLocation(location))
}

func (h *Handler) UpdateLocation(c echo.Context) error {
	const op = errors.Op("handler.Location.Update")
	ac, ok := c.(*AuthContext)
	if !ok {
		return errors.E(op, errors.KindUnexpected, "invalid echo.Context")
	}
	req := LocationUpdateRequest{}
	if err := bindAndValidate(c, &req); err != nil {
		return err
	}

	ctx := c.Request().Context()
	location, err := h.LocationService.GetLocation(ctx, req.ID, ac.MerchantID)
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

func (h *Handler) RetrieveLocation(c echo.Context) error {
	const op = errors.Op("handler.Location.Retrieve")
	ac, ok := c.(*AuthContext)
	if !ok {
		return errors.E(op, errors.KindUnexpected, "invalid echo.Context")
	}
	id := c.Param("id")
	ctx := c.Request().Context()
	location, err := h.LocationService.GetLocation(ctx, id, ac.MerchantID)
	if err != nil {
		return errors.E(op, err)
	}
	return c.JSON(http.StatusOK, NewLocation(location))
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
	var limit, offset int64 = LocationListDefaultSize, 0
	if req.Limit <= LocationListMaxSize {
		limit = req.Limit
	}
	if req.Limit > LocationListMaxSize {
		limit = LocationListMaxSize
	}
	if req.Offset != 0 {
		offset = req.Offset
	}

	locations, err := h.LocationService.ListLocation(c.Request().Context(), core.LocationFilter{
		Limit:      limit,
		Offset:     offset,
		MerchantID: ac.MerchantID,
	})
	if err != nil {
		return errors.E(op, err)
	}
	resp := LocationListResponse{Locations: make([]Location, len(locations))}
	for i, loc := range locations {
		resp.Locations[i] = NewLocation(loc)
	}
	return c.JSON(http.StatusOK, resp)
}

func (h *Handler) DeleteLocation(c echo.Context) error {
	const op = errors.Op("handler.Location.Delete")
	ac, ok := c.(*AuthContext)
	if !ok {
		return errors.E(op, errors.KindUnexpected, "invalid echo.Context")
	}
	id := c.Param("id")
	ctx := c.Request().Context()
	location, err := h.LocationService.DeleteLocation(ctx, id, ac.MerchantID)
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

type LocationCreateRequest struct {
	Name         string `json:"name" validate:"required"`
	BusinessName string `json:"business_name"`
	Image        string `json:"image"`
}

type LocationUpdateRequest struct {
	ID           string  `json:"id" param:"id" validate:"required"`
	Name         *string `json:"name" validate:"omitempty,min=1"`
	BusinessName *string `json:"business_name" validate:"omitempty"`
	Image        *string `json:"image"`
}

type LocationListRequest struct {
	Limit  int64 `query:"limit" validate:"gte=0"`
	Offset int64 `query:"offset" validate:"gte=0"`
}

type LocationListResponse struct {
	Locations []Location `json:"locations"`
}
