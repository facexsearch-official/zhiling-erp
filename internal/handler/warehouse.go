package handler

import (
	"pisa_server/internal/model"
	"pisa_server/internal/pkg/context"
	"pisa_server/internal/pkg/response"
	"pisa_server/internal/repository"
	"strconv"

	"github.com/gin-gonic/gin"
)

type WarehouseHandler struct {
	repo *repository.WarehouseRepository
}

func NewWarehouseHandler(repo *repository.WarehouseRepository) *WarehouseHandler {
	return &WarehouseHandler{repo: repo}
}

func (h *WarehouseHandler) List(c *gin.Context) {
	ctx := c.Request.Context()
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	keyword := c.Query("keyword")
	list, total := h.repo.List(ctx, page, pageSize, keyword)
	response.OKPage(c, list, total, page, pageSize)
}

func (h *WarehouseHandler) ListAll(c *gin.Context) {
	ctx := c.Request.Context()
	list, err := h.repo.ListAll(ctx)
	if err != nil {
		response.ServerError(c, "查询仓库失败")
		return
	}
	response.OK(c, list)
}

func (h *WarehouseHandler) GetByID(c *gin.Context) {
	ctx := c.Request.Context()
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	warehouse, err := h.repo.GetByID(ctx, id)
	if err != nil {
		response.NotFound(c, "仓库不存在")
		return
	}
	response.OK(c, warehouse)
}

func (h *WarehouseHandler) Create(c *gin.Context) {
	ctx := c.Request.Context()
	var warehouse model.Warehouse
	if err := c.ShouldBindJSON(&warehouse); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	if warehouse.Name == "" {
		response.BadRequest(c, "仓库名称不能为空")
		return
	}
	warehouse.Status = 1
	warehouse.TenantID = context.GetTenantID(ctx)
	if err := h.repo.Create(ctx, &warehouse); err != nil {
		response.ServerError(c, "创建仓库失败")
		return
	}
	response.OK(c, warehouse)
}

func (h *WarehouseHandler) Update(c *gin.Context) {
	ctx := c.Request.Context()
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	existing, err := h.repo.GetByID(ctx, id)
	if err != nil {
		response.NotFound(c, "仓库不存在")
		return
	}
	if err := c.ShouldBindJSON(existing); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	if err := h.repo.Update(ctx, existing); err != nil {
		response.ServerError(c, "更新仓库失败")
		return
	}
	response.OK(c, existing)
}

func (h *WarehouseHandler) Delete(c *gin.Context) {
	ctx := c.Request.Context()
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	if err := h.repo.Delete(ctx, id); err != nil {
		response.ServerError(c, "删除仓库失败")
		return
	}
	response.OKMsg(c, "删除成功")
}
