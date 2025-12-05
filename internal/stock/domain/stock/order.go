package stock

import (
	"errors"
	"fmt"

	"github.com/stripe/stripe-go/v84"
)

type Order struct {
	ID          string
	CustomerID  string
	Status      string
	PaymentLink string
	Items       []*Item
}

func NewOrder(id, customerID, status, paymentLink string, items []*Item) (*Order, error) {
	if id == "" {
		return nil, errors.New("empty id")
	}

	if customerID == "" {
		return nil, errors.New("empty customerID")
	}

	if status == "" {
		return nil, errors.New("empty status")
	}

	if len(items) == 0 {
		return nil, errors.New("items cannot be nil or empty")
	}

	return &Order{
		ID:          id,
		CustomerID:  customerID,
		Status:      status,
		PaymentLink: paymentLink,
		Items:       items,
	}, nil
}

// NewPendingOrder 创建一个待支付的订单，作为 payment 对新建订单进行消费前的状态,
// 刚创建的订单状态为 "pending"
func NewPendingOrder(customerID string, items []*Item) (*Order, error) {
	if customerID == "" {
		return nil, errors.New("empty customerID")
	}

	if len(items) == 0 {
		return nil, errors.New("items cannot be nil or empty")
	}

	return &Order{
		CustomerID: customerID,
		Status:     "pending",
		Items:      items,
	}, nil
}

func (o *Order) IsPaid() error {
	if o.Status == string(stripe.CheckoutSessionPaymentStatusPaid) {
		return nil
	}

	return fmt.Errorf("order status not paid, order_id=%s, status=%s", o.ID, o.Status)
}

type Item struct {
	Id       string
	Name     string
	Quantity int32
	PriceID  string
}

type ItemWithQuantity struct {
	Id       string
	Quantity int32
}
