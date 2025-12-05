package service

import (
	"context"

	"github.com/furutachiKurea/gorder/common/metrics"
	"github.com/furutachiKurea/gorder/stock/adapter"
	"github.com/furutachiKurea/gorder/stock/app"
	"github.com/furutachiKurea/gorder/stock/app/query"
	"github.com/furutachiKurea/gorder/stock/infrastructure/integration"

	"github.com/rs/zerolog/log"
)

func NewApplication(_ context.Context) app.Application {
	stockRepo := adapter.NewMemoryStockRepository()
	stripeAPI := integration.NewStripeAPI()
	logger := log.Logger
	metricsClient := metrics.TodoMetrics{}
	return app.Application{
		Queries: app.Queries{
			GetItems: query.NewGetItemsHandler(
				stockRepo,
				logger,
				metricsClient,
			),
			CheckIfItemsInStock: query.NewCheckIfItemsInStockHandler(
				stockRepo,
				stripeAPI,
				logger,
				metricsClient,
			),
		},
	}
}
