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
	ID              ID          `bson:"_id"`
	ItemVariationID ID          `bson:"item_variation_id"`
	Quantity        int64       `bson:"quantity"`
	Op              InventoryOp `bson:"operation"`
	Note            string      `bson:"note"`
	LocationID      ID          `bson:"location_id"`
	MerchantID      ID          `bson:"merchant_id"`
	CreatedAt       int64       `bson:"created_at"`
}

func NewInventoryAdjustment(variationID, locationID, merchantID ID) InventoryAdjustment {
	return InventoryAdjustment{
		ID:              NewID("invadj"),
		ItemVariationID: variationID,
		LocationID:      locationID,
		MerchantID:      merchantID,
	}
}

type InventoryCount struct {
	ID              ID             `bson:"_id"`
	ItemVariationID ID             `bson:"item_variation_id"`
	Quantity        int64          `bson:"quantity"`
	State           InventoryState `bson:"state"`
	CalculatedAt    int64          `bson:"calculated_at"`
	LocationID      ID             `bson:"location_id"`
	MerchantID      ID             `bson:"merchant_id"`
}

func NewInventoryCount(variationID, locationID, merchantID ID) InventoryCount {
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
	IDs              []ID
	ItemVariationIDs []ID
	LocationIDs      []ID
	MerchantID       ID
}

type InventoryStorage interface {
	PutCount(context.Context, InventoryCount) error
	PutBatchCount(context.Context, []InventoryCount) error
	PutBatchAdj(context.Context, []InventoryAdjustment) error
	ListCount(context.Context, InventoryFilter) ([]InventoryCount, int64, error)
	ListAdjustment(context.Context, InventoryFilter) ([]InventoryAdjustment, int64, error)
}

func (s *CatalogService) initializeInventory(ctx context.Context, variation ItemVariation) error {
	var inventoryCounts []InventoryCount
	locations, _, err := s.LocationStorage.List(ctx, LocationQuery{
		Filter: LocationFilter{
			MerchantID: variation.MerchantID,
		},
	})
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

	variations := make([]ID, len(adjs))
	for i, adj := range adjs {
		variations[i] = adj.ItemVariationID
	}
	counts, _, err := s.InventoryStorage.ListCount(ctx, InventoryFilter{
		ItemVariationIDs: variations,
	})
	if err != nil {
		return nil, err
	}
	if len(counts) == 0 {
		return nil, errors.E(op, errors.KindValidation, "Unknown item variations in adjustment request")
	}

	var countsToUpdate []InventoryCount
	var countIDs []ID
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

	counts, _, err = s.InventoryStorage.ListCount(ctx, InventoryFilter{
		IDs: countIDs,
	})
	if err != nil {
		return nil, errors.E(op, err)
	}

	return counts, nil
}

func (s *CatalogService) ListInventoryCounts(ctx context.Context, f InventoryFilter) ([]InventoryCount, int64, error) {
	const op = errors.Op("core/CatalogService.ListInventoryCounts")

	counts, total, err := s.InventoryStorage.ListCount(ctx, InventoryFilter{
		MerchantID:       f.MerchantID,
		LocationIDs:      f.LocationIDs,
		ItemVariationIDs: f.ItemVariationIDs,
		Limit:            f.Limit,
		Offset:           f.Offset,
	})
	if err != nil {
		return nil, 0, errors.E(op, err)
	}

	return counts, total, nil
}

func (s *CatalogService) ListInventoryAdjustment(ctx context.Context, f InventoryFilter) ([]InventoryAdjustment, int64, error) {
	const op = errors.Op("core/CatalogService.ListInventoryCounts")

	adjs, total, err := s.InventoryStorage.ListAdjustment(ctx, f)
	if err != nil {
		return nil, 0, errors.E(op, err)
	}

	return adjs, total, nil
}
