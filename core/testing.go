package core

import (
	"context"
)

type mockOrderStorage struct {
	PutFn  func(context.Context, Order) error
	GetFn  func(context.Context, ID) (Order, error)
	ListFn func(context.Context, OrderQuery) ([]Order, error)
}

func NewMockOrderStorage() *mockOrderStorage {
	return &mockOrderStorage{}
}

func (s *mockOrderStorage) Put(ctx context.Context, order Order) error {
	return s.PutFn(ctx, order)
}

func (s *mockOrderStorage) Get(ctx context.Context, id ID) (Order, error) {
	return s.GetFn(ctx, id)
}

func (s *mockOrderStorage) List(ctx context.Context, f OrderQuery) ([]Order, error) {
	return s.ListFn(ctx, f)
}

type mockItemVariationStorage struct {
	PutFn      func(context.Context, ItemVariation) error
	PutBatchFn func(context.Context, []ItemVariation) error
	GetFn      func(context.Context, ID) (ItemVariation, error)
	ListFn     func(context.Context, ItemVariationQuery) ([]ItemVariation, error)
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

func (m *mockItemVariationStorage) Get(ctx context.Context, id ID) (ItemVariation, error) {
	return m.GetFn(ctx, id)
}

func (m *mockItemVariationStorage) List(ctx context.Context, fil ItemVariationQuery) ([]ItemVariation, error) {
	return m.ListFn(ctx, fil)
}

type mockTaxStorage struct {
	PutFn      func(context.Context, Tax) error
	PutBatchFn func(context.Context, []Tax) error
	GetFn      func(context.Context, ID) (Tax, error)
	ListFn     func(context.Context, TaxQuery) ([]Tax, error)
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

func (m *mockTaxStorage) Get(ctx context.Context, id ID) (Tax, error) {
	return m.GetFn(ctx, id)
}

func (m *mockTaxStorage) List(ctx context.Context, fil TaxQuery) ([]Tax, error) {
	return m.ListFn(ctx, fil)
}

type mockDiscountStorage struct {
	PutFn      func(context.Context, Discount) error
	PutBatchFn func(context.Context, []Discount) error
	GetFn      func(context.Context, ID) (Discount, error)
	ListFn     func(context.Context, DiscountQuery) ([]Discount, error)
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

func (m *mockDiscountStorage) Get(ctx context.Context, id ID) (Discount, error) {
	return m.GetFn(ctx, id)
}

func (m *mockDiscountStorage) List(ctx context.Context, fil DiscountQuery) ([]Discount, error) {
	return m.ListFn(ctx, fil)
}

type mockCategoryStorage struct {
	PutFn      func(context.Context, Category) error
	PutBatchFn func(context.Context, []Category) error
	GetFn      func(context.Context, ID) (Category, error)
	ListFn     func(context.Context, CategoryQuery) ([]Category, error)
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

func (m *mockCategoryStorage) Get(ctx context.Context, id ID) (Category, error) {
	return m.GetFn(ctx, id)
}

func (m *mockCategoryStorage) List(ctx context.Context, fil CategoryQuery) ([]Category, error) {
	return m.ListFn(ctx, fil)
}

type mockItemStorage struct {
	PutFn      func(context.Context, Item) error
	PutBatchFn func(context.Context, []Item) error
	GetFn      func(context.Context, ID) (Item, error)
	ListFn     func(context.Context, ItemQuery) ([]Item, error)
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

func (m *mockItemStorage) Get(ctx context.Context, id ID) (Item, error) {
	return m.GetFn(ctx, id)
}

func (m *mockItemStorage) List(ctx context.Context, fil ItemQuery) ([]Item, error) {
	return m.ListFn(ctx, fil)
}
