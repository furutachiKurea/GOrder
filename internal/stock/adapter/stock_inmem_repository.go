package adapter

import (
	"context"
	"sync"

	"github.com/furutachiKurea/gorder/common/entity"
	domain "github.com/furutachiKurea/gorder/stock/domain/stock"
)

var stub = map[string]*entity.Item{
	"item_id": {
		Id:       "foo_item",
		Name:     "stub_item",
		Quantity: 100000000,
		PriceID:  "price_id",
	},
	"item1": {
		Id:       "item1",
		Name:     "stub item 1",
		Quantity: 1000000,
		PriceID:  "stub_item1_price_id",
	},
	"item2": {
		Id:       "item2",
		Name:     "stub item 2",
		Quantity: 1000000,
		PriceID:  "stub_item2_price_id",
	},
	"item3": {
		Id:       "item3",
		Name:     "stub item 3",
		Quantity: 1000000,
		PriceID:  "stub_item3_price_id",
	},
}

// Deprecated: use StockRepositoryMySQL instead.
type MemoryStockRepository struct {
	lock  *sync.RWMutex
	store map[string]*entity.Item
}

// Deprecated: use StockRepositoryMySQL.ReserveStock.
func (m MemoryStockRepository) ReserveStock(ctx context.Context, items []*entity.ItemWithQuantity) error {
	// TODO implement me
	panic("implement me")
}

// Deprecated: use NewStockRepositoryMySQL with persistent.MySQL.
func NewMemoryStockRepository() *MemoryStockRepository {
	return &MemoryStockRepository{
		lock:  &sync.RWMutex{},
		store: stub,
	}
}

// Deprecated: use StockRepositoryMySQL.GetItems.
func (m MemoryStockRepository) GetItems(ctx context.Context, ids []string) ([]*entity.Item, error) {
	m.lock.RLock()
	defer m.lock.RUnlock()

	var (
		res        []*entity.Item
		missingIDs []string
	)

	for _, id := range ids {
		if _, ok := m.store[id]; ok {
			res = append(res, stub[id])
		} else {
			missingIDs = append(missingIDs, id)
		}
	}

	if len(res) != len(ids) {
		return nil, domain.NotFoundError{Missing: missingIDs}
	}

	return res, nil
}

// Deprecated: use StockRepositoryMySQL.GetStock.
func (m MemoryStockRepository) GetStock(ctx context.Context, ids []string) ([]*entity.ItemWithQuantity, error) {
	// TODO implement me
	panic("implement me")
}
