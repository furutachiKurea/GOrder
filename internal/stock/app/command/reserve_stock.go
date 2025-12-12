package command

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

type ReserveStock struct {
	Items []*domain.ItemWithQuantity
}

type ReserveStockHandler decorator.CommandHandler[ReserveStock, []*domain.Item]

type reserveStockHandler struct {
	stockRepo     domain.Repository
	priceProvider PriceProvider
}

func NewReserveStockHandler(
	stockRepo domain.Repository,
	priceProvider PriceProvider,
	logger zerolog.Logger,
	metricsClient decorator.MetricsClient,
) ReserveStockHandler {
	if stockRepo == nil {
		panic("stockRepo is nil")
	}
	if priceProvider == nil {
		panic("priceProvider is nil")
	}

	return decorator.ApplyCommandDecorators[ReserveStock, []*domain.Item](
		reserveStockHandler{
			stockRepo:     stockRepo,
			priceProvider: priceProvider,
		},
		logger,
		metricsClient,
	)
}

func (h reserveStockHandler) Handle(ctx context.Context, query ReserveStock) ([]*domain.Item, error) {
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

func getLockKey(query ReserveStock) string {
	var ids []string
	for _, item := range query.Items {
		ids = append(ids, item.Id)
	}

	return redisLockPrefix + strings.Join(ids, "_")
}

// checkStock 检查库存是否充足并尝试预占库存避免超卖，该方法会处理重复的 item，调用时可以不用合并 item
func (h reserveStockHandler) checkStock(ctx context.Context, query []*domain.ItemWithQuantity) error {
	var ids []string
	for _, item := range query {
		ids = append(ids, item.Id)
	}

	records, err := h.stockRepo.GetStock(ctx, ids)
	if err != nil {
		return err
	}

	stocks := make(map[string]int64)
	for _, r := range records {
		stocks[r.Id] = r.Quantity
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

	for _, q := range query {
		if q.Quantity > stocks[q.Id] {
			ok = false
			failedOn = append(failedOn, struct {
				ID   string
				Want int64
				Have int64
			}{
				ID:   q.Id,
				Want: q.Quantity,
				Have: stocks[q.Id],
			})
		}
	}
	if ok {
		return h.stockRepo.ReserveStock(ctx, query, func(
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
