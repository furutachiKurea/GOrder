package adapter

import (
	"context"
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

// ReserveStock 预占库存，使用悲观锁保证一致性
func (s StockRepositoryMySQL) ReserveStock(ctx context.Context, items []*domain.ItemWithQuantity) error {
	return s.db.StartTransaction(func(tx *gorm.DB) (err error) {
		defer func() {
			if err != nil {
				log.Warn().Ctx(ctx).Err(err).Msg("reserve stock transaction failed")
			}
		}()

		stocks, err := s.getAndLockStock(ctx, tx, items)
		if err != nil {
			return err
		}

		// 如果获取到的商品库存记录数量少于请求数量，说明请求了不存在的商品
		if missingIDs := findMissingProductIDs(items, stocks); len(missingIDs) > 0 {
			return domain.NotFoundError{Missing: missingIDs}
		}

		err = s.tryReserveStock(ctx, tx, items)
		return
	})
}

func (s StockRepositoryMySQL) ConfirmStockReservation(ctx context.Context, items []*domain.ItemWithQuantity) error {
	return s.db.StartTransaction(func(tx *gorm.DB) (err error) {
		defer func() {
			if err != nil {
				log.Warn().Ctx(ctx).Err(err).Msg("confirm stock reservation transaction failed")
			}
		}()

		_, err = s.getAndLockStock(ctx, tx, items)
		if err != nil {
			return err
		}

		err = s.tryConfirmStockReservation(ctx, tx, items)

		return
	})
}

// getAndLockStock 通过 item.id 获取库存并锁定库存记录
func (s StockRepositoryMySQL) getAndLockStock(
	ctx context.Context,
	tx *gorm.DB,
	items []*domain.ItemWithQuantity,
) ([]*persistent.StockModel, error) {

	var stocks []*persistent.StockModel
	err := tx.WithContext(ctx).
		Clauses(clause.Locking{Strength: clause.LockingStrengthUpdate}).
		Model(persistent.StockModel{}).
		Where("product_id IN (?)", getIDsFromItems(items)).
		Find(&stocks).Error
	if err != nil {
		return nil, fmt.Errorf("get stock by ids from db: %w", err)
	}

	return stocks, nil
}

// tryReserveStock 尝试预占库存
func (s StockRepositoryMySQL) tryReserveStock(
	ctx context.Context,
	tx *gorm.DB,
	toReserve []*domain.ItemWithQuantity,
) error {

	requiredQuantities := make(map[string]int64)
	for _, item := range toReserve {
		requiredQuantities[item.Id] += item.Quantity
	}

	var (
		failedOn []struct {
			ID   string
			Want int64
			Have int64
		}
		failedProductIDs []string
	)
	for productID, required := range requiredQuantities {
		if required == 0 {
			continue
		}

		result := tx.Model(persistent.StockModel{}).
			Where("product_id = ? AND quantity - reserved >= ?", productID, required).
			Update("reserved", gorm.Expr("reserved + ?", required))

		if result.Error != nil {
			return fmt.Errorf("update stock in db: %w", result.Error)
		}

		// 未更新成功，说明库存不足，记录失败信息
		if result.RowsAffected == 0 {
			failedProductIDs = append(failedProductIDs, productID)
			failedOn = append(failedOn, struct {
				ID   string
				Want int64
				Have int64
			}{ID: productID, Want: required, Have: 0})
		}
	}

	// 如果存在失败商品
	if len(failedProductIDs) > 0 {
		var failed []persistent.StockModel
		if err := tx.WithContext(ctx).
			Model(persistent.StockModel{}).
			Where("product_id IN (?)", failedProductIDs).
			Find(&failed).Error; err != nil {
			return fmt.Errorf("get failed stocks from db: %w", err)
		}

		have := make(map[string]int64)
		for _, item := range failed {
			have[item.ProductID] += item.Quantity - item.Reserved
		}

		for i, fail := range failedOn {
			failedOn[i].Have = have[fail.ID]
		}
		return domain.ExceedStockError{FailedOn: failedOn}
	}
	return nil
}

func (s StockRepositoryMySQL) tryConfirmStockReservation(
	ctx context.Context,
	tx *gorm.DB,
	toConfirm []*domain.ItemWithQuantity,
) error {

	confirmedReservation := make(map[string]int64)
	for _, item := range toConfirm {
		confirmedReservation[item.Id] += item.Quantity
	}

	var ()

	for productID, quantity := range confirmedReservation {
		if quantity == 0 {
			continue
		}

		result := tx.WithContext(ctx).Model(persistent.StockModel{}).
			Where("product_id = ? AND quantity >= reserved AND reserved >= ?", productID, quantity).
			Updates(map[string]any{
				"quantity": gorm.Expr("quantity - ?", quantity),
				"reserved": gorm.Expr("reserved - ?", quantity),
			})
		if result.Error != nil {
			return fmt.Errorf("update stock in db: %w", result.Error)
		}
		if result.RowsAffected == 0 {
			return fmt.Errorf("confirm stock reservation failed for product_id=%s", productID)
		}
	}

	return nil
}

// findMissingProductIDs 比较期望的商品列表和实际从数据库获取的库存列表，返回缺失的商品 ID 列表
func findMissingProductIDs(requested []*domain.ItemWithQuantity, stocks []*persistent.StockModel) []string {
	var missingIDs []string
	gotSet := make(map[string]struct{})
	for _, item := range stocks {
		gotSet[item.ProductID] = struct{}{}
	}

	for _, item := range requested {
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
