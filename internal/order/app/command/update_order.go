package command

import (
	"context"
	"fmt"

	"github.com/furutachiKurea/gorder/common/decorator"
	"github.com/furutachiKurea/gorder/common/tracing"
	domain "github.com/furutachiKurea/gorder/order/domain/order"

	"github.com/rs/zerolog"
)

type UpdateOrder struct {
	Order *domain.Order
}

// UpdateOrderHandler 使用给定的 UpdateFn 更新订单
type UpdateOrderHandler decorator.CommandHandler[UpdateOrder, any]

type updateOrderHandler struct {
	orderRepo domain.Repository
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

	if err := c.orderRepo.Update(ctx, cmd.Order); err != nil {
		return nil, fmt.Errorf("create order: %w", err)
	}

	return nil, nil
}
