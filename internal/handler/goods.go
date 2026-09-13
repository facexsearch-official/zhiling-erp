package handler

import (
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

func (h *GoodsHandler) GetByID(c *gin.Context) {
	ctx := c.Request.Context()
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	goods, err := h.repo.GetByID(ctx, id)
	if err != nil {
		response.NotFound(c, "货品不存在")
		return
	}
	response.OK(c, goods)
}

func (h *GoodsHandler) Create(c *gin.Context) {
	ctx := c.Request.Context()
	var goods model.Goods
	if err := c.ShouldBindJSON(&goods); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	if goods.Name == "" {
		response.BadRequest(c, "货品名称不能为空")
		return
	}
	goods.Status = 1
	goods.TenantID = context.GetTenantID(ctx)
	if err := h.repo.Create(ctx, &goods); err != nil {
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
	if err := c.ShouldBindJSON(existing); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	if err := h.repo.Update(ctx, existing); err != nil {
		response.ServerError(c, "更新货品失败")
		return
	}
	response.OK(c, existing)
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
