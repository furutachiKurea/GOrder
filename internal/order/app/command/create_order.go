package command

import (
	"context"
	"fmt"

	"github.com/furutachiKurea/gorder/common/decorator"
	"github.com/furutachiKurea/gorder/common/genproto/orderpb"
	domain "github.com/furutachiKurea/gorder/order/domain/order"
	"github.com/sirupsen/logrus"
)

type CreateOrder struct {
	CustomerID string
	Items      []*orderpb.ItemWithQuantity
}

type CreateOrderResult struct {
	OrderID string
}

type CreateOrderHandler decorator.CommandHandler[CreateOrder, *CreateOrderResult]

type createOrderHandler struct {
	orderRepo domain.Repository
	// TODO stock gRPC
}

func NewCreateOrderHandler(
	orderRepo domain.Repository,
	logger *logrus.Entry,
	metricsClient decorator.MetricsClient,
) CreateOrderHandler {
	if orderRepo == nil {
		panic("orderRepo is nil")
	}

	return decorator.ApplyCommandDecorators[CreateOrder, *CreateOrderResult](
		createOrderHandler{orderRepo: orderRepo},
		logger,
		metricsClient,
	)
}

func (c createOrderHandler) Handle(ctx context.Context, command CreateOrder) (*CreateOrderResult, error) {
	// TODO 通过 Stock gPRC查询库存
	var stockResponse []*orderpb.Item
	for _, item := range command.Items {
		stockResponse = append(stockResponse, &orderpb.Item{
			Id:       item.Id,
			Name:     "", // TODO
			Quantity: item.Quantity,
			PriceId:  "", // TODO
		})
	}

	order, err := c.orderRepo.Create(ctx, &domain.Order{
		CustomerID: command.CustomerID,
		Items:      stockResponse, // TODO get items from stock
	})
	if err != nil {
		return nil, fmt.Errorf("create order: %w", err)
	}

	return &CreateOrderResult{
		OrderID: order.ID,
	}, nil

}
