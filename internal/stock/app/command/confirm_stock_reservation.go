package command

import (
	"context"
	"fmt"

	"github.com/furutachiKurea/gorder/common/decorator"
	"github.com/furutachiKurea/gorder/common/logging"
	domain "github.com/furutachiKurea/gorder/stock/domain/stock"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type ConfirmStockReservation struct {
	Items []*domain.ItemWithQuantity
}

// ConfirmStockReservationHandler 为支付成功的订单更新库存状态，扣减库存并扣除预留库存
type ConfirmStockReservationHandler decorator.CommandHandler[ConfirmStockReservation, []*domain.Item]

type confirmStockReservationHandler struct {
	stockRepo domain.Repository
}

func NewConfirmStockReservation(
	stockRepo domain.Repository,
	logger zerolog.Logger,
	metricsClient decorator.MetricsClient,
) ConfirmStockReservationHandler {
	if stockRepo == nil {
		panic("stockRepo is nil")
	}

	return decorator.ApplyCommandDecorators[ConfirmStockReservation, []*domain.Item](
		confirmStockReservationHandler{
			stockRepo: stockRepo,
		},
		logger,
		metricsClient,
	)
}

func (h confirmStockReservationHandler) Handle(ctx context.Context, command ConfirmStockReservation) ([]*domain.Item, error) {
	var err error
	defer logging.WhenCommandExecute(ctx, "ConfirmStockReservationHandler", command, err)

	if err := lock(ctx, getLockKey(command.Items)); err != nil {
		return nil, fmt.Errorf("redis lock, key=%s: %w", getLockKey(command.Items), err)
	}
	defer func() {
		if err := unlock(ctx, getLockKey(command.Items)); err != nil {
			log.Warn().Err(err).Msg("redis unlock fail")
		}
	}()

	items := packItems(command.Items)
	if err := h.stockRepo.ConfirmStockReservation(ctx, items); err != nil {
		return nil, err
	}

	return nil, nil
}
