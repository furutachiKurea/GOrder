package command

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/furutachiKurea/gorder/common/broker"
	"github.com/furutachiKurea/gorder/common/decorator"
	"github.com/furutachiKurea/gorder/common/genproto/orderpb"
	"github.com/furutachiKurea/gorder/order/app/query"
	domain "github.com/furutachiKurea/gorder/order/domain/order"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/sirupsen/logrus"
)

type CreateOrder struct {
	CustomerID string
	Items      []*orderpb.ItemWithQuantity
}

type CreateOrderResult struct {
	OrderID string
}

type CreateOrderHandler decorator.CommandHandler[CreateOrder, *CreateOrderResult]

type createOrderHandler struct {
	orderRepo domain.Repository
	stockGRPC query.StockInterface
	channel   *amqp.Channel
}

func NewCreateOrderHandler(
	orderRepo domain.Repository,
	stockGRPC query.StockInterface,
	channel *amqp.Channel,
	logger *logrus.Entry,
	metricsClient decorator.MetricsClient,
) CreateOrderHandler {
	if orderRepo == nil {
		panic("orderRepo is nil")
	}

	if stockGRPC == nil {
		panic("stockGRPC is nil")
	}

	if channel == nil {
		panic("channel is nil")
	}
	return decorator.ApplyCommandDecorators[CreateOrder, *CreateOrderResult](
		createOrderHandler{
			orderRepo: orderRepo,
			stockGRPC: stockGRPC,
			channel:   channel,
		},
		logger,
		metricsClient,
	)
}

func (c createOrderHandler) Handle(ctx context.Context, cmd CreateOrder) (*CreateOrderResult, error) {
	validItems, err := c.validate(ctx, cmd.Items)
	if err != nil {
		return nil, fmt.Errorf("validate order items: %w", err)
	}

	order, err := c.orderRepo.Create(ctx, &domain.Order{
		CustomerID: cmd.CustomerID,
		Items:      validItems,
	})
	if err != nil {
		return nil, fmt.Errorf("create order: %w", err)
	}

	// send order created event to RabbitMQ
	q, err := c.channel.QueueDeclare(broker.EventOrderCreated, true, false, false, false, nil)
	if err != nil {
		return nil, err
	}

	marshalledOrder, err := json.Marshal(order)
	if err != nil {
		return nil, err
	}

	err = c.channel.PublishWithContext(ctx, "", q.Name, false, false, amqp.Publishing{
		ContentType:  "application/json",
		DeliveryMode: amqp.Persistent,
		Body:         marshalledOrder,
	})
	if err != nil {
		return nil, err
	}

	return &CreateOrderResult{
		OrderID: order.ID,
	}, nil

}

// validate 校验订单是否合法，合并商品数量，检查库存后返回 Item
func (c createOrderHandler) validate(ctx context.Context, items []*orderpb.ItemWithQuantity) ([]*orderpb.Item, error) {
	if len(items) == 0 {
		return nil, errors.New("must have at least one item")
	}

	items = packItems(items)

	resp, err := c.stockGRPC.CheckIfItemsInStock(ctx, items)
	if err != nil {
		return nil, err
	}
	return resp.Items, nil
}

// packItems 合并相同商品的数量
func packItems(items []*orderpb.ItemWithQuantity) []*orderpb.ItemWithQuantity {
	merged := make(map[string]int32)
	for _, item := range items {
		merged[item.Id] += item.Quantity
	}

	var packed []*orderpb.ItemWithQuantity
	for id, quantity := range merged {
		packed = append(packed, &orderpb.ItemWithQuantity{
			Id:       id,
			Quantity: quantity,
		})
	}

	return packed
}
