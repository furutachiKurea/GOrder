package ports

import (
	"context"

	"github.com/furutachiKurea/gorder/common/genproto/stockpb"
	"github.com/furutachiKurea/gorder/stock/app"
	"github.com/furutachiKurea/gorder/stock/app/command"
	"github.com/furutachiKurea/gorder/stock/app/query"
	"github.com/furutachiKurea/gorder/stock/convertor"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type GRPCServer struct {
	app app.Application
}

func NewGRPCServer(app app.Application) *GRPCServer {
	return &GRPCServer{app: app}
}

func (G GRPCServer) GetItems(ctx context.Context, request *stockpb.GetItemsRequest) (*stockpb.GetItemsResponse, error) {
	items, err := G.app.Queries.GetItems.Handle(ctx, query.GetItems{ItemIDs: request.ItemIds})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &stockpb.GetItemsResponse{
		Items: convertor.NewItemConvertor().DomainsToProtos(items),
	}, nil
}

func (G GRPCServer) ReserveStock(ctx context.Context, request *stockpb.ReserveStockRequest) (*stockpb.ReserveStockResponse, error) {
	items, err := G.app.Commands.ReserveStock.Handle(
		ctx,
		command.ReserveStock{
			Items: convertor.NewItemWithQuantityConvertor().ProtosToDomains(request.Items),
		},
	)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &stockpb.ReserveStockResponse{
		Items: convertor.NewItemConvertor().DomainsToProtos(items),
	}, nil
}

func (G GRPCServer) ConfirmStockReservation(ctx context.Context, request *stockpb.ConfirmStockReservationRequest) (*stockpb.ConfirmStockReservationResponse, error) {
	_, err := G.app.Commands.ConfirmStockReservation.Handle(ctx, command.ConfirmStockReservation{
		Items: convertor.NewItemWithQuantityConvertor().ProtosToDomains(request.Items),
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &stockpb.ConfirmStockReservationResponse{}, nil
}
