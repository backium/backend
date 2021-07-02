package core

import "context"

type Authorizer struct {
	ItemVariationStorage ItemVariationStorage
}

func (auth *Authorizer) canChangeInventory(ctx context.Context, variationIDs []ID) bool {
	merchant := MerchantFromContext(ctx)

	variations, err := auth.ItemVariationStorage.List(ctx, ItemVariationFilter{
		IDs: variationIDs,
	})
	if err != nil {
		return false
	}

	for _, v := range variations {
		if v.MerchantID != merchant.ID {
			return false
		}
	}

	return true
}
