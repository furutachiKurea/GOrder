package order

import "context"

type Repository interface {
	Create(context.Context, *Order) (*Order, error)
	Get(ctx context.Context, orderID, customerID string) (*Order, error)
	// Update 更新订单
	Update(ctx context.Context, updates *Order) error
}

type NotFoundError struct {
	OrderID string
}

func (e NotFoundError) Error() string {
	return "order " + e.OrderID + " not found"
}
