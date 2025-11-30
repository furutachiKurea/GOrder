package service

import (
	"context"

	"github.com/furutachiKurea/gorder/common/broker"
	grpcclient "github.com/furutachiKurea/gorder/common/client"
	"github.com/furutachiKurea/gorder/common/metrics"
	"github.com/furutachiKurea/gorder/order/adapter"
	"github.com/furutachiKurea/gorder/order/adapter/grpc"
	"github.com/furutachiKurea/gorder/order/app"
	"github.com/furutachiKurea/gorder/order/app/command"
	"github.com/furutachiKurea/gorder/order/app/query"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

func NewApplication(ctx context.Context) (app app.Application, close func()) {
	stockClient, closeStockClient, err := grpcclient.NewStockGRPCClient(ctx)
	if err != nil {
		panic(err)
	}
	stockGRPC := grpc.NewStockGRPC(stockClient)

	ch, closeCoon := broker.Connect(
		viper.GetString("rabbitmq.user"),
		viper.GetString("rabbitmq.password"),
		viper.GetString("rabbitmq.host"),
		viper.GetString("rabbitmq.port"),
	)

	return newApplication(ctx, stockGRPC, ch), func() {
		_ = closeStockClient()
		_ = ch.Close()
		_ = closeCoon()
	}

}

func newApplication(_ context.Context, stockClient query.StockInterface, ch *amqp.Channel) app.Application {
	orderRepo := adapter.NewMemoryOrderRepository()
	logger := log.Logger
	metricsClient := metrics.TodoMetrics{}
	return app.Application{
		Commands: app.Commands{
			CreateOrder: command.NewCreateOrderHandler(
				orderRepo,
				stockClient,
				ch,
				logger,
				metricsClient,
			),
			UpdateOrder: command.NewUpdateOrderHandler(
				orderRepo,
				logger,
				metricsClient,
			),
		},
		Queries: app.Queries{
			GetCustomerOrder: query.NewGetCustomerOrderHandler(
				orderRepo,
				logger,
				metricsClient,
			),
		},
	}

}
