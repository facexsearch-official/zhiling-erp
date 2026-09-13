package handler

import (
	"pisa_server/internal/model"
	"pisa_server/internal/pkg/context"
	"pisa_server/internal/pkg/response"
	"pisa_server/internal/repository"
	"strconv"

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
	list, total := h.repo.List(ctx, page, pageSize, keyword)
	response.OKPage(c, list, total, page, pageSize)
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
	supplier.Status = 1
	supplier.TenantID = context.GetTenantID(ctx)
	if err := h.repo.Create(ctx, &supplier); err != nil {
		response.ServerError(c, "创建供应商失败")
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
	if err := h.repo.Update(ctx, existing); err != nil {
		response.ServerError(c, "更新供应商失败")
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
