package db

import (
	"log"
	"pisa_server/internal/model"

	"gorm.io/gorm"
)

// AutoCreateTables 自动创建表结构
func AutoCreateTables(db *gorm.DB) {
	tables := []interface{}{
		// 账户体系
		&model.Plan{},
		&model.Tenant{},
		&model.User{},
		&model.UserTenant{},
		&model.Shop{},
		// 业务数据
		&model.GoodsCategory{},
		&model.Unit{},
		&model.GoodsAttribute{},
		&model.GoodsProperty{},
		&model.Warehouse{},
		&model.Goods{},
		&model.GoodsUnit{},
		&model.GoodsSpec{},
		&model.Customer{},
		&model.Supplier{},
		&model.Account{},
		&model.StockBalance{},
	}
	for _, t := range tables {
		if err := db.AutoMigrate(t); err != nil {
			log.Printf("auto migrate %T error: %v", t, err)
		}
	}
	log.Println("db tables ready")
}
