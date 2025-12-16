package command

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/furutachiKurea/gorder/common/decorator"
	"github.com/furutachiKurea/gorder/common/entity"
	"github.com/furutachiKurea/gorder/common/handler/redis"
	"github.com/furutachiKurea/gorder/common/logging"
	domain "github.com/furutachiKurea/gorder/stock/domain/stock"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

const (
	redisLockPrefix = "reserve_stock_"
)

type ReserveStock struct {
	Items []*entity.ItemWithQuantity
}

type ReserveStockHandler decorator.CommandHandler[ReserveStock, []*entity.Item]

type reserveStockHandler struct {
	stockRepo     domain.Repository
	priceProvider ProductProvider
}

func NewReserveStockHandler(
	stockRepo domain.Repository,
	priceProvider ProductProvider,
	logger zerolog.Logger,
	metricsClient decorator.MetricsClient,
) ReserveStockHandler {
	if stockRepo == nil {
		panic("stockRepo is nil")
	}
	if priceProvider == nil {
		panic("priceProvider is nil")
	}

	return decorator.ApplyCommandDecorators[ReserveStock, []*entity.Item](
		reserveStockHandler{
			stockRepo:     stockRepo,
			priceProvider: priceProvider,
		},
		logger,
		metricsClient,
	)
}

func (h reserveStockHandler) Handle(ctx context.Context, command ReserveStock) ([]*entity.Item, error) {
	var err error
	defer logging.WhenCommandExecute(ctx, "ReserveStockHandler", command, err)

	if err := lock(ctx, getLockKey(command.Items)); err != nil {
		return nil, fmt.Errorf("redis lock, key=%s: %w", getLockKey(command.Items), err)
	}
	defer func() {
		if err := unlock(ctx, getLockKey(command.Items)); err != nil {
			log.Warn().Ctx(ctx).Err(err).Msg("redis unlock fail")
		}
	}()

	// 从 stripe 获取 priceID
	var res []*entity.Item
	for _, item := range command.Items {
		p, err := h.priceProvider.GetProductByID(ctx, item.ID)
		if err != nil {
			return nil, err
		}
		valid, err := entity.NewValidItem(item.ID, p.Name, item.Quantity, p.PriceID)
		if err != nil {
			return nil, err
		}
		res = append(res, valid)
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

func getLockKey(items []*entity.ItemWithQuantity) string {
	var ids []string
	for _, item := range items {
		ids = append(ids, item.ID)
	}

	return redisLockPrefix + strings.Join(ids, "_")
}

// packItems 合并相同商品的数量
func packItems(items []*entity.ItemWithQuantity) []*entity.ItemWithQuantity {
	merged := make(map[string]int64)
	for _, item := range items {
		merged[item.ID] += item.Quantity
	}

	var packed []*entity.ItemWithQuantity
	for id, quantity := range merged {
		packed = append(packed, &entity.ItemWithQuantity{
			ID:       id,
			Quantity: quantity,
		})
	}

	return packed
}
