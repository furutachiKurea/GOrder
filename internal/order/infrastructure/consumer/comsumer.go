package consumer

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/furutachiKurea/gorder/common/broker"
	"github.com/furutachiKurea/gorder/order/app"
	"github.com/furutachiKurea/gorder/order/app/command"
	domain "github.com/furutachiKurea/gorder/order/domain/order"
	"go.opentelemetry.io/otel"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/rs/zerolog/log"
)

type Consumer struct {
	app app.Application
}

func NewConsumer(app app.Application) *Consumer {
	return &Consumer{
		app: app,
	}
}

func (c *Consumer) Listen(ch *amqp.Channel) {
	q, err := ch.QueueDeclare(broker.EventOrderPaid, true, false, true, false, nil)
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

// handleMessage 处理接收到的订单创建消息，并将更新后的订单状态储存到数据库
func (c *Consumer) handleMessage(ch *amqp.Channel, msg amqp.Delivery, q amqp.Queue) {
	log.Info().
		Str("msg", string(msg.Body)).
		Msgf("order received message from %s", q.Name)

	ctx := broker.ExtractRabbitMQHeaders(context.Background(), msg.Headers)
	t := otel.Tracer("rabbitmq")
	ctx, span := t.Start(ctx, fmt.Sprintf("rabbitmq.%s.consume", q.Name))
	defer span.End()

	var err error
	defer func() {
		if err != nil {
			_ = msg.Nack(false, false)
		} else {
			_ = msg.Ack(false)
		}
	}()

	o := &domain.Order{}
	if err := json.Unmarshal(msg.Body, o); err != nil {
		log.Info().
			Err(err).
			Msg("failed to unmarshal msg.body to domain.order")

		return
	}

	// TODO 这块从 mq 中获取到的订单信息缺失 paymentLink，但是此时订单已经支付成功，应该可以放着不管或者说支付完的订单不需要支付链接
	log.Debug().Any("unmarshalled_order", o).Msg("unmarshalled order from message")
	_, err = c.app.Commands.UpdateOrder.Handle(ctx, command.UpdateOrder{
		Order: o,
		UpdateFn: func(ctx context.Context, order *domain.Order) (*domain.Order, error) {
			if err = order.IsPaid(); err != nil {
				return nil, err
			}

			return order, nil
		},
	})
	if err != nil {
		log.Info().
			Err(err).
			Msgf("error updating order, orderID = %s", o.ID)

		if err = broker.HandlerRetry(ctx, ch, &msg); err != nil {
			log.Warn().Err(err).
				Str("message_id", msg.MessageId).
				Msg("retry_error, error handling retry")
			return
		}

		return
	}

	span.AddEvent("order.updated")

	log.Info().Msg("order consume success paid event success")
}
