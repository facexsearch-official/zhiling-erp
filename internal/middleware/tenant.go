package middleware

import (
	"pisa_server/internal/pkg/context"
	"pisa_server/internal/pkg/response"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type UserTenant struct {
	ID      int64 `json:"id"`
	UserID  int64 `json:"user_id"`
	TenantID int64 `json:"tenant_id"`
	IsOwner int8  `json:"is_owner"`
	Role    int8  `json:"role"`
	Status  int8  `json:"status"`
}

func TenantMiddleware(mainDB *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		tenantID := c.GetInt64("tenant_id")
		userID := c.GetInt64("user_id")

		if tenantID == 0 || userID == 0 {
			response.Unauthorized(c, "缺少商户上下文")
			c.Abort()
			return
		}

		// Verify membership
		var member UserTenant
		result := mainDB.Where("user_id = ? AND tenant_id = ? AND status = 1", userID, tenantID).First(&member)
		if result.Error != nil {
			response.Forbidden(c, "您已不在该商户中")
			c.Abort()
			return
		}

		// Use DB role (prevents token escalation)
		ctx := context.WithTenant(c.Request.Context(), tenantID, userID, c.GetInt64("shop_id"), int(member.Role), member.IsOwner == 1)
		c.Request = c.Request.WithContext(ctx)
		c.Set("role", int(member.Role))
		c.Set("is_owner", member.IsOwner == 1)
		c.Next()
	}
}
