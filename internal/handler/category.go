package handler

import (
	"pisa_server/internal/model"
	"pisa_server/internal/pkg/context"
	"pisa_server/internal/pkg/response"
	"pisa_server/internal/repository"

	"github.com/gin-gonic/gin"
)

type CategoryHandler struct {
	repo *repository.CategoryRepository
}

func NewCategoryHandler(repo *repository.CategoryRepository) *CategoryHandler {
	return &CategoryHandler{repo: repo}
}

func (h *CategoryHandler) ListAll(c *gin.Context) {
	list, err := h.repo.ListAll(c.Request.Context())
	if err != nil {
		response.ServerError(c, "查询分类失败")
		return
	}
	response.OK(c, list)
}

func (h *CategoryHandler) Create(c *gin.Context) {
	ctx := c.Request.Context()
	var cat model.GoodsCategory
	if err := c.ShouldBindJSON(&cat); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	if cat.Name == "" {
		response.BadRequest(c, "分类名称不能为空")
		return
	}
	cat.TenantID = context.GetTenantID(ctx)
	if err := h.repo.Create(ctx, &cat); err != nil {
		response.ServerError(c, "创建分类失败")
		return
	}
	response.OK(c, cat)
}
