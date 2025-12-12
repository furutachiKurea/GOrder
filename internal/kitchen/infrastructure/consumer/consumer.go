package consumer

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/furutachiKurea/gorder/common/broker"
	"github.com/furutachiKurea/gorder/common/genproto/orderpb"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/otel"
)

type OrderService interface {
	UpdateOrder(context.Context, *orderpb.Order) error
}

type Consumer struct {
	orderGRPC OrderService
}

type Order struct {
	ID          string
	CustomerID  string
	Status      string
	PaymentLink string
	Items       []*orderpb.Item
}

func NewConsumer(orderService OrderService) *Consumer {
	return &Consumer{
		orderGRPC: orderService,
	}
}

func (c *Consumer) Listen(ch *amqp.Channel) {
	q, err := ch.QueueDeclare("", true, false, true, false, nil)
	if err != nil {
		log.Fatal().Err(err).Msg("")
	}

	if err = ch.QueueBind(q.Name, "", broker.EventOrderPaid, false, nil); err != nil {
		log.Fatal().Err(err).Msg("")
	}

	msgs, err := ch.Consume(q.Name, "", false, false, false, false, nil)
	if err != nil {
		log.Warn().Err(err).Str("queue", q.Name).Msg("failed to consume queue")
	}

	var forever chan struct{}
	go func() {
		for msg := range msgs {
			c.handleMessage(ch, msg, q)
		}
	}()

	<-forever
}

// handleMessage 处理接收到的订单支付消息，制作订单并更新状态为 ready
func (c *Consumer) handleMessage(ch *amqp.Channel, msg amqp.Delivery, q amqp.Queue) {
	var err error
	log.Info().
		Str("msg", string(msg.Body)).
		Msgf("kitchen received message from %s", q.Name)

	ctx := broker.ExtractRabbitMQHeaders(context.Background(), msg.Headers)
	t := otel.Tracer("rabbitmq")
	ctx, span := t.Start(ctx, fmt.Sprintf("rabbitmq.%s.consume", q.Name))

	defer func() {
		span.End()
		if err != nil {
			_ = msg.Nack(false, false)
		} else {
			_ = msg.Ack(false)
		}
	}()

	o := &Order{}
	if err = json.Unmarshal(msg.Body, o); err != nil {
		log.Info().
			Err(err).
			Msg("failed to unmarshal msg to body")
		return
	}

	if o.Status != "paid" {
		err = errors.New("order not paid, cannot cook")
		return
	}

	cook(o)
	span.AddEvent(fmt.Sprintf("order_cook: %v", &o))
	// TODO 这里会和 order 的 consumer 状态更新冲突，如果 kitchen 先完成，
	//  order 服务更新状态会覆盖 kitchen 的更新，此时仍然为 paid 状态。
	//  不过现在的 kitchen 会 sleep 五秒，暂时不会出现这个问题
	if err = c.orderGRPC.UpdateOrder(ctx, &orderpb.Order{
		Id:          o.ID,
		CustomerId:  o.CustomerID,
		Status:      "ready",
		PaymentLink: o.PaymentLink,
		Items:       o.Items,
	}); err != nil {
		if err = broker.HandlerRetry(ctx, ch, &msg); err != nil {
			log.Warn().Err(err).Msg("kitchen: error handling retry")
		}
		return
	}

	log.Info().Msg("kitchen.order.finished.updated")
}

func cook(o *Order) {
	log.Info().Str("order", o.ID).Msg("cooking order")
	time.Sleep(5 * time.Second)
	log.Info().Str("order", o.ID).Msg("order done!")
}
