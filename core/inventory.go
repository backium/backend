package core

import (
	"context"

	"github.com/backium/backend/errors"
)

type InventoryState string

const (
	InventoryStateInStock InventoryState = "in_stock"
	InventoryStateSold    InventoryState = "sold"
)

type InventoryOp string

const (
	InventoryOpAddStock    InventoryOp = "add_stock"
	InventoryOpRemoveStock InventoryOp = "remove_stock"
	InventoryOpResetStock  InventoryOp = "reset_stock"
)

type InventoryAdjustment struct {
	ID              string      `bson:"_id"`
	ItemVariationID string      `bson:"item_variation_id"`
	Quantity        int64       `bson:"quantity"`
	Op              InventoryOp `bson:"operation"`
	LocationID      string      `bson:"location_id"`
	MerchantID      string      `bson:"merchant_id"`
	CreatedAt       int64       `bson:"created_at"`
}

func NewInventoryAdjustment(variationID, locationID, merchantID string) InventoryAdjustment {
	return InventoryAdjustment{
		ID:              NewID("invadj"),
		ItemVariationID: variationID,
		LocationID:      locationID,
		MerchantID:      merchantID,
	}
}

type InventoryCount struct {
	ID              string         `bson:"_id"`
	ItemVariationID string         `bson:"item_variation_id"`
	Quantity        int64          `bson:"quantity"`
	State           InventoryState `bson:"state"`
	CalculatedAt    int64          `bson:"calculated_at"`
	LocationID      string         `bson:"location_id"`
	MerchantID      string         `bson:"merchant_id"`
}

func NewInventoryCount(variationID, locationID, merchantID string) InventoryCount {
	return InventoryCount{
		ID:              NewID("invcount"),
		ItemVariationID: variationID,
		State:           InventoryStateSold,
		LocationID:      locationID,
		MerchantID:      merchantID,
	}
}

func (count *InventoryCount) applyAdjustments(adjs []InventoryAdjustment) (bool, error) {
	changed := false
	count.State = InventoryStateInStock
	for _, adj := range adjs {
		if adj.ItemVariationID != count.ItemVariationID ||
			adj.LocationID != count.LocationID {
			continue
		}
		changed = true
		switch adj.Op {
		case InventoryOpAddStock:
			count.Quantity += adj.Quantity
		case InventoryOpRemoveStock:
			count.Quantity -= adj.Quantity
		case InventoryOpResetStock:
			count.Quantity = adj.Quantity
		default:
			return false, errors.E(errors.KindValidation, "Invalid inventory adjusment operation")
		}
	}
	if count.Quantity == 0 {
		count.State = InventoryStateSold
	}
	return changed, nil
}

type InventoryFilter struct {
	Limit            int64
	Offset           int64
	IDs              []string
	ItemVariationIDs []string
	LocationIDs      []string
	MerchantID       string
}

type InventoryStorage interface {
	PutCount(context.Context, InventoryCount) error
	PutBatchCount(context.Context, []InventoryCount) error
	PutBatchAdj(context.Context, []InventoryAdjustment) error
	ListCount(context.Context, InventoryFilter) ([]InventoryCount, error)
}

func (s *CatalogService) initializeInventory(ctx context.Context, variation ItemVariation) error {
	var inventoryCounts []InventoryCount
	locations, err := s.LocationStorage.List(ctx, LocationFilter{MerchantID: variation.MerchantID})
	if err != nil {
		return err
	}
	for _, loc := range locations {
		count := NewInventoryCount(variation.ID, loc.ID, variation.MerchantID)
		inventoryCounts = append(inventoryCounts, count)
	}
	if err := s.InventoryStorage.PutBatchCount(ctx, inventoryCounts); err != nil {
		return err
	}
	return nil
}

func (s *CatalogService) ApplyInventoryAdjustments(ctx context.Context, adjs []InventoryAdjustment) ([]InventoryCount, error) {
	const op = errors.Op("core/CatalogService.PutInventoryAdjusments")
	variations := make([]string, len(adjs))
	for i, adj := range adjs {
		variations[i] = adj.ItemVariationID
	}
	counts, err := s.InventoryStorage.ListCount(ctx, InventoryFilter{
		ItemVariationIDs: variations,
	})
	if err != nil {
		return nil, err
	}
	if len(counts) == 0 {
		return nil, errors.E(op, errors.KindValidation, "Unknown item variations in adjustment request")
	}

	var countsToUpdate []InventoryCount
	var countIDs []string
	for _, count := range counts {
		changed, err := count.applyAdjustments(adjs)
		if err != nil {
			return nil, errors.E(op, err)
		}
		if changed {
			countsToUpdate = append(countsToUpdate, count)
			countIDs = append(countIDs, count.ID)
		}
	}
	if err := s.InventoryStorage.PutBatchCount(ctx, countsToUpdate); err != nil {
		return nil, errors.E(op, err)
	}
	if err := s.InventoryStorage.PutBatchAdj(ctx, adjs); err != nil {
		return nil, errors.E(op, err)
	}

	counts, err = s.InventoryStorage.ListCount(ctx, InventoryFilter{
		IDs: countIDs,
	})
	if err != nil {
		return nil, errors.E(op, err)
	}
	return counts, nil
}

func (s *CatalogService) ListInventoryCounts(ctx context.Context, f InventoryFilter) ([]InventoryCount, error) {
	const op = errors.Op("core/CatalogService.ListInventoryCounts")
	counts, err := s.InventoryStorage.ListCount(ctx, InventoryFilter{
		MerchantID:       f.MerchantID,
		LocationIDs:      f.LocationIDs,
		ItemVariationIDs: f.ItemVariationIDs,
		Limit:            f.Limit,
		Offset:           f.Offset,
	})
	if err != nil {
		return nil, errors.E(op, err)
	}
	return counts, nil
}
