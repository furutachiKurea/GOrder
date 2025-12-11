package adapter

import (
	"context"
	"errors"
	"fmt"

	domain "github.com/furutachiKurea/gorder/stock/domain/stock"
	"github.com/furutachiKurea/gorder/stock/infrastructure/persistent"

	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
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

// UpdateStock 更新库存，使用悲观锁保证一致性
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

		var dest []*persistent.StockModel
		if err = tx.WithContext(ctx).
			Clauses(clause.Locking{Strength: clause.LockingStrengthUpdate}).
			Model(persistent.StockModel{}).
			Where("product_id IN (?)", getIDsFromItems(data)).
			Find(&dest).Error; err != nil {
			return fmt.Errorf("get stock by ids from db: %w", err)
		}

		// 如果存在商品库存记录数量少于请求数量，说明请求了不存在的商品
		if missingIDs := getMissingItemIDs(data, dest); len(missingIDs) > 0 {
			return domain.NotFoundError{Missing: missingIDs}
		}

		existing := s.unmarshalFromModels(dest)
		updated, err := updateFn(ctx, existing, data)
		if err != nil {
			return err
		}

		queryMap := make(map[string]int64)
		for _, item := range data {
			queryMap[item.Id] += item.Quantity
		}

		var (
			failedOn []struct {
				ID   string
				Want int64
				Have int64
			}
			failedProductIDs []string
		)

		for _, upd := range updated {
			toDeduct := queryMap[upd.Id]
			result := tx.Model(persistent.StockModel{}).
				Where("product_id = ? AND quantity >= ?", upd.Id, toDeduct).
				Update("quantity", gorm.Expr("quantity - ?", toDeduct))
			if result.Error != nil {
				if errors.Is(result.Error, gorm.ErrRecordNotFound) {
					return fmt.Errorf("product id %s not found: %w", upd.Id, result.Error)
				}
				return fmt.Errorf("update stock in db: %w", result.Error)
			}
			// 未更新成功，说明库存不足，记录失败信息
			if result.RowsAffected == 0 {
				failedProductIDs = append(failedProductIDs, upd.Id)
				failedOn = append(failedOn, struct {
					ID   string
					Want int64
					Have int64
				}{ID: upd.Id, Want: toDeduct, Have: 0})
			}
		}

		// 如果存在失败商品
		if len(failedProductIDs) > 0 {
			var failedStocks []persistent.StockModel
			if err = tx.WithContext(ctx).
				Model(persistent.StockModel{}).
				Where("product_id IN (?)", failedProductIDs).
				Find(&failedStocks).Error; err != nil {
				return fmt.Errorf("get failed stocks from db: %w", err)
			}

			stockHave := make(map[string]int64)
			for _, stock := range failedStocks {
				stockHave[stock.ProductID] = stock.Quantity
			}

			for i, fail := range failedOn {
				failedOn[i].Have = stockHave[fail.ID]
			}

			return domain.ExceedStockError{FailedOn: failedOn}
		}
		return nil
	})
}

// getMissingItemIDs 比较期望的商品列表和实际从数据库获取的库存列表，返回缺失的商品 ID 列表
func getMissingItemIDs(expect []*domain.ItemWithQuantity, got []*persistent.StockModel) []string {
	var missingIDs []string
	gotSet := make(map[string]struct{})
	for _, item := range got {
		gotSet[item.ProductID] = struct{}{}
	}

	for _, item := range expect {
		if _, ok := gotSet[item.Id]; !ok {
			missingIDs = append(missingIDs, item.Id)
		}
	}

	return missingIDs
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
