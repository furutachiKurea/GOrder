package processor

import (
	"context"

	"github.com/furutachiKurea/gorder/common/entity"
)

type InmemProcessor struct{}

func NewInmemProcessor() *InmemProcessor {
	return &InmemProcessor{}
}

func (i InmemProcessor) CreatePaymentLink(ctx context.Context, order *entity.Order) (string, error) {
	return "inmem_payment_link_for_order", nil
}
