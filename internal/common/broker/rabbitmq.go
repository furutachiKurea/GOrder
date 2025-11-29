package broker

import (
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/sirupsen/logrus"
)

// Connect 连接到 RabbitMQ 并创建 Exchange
func Connect(user, password, host, port string) (ch *amqp.Channel, closeCoon func() error) {
	addr := fmt.Sprintf("amqp://%s:%s@%s:%s/", user, password, host, port)
	coon, err := amqp.Dial(addr)
	if err != nil {
		logrus.Fatal(err)
	}

	ch, err = coon.Channel()
	if err != nil {
		logrus.Fatal(err)
	}

	if err := ch.ExchangeDeclare(
		EventOrderCreated, amqp.ExchangeDirect,
		true, false, false, false, nil,
	); err != nil {
		logrus.Fatal(err)
	}

	if err := ch.ExchangeDeclare(
		EventOrderPaid, amqp.ExchangeFanout,
		true, false, false, false, nil,
	); err != nil {
		logrus.Fatal(err)
	}

	return ch, coon.Close
}
