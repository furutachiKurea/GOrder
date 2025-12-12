package command

import "context"

type PriceProvider interface {
	GetPriceByProductID(context.Context, string) (string, error)
}
