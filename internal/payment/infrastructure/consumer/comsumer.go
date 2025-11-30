package consumer

import (
	"github.com/furutachiKurea/gorder/common/broker"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/rs/zerolog/log"
)

type Consumer struct {
}

func NewConsumer() *Consumer {
	return &Consumer{}
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

func (c *Consumer) handleMessage(msg amqp.Delivery, q amqp.Queue, ch *amqp.Channel) {
	log.Info().
		Str("msg", string(msg.Body)).
		Msgf("Payment received message from %s", q.Name)

	_ = msg.Ack(false)
}
