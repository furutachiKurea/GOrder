package domain

import (
	"context"

	"github.com/furutachiKurea/gorder/common/entity"
)

type Processor interface {
	CreatePaymentLink(ctx context.Context, order *entity.Order) (string, error)
}
type Order struct {
	ID          string
	CustomerID  string
	Status      string
	PaymentLink string
	Items       []*entity.Item
}
