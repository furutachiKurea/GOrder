package service

import (
	"context"

	"github.com/furutachiKurea/gorder/common/metrics"
	"github.com/furutachiKurea/gorder/stock/adapter"
	"github.com/furutachiKurea/gorder/stock/app"
	"github.com/furutachiKurea/gorder/stock/app/command"
	"github.com/furutachiKurea/gorder/stock/app/query"
	"github.com/furutachiKurea/gorder/stock/infrastructure/integration"
	"github.com/furutachiKurea/gorder/stock/infrastructure/persistent"

	"github.com/rs/zerolog/log"
)

func NewApplication(_ context.Context) app.Application {
	db := persistent.NewMySQL()
	stockRepo := adapter.NewStockRepositoryMySQL(db)
	stripeAPI := integration.NewStripeAPI()
	logger := log.Logger
	metricsClient := metrics.TodoMetrics{}
	return app.Application{
		Commands: app.Commands{
			ReserveStock: command.NewReserveStockHandler(
				stockRepo,
				stripeAPI,
				logger,
				metricsClient,
			),
		},
		Queries: app.Queries{
			GetItems: query.NewGetItemsHandler(
				stockRepo,
				logger,
				metricsClient,
			),
		},
	}
}
