package http

import (
	"net/http"

	"github.com/backium/backend/core"
	"github.com/backium/backend/errors"
	"github.com/labstack/echo/v4"
)

const (
	InventoryCountListDefaultSize = 10
	InventoryCountListMaxSize     = 50
)

func (h *Handler) HandleChangeInventory(c echo.Context) error {
	const op = errors.Op("http/Handler.ChangeInventory")

	type adjustment struct {
		ItemVariationID core.ID          `json:"item_variation_id" validate:"required"`
		Op              core.InventoryOp `json:"op" validate:"required"`
		Quantity        *int64           `json:"quantity" validate:"required"`
		LocationID      core.ID          `json:"location_id" validate:"required"`
	}

	type request struct {
		Adjustments []adjustment `json:"adjustments" validate:"required,dive"`
	}

	type response struct {
		Counts []InventoryCount `json:"counts"`
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

	adjs := make([]core.InventoryAdjustment, len(req.Adjustments))
	for i, adj := range req.Adjustments {
		adjs[i] = core.NewInventoryAdjustment(adj.ItemVariationID, adj.LocationID, merchant.ID)
		adjs[i].Op = adj.Op
		adjs[i].Quantity = *adj.Quantity
	}

	counts, err := h.CatalogService.ApplyInventoryAdjustments(ctx, adjs)
	if err != nil {
		return errors.E(op, err)
	}

	resp := response{
		Counts: NewInventoryCounts(counts),
	}

	return c.JSON(http.StatusOK, resp)
}

func (h *Handler) HandleBatchRetrieveInventory(c echo.Context) error {
	const op = errors.Op("http/Handler.ListInventoryCounts")

	type request struct {
		ItemVariationIDs []core.ID `json:"item_variation_ids"`
		LocationIDs      []core.ID `json:"location_ids"`
		Limit            int64     `json:"limit" validate:"gte=0"`
		Offset           int64     `json:"offset" validate:"gte=0"`
	}

	type response struct {
		Counts []InventoryCount `json:"counts"`
		Total  int64            `json:"total_count"`
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

	var limit int64 = InventoryCountListDefaultSize
	if req.Limit <= InventoryCountListMaxSize {
		limit = req.Limit
	} else {
		limit = InventoryCountListMaxSize
	}

	counts, totalCount, err := h.CatalogService.ListInventoryCounts(ctx, core.InventoryFilter{
		Limit:            limit,
		Offset:           req.Offset,
		MerchantID:       merchant.ID,
		LocationIDs:      req.LocationIDs,
		ItemVariationIDs: req.ItemVariationIDs,
	})
	if err != nil {
		return errors.E(op, err)
	}

	resp := response{
		Counts: NewInventoryCounts(counts),
		Total:  totalCount,
	}

	return c.JSON(http.StatusOK, resp)
}

type InventoryCount struct {
	ItemVariationID core.ID `json:"item_variation_id"`
	Quantity        int64   `json:"quantity"`
	CalculatedAt    int64   `json:"calculated_at"`
	LocationID      core.ID `json:"location_id"`
}

func NewInventoryCount(count core.InventoryCount) InventoryCount {
	return InventoryCount{
		ItemVariationID: count.ItemVariationID,
		Quantity:        count.Quantity,
		CalculatedAt:    count.CalculatedAt,
		LocationID:      count.LocationID,
	}
}

func NewInventoryCounts(counts []core.InventoryCount) []InventoryCount {
	resp := make([]InventoryCount, len(counts))
	for i, count := range counts {
		resp[i] = NewInventoryCount(count)
	}
	return resp
}
