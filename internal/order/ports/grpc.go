package ports

import (
	"context"

	"github.com/furutachiKurea/gorder/common/consts"
	"github.com/furutachiKurea/gorder/common/convertor"
	"github.com/furutachiKurea/gorder/common/genproto/orderpb"
	"github.com/furutachiKurea/gorder/order/app"
	"github.com/furutachiKurea/gorder/order/app/command"
	"github.com/furutachiKurea/gorder/order/app/query"
	domain "github.com/furutachiKurea/gorder/order/domain/order"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type GRPCServer struct {
	app app.Application
}

func NewGRPCServer(app app.Application) *GRPCServer {
	return &GRPCServer{app: app}
}

func (G GRPCServer) CreateOrder(ctx context.Context, request *orderpb.CreateOrderRequest) (*emptypb.Empty, error) {
	_, err := G.app.Commands.CreateOrder.Handle(ctx, command.CreateOrder{
		CustomerID: request.CustomerId,
		Items:      convertor.NewItemWithQuantityConvertor().ProtosToEntities(request.Items),
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &emptypb.Empty{}, nil
}

func (G GRPCServer) GetOrder(ctx context.Context, request *orderpb.GetOrderRequest) (*orderpb.Order, error) {
	order, err := G.app.Queries.GetCustomerOrder.Handle(ctx, query.GetCustomerOrder{
		CustomerID: request.CustomerId,
		OrderID:    request.OrderId,
	})

	if err != nil {
		return nil, status.Errorf(codes.NotFound, err.Error())
	}

	return &orderpb.Order{
		Id:          order.ID,
		CustomerId:  order.CustomerID,
		Status:      string(order.Status),
		PaymentLink: order.PaymentLink,
		Items:       convertor.NewItemConvertor().EntitiesToProtos(order.Items),
	}, nil
}

func (G GRPCServer) UpdateOrder(ctx context.Context, request *orderpb.Order) (*emptypb.Empty, error) {
	log.Info().Ctx(ctx).Any("request", request).Msg("order_grpc||request_in")
	newOrder, err := domain.NewOrder(
		request.Id,
		request.CustomerId,
		request.PaymentLink,
		consts.OrderStatus(request.Status),
		convertor.NewItemConvertor().ProtosToEntities(request.Items))
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	_, err = G.app.Commands.UpdateOrder.Handle(ctx, command.UpdateOrder{Order: newOrder})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &emptypb.Empty{}, nil
}
