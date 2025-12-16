package order

import (
	"errors"
	"fmt"

	"github.com/furutachiKurea/gorder/common/consts"
	"github.com/furutachiKurea/gorder/common/entity"

	"github.com/rs/zerolog/log"
	"github.com/stripe/stripe-go/v84"
)

type Order struct {
	ID          string
	CustomerID  string
	Status      consts.OrderStatus
	PaymentLink string
	Items       []*entity.Item
}

func (o *Order) ToProto() *entity.Order {
	items := make([]*entity.Item, len(o.Items))
	for i, item := range o.Items {
		items[i] = &entity.Item{
			ID:       item.ID,
			Name:     item.Name,
			Quantity: item.Quantity,
			PriceID:  item.PriceID,
		}
	}

	return &entity.Order{
		ID:          o.ID,
		CustomerID:  o.CustomerID,
		Status:      o.Status,
		PaymentLink: o.PaymentLink,
		Items:       items,
	}
}

func NewOrder(id, customerID, paymentLink string, status consts.OrderStatus, items []*entity.Item) (*Order, error) {
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
// 刚创建的订单状态为 consts.OrderStatusPending
func NewPendingOrder(customerID string, items []*entity.Item) (*Order, error) {
	if customerID == "" {
		return nil, errors.New("empty customerID")
	}

	if len(items) == 0 {
		return nil, errors.New("items cannot be nil or empty")
	}

	return &Order{
		CustomerID: customerID,
		Status:     consts.OrderStatusPending,
		Items:      items,
	}, nil
}

// UpdateTo 使用 order 的值更新 o, ID, CustomerID, Items 不可变
func (o *Order) UpdateTo(order *Order) (err error) {
	if order.Status != "" {
		err = o.UpdateStatusTo(order.Status)
		if err != nil {
			return err
		}
	}

	err = o.UpdatePaymentLink(order.PaymentLink)
	if err != nil {
		return err
	}

	return nil
}

// UpdateStatusTo 更新订单状态
func (o *Order) UpdateStatusTo(status consts.OrderStatus) error {
	if status == "" {
		return errors.New("order status cannot be empty")
	}

	notAllowedTrans := map[consts.OrderStatus][]consts.OrderStatus{
		consts.OrderStatusPending:           {},
		consts.OrderStatusWaitingForPayment: {consts.OrderStatusPending},
		consts.OrderStatusPaid:              {consts.OrderStatusPending, consts.OrderStatusWaitingForPayment},
		consts.OrderStatusReady:             {consts.OrderStatusPending, consts.OrderStatusWaitingForPayment, consts.OrderStatusPaid},
	}

	invalidStatus, ok := notAllowedTrans[o.Status]
	if !ok {
		// 应该不会发生
		log.Warn().Str("order_id", o.ID).Str("unknow_status", string(o.Status)).Msg("current order status is unknown")
		return fmt.Errorf("unknown current order status, order_id=%s, status=%s", o.ID, o.Status)
	}

	for _, notAllow := range invalidStatus {
		if status == notAllow {
			log.Warn().
				Str("order_id", o.ID).
				Str("current_status", string(o.Status)).
				Str("tried_status", string(status)).
				Msg("tried to update order status to an invalid status")
			return fmt.Errorf("update order status to %s from %s not allowed, order_id=%s", status, o.Status, o.ID)

		}
	}

	o.Status = status
	return nil
}

// UpdatePaymentLink 更新订单的支付链接，
func (o *Order) UpdatePaymentLink(paymentLink string) error {
	// 由于 domain.Repository 现在的设计会将传入的 updates 全盘更新给 order，
	// 这导致 UpdatesPaymentLink 会在传入的 order PaymentLink 为空时将原有的 PaymentLink 覆盖掉，
	// 这个情况会在支付完成后恰好被触发，我们暂且认为支付完成后移除 PaymentLink 是合理的，
	// 因此这里注释掉这个检查 💩
	// if paymentLink == "" {
	// 	return nil
	// }

	o.PaymentLink = paymentLink
	return nil
}

func (o *Order) IsPaid() error {
	if o.Status == consts.OrderStatus(stripe.CheckoutSessionPaymentStatusPaid) {
		return nil
	}

	return fmt.Errorf("order status not paid, order_id=%s, status=%s", o.ID, o.Status)
}
