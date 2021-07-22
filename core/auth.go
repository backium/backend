package core

import (
	"context"
)

type Authorizer struct {
	ItemVariationStorage ItemVariationStorage
	ItemStorage          ItemStorage
	CategoryStorage      CategoryStorage
	EmployeeStorage      EmployeeStorage
}

func (auth *Authorizer) CanCreateEmployee(ctx context.Context, empl Employee) bool {
	merchant := MerchantFromContext(ctx)
	employee := EmployeeFromContext(ctx)

	return empl.MerchantID == merchant.ID && employee.IsOwner
}

func (auth *Authorizer) CanUpdateEmployee(ctx context.Context, id ID) bool {
	merchant := MerchantFromContext(ctx)
	employee := EmployeeFromContext(ctx)

	if !employee.IsOwner {
		return false
	}

	empl, err := auth.EmployeeStorage.Get(ctx, id)
	if err != nil {
		return true
	}

	return empl.MerchantID == merchant.ID
}

func (auth *Authorizer) CanGetItem(ctx context.Context, id ID) bool {
	merchant := MerchantFromContext(ctx)
	employee := EmployeeFromContext(ctx)

	if !(Can(employee.Permissions, CatalogRead) || employee.IsOwner) {
		return false
	}

	item, err := auth.ItemStorage.Get(ctx, id)
	if err != nil {
		return true
	}

	return item.MerchantID == merchant.ID &&
		(ContainsOneID(item.LocationIDs, employee.LocationIDs) || employee.IsOwner)
}

func (auth *Authorizer) CanCreateItem(ctx context.Context, item Item) bool {
	merchant := MerchantFromContext(ctx)
	employee := EmployeeFromContext(ctx)

	if !(Can(employee.Permissions, CatalogWrite) || employee.IsOwner) {
		return false
	}

	return item.MerchantID == merchant.ID &&
		(ContainsAllID(employee.LocationIDs, item.LocationIDs) || employee.IsOwner)
}

func (auth *Authorizer) CanUpdateItem(ctx context.Context, id ID) bool {
	merchant := MerchantFromContext(ctx)
	employee := EmployeeFromContext(ctx)

	if !(Can(employee.Permissions, CatalogWrite) || employee.IsOwner) {
		return false
	}

	item, err := auth.ItemStorage.Get(ctx, id)
	if err != nil {
		return true
	}

	return item.MerchantID == merchant.ID
}

func (auth *Authorizer) CanSearchItem(ctx context.Context, f ItemFilter) bool {
	merchant := MerchantFromContext(ctx)
	employee := EmployeeFromContext(ctx)

	if !(Can(employee.Permissions, CatalogRead) || employee.IsOwner) {
		return false
	}

	items, _, _ := auth.ItemStorage.List(ctx, ItemQuery{
		Filter: f,
	})

	for _, item := range items {
		if item.MerchantID != merchant.ID {
			return false
		}
		if !ContainsOneID(employee.LocationIDs, item.LocationIDs) && !employee.IsOwner {
			return false
		}
	}

	return true
}

func (auth *Authorizer) CanChangeInventory(ctx context.Context, variationIDs []ID) bool {
	merchant := MerchantFromContext(ctx)

	variations, _, err := auth.ItemVariationStorage.List(ctx, ItemVariationQuery{
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
