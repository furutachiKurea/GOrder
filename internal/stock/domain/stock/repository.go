package stock

import (
	"context"
	"fmt"
	"strings"

	"github.com/furutachiKurea/gorder/common/entity"
)

type Repository interface {
	GetItems(ctx context.Context, ids []string) ([]*entity.Item, error)
	GetStock(ctx context.Context, ids []string) ([]*entity.ItemWithQuantity, error)
	// ReserveStock 预扣库存
	ReserveStock(ctx context.Context, items []*entity.ItemWithQuantity) error
	// ConfirmStockReservation 订单支付成功后，更新实际库存和预扣库存
	ConfirmStockReservation(ctx context.Context, items []*entity.ItemWithQuantity) error
}

type NotFoundError struct {
	Missing []string
}

func (e NotFoundError) Error() string {
	return fmt.Sprintf("not found in stock: %s", strings.Join(e.Missing, ","))
}

type ExceedStockError struct {
	FailedOn []struct {
		ID   string
		Want int64
		Have int64
	}
}

func (e ExceedStockError) Error() string {
	var info []string
	for _, v := range e.FailedOn {
		info = append(info, fmt.Sprintf("product_id=%s, want %d, have %d", v.ID, v.Want, v.Have))
	}
	return fmt.Sprintf("not enough stock for [%s]", strings.Join(info, ","))
}
