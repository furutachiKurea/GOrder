package query

import (
	"context"

	"github.com/furutachiKurea/gorder/common/decorator"
	domain "github.com/furutachiKurea/gorder/stock/domain/stock"

	"github.com/rs/zerolog"
)

type GetItems struct {
	ItemIDs []string
}

type GetItemsHandler decorator.QueryHandler[GetItems, []*domain.Item]

type getItemsHandler struct {
	stockRepo domain.Repository
}

func NewGetItemsHandler(
	stockRepo domain.Repository,
	logger zerolog.Logger,
	metricsClient decorator.MetricsClient,
) GetItemsHandler {
	if stockRepo == nil {
		panic("stockRepo is nil")
	}

	return decorator.ApplyCommandDecorators[GetItems, []*domain.Item](
		getItemsHandler{stockRepo: stockRepo},
		logger,
		metricsClient,
	)
}

func (g getItemsHandler) Handle(ctx context.Context, query GetItems) ([]*domain.Item, error) {
	return g.stockRepo.GetItems(ctx, query.ItemIDs)
}
