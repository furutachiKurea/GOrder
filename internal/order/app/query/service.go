package query

import (
	"context"

	"github.com/furutachiKurea/gorder/common/genproto/orderpb"
	"github.com/furutachiKurea/gorder/common/genproto/stockpb"
)

type StockInterface interface {
	GetItems(ctx context.Context, itemIDs []string) ([]*orderpb.Item, error)
	ReserveStock(ctx context.Context, items []*orderpb.ItemWithQuantity) (*stockpb.ReserveStockResponse, error)
}
