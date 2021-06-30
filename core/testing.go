package core

import (
	"context"
)

type mockOrderStorage struct {
	PutFn  func(context.Context, Order) error
	GetFn  func(context.Context, string, string, []string) (Order, error)
	ListFn func(context.Context, OrderFilter) ([]Order, error)
}

func NewMockOrderStorage() *mockOrderStorage {
	return &mockOrderStorage{}
}

func (s *mockOrderStorage) Put(ctx context.Context, order Order) error {
	return s.PutFn(ctx, order)
}

func (s *mockOrderStorage) Get(ctx context.Context, id, merchantID string, locationIDs []string) (Order, error) {
	return s.GetFn(ctx, id, merchantID, locationIDs)
}

func (s *mockOrderStorage) List(ctx context.Context, f OrderFilter) ([]Order, error) {
	return s.ListFn(ctx, f)
}

type mockItemVariationStorage struct {
	PutFn      func(context.Context, ItemVariation) error
	PutBatchFn func(context.Context, []ItemVariation) error
	GetFn      func(context.Context, string, string, []string) (ItemVariation, error)
	ListFn     func(context.Context, ItemVariationFilter) ([]ItemVariation, error)
}

func NewMockItemVariationStorage() *mockItemVariationStorage {
	return &mockItemVariationStorage{}
}

func (m *mockItemVariationStorage) Put(ctx context.Context, itvar ItemVariation) error {
	return m.PutFn(ctx, itvar)
}

func (m *mockItemVariationStorage) PutBatch(ctx context.Context, batch []ItemVariation) error {
	return m.PutBatchFn(ctx, batch)
}

func (m *mockItemVariationStorage) Get(ctx context.Context, id, merchantID string, locationIDs []string) (ItemVariation, error) {
	return m.GetFn(ctx, id, merchantID, locationIDs)
}

func (m *mockItemVariationStorage) List(ctx context.Context, fil ItemVariationFilter) ([]ItemVariation, error) {
	return m.ListFn(ctx, fil)
}

type mockTaxStorage struct {
	PutFn      func(context.Context, Tax) error
	PutBatchFn func(context.Context, []Tax) error
	GetFn      func(context.Context, string, string, []string) (Tax, error)
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

func (m *mockTaxStorage) Get(ctx context.Context, id, merchantID string, locationIDs []string) (Tax, error) {
	return m.GetFn(ctx, id, merchantID, locationIDs)
}

func (m *mockTaxStorage) List(ctx context.Context, fil TaxFilter) ([]Tax, error) {
	return m.ListFn(ctx, fil)
}

type mockDiscountStorage struct {
	PutFn      func(context.Context, Discount) error
	PutBatchFn func(context.Context, []Discount) error
	GetFn      func(context.Context, string) (Discount, error)
	ListFn     func(context.Context, DiscountFilter) ([]Discount, error)
}

func NewMockDiscountStorage() *mockDiscountStorage {
	return &mockDiscountStorage{}
}

func (m *mockDiscountStorage) Put(ctx context.Context, d Discount) error {
	return m.PutFn(ctx, d)
}

func (m *mockDiscountStorage) PutBatch(ctx context.Context, batch []Discount) error {
	return m.PutBatchFn(ctx, batch)
}

func (m *mockDiscountStorage) Get(ctx context.Context, id, merchantID string, locationIDs []string) (Discount, error) {
	return m.Get(ctx, id, merchantID, locationIDs)
}

func (m *mockDiscountStorage) List(ctx context.Context, fil DiscountFilter) ([]Discount, error) {
	return m.ListFn(ctx, fil)
}

type mockCategoryStorage struct {
	PutFn      func(context.Context, Category) error
	PutBatchFn func(context.Context, []Category) error
	GetFn      func(context.Context, string, string, []string) (Category, error)
	ListFn     func(context.Context, CategoryFilter) ([]Category, error)
}

func NewMockCategoryStorage() *mockCategoryStorage {
	return &mockCategoryStorage{}
}

func (m *mockCategoryStorage) Put(ctx context.Context, t Category) error {
	return m.PutFn(ctx, t)
}

func (m *mockCategoryStorage) PutBatch(ctx context.Context, batch []Category) error {
	return m.PutBatchFn(ctx, batch)
}

func (m *mockCategoryStorage) Get(ctx context.Context, id, merchantID string, locationIDs []string) (Category, error) {
	return m.GetFn(ctx, id, merchantID, locationIDs)
}

func (m *mockCategoryStorage) List(ctx context.Context, fil CategoryFilter) ([]Category, error) {
	return m.ListFn(ctx, fil)
}

type mockItemStorage struct {
	PutFn      func(context.Context, Item) error
	PutBatchFn func(context.Context, []Item) error
	GetFn      func(context.Context, string, string, []string) (Item, error)
	ListFn     func(context.Context, ItemFilter) ([]Item, error)
}

func NewMockItemStorage() *mockItemStorage {
	return &mockItemStorage{}
}

func (m *mockItemStorage) Put(ctx context.Context, t Item) error {
	return m.PutFn(ctx, t)
}

func (m *mockItemStorage) PutBatch(ctx context.Context, batch []Item) error {
	return m.PutBatchFn(ctx, batch)
}

func (m *mockItemStorage) Get(ctx context.Context, id, merchantID string, locationIDs []string) (Item, error) {
	return m.GetFn(ctx, id, merchantID, locationIDs)
}

func (m *mockItemStorage) List(ctx context.Context, fil ItemFilter) ([]Item, error) {
	return m.ListFn(ctx, fil)
}
