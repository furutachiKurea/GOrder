package command

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/furutachiKurea/gorder/common/broker"
	"github.com/furutachiKurea/gorder/common/decorator"
	"github.com/furutachiKurea/gorder/order/app/query"
	"github.com/furutachiKurea/gorder/order/convertor"
	domain "github.com/furutachiKurea/gorder/order/domain/order"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/otel"
	"google.golang.org/grpc/status"
)

type CreateOrder struct {
	CustomerID string
	Items      []*domain.ItemWithQuantity
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
	logger zerolog.Logger,
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
	q, err := c.channel.QueueDeclare(broker.EventOrderCreated, true, false, false, false, nil)
	if err != nil {
		return nil, err
	}

	t := otel.Tracer("rabbitmq")
	ctx, span := t.Start(ctx, fmt.Sprintf("rabbitmq.%s.publish", q.Name))
	defer span.End()

	validItems, err := c.validate(ctx, cmd.Items)
	if err != nil {
		return nil, fmt.Errorf("validate order items: %w", err)
	}

	log.Debug().
		Int("validItems", len(validItems)).
		Msg("get valid items for stock")

	pendingOrder, err := domain.NewPendingOrder(cmd.CustomerID, validItems)
	if err != nil {
		return nil, err
	}
	order, err := c.orderRepo.Create(ctx, pendingOrder)
	if err != nil {
		return nil, fmt.Errorf("create order: %w", err)
	}

	// TODO orderpb 生成的 Order struct tag 与 Go 默认序列化生成的 json 字段名不同，需要转换成 proto 版本的 Order 再序列化。
	marshalledOrder, err := json.Marshal(convertor.NewOrderConvertor().DomainToProto(order))
	if err != nil {
		return nil, err
	}

	header := broker.InjectRabbitMQHeaders(ctx)

	// send order created event to RabbitMQ
	err = c.channel.PublishWithContext(ctx, "", q.Name, false, false, amqp.Publishing{
		ContentType:  "application/json",
		DeliveryMode: amqp.Persistent,
		Body:         marshalledOrder,
		Headers:      header,
	})
	if err != nil {
		return nil, fmt.Errorf("publish event error q.Name=%s, err:%w", q.Name, err)
	}

	return &CreateOrderResult{
		OrderID: order.ID,
	}, nil

}

// validate 校验订单是否合法，合并商品数量，检查库存后返回 Item
func (c createOrderHandler) validate(ctx context.Context, items []*domain.ItemWithQuantity) ([]*domain.Item, error) {
	if len(items) == 0 {
		return nil, errors.New("must have at least one item")
	}

	items = packItems(items)

	log.Debug().Any("items", items).Msg("packed items")
	resp, err := c.stockGRPC.CheckIfItemsInStock(ctx, convertor.NewItemWithQuantityConvertor().DomainsToProtos(items))
	if err != nil {
		return nil, status.Convert(err).Err()
	}

	if len(resp.Items) == 0 {
		return nil, errors.New("no valid items in order")
	}
	return convertor.NewItemConvertor().ProtosToDomains(resp.Items), nil
}

// packItems 合并相同商品的数量
func packItems(items []*domain.ItemWithQuantity) []*domain.ItemWithQuantity {
	merged := make(map[string]int32)
	for _, item := range items {
		merged[item.Id] += item.Quantity
	}

	var packed []*domain.ItemWithQuantity
	for id, quantity := range merged {
		packed = append(packed, &domain.ItemWithQuantity{
			Id:       id,
			Quantity: quantity,
		})
	}

	return packed
}
