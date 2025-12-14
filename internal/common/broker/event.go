package broker

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/furutachiKurea/gorder/common/logging"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/rs/zerolog/log"
)

const (
	EventOrderCreated = "order.created"
	EventOrderPaid    = "order.paid"
)

type RoutingType string

const (
	Direct = "direct"
	FanOut = "fan-out"
)

type PublishEventReq struct {
	Channel  *amqp.Channel
	Routing  RoutingType
	Queue    string
	Exchange string
	Body     any
}

func PublishEvent(ctx context.Context, req *PublishEventReq) (err error) {
	_, deferlog := logging.WhenEventPublish(ctx, req)
	defer deferlog(nil, &err)

	if err = checkParam(req); err != nil {
		return err
	}

	switch req.Routing {
	case Direct:
		return directQueue(ctx, req)
	case FanOut:
		return fanOutQueue(ctx, req)
	default:
		log.Panic().Ctx(ctx).Str("routing_type", string(req.Routing)).Msg("unsupported_routing_type")
	}
	return nil
}

func directQueue(ctx context.Context, req *PublishEventReq) error {
	_, err := req.Channel.QueueDeclare(req.Queue, true, false, false, false, nil)
	if err != nil {
		return err
	}

	jsonBody, err := json.Marshal(req.Body)
	if err != nil {
		return fmt.Errorf("marshalling body in publish event: %w", err)
	}

	return doPublish(ctx, req.Channel, req.Exchange, req.Queue, false, false, amqp.Publishing{
		ContentType:  "application/json",
		DeliveryMode: amqp.Persistent,
		Body:         jsonBody,
		Headers:      InjectRabbitMQHeaders(ctx),
	})
}

func fanOutQueue(ctx context.Context, req *PublishEventReq) error {
	jsonBody, err := json.Marshal(req.Body)
	if err != nil {
		return fmt.Errorf("marshalling body in publish event: %w", err)
	}

	return doPublish(ctx, req.Channel, req.Exchange, "", false, false, amqp.Publishing{
		ContentType:  "application/json",
		DeliveryMode: amqp.Persistent,
		Body:         jsonBody,
		Headers:      InjectRabbitMQHeaders(ctx),
	})

}

// doPublish -
func doPublish(
	ctx context.Context,
	ch *amqp.Channel,
	exchange, key string,
	mandatory, immediate bool,
	msg amqp.Publishing,
) error {
	if err := ch.PublishWithContext(ctx, exchange, key, mandatory, immediate, msg); err != nil {
		log.Warn().Ctx(ctx).Msgf("_publish_event_failed||exchange=%s||key=%s||msg=%v", exchange, key, msg)
		return fmt.Errorf("publish event: %w", err)
	}
	return nil
}

func checkParam(r *PublishEventReq) error {
	if r.Channel == nil {
		return errors.New("nil channel")
	}
	return nil
}
