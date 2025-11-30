package consumer

import (
	"context"
	"encoding/json"

	"github.com/furutachiKurea/gorder/common/broker"
	"github.com/furutachiKurea/gorder/common/genproto/orderpb"
	"github.com/furutachiKurea/gorder/payment/app"
	"github.com/furutachiKurea/gorder/payment/app/command"

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
			c.handleMessage(msg, q, ch)
		}
	}()

	<-forever
}

// handleMessage 处理接收到的订单创建消息，创建支付链接
func (c *Consumer) handleMessage(msg amqp.Delivery, q amqp.Queue, ch *amqp.Channel) {
	log.Info().
		Str("msg", string(msg.Body)).
		Msgf("payment received message from %s", q.Name)

	o := &orderpb.Order{}
	if err := json.Unmarshal(msg.Body, o); err != nil {
		log.Info().
			Err(err).
			Msg("failed to unmarshal msg to body")
		_ = msg.Nack(false, false)
		return
	}

	log.Debug().Any("unmarshalled_order", o).Msg("unmarshalled order from message")
	_, err := c.app.Commands.CreatePayment.Handle(context.TODO(), command.CreatePayment{Order: o})
	if err != nil {
		// TODO 重试
		log.Info().
			Err(err).
			Msg("failed to create order payment")
		_ = msg.Nack(false, false)
		return
	}

	_ = msg.Ack(false)
	log.Info().Msg("consume success")
}
