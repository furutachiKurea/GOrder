package query

import (
	"context"

	"github.com/furutachiKurea/gorder/common/decorator"
	"github.com/furutachiKurea/gorder/common/genproto/orderpb"
	domain "github.com/furutachiKurea/gorder/stock/domain/stock"
	"github.com/rs/zerolog"
)

type CheckIfItemsInStock struct {
	Items []*orderpb.ItemWithQuantity
}

type CheckIfItemsInStockHandler decorator.QueryHandler[CheckIfItemsInStock, []*orderpb.Item]

type checkIfItemsInStockHandler struct {
	stockRepo domain.Repository
}

func NewCheckIfItemsInStockHandler(
	stockRepo domain.Repository,
	logger zerolog.Logger,
	metricsClient decorator.MetricsClient,
) CheckIfItemsInStockHandler {
	if stockRepo == nil {
		panic("stockRepo is nil")
	}

	return decorator.ApplyQueryDecorators[CheckIfItemsInStock, []*orderpb.Item](
		checkIfItemsInStockHandler{stockRepo: stockRepo},
		logger,
		metricsClient,
	)
}

var stub = map[string]string{
	"1": "price_1SZBTrKQ4HJNAIH7x1izqMNh",
	"2": "price_1SZBe7KQ4HJNAIH7CDXMbh0X",
}

func (c checkIfItemsInStockHandler) Handle(ctx context.Context, query CheckIfItemsInStock) ([]*orderpb.Item, error) {
	// TODO 当前只是单纯的用 query 构造 item，并没有实际检查库存，后续补充库存检查逻辑
	// 从数据库或者 stripe 获取
	var res []*orderpb.Item
	for _, item := range query.Items {
		priceID, ok := stub[item.Id]
		if !ok {
			priceID = stub["1"]
		}
		res = append(res, &orderpb.Item{
			Id:       item.Id,
			Quantity: item.Quantity,
			PriceId:  priceID,
		})
	}
	return res, nil
}
