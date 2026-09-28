package handler

import (
	"pisa_server/internal/model"
	"pisa_server/internal/pkg/jwt"
	"pisa_server/internal/pkg/perm"
	"pisa_server/internal/pkg/response"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AuthHandler struct {
	db    *gorm.DB
	token *jwt.TokenManager
}

func NewAuthHandler(db *gorm.DB, token *jwt.TokenManager) *AuthHandler {
	return &AuthHandler{db: db, token: token}
}

type LoginRequest struct {
	Phone    string `json:"phone" binding:"required"`
	Password string `json:"password"`
	Code     string `json:"code"`
}

func (h *AuthHandler) Login(c *gin.Context) {
	if h.db == nil {
		response.ServerError(c, "数据库未连接")
		return
	}

	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	var user model.User
	if err := h.db.Where("phone = ?", req.Phone).First(&user).Error; err != nil {
		response.Fail(c, 1, "账号不存在")
		return
	}

	if user.Status != 1 {
		response.Fail(c, 2, "账号已停用")
		return
	}

	// Password login
	if req.Password != "" {
		if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
			response.Fail(c, 3, "密码错误")
			return
		}
	} else if req.Code != "" {
		// SMS login - accept any 6-digit code in dev
		if len(req.Code) != 6 {
			response.Fail(c, 4, "验证码错误")
			return
		}
	} else {
		response.BadRequest(c, "请输入密码或验证码")
		return
	}

	// Update last login
	now := time.Now()
	h.db.Model(&user).Update("last_login_at", &now)

	// Find user's tenants
	var userTenants []model.UserTenant
	h.db.Where("user_id = ? AND status = 1", user.ID).Find(&userTenants)

	type TenantInfo struct {
		TenantID int64  `json:"tenant_id"`
		Name     string `json:"name"`
		Role     int8   `json:"role"`
		RoleName string `json:"role_name"`
	}
	var tenants []TenantInfo
	var firstTenantID int64
	for _, ut := range userTenants {
		var t model.Tenant
		if err := h.db.Where("id = ? AND status = 1", ut.TenantID).First(&t).Error; err != nil {
			continue
		}
		roleName := "操作员"
		switch ut.Role {
		case 1:
			roleName = "主账号"
		case 2:
			roleName = "管理员"
		}
		tenants = append(tenants, TenantInfo{
			TenantID: t.ID,
			Name:     t.Name,
			Role:     ut.Role,
			RoleName: roleName,
		})
		if firstTenantID == 0 {
			firstTenantID = t.ID
		}
	}

	// Generate token for first tenant
	var shopID int64
	if firstTenantID > 0 {
		var shop model.Shop
		if h.db.Where("tenant_id = ? AND is_main = 1", firstTenantID).First(&shop).Error == nil {
			shopID = shop.ID
		}
	}
	role := 1
	if len(userTenants) > 0 {
		role = int(userTenants[0].Role)
	}
	tokenStr, err := h.token.GenerateToken(user.ID, firstTenantID, shopID, role)
	if err != nil {
		response.ServerError(c, "生成Token失败")
		return
	}

	response.OK(c, gin.H{
		"token": tokenStr,
		"user": gin.H{
			"id":       user.ID,
			"phone":    user.Phone,
			"name":     user.Nickname,
			"avatar":   user.Avatar,
			"role":     role,
		},
		"tenants":        tenants,
		"permissions":    perm.Load(h.db, user.ID, firstTenantID).Raw(),
		"sensitive_data": perm.LoadSensitive(h.db, user.ID, firstTenantID).Raw(),
	})
}

type SwitchTenantRequest struct {
	TenantID int64 `json:"tenant_id" binding:"required"`
}

// Me 返回当前登录用户在 token 所属商户下的权限信息
func (h *AuthHandler) Me(c *gin.Context) {
	if h.db == nil {
		response.ServerError(c, "数据库未连接")
		return
	}
	userID := c.GetInt64("user_id")
	tenantID := c.GetInt64("tenant_id")
	var user model.User
	if err := h.db.Where("id = ?", userID).First(&user).Error; err != nil {
		response.Unauthorized(c, "用户不存在")
		return
	}
	var ut model.UserTenant
	h.db.Where("user_id = ? AND tenant_id = ?", userID, tenantID).First(&ut)
	response.OK(c, gin.H{
		"user": gin.H{
			"id": user.ID, "phone": user.Phone, "name": user.Nickname, "avatar": user.Avatar,
		},
		"role":           ut.Role,
		"is_owner":       ut.IsOwner,
		"permissions":    perm.Load(h.db, userID, tenantID).Raw(),
		"sensitive_data": perm.LoadSensitive(h.db, userID, tenantID).Raw(),
	})
}

func (h *AuthHandler) SwitchTenant(c *gin.Context) {
	if h.db == nil {
		response.ServerError(c, "数据库未连接")
		return
	}

	var req SwitchTenantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	userID := c.GetInt64("user_id")

	// Verify user belongs to this tenant
	var ut model.UserTenant
	if err := h.db.Where("user_id = ? AND tenant_id = ? AND status = 1", userID, req.TenantID).First(&ut).Error; err != nil {
		response.Fail(c, 1, "无权访问该商户")
		return
	}

	// Get shop
	var shop model.Shop
	shopID := int64(0)
	if h.db.Where("tenant_id = ? AND is_main = 1", req.TenantID).First(&shop).Error == nil {
		shopID = shop.ID
	}

	tokenStr, err := h.token.GenerateToken(userID, req.TenantID, shopID, int(ut.Role))
	if err != nil {
		response.ServerError(c, "生成Token失败")
		return
	}

	response.OK(c, gin.H{
		"token":          tokenStr,
		"tenant_id":      req.TenantID,
		"shop_id":        shopID,
		"role":           ut.Role,
		"permissions":    perm.Load(h.db, userID, req.TenantID).Raw(),
		"sensitive_data": perm.LoadSensitive(h.db, userID, req.TenantID).Raw(),
	})
}
