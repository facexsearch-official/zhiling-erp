package middleware

import (
	"pisa_server/internal/pkg/context"
	"pisa_server/internal/pkg/response"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type TenantPlan struct {
	MaxGoods  int `json:"max_goods"`
	MaxStaff  int `json:"max_staff"`
	MaxShops  int `json:"max_shops"`
	MaxOrders int `json:"max_orders"`
}

type LimitChecker struct {
	mainDB *gorm.DB
	shopDB *gorm.DB
}

func NewLimitChecker(mainDB, shopDB *gorm.DB) *LimitChecker {
	return &LimitChecker{mainDB: mainDB, shopDB: shopDB}
}

func (lc *LimitChecker) Check(tenantID int64, resource string) bool {
	var plan TenantPlan
	lc.mainDB.Where("id = (SELECT plan_id FROM tenants WHERE id = ?)", tenantID).First(&plan)

	var count int64
	switch resource {
	case "goods":
		lc.shopDB.Model(&struct{ TenantID int64 }{}).Table("goods").Where("tenant_id = ?", tenantID).Count(&count)
		return count < int64(plan.MaxGoods) || plan.MaxGoods <= 0
	case "staff":
		lc.mainDB.Model(&struct{ TenantID int64 }{}).Table("user_tenants").Where("tenant_id = ?", tenantID).Count(&count)
		return count < int64(plan.MaxStaff) || plan.MaxStaff <= 0
	case "shop":
		lc.mainDB.Model(&struct{ TenantID int64 }{}).Table("shops").Where("tenant_id = ?", tenantID).Count(&count)
		return count < int64(plan.MaxShops) || plan.MaxShops <= 0
	case "order":
		return plan.MaxOrders <= 0 // 0 = unlimited
	default:
		return true
	}
}

func PlanLimitMiddleware(checker *LimitChecker, resource string) gin.HandlerFunc {
	return func(c *gin.Context) {
		tenantID := context.GetTenantID(c.Request.Context())
		if !checker.Check(tenantID, resource) {
			response.Fail(c, 403, "套餐限额已用完，请升级套餐")
			c.Abort()
			return
		}
		c.Next()
	}
}
