package query

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/furutachiKurea/gorder/common/decorator"
	"github.com/furutachiKurea/gorder/common/handler/redis"
	domain "github.com/furutachiKurea/gorder/stock/domain/stock"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

const (
	redisLockPrefix = "check_stock_"
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
	if err := lock(ctx, getLockKey(query)); err != nil {
		return nil, fmt.Errorf("redis lock, key=%s: %w", getLockKey(query), err)
	}
	defer func() {
		if err := unlock(ctx, getLockKey(query)); err != nil {
			log.Warn().Err(err).Msg("redis unlock fail")
		}
	}()

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

	// TODO 在查询库存是否足够之后就扣除库存，没有处理当用户没有实际支付成功的情况
	if err := h.checkStock(ctx, query.Items); err != nil {
		return nil, err
	}

	return res, nil
}

func lock(ctx context.Context, key string) error {
	return redis.SetNX(ctx, redis.LocalClient(), key, "1", 5*time.Minute)
}

func unlock(ctx context.Context, key string) error {
	return redis.Del(ctx, redis.LocalClient(), key)
}

func getLockKey(query CheckIfItemsInStock) string {
	var ids []string
	for _, item := range query.Items {
		ids = append(ids, item.Id)
	}

	return redisLockPrefix + strings.Join(ids, "_")
}

// checkStock 检查库存是否充足并尝试扣除库存，该方法会处理重复的 item，调用时可以不用合并 item
func (h checkIfItemsInStockHandler) checkStock(ctx context.Context, query []*domain.ItemWithQuantity) error {
	var ids []string
	for _, item := range query {
		ids = append(ids, item.Id)
	}

	records, err := h.stockRepo.GetStock(ctx, ids)
	if err != nil {
		return err
	}

	idQuantityMap := make(map[string]int64)
	for _, r := range records {
		idQuantityMap[r.Id] = r.Quantity
	}
	var (
		ok       = true
		failedOn []struct {
			ID   string
			Want int64
			Have int64
		}
	)

	// 合并相同商品数量
	query = packItems(query)

	for _, item := range query {
		if item.Quantity > idQuantityMap[item.Id] {
			ok = false
			failedOn = append(failedOn, struct {
				ID   string
				Want int64
				Have int64
			}{
				ID:   item.Id,
				Want: item.Quantity,
				Have: idQuantityMap[item.Id],
			})
		}
	}
	if ok {
		return h.stockRepo.UpdateStock(ctx, query, func(
			ctx context.Context,
			existing,
			require []*domain.ItemWithQuantity,
		) ([]*domain.ItemWithQuantity, error) {
			var updated []*domain.ItemWithQuantity
			for _, e := range existing {
				for _, r := range require {
					if e.Id == r.Id {
						updated = append(updated, &domain.ItemWithQuantity{
							Id:       e.Id,
							Quantity: e.Quantity - r.Quantity,
						})
					}
				}
			}
			return updated, nil
		})
	}

	return domain.ExceedStockError{FailedOn: failedOn}
}

// packItems 合并相同商品的数量
func packItems(items []*domain.ItemWithQuantity) []*domain.ItemWithQuantity {
	merged := make(map[string]int64)
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
