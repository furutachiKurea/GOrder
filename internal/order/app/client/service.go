package client

import (
	"context"

	"github.com/furutachiKurea/gorder/common/genproto/orderpb"
	"github.com/furutachiKurea/gorder/common/genproto/stockpb"
)

type StockService interface {
	GetItems(ctx context.Context, itemIDs []string) ([]*orderpb.Item, error)
	ReserveStock(ctx context.Context, items []*orderpb.ItemWithQuantity) (*stockpb.ReserveStockResponse, error)
	ConfirmStockReservation(ctx context.Context, items []*orderpb.ItemWithQuantity) (*stockpb.ConfirmStockReservationResponse, error)
}
