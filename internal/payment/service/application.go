package service

import (
	"context"

	grpcclient "github.com/furutachiKurea/gorder/common/client"
	"github.com/furutachiKurea/gorder/common/metrics"
	"github.com/furutachiKurea/gorder/payment/adapter"
	"github.com/furutachiKurea/gorder/payment/app"
	"github.com/furutachiKurea/gorder/payment/app/command"
	"github.com/furutachiKurea/gorder/payment/domain"
	"github.com/furutachiKurea/gorder/payment/infrastructure/processor"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

func NewApplication(ctx context.Context) (app app.Application, close func()) {
	orderClient, closeOrderClient, err := grpcclient.NewOrderGRPCClient(ctx)
	if err != nil {
		panic(err)
	}

	orderGRPC := adapter.NewOderGRPC(orderClient)
	// memoryProcessor := processor.NewInmemProcessor()
	stripeProcessor := processor.NewStripeProcessor(viper.GetString("stripe-key"))

	return newApplication(ctx, stripeProcessor, orderGRPC), func() {
		_ = closeOrderClient()
	}
}

func newApplication(_ context.Context, processor domain.Processor, orderGRPC command.OrderService) app.Application {
	logger := log.Logger
	metricsClient := metrics.TodoMetrics{}
	return app.Application{
		Commands: app.Commands{
			CreatePayment: command.NewCreatePaymentHandler(
				processor,
				orderGRPC,
				logger,
				metricsClient,
			),
		},
	}
}
