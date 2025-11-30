package processor

import (
	"context"

	"github.com/furutachiKurea/gorder/common/genproto/orderpb"
)

type InmemProcessor struct{}

func NewInmemProcessor() *InmemProcessor {
	return &InmemProcessor{}
}

func (i InmemProcessor) CreatePaymentLink(_ context.Context, _ *orderpb.Order) (string, error) {
	return "inmem_payment_link_for_order", nil
}
