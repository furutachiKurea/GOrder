package command

import (
	"context"

	"github.com/furutachiKurea/gorder/common/decorator"
	"github.com/furutachiKurea/gorder/common/genproto/orderpb"
	"github.com/furutachiKurea/gorder/common/logging"
	"github.com/furutachiKurea/gorder/common/tracing"
	"github.com/furutachiKurea/gorder/payment/domain"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type CreatePayment struct {
	Order *orderpb.Order
}

type CreatePaymentHandler decorator.CommandHandler[CreatePayment, string]

type createPaymentHandler struct {
	processor domain.Processor
	orderGRPC OrderService
}

func NewCreatePaymentHandler(
	processor domain.Processor,
	orderGRPC OrderService,
	logger zerolog.Logger,
	metricsClient decorator.MetricsClient,
) CreatePaymentHandler {
	if processor == nil {
		panic("processor is nil")
	}

	if orderGRPC == nil {
		panic("orderGRPC is nil")
	}

	return decorator.ApplyCommandDecorators[CreatePayment, string](
		createPaymentHandler{
			processor: processor,
			orderGRPC: orderGRPC,
		},
		logger,
		metricsClient,
	)
}

// Handle 创建支付链接并更新订单状态为等待支付，返回支付链接，如果更新订单是失败会同时返回 error
func (c createPaymentHandler) Handle(ctx context.Context, cmd CreatePayment) (string, error) {
	var err error
	defer logging.WhenCommandExecute(ctx, "CreatePaymentHandler", cmd, err)

	ctx, span := tracing.Start(ctx, "createPaymentHandle")
	defer span.End()

	link, err := c.processor.CreatePaymentLink(ctx, cmd.Order)
	if err != nil {
		return "", err
	}

	log.Info().Ctx(ctx).
		Str("payment_link", link).
		Any("order_id", cmd.Order.Id).
		Msg("create payment link for order")

	newOrder := &orderpb.Order{
		Id:          cmd.Order.Id,
		CustomerId:  cmd.Order.CustomerId,
		Status:      "waiting_for_payment",
		Items:       cmd.Order.Items,
		PaymentLink: link,
	}

	log.Debug().Any("new_order", newOrder).Msg("updating order with payment link")

	err = c.orderGRPC.UpdateOrder(ctx, newOrder)
	return link, err
}
