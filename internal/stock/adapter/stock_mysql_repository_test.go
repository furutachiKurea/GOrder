package adapter

import (
	"context"
	"fmt"
	"sync"
	"testing"

	_ "github.com/furutachiKurea/gorder/common/config"
	domain "github.com/furutachiKurea/gorder/stock/domain/stock"
	"github.com/furutachiKurea/gorder/stock/infrastructure/persistent"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/sync/errgroup"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// deductStockUpdateFunc 用于测试，库存更新函数，执行扣减操作
var deductStockUpdateFunc = func(ctx context.Context,
	existing []*domain.ItemWithQuantity,
	query []*domain.ItemWithQuantity,
) ([]*domain.ItemWithQuantity, error) {
	var updated []*domain.ItemWithQuantity
	for _, e := range existing {
		for _, r := range query {
			if e.Id == r.Id {
				updated = append(updated, &domain.ItemWithQuantity{
					Id:       e.Id,
					Quantity: e.Quantity - r.Quantity,
				})
			}
		}
	}
	return updated, nil
}

func setupTestDB(t *testing.T) *persistent.MySQL {
	cfg := viper.Sub("mysql")

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.GetString("user"),
		cfg.GetString("password"),
		cfg.GetString("host"),
		cfg.GetString("port"),
		"",
	)
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	assert.NoError(t, err)

	testDB := cfg.GetString("database") + "_shadow"
	assert.NoError(t, db.Exec("DROP DATABASE IF EXISTS "+testDB).Error)
	assert.NoError(t, db.Exec("CREATE DATABASE IF NOT EXISTS "+testDB).Error)

	dsn = fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.GetString("user"),
		cfg.GetString("password"),
		cfg.GetString("host"),
		cfg.GetString("port"),
		testDB,
	)
	db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	assert.NoError(t, err)
	assert.NoError(t, db.AutoMigrate(persistent.StockModel{}))

	return persistent.NewMySQLWithDB(db)
}

func TestStockRepositoryMySQL_ReserveStock_Race(t *testing.T) {
	db := setupTestDB(t)

	var (
		ctx          = context.Background()
		testItem     = "test-race-item"
		initialStock = 10000
	)

	err := db.CreateBatch(ctx, []*persistent.StockModel{
		{
			ProductID: testItem,
			Quantity:  int64(initialStock),
		},
	})
	assert.NoError(t, err)

	repo := NewStockRepositoryMySQL(db)
	var g errgroup.Group
	concurrentGoroutines := 50

	for range concurrentGoroutines {
		g.Go(func() error {
			err := repo.ReserveStock(
				ctx,
				[]*domain.ItemWithQuantity{{Id: testItem, Quantity: 1}},
				deductStockUpdateFunc,
			)
			return err
		})
	}
	err = g.Wait()
	assert.NoError(t, err)

	res, err := db.BatchGetStockByID(ctx, []string{testItem})
	assert.NoError(t, err)
	assert.NotEmpty(t, res)

	expected := initialStock - concurrentGoroutines
	assert.Equal(t, int64(expected), res[0].Quantity-res[0].Reserved)
}

func TestStockRepositoryMySQL_ReserveStock_OverSell(t *testing.T) {
	db := setupTestDB(t)

	var (
		ctx          = context.Background()
		testItem     = "test-oversell-item"
		initialStock = 5
	)

	err := db.CreateBatch(ctx, []*persistent.StockModel{
		{
			ProductID: testItem,
			Quantity:  int64(initialStock),
		},
	})
	assert.NoError(t, err)

	repo := NewStockRepositoryMySQL(db)
	var wg sync.WaitGroup
	concurrentGoroutines := 50

	for range concurrentGoroutines {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = repo.ReserveStock(
				ctx,
				[]*domain.ItemWithQuantity{{Id: testItem, Quantity: 1}},
				deductStockUpdateFunc)
		}()
	}
	wg.Wait()

	res, err := db.BatchGetStockByID(ctx, []string{testItem})
	assert.NoError(t, err)
	assert.NotEmpty(t, res)
	t.Log("stock remain: ", res[0].Quantity)
	assert.GreaterOrEqual(t, res[0].Quantity-res[0].Reserved, int64(0))
}

func TestStockRepositoryMySQL_ReserveStock(t *testing.T) {
	tests := []struct {
		name             string
		stock            []*persistent.StockModel
		toUpdate         []*domain.ItemWithQuantity
		expectedReserved map[string]int64
		wantErr          bool
	}{
		{
			name: "success",
			stock: []*persistent.StockModel{
				{
					ProductID: "item-1",
					Quantity:  100,
					Reserved:  0,
				},
				{
					ProductID: "item-2",
					Quantity:  2,
					Reserved:  0,
				},
			},
			toUpdate: []*domain.ItemWithQuantity{
				{Id: "item-2", Quantity: 2},
				{Id: "item-1", Quantity: 2},
			},
			expectedReserved: map[string]int64{
				"item-1": 2,
				"item-2": 2,
			},
			wantErr: false,
		},
		{
			name: "insufficient_stock",
			stock: []*persistent.StockModel{
				{
					ProductID: "item-1",
					Quantity:  100,
					Reserved:  0,
				},
				{
					ProductID: "item-2",
					Quantity:  2,
					Reserved:  0,
				},
			},
			toUpdate: []*domain.ItemWithQuantity{
				{Id: "item-2", Quantity: 1000},
				{Id: "item-1", Quantity: 200000},
			},
			expectedReserved: map[string]int64{
				"item-1": 0,
				"item-2": 0,
			},
			wantErr: true,
		},
		{
			name: "insufficient_stock_partial",
			stock: []*persistent.StockModel{
				{
					ProductID: "item-1",
					Quantity:  100,
					Reserved:  0,
				},
				{
					ProductID: "item-2",
					Quantity:  2,
					Reserved:  0,
				},
			},
			toUpdate: []*domain.ItemWithQuantity{
				{Id: "item-2", Quantity: 1000},
				{Id: "item-1", Quantity: 1},
			},
			expectedReserved: map[string]int64{
				"item-1": 0,
				"item-2": 0,
			},
			wantErr: true,
		},
		{
			name:  "un_exists_item",
			stock: []*persistent.StockModel{},
			toUpdate: []*domain.ItemWithQuantity{
				{Id: "item-3", Quantity: 1000},
				{Id: "item-1", Quantity: 1},
			},
			expectedReserved: map[string]int64{
				"item-3": 0,
				"item-1": 0,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := setupTestDB(t)
			ctx := context.Background()

			err := db.CreateBatch(ctx, tt.stock)
			require.NoError(t, err)

			repo := NewStockRepositoryMySQL(db)
			err = repo.ReserveStock(ctx, tt.toUpdate, deductStockUpdateFunc)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			// 验证预占库存：用期望值的 key 作为查询范围
			ids := make([]string, 0, len(tt.expectedReserved))
			for id := range tt.expectedReserved {
				ids = append(ids, id)
			}

			stocks, err := db.BatchGetStockByID(ctx, ids)
			require.NoError(t, err)

			// 将结果转成 map，缺失的商品视为 0
			gotReserved := make(map[string]int64, len(tt.expectedReserved))
			for _, id := range ids {
				gotReserved[id] = 0
			}
			for _, stock := range stocks {
				gotReserved[stock.ProductID] = stock.Reserved
			}

			assert.Equal(t, tt.expectedReserved, gotReserved)
		})
	}
}
