package consts

type OrderStatus string

const (
	OrderStatusPending           OrderStatus = "pending"
	OrderStatusWaitingForPayment OrderStatus = "waiting_for_payment"
	OrderStatusPaid              OrderStatus = "paid"
	OrderStatusReady             OrderStatus = "ready"
)
