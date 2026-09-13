package db

import (
	"log"
	"pisa_server/internal/model"

	"gorm.io/gorm"
)

// AutoCreateTables 自动创建主库表结构
func AutoCreateTables(mainDB *gorm.DB) {
	tables := []interface{}{
		&model.Plan{},
		&model.Tenant{},
		&model.User{},
		&model.UserTenant{},
		&model.Shop{},
	}
	for _, t := range tables {
		if err := mainDB.AutoMigrate(t); err != nil {
			log.Printf("auto migrate %T error: %v", t, err)
		}
	}
	log.Println("main db tables ready")
}

// AutoCreateShopTables 自动创建业务库非分片表结构
func AutoCreateShopTables(shopDB *gorm.DB) {
	tables := []interface{}{
		&model.GoodsCategory{},
		&model.Unit{},
		&model.GoodsAttribute{},
		&model.Warehouse{},
		&model.Goods{},
		&model.Customer{},
		&model.Supplier{},
		&model.Account{},
		&model.StockBalance{},
	}
	for _, t := range tables {
		if err := shopDB.AutoMigrate(t); err != nil {
			log.Printf("auto migrate shop %T error: %v", t, err)
		}
	}
	log.Println("shop db tables ready")
}
