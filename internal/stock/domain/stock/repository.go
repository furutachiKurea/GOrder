package stock

import (
	"context"
	"fmt"
	"strings"
)

type Repository interface {
	GetItems(ctx context.Context, ids []string) ([]*Item, error)
	GetStock(ctx context.Context, ids []string) ([]*ItemWithQuantity, error)
	// ReserveStock 更新库存，updateFn 接收 existing item , 应返回 item 的期望状态
	//
	// TODO 现在的设计中，返回 item 的期望状态，但是在实际实现 (@stock_mysql_repository.go) 中
	//  是依据 SQL 查询来动态计算更新后的库存数量，与设计不一致 (目前的调用者还在 updateFn 中写了计算逻辑)。
	//  设计问题，优先级稍低
	ReserveStock(
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
