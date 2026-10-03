package main

import (
	"log"

	"pisa_server/internal/config"
	"pisa_server/internal/db"
	"pisa_server/internal/model"
	"pisa_server/internal/pkg/snowflake"

	"gorm.io/gorm/logger"
)

// 独立 mock 命令：清空并重建演示商户的全部测试数据。
// 用法：go run ./cmd/mock
func main() {
	if err := config.Load("config.yaml"); err != nil {
		log.Fatalf("load config: %v", err)
	}
	if err := snowflake.Init(1); err != nil {
		log.Printf("snowflake init warning: %v", err)
	}
	database, err := db.Connect(config.GlobalConfig.Database)
	if err != nil {
		log.Fatalf("db connect: %v", err)
	}
	database.Logger = logger.Default.LogMode(logger.Silent)
	db.AutoCreateTables(database)
	db.Seed(database)

	var tenant model.Tenant
	if err := database.Where("name = ?", "演示商户").First(&tenant).Error; err != nil || tenant.ID == 0 {
		log.Fatalf("演示商户不存在，无法 mock")
	}
	db.SeedMock(database, tenant.ID)
}
