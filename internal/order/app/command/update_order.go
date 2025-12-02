package command

import (
	"context"
	"fmt"

	"github.com/furutachiKurea/gorder/common/decorator"
	"github.com/furutachiKurea/gorder/common/tracing"
	domain "github.com/furutachiKurea/gorder/order/domain/order"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type UpdateOrder struct {
	Order    *domain.Order
	UpdateFn func(ctx context.Context, order *domain.Order) (*domain.Order, error)
}

type UpdateOrderHandler decorator.CommandHandler[UpdateOrder, any]

type updateOrderHandler struct {
	orderRepo domain.Repository
	// TODO stock gRPC
}

func NewUpdateOrderHandler(
	orderRepo domain.Repository,
	logger zerolog.Logger,
	metricsClient decorator.MetricsClient,
) UpdateOrderHandler {
	if orderRepo == nil {
		panic("orderRepo is nil")
	}

	return decorator.ApplyCommandDecorators[UpdateOrder, any](
		updateOrderHandler{orderRepo: orderRepo},
		logger,
		metricsClient,
	)
}

func (c updateOrderHandler) Handle(ctx context.Context, cmd UpdateOrder) (any, error) {
	ctx, span := tracing.Start(ctx, "updateOrderHandler")
	defer span.End()

	if cmd.UpdateFn == nil {
		log.Warn().Msgf("updateOrderHandler got nil UpdateFunc, order=%#v", cmd.Order)
		cmd.UpdateFn = func(_ context.Context, order *domain.Order) (*domain.Order, error) {
			return order, nil
		}
	}

	// TODO 为什么没有查询库存再更新订单？

	if err := c.orderRepo.Update(ctx, cmd.Order, cmd.UpdateFn); err != nil {
		return nil, fmt.Errorf("create order: %w", err)
	}

	return nil, nil
}
