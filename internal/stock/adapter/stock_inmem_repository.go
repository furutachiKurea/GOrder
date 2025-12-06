package adapter

import (
	"context"
	"sync"

	domain "github.com/furutachiKurea/gorder/stock/domain/stock"
)

// stub data
var stub = map[string]*domain.Item{
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

type MemoryStockRepository struct {
	lock  *sync.RWMutex
	store map[string]*domain.Item
}

func (m MemoryStockRepository) UpdateStock(ctx context.Context, data []*domain.ItemWithQuantity,
	updateFn func(ctx context.Context,
		existing []*domain.ItemWithQuantity,
		query []*domain.ItemWithQuantity,
	) ([]*domain.ItemWithQuantity, error),
) error {
	// TODO implement me
	panic("implement me")
}

func NewMemoryStockRepository() *MemoryStockRepository {
	return &MemoryStockRepository{
		lock:  &sync.RWMutex{},
		store: stub,
	}
}

func (m MemoryStockRepository) GetItems(ctx context.Context, ids []string) ([]*domain.Item, error) {
	m.lock.RLock()
	defer m.lock.RUnlock()

	var (
		res        []*domain.Item
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
		return res, domain.NotFoundError{Missing: missingIDs} // TODO res 不应该被使用或者为 nil
	}

	return res, nil
}

func (m MemoryStockRepository) GetStock(ctx context.Context, ids []string) ([]*domain.ItemWithQuantity, error) {
	// TODO implement me
	panic("implement me")
}
