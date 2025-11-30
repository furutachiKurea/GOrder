package broker

import (
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/rs/zerolog/log"
)

// Connect 连接到 RabbitMQ 并创建 Exchange
func Connect(user, password, host, port string) (ch *amqp.Channel, closeCoon func() error) {
	addr := fmt.Sprintf("amqp://%s:%s@%s:%s/", user, password, host, port)
	coon, err := amqp.Dial(addr)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to RabbitMQ")
	}

	ch, err = coon.Channel()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to get open RabbitMQ channel")
	}

	if err := ch.ExchangeDeclare(
		EventOrderCreated, amqp.ExchangeDirect,
		true, false, false, false, nil,
	); err != nil {
		log.Fatal().Err(err).Str("exchange", EventOrderCreated).Msg("failed to declare exchange")
	}

	if err := ch.ExchangeDeclare(
		EventOrderPaid, amqp.ExchangeFanout,
		true, false, false, false, nil,
	); err != nil {
		log.Fatal().Err(err).Str("exchange", EventOrderPaid).Msg("failed to declare exchange")
	}

	return ch, coon.Close
}
