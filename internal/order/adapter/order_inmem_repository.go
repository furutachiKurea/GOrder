package adapter

import (
	"context"
	"fmt"
	"strconv"
	"sync"
	"time"

	domain "github.com/furutachiKurea/gorder/order/domain/order"
	"github.com/rs/zerolog/log"
)

type MemoryOrderRepository struct {
	lock  *sync.RWMutex
	store []*domain.Order
}

func NewMemoryOrderRepository() *MemoryOrderRepository {
	s := make([]*domain.Order, 0)
	s = append(s, &domain.Order{
		ID:          "fake-ID",
		CustomerID:  "fake-customer-id",
		Status:      "fake-status",
		PaymentLink: "fake-payment-link",
		Items:       nil,
	})
	return &MemoryOrderRepository{
		lock:  &sync.RWMutex{},
		store: s, // TODO remove hard code data, replace to make([]*domain.Order, 0),
	}
}

func (m *MemoryOrderRepository) Create(_ context.Context, order *domain.Order) (*domain.Order, error) {
	m.lock.Lock()
	defer m.lock.Unlock()

	newOrder := &domain.Order{
		ID:          strconv.FormatInt(time.Now().UnixNano(), 10),
		CustomerID:  order.CustomerID,
		Status:      order.Status,
		PaymentLink: order.PaymentLink,
		Items:       order.Items,
	}

	m.store = append(m.store, newOrder)

	// Debug 转换 store 内容为值类型切片
	if log.Debug().Enabled() {
		storeValues := make([]domain.Order, len(m.store))
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
}

func (m *MemoryOrderRepository) Get(_ context.Context, orderID, customerID string) (*domain.Order, error) {
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

func (m *MemoryOrderRepository) Update(ctx context.Context, order *domain.Order, updateFn func(context.Context, *domain.Order) (*domain.Order, error)) error {
	m.lock.Lock()
	defer m.lock.Unlock()

	isFound := false
	for i, o := range m.store {
		if o.ID == order.ID && o.CustomerID == order.CustomerID {
			isFound = true
			updatedOrder, err := updateFn(ctx, o)
			if err != nil {
				return fmt.Errorf("memory order repository update: %w", err)
			}

			m.store[i] = updatedOrder
		}
	}

	if !isFound {
		return domain.NotFoundError{OrderID: order.ID}
	}

	return nil
}
