package db

import (
	"log"
	"pisa_server/internal/model"
	"pisa_server/internal/pkg/snowflake"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func Seed(db *gorm.DB) {
	seedPlans(db)
	seedUsers(db)
	seedTenants(db)
	seedShops(db)
	seedGoodsCategories(db)
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

	ut := model.UserTenant{
		UserID:    admin.ID,
		TenantID:  tenant.ID,
		IsOwner:   1,
		Role:      1,
		StaffName: "管理员",
		JoinedAt:  time.Now(),
		Status:    1,
	}
	if err := db.Create(&ut).Error; err != nil {
		log.Printf("seed user_tenant error: %v", err)
		return
	}

	db.Model(&admin).Update("default_tenant_id", tenant.ID)
	log.Printf("seed tenant: name=演示商户, id=%d", tenant.ID)
}

func seedShops(db *gorm.DB) {
	var count int64
	db.Model(&model.Shop{}).Where("name = ?", "总店").Count(&count)
	if count > 0 {
		return
	}

	var tenant model.Tenant
	db.Where("name = ?", "演示商户").First(&tenant)
	if tenant.ID == 0 {
		return
	}

	shop := model.Shop{
		TenantID: tenant.ID,
		Name:     "总店",
		Address:  "演示地址",
		IsMain:   1,
		Status:   1,
	}
	if err := db.Create(&shop).Error; err != nil {
		log.Printf("seed shop error: %v", err)
		return
	}
	log.Printf("seed shop: name=总店, id=%d", shop.ID)
}

func seedGoodsCategories(db *gorm.DB) {
	var tenant model.Tenant
	db.Where("name = ?", "演示商户").First(&tenant)
	if tenant.ID == 0 {
		return
	}

	// 先清空该商户所有商品分类
	db.Where("tenant_id = ?", tenant.ID).Delete(&model.GoodsCategory{})

	type catDef struct {
		Name  string
		Sort  int
		Kids  []catDef
	}
	tree := []catDef{
		{Name: "食品饮料", Sort: 1, Kids: []catDef{
			{Name: "休闲零食", Sort: 1},
			{Name: "乳制品", Sort: 2},
			{Name: "饮料冲调", Sort: 3},
			{Name: "调味品", Sort: 4},
		}},
		{Name: "日用百货", Sort: 2, Kids: []catDef{
			{Name: "清洁用品", Sort: 1},
			{Name: "纸品家居", Sort: 2},
			{Name: "个人护理", Sort: 3},
		}},
		{Name: "数码家电", Sort: 3, Kids: []catDef{
			{Name: "手机配件", Sort: 1},
			{Name: "电脑配件", Sort: 2},
			{Name: "厨房电器", Sort: 3},
		}},
		{Name: "服装鞋帽", Sort: 4, Kids: []catDef{
			{Name: "男装", Sort: 1},
			{Name: "女装", Sort: 2},
			{Name: "鞋靴", Sort: 3},
		}},
		{Name: "办公用品", Sort: 5, Kids: []catDef{
			{Name: "文具", Sort: 1},
			{Name: "打印耗材", Sort: 2},
		}},
	}

	var created int
	for _, t := range tree {
		parent := model.GoodsCategory{
			TenantID: tenant.ID,
			Name:     t.Name,
			ParentID: 0,
			Sort:     t.Sort,
		}
		if err := db.Create(&parent).Error; err != nil {
			log.Printf("seed category %s error: %v", t.Name, err)
			continue
		}
		created++
		for _, k := range t.Kids {
			child := model.GoodsCategory{
				TenantID: tenant.ID,
				Name:     k.Name,
				ParentID: parent.ID,
				Sort:     k.Sort,
			}
			if err := db.Create(&child).Error; err != nil {
				log.Printf("seed category %s error: %v", k.Name, err)
				continue
			}
			created++
		}
	}
	log.Printf("seed goods categories: %d 条", created)
}
