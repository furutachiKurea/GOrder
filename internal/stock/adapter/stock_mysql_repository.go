package adapter

import (
	"context"
	"fmt"

	domain "github.com/furutachiKurea/gorder/stock/domain/stock"
	"github.com/furutachiKurea/gorder/stock/infrastructure/persistent"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
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
		return nil, fmt.Errorf("batch get stock by id: %w", err)
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

func (s StockRepositoryMySQL) UpdateStock(
	ctx context.Context,
	data []*domain.ItemWithQuantity,
	updateFn func(
		ctx context.Context,
		existing []*domain.ItemWithQuantity,
		query []*domain.ItemWithQuantity,
	) ([]*domain.ItemWithQuantity, error),
) error {
	return s.db.StartTransaction(func(tx *gorm.DB) (err error) {
		defer func() {
			if err != nil {
				log.Warn().Err(err).Msg("update stock transaction failed")
			}
		}()

		var (
			dest      []*persistent.StockModel
			tableName = "o_stock"
		)
		if err = tx.Table(tableName).Where("product_id IN ?", getIDsFromItems(data)).Find(&dest).Error; err != nil {
			return fmt.Errorf("get stock by ids from db: %w", err)
		}

		existing := s.unmarshalFromModels(dest)

		updated, err := updateFn(ctx, existing, data)
		if err != nil {
			return nil
		}

		for _, upd := range updated {
			if err = tx.Table(tableName).
				Where("product_id = ?", upd.Id).
				Update("quantity", upd.Quantity).Error; err != nil {
				return fmt.Errorf("update stock in db: %w", err)
			}
		}

		return nil
	})
}

// getIDsFromItems 从 ItemWithQuantity 切片中提取 ID 切片，用于从数据库查询所有商品对应的库存
func getIDsFromItems(items []*domain.ItemWithQuantity) []string {
	var ids []string
	for _, item := range items {
		ids = append(ids, item.Id)
	}

	return ids
}

func (s StockRepositoryMySQL) unmarshalFromModels(dest []*persistent.StockModel) []*domain.ItemWithQuantity {
	var res []*domain.ItemWithQuantity
	for _, item := range dest {
		res = append(res, &domain.ItemWithQuantity{
			Id:       item.ProductID,
			Quantity: item.Quantity,
		})
	}

	return res
}
