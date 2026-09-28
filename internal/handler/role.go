package handler

import (
	"strconv"
	"strings"
	"unicode/utf8"

	"pisa_server/internal/model"
	"pisa_server/internal/pkg/context"
	"pisa_server/internal/pkg/response"
	"pisa_server/internal/pkg/snowflake"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type RoleHandler struct{ db *gorm.DB }

func NewRoleHandler(db *gorm.DB) *RoleHandler { return &RoleHandler{db: db} }

func (h *RoleHandler) seed(tenantID int64) {
	var cnt int64
	h.db.Model(&model.Role{}).Where("tenant_id = ?", tenantID).Count(&cnt)
	if cnt > 0 {
		return
	}
	roles := []model.Role{
		{ID: snowflake.GenID(), TenantID: tenantID, Name: "店长", Description: "所有权限", Permissions: `{"*":true}`, SensitiveData: `{"*":true}`, IsSystem: 1, Status: 1},
		{ID: snowflake.GenID(), TenantID: tenantID, Name: "销售", Description: "仅销售、查报价全选", Permissions: `{"sale.*":true,"customer.quote.*":true}`, SensitiveData: `{}`, IsSystem: 1, Status: 1},
	}
	h.db.Table("roles").Create(&roles)
}

func (h *RoleHandler) List(c *gin.Context) {
	ctx := c.Request.Context()
	tenantID := context.GetTenantID(ctx)
	h.seed(tenantID)
	var list []model.Role
	h.db.Where("tenant_id = ?", tenantID).Order("is_system DESC, id ASC").Find(&list)

	type countRow struct {
		RoleID int64
		Cnt    int64
	}
	var rows []countRow
	h.db.Table("user_tenants").Select("role_id, count(*) as cnt").
		Where("tenant_id = ? AND role_id > 0", tenantID).Group("role_id").Scan(&rows)
	m := map[int64]int64{}
	for _, r := range rows {
		m[r.RoleID] = r.Cnt
	}
	for i := range list {
		list[i].EmployeeCount = m[list[i].ID]
	}
	response.OK(c, list)
}

func (h *RoleHandler) GetByID(c *gin.Context) {
	ctx := c.Request.Context()
	tenantID := context.GetTenantID(ctx)
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var r model.Role
	if err := h.db.Where("id = ? AND tenant_id = ?", id, tenantID).First(&r).Error; err != nil {
		response.NotFound(c, "角色不存在")
		return
	}
	response.OK(c, r)
}

type roleReq struct {
	Name          string `json:"name"`
	Description   string `json:"description"`
	Permissions   string `json:"permissions"`
	SensitiveData string `json:"sensitive_data"`
}

func (h *RoleHandler) Create(c *gin.Context) {
	ctx := c.Request.Context()
	tenantID := context.GetTenantID(ctx)
	var req roleReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	req.Name = strings.TrimSpace(req.Name)
	if n := utf8.RuneCountInString(req.Name); n < 2 || n > 50 {
		response.BadRequest(c, "角色名称需 2-50 个字")
		return
	}
	if req.Permissions == "" {
		req.Permissions = "{}"
	}
	if req.SensitiveData == "" {
		req.SensitiveData = "{}"
	}
	r := model.Role{
		ID:            snowflake.GenID(),
		TenantID:      tenantID,
		Name:          req.Name,
		Description:   req.Description,
		Permissions:   req.Permissions,
		SensitiveData: req.SensitiveData,
		Status:        1,
	}
	h.db.Table("roles").Create(&r)
	response.OK(c, r)
}

func (h *RoleHandler) Update(c *gin.Context) {
	ctx := c.Request.Context()
	tenantID := context.GetTenantID(ctx)
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var r model.Role
	if err := h.db.Where("id = ? AND tenant_id = ?", id, tenantID).First(&r).Error; err != nil {
		response.NotFound(c, "角色不存在")
		return
	}
	var req roleReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	req.Name = strings.TrimSpace(req.Name)
	if n := utf8.RuneCountInString(req.Name); n < 2 || n > 50 {
		response.BadRequest(c, "角色名称需 2-50 个字")
		return
	}
	upd := map[string]interface{}{
		"name":           req.Name,
		"description":    req.Description,
		"permissions":    req.Permissions,
		"sensitive_data": req.SensitiveData,
	}
	if req.Permissions == "" {
		upd["permissions"] = "{}"
	}
	if req.SensitiveData == "" {
		upd["sensitive_data"] = "{}"
	}
	h.db.Table("roles").Where("id = ? AND tenant_id = ?", id, tenantID).Updates(upd)
	response.OKMsg(c, "保存成功")
}

func (h *RoleHandler) Delete(c *gin.Context) {
	ctx := c.Request.Context()
	tenantID := context.GetTenantID(ctx)
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var r model.Role
	if err := h.db.Where("id = ? AND tenant_id = ?", id, tenantID).First(&r).Error; err != nil {
		response.NotFound(c, "角色不存在")
		return
	}
	if r.IsSystem == 1 {
		response.BadRequest(c, "系统内置角色不可删除")
		return
	}
	var cnt int64
	h.db.Table("user_tenants").Where("tenant_id = ? AND role_id = ?", tenantID, id).Count(&cnt)
	if cnt > 0 {
		response.BadRequest(c, "该角色下还有员工，无法删除")
		return
	}
	h.db.Table("roles").Where("id = ? AND tenant_id = ?", id, tenantID).Delete(&model.Role{})
	response.OKMsg(c, "删除成功")
}
