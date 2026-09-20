package handler

import (
	"strconv"
	"strings"
	"unicode/utf8"

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
	cat.Name = strings.TrimSpace(cat.Name)
	if cat.Name == "" {
		response.BadRequest(c, "分类名称不能为空")
		return
	}
	if utf8.RuneCountInString(cat.Name) > 30 {
		response.BadRequest(c, "类别名称不能超过30个字符")
		return
	}
	if cat.ParentID != 0 {
		if _, err := h.repo.GetByID(ctx, cat.ParentID); err != nil {
			response.BadRequest(c, "上级类别不存在")
			return
		}
		parentDepth, err := h.repo.GetDepth(ctx, cat.ParentID)
		if err != nil {
			response.ServerError(c, "查询分类层级失败")
			return
		}
		if parentDepth >= 5 {
			response.BadRequest(c, "最多支持5级分类")
			return
		}
	}
	cat.ID = 0
	cat.TenantID = context.GetTenantID(ctx)
	if err := h.repo.Create(ctx, &cat); err != nil {
		response.ServerError(c, "创建分类失败")
		return
	}
	response.OK(c, cat)
}

func (h *CategoryHandler) Update(c *gin.Context) {
	ctx := c.Request.Context()
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	existing, err := h.repo.GetByID(ctx, id)
	if err != nil {
		response.NotFound(c, "分类不存在")
		return
	}
	var body struct {
		Name     string `json:"name"`
		ParentID int64  `json:"parent_id"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	body.Name = strings.TrimSpace(body.Name)
	if body.Name == "" {
		response.BadRequest(c, "分类名称不能为空")
		return
	}
	if utf8.RuneCountInString(body.Name) > 30 {
		response.BadRequest(c, "类别名称不能超过30个字符")
		return
	}
	if body.ParentID == id {
		response.BadRequest(c, "上级类别不能是自身")
		return
	}
	if body.ParentID != 0 {
		if _, err := h.repo.GetByID(ctx, body.ParentID); err != nil {
			response.BadRequest(c, "上级类别不存在")
			return
		}
		if body.ParentID != existing.ParentID {
			newParentDepth, err := h.repo.GetDepth(ctx, body.ParentID)
			if err != nil {
				response.ServerError(c, "查询分类层级失败")
				return
			}
			if newParentDepth >= 5 {
				response.BadRequest(c, "最多支持5级分类")
				return
			}
		}
	}
	existing.Name = body.Name
	existing.ParentID = body.ParentID
	if err := h.repo.Update(ctx, existing); err != nil {
		response.ServerError(c, "更新分类失败")
		return
	}
	response.OK(c, existing)
}

func (h *CategoryHandler) Delete(c *gin.Context) {
	ctx := c.Request.Context()
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	if _, err := h.repo.GetByID(ctx, id); err != nil {
		response.NotFound(c, "分类不存在")
		return
	}
	hasChildren, err := h.repo.HasChildren(ctx, id)
	if err != nil {
		response.ServerError(c, "查询分类失败")
		return
	}
	if hasChildren {
		response.BadRequest(c, "该分类下有子分类，无法删除")
		return
	}
	hasGoods, err := h.repo.HasGoods(ctx, id)
	if err != nil {
		response.ServerError(c, "查询分类失败")
		return
	}
	if hasGoods {
		response.BadRequest(c, "该分类下有商品，无法删除")
		return
	}
	if err := h.repo.Delete(ctx, id); err != nil {
		response.ServerError(c, "删除分类失败")
		return
	}
	response.OKMsg(c, "删除成功")
}
