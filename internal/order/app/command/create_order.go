package command

import (
	"context"
	"fmt"

	"github.com/furutachiKurea/gorder/common/decorator"
	"github.com/furutachiKurea/gorder/common/genproto/orderpb"
	"github.com/furutachiKurea/gorder/order/app/query"
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
	stockGRPC query.StockInterface
}

func NewCreateOrderHandler(
	orderRepo domain.Repository,
	stockGRPC query.StockInterface,
	logger *logrus.Entry,
	metricsClient decorator.MetricsClient,
) CreateOrderHandler {
	if orderRepo == nil {
		panic("orderRepo is nil")
	}

	return decorator.ApplyCommandDecorators[CreateOrder, *CreateOrderResult](
		createOrderHandler{orderRepo: orderRepo, stockGRPC: stockGRPC},
		logger,
		metricsClient,
	)
}

func (c createOrderHandler) Handle(ctx context.Context, command CreateOrder) (*CreateOrderResult, error) {
	// TODO 通过 Stock gPRC查询库存
	err := c.stockGRPC.CheckIfItemsInStock(ctx, command.Items)
	logrus.Info("createOrderHandler || err from stockGRPC", err)

	resp, err := c.stockGRPC.GetItems(ctx, []string{"123"})
	logrus.Info("createOrderHandler || resp from stockGRPC.GetItems:", resp)

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
