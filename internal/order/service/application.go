package service

import (
	"context"

	grpcclient "github.com/furutachiKurea/gorder/common/client"
	"github.com/furutachiKurea/gorder/common/metrics"
	"github.com/furutachiKurea/gorder/order/adapter"
	"github.com/furutachiKurea/gorder/order/adapter/grpc"
	"github.com/furutachiKurea/gorder/order/app"
	"github.com/furutachiKurea/gorder/order/app/command"
	"github.com/furutachiKurea/gorder/order/app/query"

	"github.com/sirupsen/logrus"
)

func NewApplication(ctx context.Context) (app app.Application, close func()) {
	stockClient, closeStockClient, err := grpcclient.NewStockGRPCClient(ctx)
	if err != nil {
		panic(err)
	}
	stockGRPC := grpc.NewStockGRPC(stockClient)

	return newApplication(ctx, stockGRPC), func() {
		_ = closeStockClient()
	}

}

func newApplication(_ context.Context, stockClient query.StockInterface) app.Application {
	orderRepo := adapter.NewMemoryOrderRepository()
	logger := logrus.NewEntry(logrus.StandardLogger())
	metricsClient := metrics.TodoMetrics{}
	return app.Application{
		Commands: app.Commands{
			CreateOrder: command.NewCreateOrderHandler(
				orderRepo,
				stockClient,
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
