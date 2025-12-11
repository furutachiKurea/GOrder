package persistent

import (
	"context"
	"fmt"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

const (
	SockModelTable = "o_stock"
)

type StockModel struct {
	ID        int64     `gorm:"column:id"`
	ProductID string    `gorm:"column:product_id"`
	Quantity  int64     `gorm:"column:quantity"`
	CreatedAt time.Time `gorm:"column:created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at"`
}

func (s StockModel) TableName() string {
	return SockModelTable
}

type MySQL struct {
	db *gorm.DB
}

func NewMySQL() *MySQL {
	cfg := viper.Sub("mysql")

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.GetString("user"),
		cfg.GetString("password"),
		cfg.GetString("host"),
		cfg.GetString("port"),
		cfg.GetString("database"),
	)
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Panic().Err(err).Msg("connect to mysql failed")
	}

	return &MySQL{db: db}
}

func NewMySQLWithDB(db *gorm.DB) *MySQL {
	return &MySQL{db: db}
}

func (d MySQL) StartTransaction(fc func(tx *gorm.DB) error) error {
	return d.db.Transaction(fc)
}

// BatchGetStockByID 从数据库中使用 product IDs 批量获取库存信息
func (d MySQL) BatchGetStockByID(ctx context.Context, productIDs []string) ([]StockModel, error) {
	var res []StockModel
	tx := d.db.WithContext(ctx).Where("product_id IN ?", productIDs).Find(&res)
	if tx.Error != nil {
		return nil, tx.Error
	}

	return res, nil
}

func (d MySQL) CreateBatch(ctx context.Context, create []*StockModel) error {
	return gorm.G[*StockModel](d.db).CreateInBatches(ctx, &create, 100)
}
