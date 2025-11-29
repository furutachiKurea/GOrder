package query

import (
	"context"

	"github.com/furutachiKurea/gorder/common/genproto/orderpb"
)

type StockInterface interface {
	GetItems(ctx context.Context, itemIDs []string) ([]*orderpb.Item, error)
	CheckIfItemsInStock(ctx context.Context, items []*orderpb.ItemWithQuantity) error
}
