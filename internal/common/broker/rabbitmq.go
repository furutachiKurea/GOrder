package broker

import (
	"context"
	"errors"
	"fmt"
	"time"

	_ "github.com/furutachiKurea/gorder/common/config"
	"github.com/furutachiKurea/gorder/common/logging"

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

	if err = ch.ExchangeDeclare(
		EventOrderCreated, amqp.ExchangeDirect,
		true, false, false, false, nil,
	); err != nil {
		log.Fatal().Err(err).Str("exchange", EventOrderCreated).Msg("failed to declare exchange")
	}

	if err = ch.ExchangeDeclare(
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

func HandlerRetry(ctx context.Context, ch *amqp.Channel, d *amqp.Delivery) (err error) {
	var retryCount int64

	l, deferlog := logging.WhenRequest(ctx, "HandleRetry", map[string]any{
		"delivery":        d,
		"max_retry_count": maxRetryCount,
	})
	defer func() {
		deferlog(nil, &err)
	}()

	log.Info().Ctx(ctx).
		Any("delivery", d).
		Msg("handle_retry_start")
	if d.Headers == nil {
		d.Headers = make(amqp.Table)
	}

	retryCount, ok := d.Headers[amqpRetryHeaderKey].(int64)
	if !ok {
		retryCount = 0
	}

	retryCount++
	d.Headers[amqpRetryHeaderKey] = retryCount
	l = l.With().Int64("retry_count", retryCount).Logger()

	if retryCount > maxRetryCount {
		log.Info().Ctx(ctx).Str("message_id", d.MessageId).Msg("moving message to dlq")
		err = doPublish(ctx, ch, "", DLQ, false, false, amqp.Publishing{
			Headers:      d.Headers,
			ContentType:  "application/json",
			Body:         d.Body,
			DeliveryMode: amqp.Persistent,
		})
		if err != nil {
			err = fmt.Errorf("publish to dlq: %w", err)
			return err
		}
		return errors.New("max retry count exceeded, message moved to dlq")
	}

	log.Debug().Ctx(ctx).Str("message_id", d.MessageId).Int64("retry_count", retryCount).Msg("retrying message")
	time.Sleep(time.Second * time.Duration(retryCount))
	return doPublish(ctx, ch, d.Exchange, d.RoutingKey, false, false, amqp.Publishing{
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

// InjectRabbitMQHeaders 将 context 注入 RabbitMQ header 以进行分布式追踪
func InjectRabbitMQHeaders(ctx context.Context) map[string]any {
	carrier := make(RabbitMQHeaderCarrier)
	otel.GetTextMapPropagator().Inject(ctx, &carrier)
	return carrier
}

func ExtractRabbitMQHeaders(ctx context.Context, headers map[string]any) context.Context {
	return otel.GetTextMapPropagator().Extract(ctx, RabbitMQHeaderCarrier(headers))
}
