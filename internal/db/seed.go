package db

import (
	"log"
	"pisa_server/internal/model"
	"pisa_server/internal/pkg/snowflake"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// Seed 初始化种子数据
func Seed(mainDB, shopDB *gorm.DB) {
	seedUsers(mainDB)
	seedTenants(mainDB)
	seedShops(mainDB)
	seedPlans(mainDB)
}

func seedPlans(db *gorm.DB) {
	plans := []model.Plan{
		{ID: 1, Code: "free", Name: "免费版", Price: 0, MaxGoods: 100, MaxStaff: 1, MaxShops: 1, MaxOrders: 100, Sort: 1, Status: 1},
		{ID: 2, Code: "basic", Name: "开单版", Price: 398, MaxGoods: 500, MaxStaff: 3, MaxShops: 1, MaxOrders: 0, Sort: 2, Status: 1},
		{ID: 3, Code: "single", Name: "单店版", Price: 998, MaxGoods: 2000, MaxStaff: 10, MaxShops: 1, MaxOrders: 0, Sort: 3, Status: 1},
		{ID: 4, Code: "multi", Name: "多店版", Price: 1998, MaxGoods: 0, MaxStaff: 0, MaxShops: 0, MaxOrders: 0, Sort: 4, Status: 1},
	}
	for _, p := range plans {
		db.Where("id = ?", p.ID).FirstOrCreate(&p)
	}
}

func seedUsers(db *gorm.DB) {
	// Check if admin exists
	var count int64
	db.Model(&model.User{}).Where("phone = ?", "admin").Count(&count)
	if count > 0 {
		return
	}

	hash, _ := bcrypt.GenerateFromPassword([]byte("123456"), bcrypt.DefaultCost)
	admin := model.User{
		ID:           snowflake.GenID(),
		Phone:        "admin",
		PasswordHash: string(hash),
		Nickname:     "管理员",
		Status:       1,
	}
	if err := db.Create(&admin).Error; err != nil {
		log.Printf("seed admin user error: %v", err)
		return
	}
	log.Printf("seed admin user: phone=admin, password=123456, id=%d", admin.ID)
}

func seedTenants(db *gorm.DB) {
	var count int64
	db.Model(&model.Tenant{}).Where("name = ?", "演示商户").Count(&count)
	if count > 0 {
		return
	}

	var admin model.User
	db.Where("phone = ?", "admin").First(&admin)
	if admin.ID == 0 {
		return
	}

	tenant := model.Tenant{
		ID:           snowflake.GenID(),
		Name:         "演示商户",
		Type:         1,
		ContactName:  "管理员",
		ContactPhone: "admin",
		OwnerUserID:  admin.ID,
		PlanID:       3,
		MaxGoods:     2000,
		MaxStaff:     10,
		MaxShops:     1,
		MaxOrders:    0,
		Status:       1,
	}
	if err := db.Create(&tenant).Error; err != nil {
		log.Printf("seed tenant error: %v", err)
		return
	}

	// Link user to tenant
	ut := model.UserTenant{
		ID:        snowflake.GenID(),
		UserID:    admin.ID,
		TenantID:  tenant.ID,
		IsOwner:   1,
		Role:      1,
		StaffName: "管理员",
		Status:    1,
	}
	db.Create(&ut)

	// Update user default tenant
	db.Model(&admin).Update("default_tenant_id", tenant.ID)

	log.Printf("seed tenant: name=演示商户, id=%d", tenant.ID)
}

func seedShops(db *gorm.DB) {
	var count int64
	db.Model(&model.Shop{}).Where("tenant_id = ? AND name = ?", 1, "总店").Count(&count)
	if count > 0 {
		return
	}

	var tenant model.Tenant
	db.Where("name = ?", "演示商户").First(&tenant)
	if tenant.ID == 0 {
		return
	}

	shop := model.Shop{
		ID:       snowflake.GenID(),
		TenantID: tenant.ID,
		Name:     "总店",
		Address:  "演示地址",
		IsMain:   1,
		Status:   1,
	}
	db.Create(&shop)
	log.Printf("seed shop: name=总店, id=%d", shop.ID)
}
