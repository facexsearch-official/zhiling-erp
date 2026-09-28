package handler

import (
	"pisa_server/internal/model"
	"pisa_server/internal/pkg/context"
	"pisa_server/internal/pkg/response"
	"pisa_server/internal/repository"
	"strconv"
	"unicode/utf8"

	"github.com/gin-gonic/gin"
)

type SupplierHandler struct {
	repo *repository.SupplierRepository
}

func NewSupplierHandler(repo *repository.SupplierRepository) *SupplierHandler {
	return &SupplierHandler{repo: repo}
}

func (h *SupplierHandler) List(c *gin.Context) {
	ctx := c.Request.Context()
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	keyword := c.Query("keyword")
	categoryID, _ := strconv.ParseInt(c.Query("category_id"), 10, 64)
	hideDisabled := c.Query("hide_disabled") == "1"
	hideZero := c.Query("hide_zero") == "1"
	list, total := h.repo.List(ctx, page, pageSize, keyword, categoryID, hideDisabled, hideZero)
	for i := range list {
		list[i].IDStr = strconv.FormatInt(list[i].ID, 10)
	}
	response.OKPage(c, list, total, page, pageSize)
}

type supplierBatchFilters struct {
	Keyword      string `json:"keyword"`
	CategoryID   int64  `json:"category_id"`
	HideDisabled bool   `json:"hide_disabled"`
	HideZero     bool   `json:"hide_zero"`
}

type supplierBatchReq struct {
	Action  string               `json:"action"` // update | delete
	Scope   string               `json:"scope"`  // selected | query
	IDs     []string             `json:"ids"`
	Filters supplierBatchFilters `json:"filters"`
	Patch   struct {
		Status     *int8   `json:"status"`
		CategoryID *string `json:"category_id"`
	} `json:"patch"`
}

// BatchUpdate 批量修改/删除供应商
func (h *SupplierHandler) BatchUpdate(c *gin.Context) {
	ctx := c.Request.Context()
	tenantID := context.GetTenantID(ctx)
	var req supplierBatchReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	q := h.repo.DB.Table("suppliers").Where("tenant_id = ?", tenantID)
	if req.Scope == "query" {
		f := req.Filters
		if f.Keyword != "" {
			kw := "%" + f.Keyword + "%"
			q = q.Where("(name LIKE ? OR code LIKE ?)", kw, kw)
		}
		if f.CategoryID == -1 {
			q = q.Where("category_id IS NULL OR category_id = 0")
		} else if f.CategoryID > 0 {
			q = q.Where("category_id = ?", f.CategoryID)
		}
		if f.HideDisabled {
			q = q.Where("status = 1")
		}
		if f.HideZero {
			q = q.Where("total_payable <> 0")
		}
	} else {
		ids := make([]int64, 0, len(req.IDs))
		for _, s := range req.IDs {
			if n, err := strconv.ParseInt(s, 10, 64); err == nil && n > 0 {
				ids = append(ids, n)
			}
		}
		if len(ids) == 0 {
			response.BadRequest(c, "请选择供应商")
			return
		}
		q = q.Where("id IN ?", ids)
	}

	if req.Action == "delete" {
		res := q.Delete(&model.Supplier{})
		if res.Error != nil {
			response.ServerError(c, "批量删除失败")
			return
		}
		response.OK(c, gin.H{"count": res.RowsAffected})
		return
	}

	upd := map[string]interface{}{}
	if req.Patch.Status != nil {
		upd["status"] = *req.Patch.Status
	}
	if req.Patch.CategoryID != nil {
		if *req.Patch.CategoryID == "" || *req.Patch.CategoryID == "0" {
			upd["category_id"] = nil
		} else if n, err := strconv.ParseInt(*req.Patch.CategoryID, 10, 64); err == nil {
			upd["category_id"] = n
		}
	}
	if len(upd) == 0 {
		response.BadRequest(c, "没有需要修改的内容")
		return
	}
	res := q.Updates(upd)
	if res.Error != nil {
		response.ServerError(c, "批量修改失败")
		return
	}
	response.OK(c, gin.H{"count": res.RowsAffected})
}

func (h *SupplierHandler) ListAll(c *gin.Context) {
	ctx := c.Request.Context()
	list, err := h.repo.ListAll(ctx)
	if err != nil {
		response.ServerError(c, "查询供应商失败")
		return
	}
	response.OK(c, list)
}

func (h *SupplierHandler) GetByID(c *gin.Context) {
	ctx := c.Request.Context()
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	supplier, err := h.repo.GetByID(ctx, id)
	if err != nil {
		response.NotFound(c, "供应商不存在")
		return
	}
	if addrs, err := h.repo.Addresses(ctx, id); err == nil {
		supplier.Addresses = addrs
	}
	response.OK(c, supplier)
}

func (h *SupplierHandler) Create(c *gin.Context) {
	ctx := c.Request.Context()
	var supplier model.Supplier
	if err := c.ShouldBindJSON(&supplier); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	if supplier.Name == "" {
		response.BadRequest(c, "供应商名称不能为空")
		return
	}
	if utf8.RuneCountInString(supplier.Name) > 220 {
		response.BadRequest(c, "供应商名称不能超过220个字符")
		return
	}
	supplier.Status = 1
	supplier.TenantID = context.GetTenantID(ctx)
	if err := h.repo.Create(ctx, &supplier); err != nil {
		response.ServerError(c, "创建供应商失败")
		return
	}
	if err := h.repo.ReplaceAddresses(ctx, supplier.ID, supplier.Addresses); err != nil {
		response.ServerError(c, "保存收货地址失败")
		return
	}
	response.OK(c, supplier)
}

func (h *SupplierHandler) Update(c *gin.Context) {
	ctx := c.Request.Context()
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	existing, err := h.repo.GetByID(ctx, id)
	if err != nil {
		response.NotFound(c, "供应商不存在")
		return
	}
	if err := c.ShouldBindJSON(existing); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	if existing.Name == "" {
		response.BadRequest(c, "供应商名称不能为空")
		return
	}
	if utf8.RuneCountInString(existing.Name) > 220 {
		response.BadRequest(c, "供应商名称不能超过220个字符")
		return
	}
	if err := h.repo.Update(ctx, existing); err != nil {
		response.ServerError(c, "更新供应商失败")
		return
	}
	if err := h.repo.ReplaceAddresses(ctx, existing.ID, existing.Addresses); err != nil {
		response.ServerError(c, "保存收货地址失败")
		return
	}
	response.OK(c, existing)
}

func (h *SupplierHandler) Delete(c *gin.Context) {
	ctx := c.Request.Context()
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	if err := h.repo.Delete(ctx, id); err != nil {
		response.ServerError(c, "删除供应商失败")
		return
	}
	response.OKMsg(c, "删除成功")
}
