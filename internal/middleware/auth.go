package middleware

import (
	"strings"

	"pisa_server/internal/pkg/jwt"
	"pisa_server/internal/pkg/response"

	"github.com/gin-gonic/gin"
)

func Auth(tm *jwt.TokenManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Unauthorized(c, "未登录")
			c.Abort()
			return
		}
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			response.Unauthorized(c, "Token格式错误")
			c.Abort()
			return
		}
		claims, err := tm.ParseToken(parts[1])
		if err != nil {
			response.Unauthorized(c, "Token无效或已过期")
			c.Abort()
			return
		}
		c.Set("user_id", claims.UserID)
		c.Set("tenant_id", claims.TenantID)
		c.Set("shop_id", claims.ShopID)
		c.Set("role", claims.Role)
		c.Next()
	}
}
