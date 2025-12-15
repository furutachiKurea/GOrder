package command

import (
	"context"
	"fmt"

	"github.com/furutachiKurea/gorder/common/convertor"
	"github.com/furutachiKurea/gorder/common/decorator"
	"github.com/furutachiKurea/gorder/common/entity"
	"github.com/furutachiKurea/gorder/common/logging"
	"github.com/furutachiKurea/gorder/common/tracing"
	"github.com/furutachiKurea/gorder/order/app/client"
	domain "github.com/furutachiKurea/gorder/order/domain/order"

	"github.com/rs/zerolog"
)

type ConfirmOrderPaid struct {
	Order *domain.Order
}

// ConfirmOrderPaidHandler 确认订单支付成功，更新订单状态并完成订单的实际减扣
type ConfirmOrderPaidHandler decorator.CommandHandler[ConfirmOrderPaid, any]

type confirmOrderPaidHandler struct {
	orderRepo domain.Repository
	stockGRPC client.StockService
}

func NewConfirmOrderPaidHandler(
	orderRepo domain.Repository,
	stockGrpc client.StockService,
	logger zerolog.Logger,
	metricsClient decorator.MetricsClient,
) ConfirmOrderPaidHandler {
	if orderRepo == nil {
		panic("orderRepo is nil")
	}

	return decorator.ApplyCommandDecorators[ConfirmOrderPaid, any](
		confirmOrderPaidHandler{
			orderRepo: orderRepo,
			stockGRPC: stockGrpc,
		},
		logger,
		metricsClient,
	)
}

func (c confirmOrderPaidHandler) Handle(ctx context.Context, cmd ConfirmOrderPaid) (any, error) {
	var err error
	defer logging.WhenCommandExecute(ctx, "ConfirmOrderPaidHandler", cmd, err)

	ctx, span := tracing.Start(ctx, "confirmOrderPaidHandler")
	defer span.End()
	if err := cmd.Order.IsPaid(); err != nil {
		return nil, err
	}
	if err := c.orderRepo.Update(ctx, cmd.Order); err != nil {
		return nil, err
	}

	var itemWithQuantities []*entity.ItemWithQuantity
	for _, orderItem := range cmd.Order.Items {
		itemWithQuantities = append(itemWithQuantities, &entity.ItemWithQuantity{
			ID:       orderItem.ID,
			Quantity: orderItem.Quantity,
		})
	}

	_, err = c.stockGRPC.ConfirmStockReservation(
		ctx,
		convertor.NewItemWithQuantityConvertor().EntitiesToProtos(itemWithQuantities),
	)
	if err != nil {
		return nil, fmt.Errorf("confirm stock reservation: %w", err)
	}

	return nil, nil
}
