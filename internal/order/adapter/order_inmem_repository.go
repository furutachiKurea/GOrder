package adapter

import (
	"context"
	"strconv"
	"sync"
	"time"

	"github.com/furutachiKurea/gorder/common/entity"
	domain "github.com/furutachiKurea/gorder/order/domain/order"
	"github.com/rs/zerolog/log"
)

var stub = []*entity.Order{
	{
		ID:          "fake-ID",
		CustomerID:  "fake-customer-id",
		Status:      "fake-status",
		PaymentLink: "fake-payment-link",
		Items:       nil,
	},
}

type MemoryOrderRepository struct {
	lock  *sync.RWMutex
	store []*entity.Order
}

func NewMemoryOrderRepository() *MemoryOrderRepository {
	return &MemoryOrderRepository{
		lock:  &sync.RWMutex{},
		store: stub,
	}
}

func (m *MemoryOrderRepository) Create(_ context.Context, order *entity.Order) (*entity.Order, error) {
	m.lock.Lock()
	defer m.lock.Unlock()

	newOrder := &entity.Order{
		ID:          strconv.FormatInt(time.Now().UnixNano(), 10),
		CustomerID:  order.CustomerID,
		Status:      order.Status,
		PaymentLink: order.PaymentLink,
		Items:       order.Items,
	}

	m.store = append(m.store, newOrder)

	// Debug 转换 store 内容为值类型切片
	if log.Debug().Enabled() {
		storeValues := make([]entity.Order, len(m.store))
		for i, o := range m.store {
			storeValues[i] = *o
		}

		log.Debug().
			Any("input_order", order).
			Any("new_order", newOrder).
			Any("store_after_create", storeValues).
			Msg("memory_order_repo_create")
	}

	return newOrder, nil
	/*			updatedOrder, err := updateFn(ctx, order)
				if err != nil {
					return fmt.Errorf("memory order repository update: %w", err)
				}
	*/
}

func (m *MemoryOrderRepository) Get(_ context.Context, orderID, customerID string) (*entity.Order, error) {
	m.lock.RLock()
	defer m.lock.RUnlock()

	for _, o := range m.store {
		if o.ID == orderID && o.CustomerID == customerID {
			log.Debug().Msgf("memory_order_repo_get||found||id=%s||customID=%s||res=%+v", orderID, customerID, *o)
			return o, nil
		}
	}

	return nil, domain.NotFoundError{OrderID: orderID}
}

func (m *MemoryOrderRepository) Update(ctx context.Context, updates *entity.Order) error {
	m.lock.Lock()
	defer m.lock.Unlock()

	isFound := false
	for i, o := range m.store {
		if o.ID == updates.ID && o.CustomerID == updates.CustomerID {
			isFound = true
			m.store[i] = updates
		}
	}

	if !isFound {
		return domain.NotFoundError{OrderID: updates.ID}
	}

	return nil
}
