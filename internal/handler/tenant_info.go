package handler

import (
	"pisa_server/internal/model"
	"pisa_server/internal/pkg/context"
	"pisa_server/internal/pkg/response"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type TenantInfoHandler struct{ db *gorm.DB }

func NewTenantInfoHandler(db *gorm.DB) *TenantInfoHandler { return &TenantInfoHandler{db: db} }

func (h *TenantInfoHandler) Get(c *gin.Context) {
	ctx := c.Request.Context()
	tenantID := context.GetTenantID(ctx)
	var t model.Tenant
	if err := h.db.Where("id = ?", tenantID).First(&t).Error; err != nil {
		response.NotFound(c, "商户不存在")
		return
	}
	var shop model.Shop
	h.db.Where("tenant_id = ? AND is_main = 1", tenantID).First(&shop)
	if shop.Name != "" {
		t.Name = shop.Name
	}
	if shop.Logo != "" {
		t.Logo = shop.Logo
	}
	if shop.Phone != "" {
		t.ContactPhone = shop.Phone
	}
	if shop.Address != "" {
		t.Address = shop.Address
	}
	response.OK(c, t)
}

type tenantInfoReq struct {
	Name          string `json:"name"`
	Logo          string `json:"logo"`
	BusinessMode  int8   `json:"business_mode"`
	Industry      string `json:"industry"`
	ContactPhone  string `json:"contact_phone"`
	Address       string `json:"address"`
	AddressDetail string `json:"address_detail"`
}

func (h *TenantInfoHandler) Update(c *gin.Context) {
	ctx := c.Request.Context()
	tenantID := context.GetTenantID(ctx)
	var req tenantInfoReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	if req.Name == "" {
		response.BadRequest(c, "请输入商户名称")
		return
	}
	h.db.Model(&model.Tenant{}).Where("id = ?", tenantID).Updates(map[string]interface{}{
		"name": req.Name, "logo": req.Logo, "business_mode": req.BusinessMode, "industry": req.Industry,
		"contact_phone": req.ContactPhone, "address": req.Address, "address_detail": req.AddressDetail,
	})
	var shop model.Shop
	if h.db.Where("tenant_id = ? AND is_main = 1", tenantID).First(&shop).Error == nil {
		h.db.Table("shops").Where("id = ?", shop.ID).Updates(map[string]interface{}{
			"name": req.Name, "logo": req.Logo, "phone": req.ContactPhone,
			"address": req.Address + " " + req.AddressDetail,
		})
	}
	response.OKMsg(c, "保存成功")
}
