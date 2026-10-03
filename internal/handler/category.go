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
	// 附带字符串 ID，供前端级联选择器精确匹配
	out := make([]gin.H, 0, len(list))
	for _, x := range list {
		out = append(out, gin.H{
			"id":            x.ID,
			"id_str":        strconv.FormatInt(x.ID, 10),
			"parent_id":     x.ParentID,
			"parent_id_str": strconv.FormatInt(x.ParentID, 10),
			"name":          x.Name,
			"sort":          x.Sort,
		})
	}
	response.OK(c, out)
}

func (h *CategoryHandler) Create(c *gin.Context) {
	ctx := c.Request.Context()
	var req struct {
		Name     string          `json:"name"`
		ParentID model.FlexInt64 `json:"parent_id"`
		Sort     int             `json:"sort"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	name := strings.TrimSpace(req.Name)
	if name == "" {
		response.BadRequest(c, "分类名称不能为空")
		return
	}
	if utf8.RuneCountInString(name) > 30 {
		response.BadRequest(c, "类别名称不能超过30个字符")
		return
	}
	parentID := int64(req.ParentID)
	if parentID != 0 {
		if _, err := h.repo.GetByID(ctx, parentID); err != nil {
			response.BadRequest(c, "上级类别不存在")
			return
		}
		parentDepth, err := h.repo.GetDepth(ctx, parentID)
		if err != nil {
			response.ServerError(c, "查询分类层级失败")
			return
		}
		if parentDepth >= 5 {
			response.BadRequest(c, "最多支持5级分类")
			return
		}
	}
	if exists, err := h.repo.ExistsByName(ctx, name, parentID, 0); err != nil {
		response.ServerError(c, "查询分类失败")
		return
	} else if exists {
		response.BadRequest(c, "同级分类名称已存在")
		return
	}
	cat := model.GoodsCategory{
		TenantID: context.GetTenantID(ctx),
		Name:     name,
		ParentID: parentID,
		Sort:     req.Sort,
	}
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
		Name     string          `json:"name"`
		ParentID model.FlexInt64 `json:"parent_id"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	name := strings.TrimSpace(body.Name)
	if name == "" {
		response.BadRequest(c, "分类名称不能为空")
		return
	}
	if utf8.RuneCountInString(name) > 30 {
		response.BadRequest(c, "类别名称不能超过30个字符")
		return
	}
	parentID := int64(body.ParentID)
	if parentID == id {
		response.BadRequest(c, "上级类别不能是自身")
		return
	}
	if parentID != 0 {
		if _, err := h.repo.GetByID(ctx, parentID); err != nil {
			response.BadRequest(c, "上级类别不存在")
			return
		}
		if parentID != existing.ParentID {
			newParentDepth, err := h.repo.GetDepth(ctx, parentID)
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
	if exists, err := h.repo.ExistsByName(ctx, name, parentID, id); err != nil {
		response.ServerError(c, "查询分类失败")
		return
	} else if exists {
		response.BadRequest(c, "同级分类名称已存在")
		return
	}
	existing.Name = name
	existing.ParentID = parentID
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
