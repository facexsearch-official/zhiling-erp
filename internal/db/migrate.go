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
