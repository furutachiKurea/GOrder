package order

import "context"

type Repository interface {
	Create(context.Context, *Order) (*Order, error)
	Get(ctx context.Context, orderID, customerID string) (*Order, error)
	// Update 更新订单，updateFn 接收 existing order , 应返回 order 的期望状态
	//
	// TODO 现在只接收需要更新的 order，而不返回现有的 order，导致每次更新必然会返回所有字段，
	//  尽管业务上符合 consumer 的需求，但是已经导致了 consumer 在更新支付状态时使得 repository 中的 payment_link 丢失。
	//  设计问题，优先级稍低
	Update(
		ctx context.Context,
		order *Order,
		updateFn func(ctx context.Context, update *Order) (*Order, error),
	) error
}

type NotFoundError struct {
	OrderID string
}

func (e NotFoundError) Error() string {
	return "order " + e.OrderID + " not found"
}
