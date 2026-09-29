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

type UnitHandler struct {
	repo *repository.UnitRepository
}

func NewUnitHandler(repo *repository.UnitRepository) *UnitHandler {
	return &UnitHandler{repo: repo}
}

func (h *UnitHandler) ListAll(c *gin.Context) {
	list, err := h.repo.ListAll(c.Request.Context())
	if err != nil {
		response.ServerError(c, "查询单位失败")
		return
	}
	for i := range list {
		list[i].IDStr = strconv.FormatInt(list[i].ID, 10)
	}
	response.OK(c, list)
}

func (h *UnitHandler) Create(c *gin.Context) {
	ctx := c.Request.Context()
	var u model.Unit
	if err := c.ShouldBindJSON(&u); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	if u.Name == "" {
		response.BadRequest(c, "单位名称不能为空")
		return
	}
	u.Name = strings.TrimSpace(u.Name)
	if u.Name == "" {
		response.BadRequest(c, "单位名称不能为空")
		return
	}
	if utf8.RuneCountInString(u.Name) > 20 {
		response.BadRequest(c, "单位名称不能超过20个字符")
		return
	}
	u.ID = 0
	u.TenantID = context.GetTenantID(ctx)
	if err := h.repo.Create(ctx, &u); err != nil {
		response.ServerError(c, "创建单位失败")
		return
	}
	response.OK(c, u)
}

func (h *UnitHandler) Update(c *gin.Context) {
	ctx := c.Request.Context()
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	u, err := h.repo.GetByID(ctx, id)
	if err != nil {
		response.NotFound(c, "单位不存在")
		return
	}
	var body struct {
		Name string `json:"name"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	body.Name = strings.TrimSpace(body.Name)
	if body.Name == "" {
		response.BadRequest(c, "单位名称不能为空")
		return
	}
	if utf8.RuneCountInString(body.Name) > 20 {
		response.BadRequest(c, "单位名称不能超过20个字符")
		return
	}
	u.Name = body.Name
	if err := h.repo.Save(ctx, u); err != nil {
		response.ServerError(c, "更新单位失败")
		return
	}
	response.OK(c, u)
}

func (h *UnitHandler) Delete(c *gin.Context) {
	ctx := c.Request.Context()
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	if _, err := h.repo.GetByID(ctx, id); err != nil {
		response.NotFound(c, "单位不存在")
		return
	}
	if h.repo.HasGoods(ctx, id) {
		response.BadRequest(c, "该单位已被货品使用，无法删除")
		return
	}
	if err := h.repo.Delete(ctx, id); err != nil {
		response.ServerError(c, "删除单位失败")
		return
	}
	response.OKMsg(c, "删除成功")
}
