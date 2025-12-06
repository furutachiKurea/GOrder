package adapter

import (
	"context"

	domain "github.com/furutachiKurea/gorder/stock/domain/stock"
	"github.com/furutachiKurea/gorder/stock/infrastructure/persistent"
)

type StockRepositoryMySQL struct {
	db *persistent.MySQL
}

func NewStockRepositoryMySQL(db *persistent.MySQL) *StockRepositoryMySQL {
	return &StockRepositoryMySQL{db: db}
}

func (s StockRepositoryMySQL) GetItems(ctx context.Context, ids []string) ([]*domain.Item, error) {
	// TODO implement me
	panic("implement me")
}

func (s StockRepositoryMySQL) GetStock(ctx context.Context, ids []string) ([]*domain.ItemWithQuantity, error) {
	data, err := s.db.BatchGetStockByID(ctx, ids)
	if err != nil {
		return nil, err
	}

	var result []*domain.ItemWithQuantity
	for _, d := range data {
		result = append(result, &domain.ItemWithQuantity{
			Id:       d.ProductID,
			Quantity: d.Quantity,
		})
	}

	return result, nil
}
