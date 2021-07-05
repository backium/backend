package core

import (
	"context"
)

type Authorizer struct {
	ItemVariationStorage ItemVariationStorage
	ItemStorage          ItemStorage
	CategoryStorage      CategoryStorage
}

func (auth *Authorizer) CanGetItem(ctx context.Context, id ID) bool {
	merchant := MerchantFromContext(ctx)

	item, err := auth.ItemStorage.Get(ctx, id)
	if err != nil {
		return true
	}

	return item.MerchantID == merchant.ID
}

func (auth *Authorizer) CanCreateItem(ctx context.Context, item Item) bool {
	merchant := MerchantFromContext(ctx)

	return item.MerchantID == merchant.ID
}

func (auth *Authorizer) CanUpdateItem(ctx context.Context, id ID) bool {
	merchant := MerchantFromContext(ctx)

	item, err := auth.ItemStorage.Get(ctx, id)
	if err != nil {
		return true
	}

	return item.MerchantID == merchant.ID
}

func (auth *Authorizer) CanSearchItem(ctx context.Context, f ItemFilter) bool {
	merchant := MerchantFromContext(ctx)

	items, _ := auth.ItemStorage.List(ctx, ItemQuery{
		Filter: f,
	})

	for _, item := range items {
		if item.MerchantID != merchant.ID {
			return false
		}
	}

	return true
}

func (auth *Authorizer) CanChangeInventory(ctx context.Context, variationIDs []ID) bool {
	merchant := MerchantFromContext(ctx)

	variations, err := auth.ItemVariationStorage.List(ctx, ItemVariationQuery{
		Filter: ItemVariationFilter{IDs: variationIDs},
	})
	if err != nil {
		return true
	}

	for _, v := range variations {
		if v.MerchantID != merchant.ID {
			return false
		}
	}

	return true
}
