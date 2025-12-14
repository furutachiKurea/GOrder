package adapter

import (
	"context"
	"fmt"
	"sync"
	"testing"

	_ "github.com/furutachiKurea/gorder/common/config"
	"github.com/furutachiKurea/gorder/common/entity"
	"github.com/furutachiKurea/gorder/stock/infrastructure/persistent"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/sync/errgroup"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

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
			err := repo.ReserveStock(ctx, []*entity.ItemWithQuantity{{Id: testItem, Quantity: 1}})
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
			_ = repo.ReserveStock(ctx, []*entity.ItemWithQuantity{{Id: testItem, Quantity: 1}})
		}()
	}
	wg.Wait()

	res, err := db.BatchGetStockByID(ctx, []string{testItem})
	assert.NoError(t, err)
	assert.NotEmpty(t, res)
	assert.GreaterOrEqual(t, res[0].Quantity-res[0].Reserved, int64(0))
}

func TestStockRepositoryMySQL_ReserveStock(t *testing.T) {
	tests := []struct {
		name             string
		stock            []*persistent.StockModel
		toUpdate         []*entity.ItemWithQuantity
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
			toUpdate: []*entity.ItemWithQuantity{
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
			toUpdate: []*entity.ItemWithQuantity{
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
			toUpdate: []*entity.ItemWithQuantity{
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
			name: "un_exists_item",
			stock: []*persistent.StockModel{
				{
					ProductID: "item-not-exists",
					Quantity:  0,
					Reserved:  0,
				},
			},
			toUpdate: []*entity.ItemWithQuantity{
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
			err = repo.ReserveStock(ctx, tt.toUpdate)

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

func TestStockRepositoryMySQL_ConfirmStockReservation(t *testing.T) {
	tests := []struct {
		name             string
		initialStock     []*persistent.StockModel
		toConfirm        []*entity.ItemWithQuantity
		expectedQuantity map[string]int64
		expectedReserved map[string]int64
		wantErr          bool
	}{
		{
			name: "success",
			initialStock: []*persistent.StockModel{
				{ProductID: "item-1", Quantity: 100, Reserved: 10},
				{ProductID: "item-2", Quantity: 50, Reserved: 5},
			},
			toConfirm: []*entity.ItemWithQuantity{
				{Id: "item-1", Quantity: 5},
				{Id: "item-2", Quantity: 3},
			},
			expectedQuantity: map[string]int64{
				"item-1": 95,
				"item-2": 47,
			},
			expectedReserved: map[string]int64{
				"item-1": 5,
				"item-2": 2,
			},
			wantErr: false,
		},
		{
			name: "over sell",
			initialStock: []*persistent.StockModel{
				{ProductID: "item-1", Quantity: 100, Reserved: 3},
			},
			toConfirm: []*entity.ItemWithQuantity{
				{Id: "item-1", Quantity: 5}, // 超卖了
			},
			expectedQuantity: map[string]int64{
				"item-1": 100,
			},
			expectedReserved: map[string]int64{
				"item-1": 3,
			},
			wantErr: true,
		},
		{
			name: "item not exists",
			initialStock: []*persistent.StockModel{
				{ProductID: "item-1", Quantity: 100, Reserved: 10},
			},
			toConfirm: []*entity.ItemWithQuantity{
				{Id: "item-2", Quantity: 5},
			},
			expectedQuantity: map[string]int64{
				"item-1": 100,
			},
			expectedReserved: map[string]int64{
				"item-1": 10,
			},
			wantErr: true,
		},
		{
			name: "failed partial items",
			initialStock: []*persistent.StockModel{
				{ProductID: "item-1", Quantity: 100, Reserved: 20},
				{ProductID: "item-2", Quantity: 50, Reserved: 0},
				{ProductID: "item-3", Quantity: 30, Reserved: 5},
			},
			toConfirm: []*entity.ItemWithQuantity{
				{Id: "item-1", Quantity: 10},
				{Id: "item-2", Quantity: 5},
				{Id: "item-3", Quantity: 2},
			},
			expectedQuantity: map[string]int64{
				"item-1": 100,
				"item-2": 50,
				"item-3": 30,
			},
			expectedReserved: map[string]int64{
				"item-1": 20,
				"item-2": 0,
				"item-3": 5,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := setupTestDB(t)
			ctx := context.Background()

			// 设置初始数据
			err := db.CreateBatch(ctx, tt.initialStock)
			require.NoError(t, err)

			repo := NewStockRepositoryMySQL(db)
			err = repo.ConfirmStockReservation(ctx, tt.toConfirm)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			ids := make([]string, len(tt.initialStock))
			for i, stock := range tt.initialStock {
				ids[i] = stock.ProductID
			}

			stocks, err := db.BatchGetStockByID(ctx, ids)
			require.NoError(t, err)

			// 验证库存数量
			if tt.expectedQuantity != nil {
				for _, stock := range stocks {
					if expected, exists := tt.expectedQuantity[stock.ProductID]; exists {
						assert.Equal(t, expected, stock.Quantity)
					}
				}
			}

			// 验证预占数量
			if tt.expectedReserved != nil {
				for _, stock := range stocks {
					if expected, exists := tt.expectedReserved[stock.ProductID]; exists {
						assert.Equal(t, expected, stock.Reserved)
					}
				}
			}
		})
	}
}
