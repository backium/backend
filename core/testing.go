package core

import (
	"context"
)

type mockOrderStorage struct {
	PutFn func(context.Context, Order) error
	GetFn func(context.Context, string) (Order, error)
}

func NewMockOrderStorage() *mockOrderStorage {
	return &mockOrderStorage{}
}

func (s *mockOrderStorage) Put(ctx context.Context, order Order) error {
	return s.PutFn(ctx, order)
}

func (s *mockOrderStorage) Get(ctx context.Context, id string) (Order, error) {
	return s.GetFn(ctx, id)
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

type mockTaxStorage struct {
	PutFn      func(context.Context, Tax) error
	PutBatchFn func(context.Context, []Tax) error
	GetFn      func(context.Context, string) (Tax, error)
	ListFn     func(context.Context, TaxFilter) ([]Tax, error)
}

func NewMockTaxStorage() *mockTaxStorage {
	return &mockTaxStorage{}
}

func (m *mockTaxStorage) Put(ctx context.Context, t Tax) error {
	return m.PutFn(ctx, t)
}

func (m *mockTaxStorage) PutBatch(ctx context.Context, batch []Tax) error {
	return m.PutBatchFn(ctx, batch)
}

func (m *mockTaxStorage) Get(ctx context.Context, id string) (Tax, error) {
	return m.Get(ctx, id)
}

func (m *mockTaxStorage) List(ctx context.Context, fil TaxFilter) ([]Tax, error) {
	return m.ListFn(ctx, fil)
}
