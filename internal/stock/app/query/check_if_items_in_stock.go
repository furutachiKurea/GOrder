package query

import (
	"context"

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
	if err := h.checkStock(ctx, query.Items); err != nil {
		return nil, err
	}

	// 从 stripe 获取 priceID
	var res []*domain.Item
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

	// TODO 库存扣减
	return res, nil
}

// checkStock 检查库存是否充足，该方法会处理重复的 item，调用时可以不用合并 item
func (h checkIfItemsInStockHandler) checkStock(ctx context.Context, query []*domain.ItemWithQuantity) error {
	var ids []string
	for _, item := range query {
		ids = append(ids, item.Id)
	}

	records, err := h.stockRepo.GetStock(ctx, ids)
	if err != nil {
		return err
	}

	idQuantityMap := make(map[string]int32)
	for _, r := range records {
		idQuantityMap[r.Id] = r.Quantity
	}

	var failedOn []struct {
		ID   string
		Want int32
		Have int32
	}

	// 合并相同商品数量
	query = packItems(query)

	for _, item := range query {
		if item.Quantity > idQuantityMap[item.Id] {
			failedOn = append(failedOn, struct {
				ID   string
				Want int32
				Have int32
			}{
				ID:   item.Id,
				Want: item.Quantity,
				Have: idQuantityMap[item.Id],
			})
		}
	}

	if len(failedOn) > 0 {
		return domain.ExceedStockError{FailedOn: failedOn}
	}
	return nil
}

// packItems 合并相同商品的数量
func packItems(items []*domain.ItemWithQuantity) []*domain.ItemWithQuantity {
	merged := make(map[string]int32)
	for _, item := range items {
		merged[item.Id] += item.Quantity
	}

	var packed []*domain.ItemWithQuantity
	for id, quantity := range merged {
		packed = append(packed, &domain.ItemWithQuantity{
			Id:       id,
			Quantity: quantity,
		})
	}

	return packed
}
