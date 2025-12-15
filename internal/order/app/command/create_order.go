package command

import (
	"context"
	"errors"
	"fmt"

	"github.com/furutachiKurea/gorder/common/broker"
	"github.com/furutachiKurea/gorder/common/convertor"
	"github.com/furutachiKurea/gorder/common/decorator"
	"github.com/furutachiKurea/gorder/common/entity"
	"github.com/furutachiKurea/gorder/common/logging"
	"github.com/furutachiKurea/gorder/order/app/client"
	domain "github.com/furutachiKurea/gorder/order/domain/order"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/otel"
	"google.golang.org/grpc/status"
)

type CreateOrder struct {
	CustomerID string
	Items      []*entity.ItemWithQuantity
}

type CreateOrderResult struct {
	OrderID string
}

// CreateOrderHandler 创建订单，校验库存后发布订单创建事件到 RabbitMQ
type CreateOrderHandler decorator.CommandHandler[CreateOrder, *CreateOrderResult]

type createOrderHandler struct {
	orderRepo domain.Repository
	stockGRPC client.StockService
	channel   *amqp.Channel
}

func NewCreateOrderHandler(
	orderRepo domain.Repository,
	stockGRPC client.StockService,
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
	var err error
	defer logging.WhenCommandExecute(ctx, "CreateOrderHandler", cmd, err)

	t := otel.Tracer("rabbitmq")
	ctx, span := t.Start(ctx, fmt.Sprintf("rabbitmq.%s.publish", broker.EventOrderPaid))
	defer span.End()

	validItems, err := c.validate(ctx, cmd.Items)
	if err != nil {
		return nil, err
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
	log.Debug().Ctx(ctx).Any("order", order).Msg("create order in repository")
	err = broker.PublishEvent(ctx, &broker.PublishEventReq{
		Channel:  c.channel,
		Routing:  broker.Direct,
		Queue:    broker.EventOrderCreated,
		Exchange: "",
		Body:     order,
	})
	if err != nil {
		return nil, fmt.Errorf("publish event error q.Name=%s, err:%w", broker.EventOrderPaid, err)
	}

	return &CreateOrderResult{
		OrderID: order.ID,
	}, nil

}

// validate 校验订单是否合法，合并商品数量，库存充足并正确预扣库存后返回订单 Item
func (c createOrderHandler) validate(ctx context.Context, items []*entity.ItemWithQuantity) ([]*entity.Item, error) {
	if len(items) == 0 {
		return nil, errors.New("must have at least one item")
	}

	items = packItems(items)

	log.Debug().Any("items", items).Msg("packed items")
	resp, err := c.stockGRPC.ReserveStock(ctx, convertor.NewItemWithQuantityConvertor().EntitiesToProtos(items))
	if err != nil {
		return nil, fmt.Errorf("reserve stock:%w", status.Convert(err).Err())
	}

	if len(resp.Items) == 0 {
		return nil, errors.New("no valid items in order")
	}
	return convertor.NewItemConvertor().ProtosToEntities(resp.Items), nil
}

// packItems 合并相同商品的数量
func packItems(items []*entity.ItemWithQuantity) []*entity.ItemWithQuantity {
	merged := make(map[string]int64)
	for _, item := range items {
		merged[item.ID] += item.Quantity
	}

	var packed []*entity.ItemWithQuantity
	for id, quantity := range merged {
		packed = append(packed, &entity.ItemWithQuantity{
			ID:       id,
			Quantity: quantity,
		})
	}

	return packed
}
