package core

import (
	"context"
)

type mockOrderStorage struct {
	CreateFunc func(context.Context, Order) (string, error)
	OrderFunc  func(context.Context, string) (Order, error)
}

func NewMockOrderStorage() *mockOrderStorage {
	return &mockOrderStorage{}
}

func (s *mockOrderStorage) Create(ctx context.Context, order Order) (string, error) {
	return s.CreateFunc(ctx, order)
}

func (s *mockOrderStorage) Order(ctx context.Context, id string) (Order, error) {
	return s.OrderFunc(ctx, id)
}

type mockItemVariationStorage struct {
	CreateFunc        func(context.Context, ItemVariation) (string, error)
	UpdateFunc        func(context.Context, ItemVariation) error
	UpdatePartialFunc func(context.Context, string, ItemVariationPartial) error
	RetrieveFunc      func(context.Context, string) (ItemVariation, error)
	ListFunc          func(context.Context, ItemVariationFilter) ([]ItemVariation, error)
}

func NewMockItemVariationStorage() *mockItemVariationStorage {
	return &mockItemVariationStorage{}
}

func (m *mockItemVariationStorage) Create(ctx context.Context, itvar ItemVariation) (string, error) {
	return m.CreateFunc(ctx, itvar)
}

func (m *mockItemVariationStorage) Update(ctx context.Context, itvar ItemVariation) error {
	return m.UpdateFunc(ctx, itvar)
}

func (m *mockItemVariationStorage) UpdatePartial(ctx context.Context, id string, itvar ItemVariationPartial) error {
	return m.UpdatePartial(ctx, id, itvar)
}

func (m *mockItemVariationStorage) Retrieve(ctx context.Context, id string) (ItemVariation, error) {
	return m.Retrieve(ctx, id)
}

func (m *mockItemVariationStorage) List(ctx context.Context, fil ItemVariationFilter) ([]ItemVariation, error) {
	return m.ListFunc(ctx, fil)
}
