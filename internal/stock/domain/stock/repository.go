package stock

import (
	"context"
	"fmt"
	"strings"
)

type Repository interface {
	GetItems(ctx context.Context, ids []string) ([]*Item, error)
	GetStock(ctx context.Context, ids []string) ([]*ItemWithQuantity, error)
	// UpdateStock 更新库存，updateFn 接收 existing item , 应返回 item 的期望状态
	UpdateStock(
		ctx context.Context,
		data []*ItemWithQuantity,
		updateFn func(
			ctx context.Context,
			existing, query []*ItemWithQuantity,
		) ([]*ItemWithQuantity, error),
	) error
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
		Want int32
		Have int32
	}
}

func (e ExceedStockError) Error() string {
	var info []string
	for _, v := range e.FailedOn {
		info = append(info, fmt.Sprintf("product_id=%s, want %d, have %d", v.ID, v.Want, v.Have))
	}
	return fmt.Sprintf("not enough stock for [%s]", strings.Join(info, ","))
}
