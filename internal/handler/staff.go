package handler

import (
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"pisa_server/internal/model"
	"pisa_server/internal/pkg/context"
	"pisa_server/internal/pkg/response"
	"pisa_server/internal/pkg/snowflake"
	"pisa_server/internal/repository"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

type StaffHandler struct {
	repo *repository.StaffRepository
}

func NewStaffHandler(repo *repository.StaffRepository) *StaffHandler {
	return &StaffHandler{repo: repo}
}

/* ────────── 用户 ────────── */

// ListUsers 用户列表 + 套餐用量
func (h *StaffHandler) ListUsers(c *gin.Context) {
	ctx := c.Request.Context()
	rows, err := h.repo.ListUsers(ctx)
	if err != nil {
		response.ServerError(c, "查询用户失败")
		return
	}
	uid := context.GetUserID(ctx)
	for i := range rows {
		rows[i].IsCurrent = rows[i].UserID == uid
	}
	tenantID := context.GetTenantID(ctx)
	used := len(rows)
	maxStaff := 0
	if t, err := h.repo.GetTenant(ctx, tenantID); err == nil {
		maxStaff = t.MaxStaff
	}
	remaining := maxStaff - used
	if maxStaff <= 0 {
		remaining = -1 // 不限
	}
	response.OK(c, gin.H{
		"list":      rows,
		"max_staff": maxStaff,
		"used":      used,
		"remaining": remaining,
	})
}

type createUserReq struct {
	Name     string `json:"name"`
	Phone    string `json:"phone"`
	Role     int8   `json:"role"`
	Password string `json:"password"`
}

// CreateUser 新增用户（创建登录账号并加入当前商户）
func (h *StaffHandler) CreateUser(c *gin.Context) {
	ctx := c.Request.Context()
	var req createUserReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	req.Name = strings.TrimSpace(req.Name)
	if n := utf8.RuneCountInString(req.Name); n < 2 || n > 20 {
		response.BadRequest(c, "姓名需 2-20 个字")
		return
	}
	req.Phone = strings.TrimSpace(req.Phone)
	if req.Phone == "" {
		response.BadRequest(c, "请输入手机号")
		return
	}
	if req.Role != 2 && req.Role != 3 {
		req.Role = 3
	}
	if len(req.Password) < 6 {
		response.BadRequest(c, "初始密码至少 6 位")
		return
	}
	tenantID := context.GetTenantID(ctx)

	user, err := h.repo.FindUserByPhone(req.Phone)
	if err != nil {
		hash, _ := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		user = &model.User{
			ID:           snowflake.GenID(),
			Phone:        req.Phone,
			PasswordHash: string(hash),
			Nickname:     req.Name,
			Status:       1,
		}
		if err := h.repo.CreateUser(user); err != nil {
			response.ServerError(c, "创建账号失败")
			return
		}
	} else if h.repo.CountUserInTenant(user.ID, tenantID) > 0 {
		response.Fail(c, 1, "该手机号已在该商户中")
		return
	}

	ut := &model.UserTenant{
		UserID:     user.ID,
		TenantID:   tenantID,
		IsOwner:    0,
		Role:       req.Role,
		StaffName:  req.Name,
		StaffPhone: req.Phone,
		Status:     1,
		JoinedAt:   time.Now(),
	}
	if err := h.repo.CreateUserTenant(ut); err != nil {
		response.ServerError(c, "添加用户失败")
		return
	}
	response.OK(c, gin.H{"id": ut.ID})
}

type updateUserReq struct {
	Name   string `json:"name"`
	Role   int8   `json:"role"`
	Status *int8  `json:"status"`
}

// UpdateUser 修改用户（姓名 / 角色 / 状态）
func (h *StaffHandler) UpdateUser(c *gin.Context) {
	ctx := c.Request.Context()
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	ut, err := h.repo.GetUserTenant(ctx, id)
	if err != nil {
		response.NotFound(c, "用户不存在")
		return
	}
	var req updateUserReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	if req.Name != "" {
		req.Name = strings.TrimSpace(req.Name)
		if n := utf8.RuneCountInString(req.Name); n < 2 || n > 20 {
			response.BadRequest(c, "姓名需 2-20 个字")
			return
		}
		ut.StaffName = req.Name
		if u, err := h.repo.FindUserByPhone(ut.StaffPhone); err == nil {
			u.Nickname = req.Name
			_ = h.repo.UpdateUser(u)
		}
	}
	// 主账号不可变更角色/停用
	if ut.IsOwner == 0 {
		if req.Role == 2 || req.Role == 3 {
			ut.Role = req.Role
		}
		if req.Status != nil && (*req.Status == 0 || *req.Status == 1) {
			ut.Status = *req.Status
		}
	}
	if err := h.repo.UpdateUserTenant(ctx, ut); err != nil {
		response.ServerError(c, "更新用户失败")
		return
	}
	response.OK(c, ut)
}

// DeleteUser 移除用户
func (h *StaffHandler) DeleteUser(c *gin.Context) {
	ctx := c.Request.Context()
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	ut, err := h.repo.GetUserTenant(ctx, id)
	if err != nil {
		response.NotFound(c, "用户不存在")
		return
	}
	if ut.IsOwner == 1 {
		response.BadRequest(c, "主账号不可删除")
		return
	}
	if err := h.repo.DeleteUserTenant(ctx, id); err != nil {
		response.ServerError(c, "删除失败")
		return
	}
	if h.repo.CountUserTenantsByUser(ut.UserID) == 0 {
		if u, err := h.repo.FindUserByPhone(ut.StaffPhone); err == nil {
			u.Status = 0
			_ = h.repo.UpdateUser(u)
		}
	}
	response.OKMsg(c, "删除成功")
}

/* ────────── 业务员 ────────── */

func (h *StaffHandler) ListSalesmen(c *gin.Context) {
	list, err := h.repo.ListSalesmen(c.Request.Context())
	if err != nil {
		response.ServerError(c, "查询业务员失败")
		return
	}
	response.OK(c, list)
}

type salesmanReq struct {
	Name   string `json:"name"`
	Phone  string `json:"phone"`
	ShopID int64  `json:"shop_id"`
	Status *int8  `json:"status"`
}

func (h *StaffHandler) CreateSalesman(c *gin.Context) {
	ctx := c.Request.Context()
	var req salesmanReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	req.Name = strings.TrimSpace(req.Name)
	if n := utf8.RuneCountInString(req.Name); n < 2 || n > 20 {
		response.BadRequest(c, "业务员姓名需 2-20 个字")
		return
	}
	s := &model.Salesman{
		TenantID: context.GetTenantID(ctx),
		Name:     req.Name,
		Phone:    strings.TrimSpace(req.Phone),
		ShopID:   req.ShopID,
		Status:   1,
	}
	if err := h.repo.CreateSalesman(ctx, s); err != nil {
		response.ServerError(c, "新增业务员失败")
		return
	}
	response.OK(c, s)
}

func (h *StaffHandler) UpdateSalesman(c *gin.Context) {
	ctx := c.Request.Context()
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	s, err := h.repo.GetSalesman(ctx, id)
	if err != nil {
		response.NotFound(c, "业务员不存在")
		return
	}
	var req salesmanReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	if req.Name != "" {
		req.Name = strings.TrimSpace(req.Name)
		if n := utf8.RuneCountInString(req.Name); n < 2 || n > 20 {
			response.BadRequest(c, "业务员姓名需 2-20 个字")
			return
		}
		s.Name = req.Name
	}
	s.Phone = strings.TrimSpace(req.Phone)
	if req.ShopID != 0 {
		s.ShopID = req.ShopID
	}
	if req.Status != nil && (*req.Status == 0 || *req.Status == 1) {
		s.Status = *req.Status
	}
	if err := h.repo.UpdateSalesman(ctx, s); err != nil {
		response.ServerError(c, "更新业务员失败")
		return
	}
	response.OK(c, s)
}

func (h *StaffHandler) DeleteSalesman(c *gin.Context) {
	ctx := c.Request.Context()
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	if _, err := h.repo.GetSalesman(ctx, id); err != nil {
		response.NotFound(c, "业务员不存在")
		return
	}
	if err := h.repo.DeleteSalesman(ctx, id); err != nil {
		response.ServerError(c, "删除失败")
		return
	}
	response.OKMsg(c, "删除成功")
}

/* ────────── 门店 ────────── */

func (h *StaffHandler) ListShops(c *gin.Context) {
	list, err := h.repo.ListShops(c.Request.Context())
	if err != nil {
		response.ServerError(c, "查询门店失败")
		return
	}
	response.OK(c, list)
}

type shopReq struct {
	Name    string `json:"name"`
	Address string `json:"address"`
	Phone   string `json:"phone"`
	IsMain  *int8  `json:"is_main"`
	Status  *int8  `json:"status"`
}

func (h *StaffHandler) CreateShop(c *gin.Context) {
	ctx := c.Request.Context()
	var req shopReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		response.BadRequest(c, "请输入门店名称")
		return
	}
	if utf8.RuneCountInString(req.Name) > 128 {
		response.BadRequest(c, "门店名称过长")
		return
	}
	s := &model.Shop{
		TenantID: context.GetTenantID(ctx),
		Name:     req.Name,
		Address:  strings.TrimSpace(req.Address),
		Phone:    strings.TrimSpace(req.Phone),
		Status:   1,
	}
	if req.IsMain != nil && *req.IsMain == 1 {
		s.IsMain = 1
	} else {
		if shops, _ := h.repo.ListShops(ctx); len(shops) == 0 {
			s.IsMain = 1 // 首个门店自动设为主店
		}
	}
	if err := h.repo.CreateShop(ctx, s); err != nil {
		response.ServerError(c, "新增门店失败")
		return
	}
	if s.IsMain == 1 {
		h.repo.ClearMainShop(ctx, s.ID)
	}
	response.OK(c, s)
}

func (h *StaffHandler) UpdateShop(c *gin.Context) {
	ctx := c.Request.Context()
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	s, err := h.repo.GetShop(ctx, id)
	if err != nil {
		response.NotFound(c, "门店不存在")
		return
	}
	var req shopReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	if req.Name != "" {
		req.Name = strings.TrimSpace(req.Name)
		if utf8.RuneCountInString(req.Name) > 128 {
			response.BadRequest(c, "门店名称过长")
			return
		}
		s.Name = req.Name
	}
	s.Address = strings.TrimSpace(req.Address)
	s.Phone = strings.TrimSpace(req.Phone)
	if req.Status != nil && (*req.Status == 0 || *req.Status == 1) {
		s.Status = *req.Status
	}
	if req.IsMain != nil && *req.IsMain == 1 {
		s.IsMain = 1
	}
	if err := h.repo.UpdateShop(ctx, s); err != nil {
		response.ServerError(c, "更新门店失败")
		return
	}
	if s.IsMain == 1 {
		h.repo.ClearMainShop(ctx, s.ID)
	}
	response.OK(c, s)
}

func (h *StaffHandler) DeleteShop(c *gin.Context) {
	ctx := c.Request.Context()
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	s, err := h.repo.GetShop(ctx, id)
	if err != nil {
		response.NotFound(c, "门店不存在")
		return
	}
	if s.IsMain == 1 {
		response.BadRequest(c, "主店不可删除")
		return
	}
	if err := h.repo.DeleteShop(ctx, id); err != nil {
		response.ServerError(c, "删除失败")
		return
	}
	response.OKMsg(c, "删除成功")
}
