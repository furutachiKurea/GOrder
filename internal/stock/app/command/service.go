package command

import (
	"context"

	"github.com/furutachiKurea/gorder/stock/app/dto"
)

type ProductProvider interface {
	GetProductByID(ctx context.Context, pid string) (*dto.Product, error)
}
