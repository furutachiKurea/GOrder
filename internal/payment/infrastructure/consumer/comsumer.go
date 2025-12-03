package consumer

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/furutachiKurea/gorder/common/broker"
	"github.com/furutachiKurea/gorder/common/genproto/orderpb"
	"github.com/furutachiKurea/gorder/payment/app"
	"github.com/furutachiKurea/gorder/payment/app/command"
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
	q, err := ch.QueueDeclare(broker.EventOrderCreated, true, false, false, false, nil)
	if err != nil {
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

// handleMessage 处理接收到的订单创建消息，创建支付链接
func (c *Consumer) handleMessage(ch *amqp.Channel, msg amqp.Delivery, q amqp.Queue) {
	log.Info().
		Str("msg", string(msg.Body)).
		Msgf("payment received message from %s", q.Name)

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

	o := &orderpb.Order{}
	if err := json.Unmarshal(msg.Body, o); err != nil {
		log.Info().
			Err(err).
			Msg("failed to unmarshal msg to body")
		return
	}

	log.Debug().Any("unmarshalled_order", o).Msg("unmarshalled order from message")
	_, err = c.app.Commands.CreatePayment.Handle(ctx, command.CreatePayment{Order: o})
	if err != nil {
		log.Info().
			Err(err).
			Msg("failed to create payment")

		if err = broker.HandlerRetry(ctx, ch, &msg); err != nil {
			log.Warn().Err(err).
				Str("message_id", msg.MessageId).
				Msg("retry_error, error handling retry")
			return
		}
		return
	}

	span.AddEvent("payment.created")

	log.Info().Msg("consume success")
}
