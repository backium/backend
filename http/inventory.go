package http

import (
	"fmt"
	"net/http"

	"github.com/backium/backend/core"
	"github.com/backium/backend/errors"
	"github.com/labstack/echo/v4"
)

const (
	InventoryCountListDefaultSize = 10
	InventoryCountListMaxSize     = 50
)

func (h *Handler) ChangeInventory(c echo.Context) error {
	const op = errors.Op("http/Handler.ChangeInventory")
	ac, ok := c.(*AuthContext)
	if !ok {
		return errors.E(op, errors.KindUnexpected, "invalid echo.Context")
	}
	req := InventoryChangeRequest{}
	if err := bindAndValidate(c, &req); err != nil {
		return err
	}
	adjs := make([]core.InventoryAdjustment, len(req.Adjustments))
	for i, adj := range req.Adjustments {
		adjs[i] = core.NewInventoryAdjustment(adj.ItemVariationID, adj.LocationID, ac.MerchantID)
		adjs[i].Op = adj.Op
		adjs[i].Quantity = *adj.Quantity
	}

	ctx := c.Request().Context()
	counts, err := h.CatalogService.ApplyInventoryAdjustments(ctx, adjs)
	if err != nil {
		return errors.E(op, err)
	}
	resp := InventoryChangeResponse{
		Counts: NewInventoryCounts(counts),
	}
	return c.JSON(http.StatusOK, resp)
}

func (h *Handler) BatchRetrieveInventory(c echo.Context) error {
	const op = errors.Op("http/Handler.ListInventoryCounts")
	ac, ok := c.(*AuthContext)
	if !ok {
		return errors.E(op, errors.KindUnexpected, "invalid echo.Context")
	}
	req := InventoryBatchRetrieveRequest{}
	if err := bindAndValidate(c, &req); err != nil {
		return err
	}
	fmt.Println("request", req)
	var limit, offset int64 = InventoryCountListDefaultSize, 0
	if req.Limit <= InventoryCountListMaxSize {
		limit = req.Limit
	}
	if req.Limit > InventoryCountListMaxSize {
		limit = InventoryCountListMaxSize
	}
	if req.Offset != 0 {
		offset = req.Offset
	}

	ctx := c.Request().Context()
	counts, err := h.CatalogService.ListInventoryCounts(ctx, core.InventoryFilter{
		Limit:            limit,
		Offset:           offset,
		MerchantID:       ac.MerchantID,
		LocationIDs:      req.LocationIDs,
		ItemVariationIDs: req.ItemVariationIDs,
	})
	if err != nil {
		return errors.E(op, err)
	}

	resp := InventoryListResponse{Counts: NewInventoryCounts(counts)}
	return c.JSON(http.StatusOK, resp)
}

type InventoryCount struct {
	ItemVariationID string `json:"item_variation_id"`
	Quantity        int64  `json:"quantity"`
	CalculatedAt    int64  `json:"calculated_at"`
	LocationID      string `json:"location_id"`
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
	res := make([]InventoryCount, len(counts))
	for i, count := range counts {
		res[i] = NewInventoryCount(count)
	}
	return res
}

type InventoryAdjustmentRequest struct {
	ItemVariationID string           `json:"item_variation_id" validate:"required"`
	Op              core.InventoryOp `json:"op" validate:"required"`
	Quantity        *int64           `json:"quantity" validate:"required"`
	LocationID      string           `json:"location_id" validate:"required"`
}

type InventoryChangeRequest struct {
	Adjustments []InventoryAdjustmentRequest `json:"adjustments" validate:"required,dive"`
}

type InventoryChangeResponse struct {
	Counts []InventoryCount `json:"counts"`
}

type InventoryBatchRetrieveRequest struct {
	ItemVariationIDs []string `json:"item_variation_ids"`
	LocationIDs      []string `json:"location_ids"`
	Limit            int64    `json:"limit" validate:"gte=0"`
	Offset           int64    `json:"offset" validate:"gte=0"`
}

type InventoryListResponse struct {
	Counts []InventoryCount `json:"counts"`
}
