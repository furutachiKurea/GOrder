package consumer

import (
	"context"
	"encoding/json"

	"github.com/furutachiKurea/gorder/common/broker"
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
			c.handleMessage(msg, q)
		}
	}()

	<-forever
}

// handleMessage 处理接收到的订单创建消息，创建支付链接
func (c *Consumer) handleMessage(msg amqp.Delivery, q amqp.Queue) {
	log.Info().
		Str("msg", string(msg.Body)).
		Msgf("order received message from %s", q.Name)

	o := &domain.Order{}
	if err := json.Unmarshal(msg.Body, o); err != nil {
		log.Info().
			Err(err).
			Msg("failed to unmarshal msg.body to domain.order")

		_ = msg.Nack(false, false)
		return
	}

	log.Debug().Any("unmarshalled_order", o).Msg("unmarshalled order from message")
	_, err := c.app.Commands.UpdateOrder.Handle(context.TODO(), command.UpdateOrder{
		Order: o,
		UpdateFn: func(ctx context.Context, order *domain.Order) (*domain.Order, error) {
			if err := order.IsPaid(); err != nil {
				return nil, err
			}

			return order, nil
		},
	})
	if err != nil {
		// TODO 重试
		log.Info().
			Err(err).
			Msgf("error updating order, orderID = %s", o.ID)
		_ = msg.Nack(false, false)
		return
	}

	_ = msg.Ack(false)
	log.Info().Msg("order consume success paid event success")
}
