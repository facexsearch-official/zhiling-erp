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
		&model.Salesman{},
		// 业务数据
		&model.GoodsCategory{},
		&model.Unit{},
		&model.GoodsAttribute{},
		&model.GoodsProperty{},
		&model.Warehouse{},
		&model.Goods{},
		&model.GoodsUnit{},
		&model.GoodsPrice{},
		&model.GoodsStock{},
		&model.Customer{},
		&model.Supplier{},
		&model.SupplierCategory{},
		&model.SupplierAddress{},
		&model.Account{},
		&model.StockBalance{},
		// 采购
		&model.PurchaseOrder{},
		&model.PurchaseOrderItem{},
		&model.Purchase{},
		&model.PurchaseItem{},
		&model.PurchaseReturn{},
		&model.PurchaseReturnItem{},
		// 销售
		&model.SaleOrder{},
		&model.SaleOrderItem{},
		&model.Sale{},
		&model.SaleItem{},
		&model.SalesReturn{},
		&model.SalesReturnItem{},
		&model.Quote{},
		&model.QuoteItem{},
	}
	for _, t := range tables {
		if err := db.AutoMigrate(t); err != nil {
			log.Printf("auto migrate %T error: %v", t, err)
		}
	}
	// 旧的货品多规格表已废弃，结构合并进 goods_prices / goods_stocks
	if db.Migrator().HasTable("goods_specs") {
		if err := db.Migrator().DropTable("goods_specs"); err != nil {
			log.Printf("drop table goods_specs error: %v", err)
		}
	}
	log.Println("db tables ready")
}
