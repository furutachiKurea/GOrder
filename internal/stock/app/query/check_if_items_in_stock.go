package query

import (
	"context"
	"errors"

	"github.com/furutachiKurea/gorder/common/decorator"
	domain "github.com/furutachiKurea/gorder/stock/domain/stock"

	"github.com/rs/zerolog"
)

type CheckIfItemsInStock struct {
	Items []*domain.ItemWithQuantity
}

type CheckIfItemsInStockHandler decorator.QueryHandler[CheckIfItemsInStock, []*domain.Item]

type checkIfItemsInStockHandler struct {
	stockRepo     domain.Repository
	priceProvider PriceProvider
}

func NewCheckIfItemsInStockHandler(
	stockRepo domain.Repository,
	priceProvider PriceProvider,
	logger zerolog.Logger,
	metricsClient decorator.MetricsClient,
) CheckIfItemsInStockHandler {
	if stockRepo == nil {
		panic("stockRepo is nil")
	}
	if priceProvider == nil {
		panic("priceProvider is nil")
	}

	return decorator.ApplyQueryDecorators[CheckIfItemsInStock, []*domain.Item](
		checkIfItemsInStockHandler{
			stockRepo:     stockRepo,
			priceProvider: priceProvider,
		},
		logger,
		metricsClient,
	)
}

// Deprecated
var stub = map[string]string{
	"1": "price_1SZBTrKQ4HJNAIH7x1izqMNh",
	"2": "price_1SZBe7KQ4HJNAIH7CDXMbh0X",
}

func (h checkIfItemsInStockHandler) Handle(ctx context.Context, query CheckIfItemsInStock) ([]*domain.Item, error) {
	// TODO 当前只是单纯的用 query 构造 item，并没有实际检查库存，后续补充库存检查逻辑
	// 从数据库或者 stripe 获取
	var (
		res []*domain.Item
	)
	for _, item := range query.Items {
		priceID, err := h.priceProvider.GetPriceByProductID(ctx, item.Id)
		if err != nil || priceID == "" {
			return nil, err
		}

		res = append(res, &domain.Item{
			Id:       item.Id,
			Quantity: item.Quantity,
			PriceID:  priceID,
		})
	}
	if len(res) == 0 {
		return nil, errors.New("no items in stock or price not found")
	}

	return res, nil
}
