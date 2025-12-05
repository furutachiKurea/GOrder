package ports

import (
	"context"

	"github.com/furutachiKurea/gorder/common/genproto/stockpb"
	"github.com/furutachiKurea/gorder/stock/app"
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

func (G GRPCServer) CheckIfItemsInStock(ctx context.Context, request *stockpb.CheckIfItemsInStockRequest) (*stockpb.CheckIfItemsInStockResponse, error) {
	items, err := G.app.Queries.CheckIfItemsInStock.Handle(
		ctx,
		query.CheckIfItemsInStock{
			Items: convertor.NewItemWithQuantityConvertor().ProtosToDomains(request.Items),
		},
	)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &stockpb.CheckIfItemsInStockResponse{
		InStock: 1,
		Items:   convertor.NewItemConvertor().DomainsToProtos(items),
	}, nil
}
