package persistent

import (
	"context"
	"fmt"
	"time"

	"github.com/furutachiKurea/gorder/common/logging"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	SockModelTable = "o_stock"
)

type StockModel struct {
	ID        int64     `gorm:"column:id"`
	ProductID string    `gorm:"column:product_id"`
	Quantity  int64     `gorm:"column:quantity"`
	Reserved  int64     `gorm:"column:reserved"`
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
func (d MySQL) BatchGetStockByID(ctx context.Context, productIDs []string) (res []StockModel, err error) {
	_, deferlog := logging.WhenMySQL(ctx, "BatchGetStockByID", productIDs)
	defer deferlog(res, &err)

	err = d.db.WithContext(ctx).
		Model(StockModel{}).
		Clauses(clause.Returning{}).
		Where("product_id IN ?", productIDs).
		Find(&res).Error
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (d MySQL) CreateBatch(ctx context.Context, create []*StockModel) (err error) {
	var returning StockModel
	_, deferlog := logging.WhenMySQL(ctx, "CreateBatch", create)
	defer deferlog(returning, &err)
	err = d.db.WithContext(ctx).Model(&returning).Clauses(clause.Returning{}).Create(create).Error
	return err
}
