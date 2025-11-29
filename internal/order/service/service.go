package service

import (
	"context"

	"github.com/furutachiKurea/gorder/common/metrics"
	"github.com/furutachiKurea/gorder/order/adapter"
	"github.com/furutachiKurea/gorder/order/app"
	"github.com/furutachiKurea/gorder/order/app/command"
	"github.com/furutachiKurea/gorder/order/app/query"

	"github.com/sirupsen/logrus"
)

func NewApplication(ctx context.Context) app.Application {
	orderRepo := adapter.NewMemoryOrderRepository()
	logger := logrus.NewEntry(logrus.StandardLogger())
	metricsClient := metrics.TodoMetrics{}
	return app.Application{
		Commands: app.Commands{
			CreateOrder: command.NewCreateOrderHandler(
				orderRepo,
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
