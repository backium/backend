package http

import (
	"net/http"

	"github.com/backium/backend/core"
	"github.com/backium/backend/errors"
	"github.com/backium/backend/ptr"
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
	loc.Image = req.Image
	loc.MerchantID = ac.MerchantID

	ctx := c.Request().Context()
	loc, err := h.LocationService.PutLocation(ctx, loc)
	if err != nil {
		return errors.E(op, err)
	}
	return c.JSON(http.StatusOK, NewLocation(loc))
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
	loc, err := h.LocationService.GetLocation(ctx, req.ID, ac.MerchantID)
	if err != nil {
		return err
	}
	if req.Name != nil {
		loc.Name = *req.Name
	}
	if req.BusinessName != nil {
		loc.BusinessName = *req.BusinessName
	}
	if req.Image != nil {
		loc.Image = *req.Image
	}
	loc, err = h.LocationService.PutLocation(ctx, loc)
	if err != nil {
		return errors.E(op, err)
	}
	return c.JSON(http.StatusOK, NewLocation(loc))
}

func (h *Handler) RetrieveLocation(c echo.Context) error {
	const op = errors.Op("handler.Location.Retrieve")
	ac, ok := c.(*AuthContext)
	if !ok {
		return errors.E(op, errors.KindUnexpected, "invalid echo.Context")
	}
	id := c.Param("id")
	ctx := c.Request().Context()
	loc, err := h.LocationService.GetLocation(ctx, id, ac.MerchantID)
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
	locs, err := h.LocationService.ListLocation(c.Request().Context(), core.LocationFilter{
		Limit:      ptr.GetInt64(req.Limit),
		Offset:     ptr.GetInt64(req.Offset),
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
	ctx := c.Request().Context()
	loc, err := h.LocationService.DeleteLocation(ctx, id, ac.MerchantID)
	if err != nil {
		return errors.E(op, err)
	}
	return c.JSON(http.StatusOK, NewLocation(loc))
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

func NewLocation(loc core.Location) Location {
	return Location{
		ID:           loc.ID,
		Name:         loc.Name,
		BusinessName: loc.BusinessName,
		Image:        loc.Image,
		MerchantID:   loc.MerchantID,
		CreatedAt:    loc.CreatedAt,
		UpdatedAt:    loc.UpdatedAt,
		Status:       loc.Status,
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
	Limit  *int64 `query:"limit" validate:"omitempty,gte=1"`
	Offset *int64 `query:"offset"`
}

type LocationListResponse struct {
	Locations []Location `json:"locations"`
}
