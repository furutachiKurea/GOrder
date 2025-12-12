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
	redisLockPrefix = "reserve_stock_"
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

func (h reserveStockHandler) Handle(ctx context.Context, command ReserveStock) ([]*domain.Item, error) {
	if err := lock(ctx, getLockKey(command.Items)); err != nil {
		return nil, fmt.Errorf("redis lock, key=%s: %w", getLockKey(command.Items), err)
	}
	defer func() {
		if err := unlock(ctx, getLockKey(command.Items)); err != nil {
			log.Warn().Err(err).Msg("redis unlock fail")
		}
	}()

	// 从 stripe 获取 priceID
	var res []*domain.Item
	for _, item := range command.Items {
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

	// 预扣库存
	items := packItems(command.Items)
	if err := h.stockRepo.ReserveStock(ctx, items); err != nil {
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

func getLockKey(items []*domain.ItemWithQuantity) string {
	var ids []string
	for _, item := range items {
		ids = append(ids, item.Id)
	}

	return redisLockPrefix + strings.Join(ids, "_")
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
