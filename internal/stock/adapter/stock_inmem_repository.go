package adapter

import (
	"context"
	"sync"

	"github.com/furutachiKurea/gorder/common/genproto/orderpb"
	domain "github.com/furutachiKurea/gorder/stock/domain/stock"
)

// stub data
var stub = map[string]*orderpb.Item{
	"item_id": &orderpb.Item{
		Id:       "foo_item",
		Name:     "stub_item",
		Quantity: 100000000,
		PriceId:  "price_id",
	},
}

type MemoryStockRepository struct {
	lock  *sync.RWMutex
	store map[string]*orderpb.Item
}

func NewMemoryOrderRepository() *MemoryStockRepository {
	return &MemoryStockRepository{
		lock:  &sync.RWMutex{},
		store: stub,
	}
}

func (m MemoryStockRepository) GetItems(ctx context.Context, ids []string) ([]*orderpb.Item, error) {
	m.lock.RLock()
	defer m.lock.RUnlock()

	var (
		res        []*orderpb.Item
		missingIds []string
	)

	for _, id := range ids {
		if _, ok := m.store[id]; ok {
			res = append(res, stub[id])
		} else {
			missingIds = append(missingIds, id)
		}
	}

	if len(res) != len(ids) {
		return res, domain.NotFoundError{Missing: missingIds} // TODO res 不应该被使用或者为 nil
	}

	return res, nil
}
