package ports

import (
	"context"

	"github.com/furutachiKurea/gorder/common/genproto/orderpb"
	"github.com/furutachiKurea/gorder/common/genproto/stockpb"
	"github.com/furutachiKurea/gorder/stock/app"
	"github.com/sirupsen/logrus"
)

type GRPCServer struct {
	app app.Application
}

func NewGRPCServer(app app.Application) *GRPCServer {
	return &GRPCServer{app: app}
}

func (G GRPCServer) GetItems(ctx context.Context, request *stockpb.GetItemsRequest) (*stockpb.GetItemsResponse, error) {
	// TODO mock data, need to replace
	logrus.Infof("rpc_request_in, stock.GetItems")
	defer func() {
		logrus.Infof("rpc_request_out, stock.GetItems")
	}()

	fake := []*orderpb.Item{
		{
			Id:       "fake-item-from-GetItems",
			Name:     "",
			Quantity: 0,
			PriceId:  "",
		},
	}
	return &stockpb.GetItemsResponse{Items: fake}, nil
}

func (G GRPCServer) CheckIfItemsInStock(ctx context.Context, request *stockpb.CheckIfItemsInStockRequest) (*stockpb.CheckIfItemsInStockResponse, error) {
	// TODO mock data, need to replace
	logrus.Infof("rpc_request_in, stock.CheckIfItemsInStock")
	defer func() {
		logrus.Infof("rpc_request_out, stock.CheckIfItemsInStock")
	}()
	return nil, nil
}
