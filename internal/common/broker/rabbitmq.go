package broker

import (
	"context"
	"fmt"
	"time"

	_ "github.com/furutachiKurea/gorder/common/config"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"go.opentelemetry.io/otel"
)

const (
	DLX                = "dlx"
	DLQ                = "dlq"
	amqpRetryHeaderKey = "x-retry-count"
)

var (
	maxRetryCount int64 = viper.GetInt64("rabbitmq.max-retry")
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

	if err = createDLX(ch); err != nil {
		log.Fatal().Err(err).Msg("failed to create dlx")
	}

	return ch, coon.Close
}

func createDLX(ch *amqp.Channel) error {
	q, err := ch.QueueDeclare("share_queue", true, false, false, false, nil)
	if err != nil {
		return err
	}

	err = ch.ExchangeDeclare(DLX, amqp.ExchangeFanout, true, false, false, false, nil)
	if err != nil {
		return err
	}

	err = ch.QueueBind(q.Name, "", DLX, false, nil)
	if err != nil {
		return err
	}

	_, err = ch.QueueDeclare(DLQ, true, false, false, false, nil)
	return err
}

func HandlerRetry(ctx context.Context, ch *amqp.Channel, d *amqp.Delivery) error {
	if d.Headers == nil {
		d.Headers = make(amqp.Table)
	}

	retryCount, ok := d.Headers[amqpRetryHeaderKey].(int64)
	if !ok {
		retryCount = 0
	}

	retryCount++
	d.Headers[amqpRetryHeaderKey] = retryCount

	if retryCount > maxRetryCount {
		log.Info().Str("message_id", d.MessageId).Msg("moving message to dlq")
		return ch.PublishWithContext(ctx, "", DLQ, false, false, amqp.Publishing{
			Headers:      d.Headers,
			ContentType:  "application/json",
			Body:         d.Body,
			DeliveryMode: amqp.Persistent,
		})
	}

	log.Info().Str("message_id", d.MessageId).Int64("retry_count", retryCount).Msg("retrying message")
	time.Sleep(time.Second * time.Duration(retryCount))
	return ch.PublishWithContext(ctx, d.Exchange, d.RoutingKey, false, false, amqp.Publishing{
		Headers:      d.Headers,
		ContentType:  "application/json",
		Body:         d.Body,
		DeliveryMode: amqp.Persistent,
	})
}

type RabbitMQHeaderCarrier map[string]any

func (r RabbitMQHeaderCarrier) Get(key string) string {
	if v, ok := r[key]; ok {
		return v.(string)
	}

	return ""
}

func (r RabbitMQHeaderCarrier) Set(key, value string) {
	r[key] = value
}

func (r RabbitMQHeaderCarrier) Keys() []string {
	keys := make([]string, 0, len(r))

	for key := range r {
		keys = append(keys, key)
	}

	return keys
}

func InjectRabbitMQHeaders(ctx context.Context) map[string]any {
	carrier := make(RabbitMQHeaderCarrier)
	otel.GetTextMapPropagator().Inject(ctx, &carrier)
	return carrier
}

func ExtractRabbitMQHeaders(ctx context.Context, headers map[string]any) context.Context {
	return otel.GetTextMapPropagator().Extract(ctx, RabbitMQHeaderCarrier(headers))
}
