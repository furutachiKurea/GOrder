package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/furutachiKurea/gorder/common/broker"
	"github.com/furutachiKurea/gorder/common/genproto/orderpb"
	"github.com/furutachiKurea/gorder/payment/domain"

	"github.com/gin-gonic/gin"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"github.com/stripe/stripe-go/v84"
	"github.com/stripe/stripe-go/v84/webhook"
	"go.opentelemetry.io/otel"
)

type PaymentHandler struct {
	channel *amqp.Channel
}

func NewPaymentHandler(ch *amqp.Channel) *PaymentHandler {
	return &PaymentHandler{channel: ch}
}

func (h PaymentHandler) RegisterRoutes(router *gin.Engine) {
	router.POST("/api/webhook", h.handleWebhook)
}

// handleWebhook handles Stripe webhook events，并将支付成功的订单信息发布到消息队列
func (h PaymentHandler) handleWebhook(c *gin.Context) {
	var err error
	defer func() {
		if err != nil {
			log.Error().Ctx(c.Request.Context()).Err(err).Msg("handlerWebhook error")
		} else {
			log.Info().Ctx(c.Request.Context()).Msg("handlerWebhook success")
		}
	}()

	const MaxBodyBytes = int64(65536)
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, MaxBodyBytes)
	payload, err := io.ReadAll(c.Request.Body)
	if err != nil {
		err = fmt.Errorf("reading request body: %w", err)
		c.JSON(http.StatusServiceUnavailable, err.Error())
		return
	}

	event, err := webhook.ConstructEvent(payload, c.Request.Header.Get("Stripe-Signature"),
		viper.GetString("endpoint-stripe-secret"))
	if err != nil {
		err = fmt.Errorf("verifying webhook signature: %w", err)
		c.JSON(http.StatusBadRequest, err.Error()) // Return a 400 error on a bad signature
		return
	}

	if err = json.Unmarshal(payload, &event); err != nil {
		err = fmt.Errorf("unmarshalling webhook body json: %w", err)
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}

	// Unmarshal the event data into an appropriate struct depending on its Type
	switch event.Type {
	case stripe.EventTypeCheckoutSessionCompleted:
		var session stripe.CheckoutSession
		if err = json.Unmarshal(event.Data.Raw, &session); err != nil {
			err = fmt.Errorf("unmarshalling event.Data.Raw into session: %w", err)
			c.JSON(http.StatusBadRequest, err.Error())
			return
		}

		if session.PaymentStatus == stripe.CheckoutSessionPaymentStatusPaid {
			ctx, cancel := context.WithCancel(c)
			defer cancel()

			var items []*orderpb.Item
			_ = json.Unmarshal([]byte(session.Metadata["items"]), &items)

			marshalledOrder, err := json.Marshal(&domain.Order{
				ID:         session.Metadata["order_id"],
				CustomerID: session.Metadata["customer_id"],
				Status:     string(session.PaymentStatus),
				Items:      items,
			})
			if err != nil {
				err = fmt.Errorf("marshalling domain.order: %w", err)
				c.JSON(http.StatusInternalServerError, err.Error())
				return
			}

			t := otel.Tracer("rabbitmq")
			mqCtx, span := t.Start(ctx, fmt.Sprintf("rabbitmq.%s.publish", broker.EventOrderPaid))
			defer span.End()

			headers := broker.InjectRabbitMQHeaders(mqCtx)
			_ = h.channel.PublishWithContext(mqCtx, broker.EventOrderPaid, "", false, false,
				amqp.Publishing{
					ContentType:  "application/json",
					DeliveryMode: amqp.Persistent,
					Body:         marshalledOrder,
					Headers:      headers,
				})
			log.Info().Ctx(mqCtx).
				Str("message_body", string(marshalledOrder)).
				Msgf("message published to %s", broker.EventOrderPaid)
		}
	default:
		log.Warn().Ctx(c.Request.Context()).Str("event type", string(event.Type)).Msg("Unhandled event type")
	}

	c.JSON(http.StatusOK, nil)
}
