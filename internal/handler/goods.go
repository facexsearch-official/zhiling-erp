package handler

import (
	"strings"

	"pisa_server/internal/model"
	"pisa_server/internal/pkg/context"
	"pisa_server/internal/pkg/response"
	"pisa_server/internal/repository"
	"strconv"

	"github.com/gin-gonic/gin"
)

type GoodsHandler struct {
	repo *repository.GoodsRepository
}

func NewGoodsHandler(repo *repository.GoodsRepository) *GoodsHandler {
	return &GoodsHandler{repo: repo}
}

func (h *GoodsHandler) List(c *gin.Context) {
	ctx := c.Request.Context()
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	keyword := c.Query("keyword")
	categoryID, _ := strconv.ParseInt(c.Query("category_id"), 10, 64)
	list, total := h.repo.List(ctx, page, pageSize, keyword, categoryID)
	response.OKPage(c, list, total, page, pageSize)
}

func (h *GoodsHandler) ListAll(c *gin.Context) {
	ctx := c.Request.Context()
	list, err := h.repo.ListAll(ctx)
	if err != nil {
		response.ServerError(c, "查询货品失败")
		return
	}
	response.OK(c, list)
}

// GetByID 返回货品详情（含多单位 / 多规格）
func (h *GoodsHandler) GetByID(c *gin.Context) {
	ctx := c.Request.Context()
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	goods, err := h.repo.GetByIDWithChildren(ctx, id)
	if err != nil {
		response.NotFound(c, "货品不存在")
		return
	}
	response.OK(c, goods)
}

// NextCode 生成下一个货品编号
func (h *GoodsHandler) NextCode(c *gin.Context) {
	response.OK(c, gin.H{"code": h.repo.NextCode(c.Request.Context())})
}

// Brands 返回当前商户所有不重复的品牌
func (h *GoodsHandler) Brands(c *gin.Context) {
	ctx := c.Request.Context()
	tenantID := context.GetTenantID(ctx)
	var list []string
	h.repo.DB.Model(&model.Goods{}).
		Where("tenant_id = ? AND brand != '' AND brand IS NOT NULL", tenantID).
		Distinct("brand").Pluck("brand", &list)
	response.OK(c, list)
}

// Origins 返回当前商户所有不重复的产地
func (h *GoodsHandler) Origins(c *gin.Context) {
	ctx := c.Request.Context()
	tenantID := context.GetTenantID(ctx)
	var list []string
	h.repo.DB.Model(&model.Goods{}).
		Where("tenant_id = ? AND origin != '' AND origin IS NOT NULL", tenantID).
		Distinct("origin").Pluck("origin", &list)
	response.OK(c, list)
}

func (h *GoodsHandler) Create(c *gin.Context) {
	ctx := c.Request.Context()
	var goods model.Goods
	if err := c.ShouldBindJSON(&goods); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	if strings.TrimSpace(goods.Name) == "" {
		response.BadRequest(c, "货品名称不能为空")
		return
	}
	goods.ID = 0
	goods.Status = 1
	goods.TenantID = context.GetTenantID(ctx)
	if err := h.repo.CreateWithChildren(ctx, &goods, goods.Units, goods.Specs); err != nil {
		response.ServerError(c, "创建货品失败")
		return
	}
	response.OK(c, goods)
}

func (h *GoodsHandler) Update(c *gin.Context) {
	ctx := c.Request.Context()
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	existing, err := h.repo.GetByID(ctx, id)
	if err != nil {
		response.NotFound(c, "货品不存在")
		return
	}
	var payload model.Goods
	if err := c.ShouldBindJSON(&payload); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	if strings.TrimSpace(payload.Name) == "" {
		response.BadRequest(c, "货品名称不能为空")
		return
	}
	// 不可变字段以库中为准
	payload.ID = existing.ID
	payload.TenantID = existing.TenantID
	payload.CreatedAt = existing.CreatedAt
	if err := h.repo.UpdateWithChildren(ctx, &payload, payload.Units, payload.Specs); err != nil {
		response.ServerError(c, "更新货品失败")
		return
	}
	response.OK(c, payload)
}

func (h *GoodsHandler) Delete(c *gin.Context) {
	ctx := c.Request.Context()
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	if err := h.repo.Delete(ctx, id); err != nil {
		response.ServerError(c, "删除货品失败")
		return
	}
	response.OKMsg(c, "删除成功")
}
