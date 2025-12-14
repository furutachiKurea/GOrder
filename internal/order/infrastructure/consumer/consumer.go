package consumer

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/furutachiKurea/gorder/common/broker"
	"github.com/furutachiKurea/gorder/common/tracing"
	"github.com/furutachiKurea/gorder/order/app"
	"github.com/furutachiKurea/gorder/order/app/command"
	domain "github.com/furutachiKurea/gorder/order/domain/order"
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

// handleMessage 处理接收到的订单支付消息，更新订单状态并更新库存
func (c *Consumer) handleMessage(ch *amqp.Channel, msg amqp.Delivery, q amqp.Queue) {
	log.Info().
		Str("msg", string(msg.Body)).
		Msgf("order received message from %s", q.Name)

	ctx := broker.ExtractRabbitMQHeaders(context.Background(), msg.Headers)
	ctx, span := tracing.Start(ctx, fmt.Sprintf("rabbitmq.%s.consume", q.Name))
	defer span.End()

	var err error
	defer func() {
		if err != nil {
			_ = msg.Nack(false, false)
			log.Warn().Ctx(ctx).
				Err(err).
				Str("from", q.Name).
				Str("msg", string(msg.Body)).
				Msg("consume failed")
		} else {
			_ = msg.Ack(false)
			span.AddEvent("order.paid_confirmed")
			log.Info().Ctx(ctx).Msg("consume success")
		}
	}()

	o := &domain.Order{}
	if err = json.Unmarshal(msg.Body, o); err != nil {
		err = fmt.Errorf("unmarshal msg to body: %w", err)
		return
	}

	log.Debug().Any("unmarshalled_order", o).Msg("unmarshalled order from message")
	_, err = c.app.Commands.ConfirmOrderPaid.Handle(ctx, command.ConfirmOrderPaid{Order: o})
	if err != nil {
		err = fmt.Errorf("confirm order paid: %w", err)
		if err = broker.HandlerRetry(ctx, ch, &msg); err != nil {
			err = fmt.Errorf("handle retry, messageId=%s: %w", msg.MessageId, err)
		}
		return
	}
}
