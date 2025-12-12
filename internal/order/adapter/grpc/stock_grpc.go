package grpc

import (
	"context"

	"github.com/furutachiKurea/gorder/common/genproto/orderpb"
	"github.com/furutachiKurea/gorder/common/genproto/stockpb"
)

type StockGRPC struct {
	client stockpb.StockServiceClient
}

func NewStockGRPC(client stockpb.StockServiceClient) StockGRPC {
	return StockGRPC{client: client}
}

func (s StockGRPC) GetItems(ctx context.Context, itemIDs []string) ([]*orderpb.Item, error) {
	resp, err := s.client.GetItems(ctx, &stockpb.GetItemsRequest{ItemIds: itemIDs})
	if err != nil {
		return nil, err
	}

	return resp.Items, nil
}

func (s StockGRPC) ReserveStock(ctx context.Context, items []*orderpb.ItemWithQuantity) (*stockpb.ReserveStockResponse, error) {
	resp, err := s.client.ReserveStock(ctx,
		&stockpb.ReserveStockRequest{Items: items},
	)

	return resp, err
}
