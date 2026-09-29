package handler

import (
	"pisa_server/internal/model"
	"pisa_server/internal/pkg/jwt"
	"pisa_server/internal/pkg/perm"
	"pisa_server/internal/pkg/response"
	"pisa_server/internal/pkg/snowflake"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

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
		TenantID    int64  `json:"tenant_id"`
		TenantIDStr string `json:"tenant_id_str"`
		Name        string `json:"name"`
		Role        int8   `json:"role"`
		RoleName    string `json:"role_name"`
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
			TenantID:    t.ID,
			TenantIDStr: strconv.FormatInt(t.ID, 10),
			Name:        t.Name,
			Role:        ut.Role,
			RoleName:    roleName,
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
	TenantID    int64  `json:"tenant_id"`
	TenantIDStr string `json:"tenant_id_str"`
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

type CreateTenantRequest struct {
	Name         string `json:"name"`
	Type         int8   `json:"type"`
	ContactName  string `json:"contact_name"`
	ContactPhone string `json:"contact_phone"`
}

// CreateTenant 当前用户创建新商户，并成为其主账号
func (h *AuthHandler) CreateTenant(c *gin.Context) {
	if h.db == nil {
		response.ServerError(c, "数据库未连接")
		return
	}
	userID := c.GetInt64("user_id")

	var req CreateTenantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		response.BadRequest(c, "请输入商户名称")
		return
	}
	if utf8.RuneCountInString(req.Name) > 128 {
		response.BadRequest(c, "商户名称过长")
		return
	}
	if req.Type == 0 {
		req.Type = 1
	}

	var user model.User
	if err := h.db.Where("id = ?", userID).First(&user).Error; err != nil {
		response.Unauthorized(c, "用户不存在")
		return
	}

	now := time.Now()
	tenant := model.Tenant{
		ID:           snowflake.GenID(),
		Name:         req.Name,
		Type:         req.Type,
		ContactName:  strings.TrimSpace(req.ContactName),
		ContactPhone: strings.TrimSpace(req.ContactPhone),
		OwnerUserID:  userID,
		Status:       1,
	}
	shop := model.Shop{
		TenantID: tenant.ID,
		Name:     req.Name,
		IsMain:   1,
		Status:   1,
	}
	ut := model.UserTenant{
		UserID:     userID,
		TenantID:   tenant.ID,
		IsOwner:    1,
		Role:       1,
		StaffName:  user.Nickname,
		StaffPhone: user.Phone,
		JoinedAt:   now,
		Status:     1,
	}

	tx := h.db.Begin()
	if err := tx.Create(&tenant).Error; err != nil {
		tx.Rollback()
		response.ServerError(c, "创建商户失败")
		return
	}
	if err := tx.Create(&shop).Error; err != nil {
		tx.Rollback()
		response.ServerError(c, "创建商户失败")
		return
	}
	if err := tx.Create(&ut).Error; err != nil {
		tx.Rollback()
		response.ServerError(c, "创建商户失败")
		return
	}
	tx.Commit()

	tokenStr, err := h.token.GenerateToken(userID, tenant.ID, shop.ID, 1)
	if err != nil {
		response.ServerError(c, "生成Token失败")
		return
	}

	response.OK(c, gin.H{
		"token":          tokenStr,
		"tenant":         gin.H{"tenant_id": tenant.ID, "tenant_id_str": strconv.FormatInt(tenant.ID, 10), "name": tenant.Name, "role": 1, "role_name": "主账号"},
		"tenant_id":      tenant.ID,
		"shop_id":        shop.ID,
		"permissions":    perm.Load(h.db, userID, tenant.ID).Raw(),
		"sensitive_data": perm.LoadSensitive(h.db, userID, tenant.ID).Raw(),
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

	targetTenantID := req.TenantID
	if req.TenantIDStr != "" {
		if v, err := strconv.ParseInt(req.TenantIDStr, 10, 64); err == nil {
			targetTenantID = v
		}
	}
	if targetTenantID == 0 {
		response.BadRequest(c, "参数错误")
		return
	}

	// Verify user belongs to this tenant
	var ut model.UserTenant
	if err := h.db.Where("user_id = ? AND tenant_id = ? AND status = 1", userID, targetTenantID).First(&ut).Error; err != nil {
		response.Fail(c, 1, "无权访问该商户")
		return
	}

	// Get shop
	var shop model.Shop
	shopID := int64(0)
	if h.db.Where("tenant_id = ? AND is_main = 1", targetTenantID).First(&shop).Error == nil {
		shopID = shop.ID
	}

	tokenStr, err := h.token.GenerateToken(userID, targetTenantID, shopID, int(ut.Role))
	if err != nil {
		response.ServerError(c, "生成Token失败")
		return
	}

	response.OK(c, gin.H{
		"token":          tokenStr,
		"tenant_id":      targetTenantID,
		"shop_id":        shopID,
		"role":           ut.Role,
		"permissions":    perm.Load(h.db, userID, targetTenantID).Raw(),
		"sensitive_data": perm.LoadSensitive(h.db, userID, targetTenantID).Raw(),
	})
}
