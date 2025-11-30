package domain

import (
	"context"

	"github.com/furutachiKurea/gorder/common/genproto/orderpb"
)

type Processor interface {
	CreatePaymentLink(ctx context.Context, order *orderpb.Order) (string, error)
}
