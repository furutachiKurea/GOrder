package grpc

import (
	"context"

	"github.com/furutachiKurea/gorder/common/genproto/orderpb"
	"github.com/furutachiKurea/gorder/common/genproto/stockpb"
	"github.com/furutachiKurea/gorder/common/logging"
)

type StockGRPC struct {
	client stockpb.StockServiceClient
}

func NewStockGRPC(client stockpb.StockServiceClient) StockGRPC {
	return StockGRPC{client: client}
}

func (s StockGRPC) GetItems(ctx context.Context, itemIDs []string) (items []*orderpb.Item, err error) {
	_, deferlog := logging.WhenRequest(ctx, "StockGRPC.GetItems", items)
	defer deferlog(items, &err)
	resp, err := s.client.GetItems(ctx, &stockpb.GetItemsRequest{ItemIds: itemIDs})
	if err != nil {
		return nil, err
	}

	return resp.Items, nil
}

func (s StockGRPC) ReserveStock(ctx context.Context, items []*orderpb.ItemWithQuantity) (resp *stockpb.ReserveStockResponse, err error) {
	_, deferlog := logging.WhenRequest(ctx, "StockGRPC.ReserveStock", items)
	defer deferlog(resp, &err)

	return s.client.ReserveStock(ctx,
		&stockpb.ReserveStockRequest{Items: items},
	)
}

func (s StockGRPC) ConfirmStockReservation(ctx context.Context, items []*orderpb.ItemWithQuantity) (resp *stockpb.ConfirmStockReservationResponse, err error) {
	_, deferlog := logging.WhenRequest(ctx, "StockGRPC.ConfirmStockReservation", items)
	defer deferlog(resp, &err)

	return s.client.ConfirmStockReservation(
		ctx,
		&stockpb.ConfirmStockReservationRequest{Items: items},
	)
}
