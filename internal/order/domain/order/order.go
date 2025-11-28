package order

import (
	"github.com/furutachiKurea/gorder/common/genproto/orderpb"
)

type Order struct {
	ID          string
	CustomerID  string
	Status      string
	PaymentLink string
	Items       []*orderpb.Item
}

type NotFoundError struct {
	OrderID string
}

func (e NotFoundError) Error() string {
	return "order " + e.OrderID + " not found"
}
