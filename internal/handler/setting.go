package handler

import (
	"encoding/json"
	"io"
	"time"

	"pisa_server/internal/model"
	"pisa_server/internal/pkg/context"
	"pisa_server/internal/pkg/response"
	"pisa_server/internal/pkg/snowflake"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type SystemSettingHandler struct{ db *gorm.DB }

func NewSystemSettingHandler(db *gorm.DB) *SystemSettingHandler {
	return &SystemSettingHandler{db: db}
}

// Get 返回当前商户的系统设置（无则返回空对象）
func (h *SystemSettingHandler) Get(c *gin.Context) {
	ctx := c.Request.Context()
	tenantID := context.GetTenantID(ctx)
	var s model.TenantSetting
	if err := h.db.Where("tenant_id = ?", tenantID).First(&s).Error; err != nil {
		response.OK(c, gin.H{})
		return
	}
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(s.Data), &data); err != nil || data == nil {
		data = map[string]interface{}{}
	}
	response.OK(c, data)
}

// Update 整体保存系统设置
func (h *SystemSettingHandler) Update(c *gin.Context) {
	ctx := c.Request.Context()
	tenantID := context.GetTenantID(ctx)
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	var obj map[string]interface{}
	if err := json.Unmarshal(body, &obj); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	var s model.TenantSetting
	if err := h.db.Where("tenant_id = ?", tenantID).First(&s).Error; err != nil {
		s = model.TenantSetting{ID: snowflake.GenID(), TenantID: tenantID, Data: string(body), UpdatedAt: time.Now()}
		h.db.Table("tenant_settings").Create(&s)
	} else {
		h.db.Table("tenant_settings").Where("tenant_id = ?", tenantID).
			Updates(map[string]interface{}{"data": string(body), "updated_at": time.Now()})
	}
	response.OKMsg(c, "保存成功")
}

type UserPreferenceHandler struct{ db *gorm.DB }

func NewUserPreferenceHandler(db *gorm.DB) *UserPreferenceHandler {
	return &UserPreferenceHandler{db: db}
}

// Get 返回当前登录账号的偏好设置（无则返回空对象）
func (h *UserPreferenceHandler) Get(c *gin.Context) {
	ctx := c.Request.Context()
	tenantID := context.GetTenantID(ctx)
	userID := context.GetUserID(ctx)
	var s model.UserPreference
	if err := h.db.Where("user_id = ? AND tenant_id = ?", userID, tenantID).First(&s).Error; err != nil {
		response.OK(c, gin.H{})
		return
	}
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(s.Data), &data); err != nil || data == nil {
		data = map[string]interface{}{}
	}
	response.OK(c, data)
}

// Update 整体保存当前登录账号的偏好设置
func (h *UserPreferenceHandler) Update(c *gin.Context) {
	ctx := c.Request.Context()
	tenantID := context.GetTenantID(ctx)
	userID := context.GetUserID(ctx)
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	var obj map[string]interface{}
	if err := json.Unmarshal(body, &obj); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	var s model.UserPreference
	if err := h.db.Where("user_id = ? AND tenant_id = ?", userID, tenantID).First(&s).Error; err != nil {
		s = model.UserPreference{ID: snowflake.GenID(), UserID: userID, TenantID: tenantID, Data: string(body), UpdatedAt: time.Now()}
		h.db.Table("user_preferences").Create(&s)
	} else {
		h.db.Table("user_preferences").Where("user_id = ? AND tenant_id = ?", userID, tenantID).
			Updates(map[string]interface{}{"data": string(body), "updated_at": time.Now()})
	}
	response.OKMsg(c, "保存成功")
}

// PointsSettingHandler 积分设置（存储于 tenant_settings，按 key 合并）
type PointsSettingHandler struct{ db *gorm.DB }

func NewPointsSettingHandler(db *gorm.DB) *PointsSettingHandler {
	return &PointsSettingHandler{db: db}
}

var pointsKeys = map[string]bool{"points_enabled": true, "points_per": true, "points_reward": true}

func (h *PointsSettingHandler) Get(c *gin.Context) {
	ctx := c.Request.Context()
	tenantID := context.GetTenantID(ctx)
	var s model.TenantSetting
	if err := h.db.Where("tenant_id = ?", tenantID).First(&s).Error; err != nil {
		response.OK(c, gin.H{})
		return
	}
	var data map[string]interface{}
	json.Unmarshal([]byte(s.Data), &data)
	out := gin.H{}
	for k := range pointsKeys {
		if v, ok := data[k]; ok {
			out[k] = v
		}
	}
	response.OK(c, out)
}

func (h *PointsSettingHandler) Update(c *gin.Context) {
	ctx := c.Request.Context()
	tenantID := context.GetTenantID(ctx)
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	var obj map[string]interface{}
	if err := json.Unmarshal(body, &obj); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	var s model.TenantSetting
	existing := map[string]interface{}{}
	if err := h.db.Where("tenant_id = ?", tenantID).First(&s).Error; err == nil {
		json.Unmarshal([]byte(s.Data), &existing)
	}
	for k, v := range obj {
		existing[k] = v
	}
	merged, _ := json.Marshal(existing)
	if s.ID == 0 {
		s = model.TenantSetting{ID: snowflake.GenID(), TenantID: tenantID, Data: string(merged), UpdatedAt: time.Now()}
		h.db.Table("tenant_settings").Create(&s)
	} else {
		h.db.Table("tenant_settings").Where("tenant_id = ?", tenantID).
			Updates(map[string]interface{}{"data": string(merged), "updated_at": time.Now()})
	}
	response.OKMsg(c, "保存成功")
}

// PrintSettingHandler 打印设置（存储于 tenant_settings 的 print 键）
type PrintSettingHandler struct{ db *gorm.DB }

func NewPrintSettingHandler(db *gorm.DB) *PrintSettingHandler {
	return &PrintSettingHandler{db: db}
}

func (h *PrintSettingHandler) Get(c *gin.Context) {
	ctx := c.Request.Context()
	tenantID := context.GetTenantID(ctx)
	var s model.TenantSetting
	if err := h.db.Where("tenant_id = ?", tenantID).First(&s).Error; err != nil {
		response.OK(c, gin.H{})
		return
	}
	var data map[string]interface{}
	json.Unmarshal([]byte(s.Data), &data)
	if v, ok := data["print"].(map[string]interface{}); ok {
		response.OK(c, v)
		return
	}
	response.OK(c, gin.H{})
}

func (h *PrintSettingHandler) Update(c *gin.Context) {
	ctx := c.Request.Context()
	tenantID := context.GetTenantID(ctx)
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	var obj map[string]interface{}
	if err := json.Unmarshal(body, &obj); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	var s model.TenantSetting
	existing := map[string]interface{}{}
	if err := h.db.Where("tenant_id = ?", tenantID).First(&s).Error; err == nil {
		json.Unmarshal([]byte(s.Data), &existing)
	}
	existing["print"] = obj
	merged, _ := json.Marshal(existing)
	if s.ID == 0 {
		s = model.TenantSetting{ID: snowflake.GenID(), TenantID: tenantID, Data: string(merged), UpdatedAt: time.Now()}
		h.db.Table("tenant_settings").Create(&s)
	} else {
		h.db.Table("tenant_settings").Where("tenant_id = ?", tenantID).
			Updates(map[string]interface{}{"data": string(merged), "updated_at": time.Now()})
	}
	response.OKMsg(c, "保存成功")
}
