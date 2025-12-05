package stock

import (
	"context"
	"fmt"
	"strings"
)

type Repository interface {
	GetItems(ctx context.Context, ids []string) ([]*Item, error)
}

type NotFoundError struct {
	Missing []string
}

func (e NotFoundError) Error() string {
	return fmt.Sprintf("not found in stock: %s", strings.Join(e.Missing, ","))
}
