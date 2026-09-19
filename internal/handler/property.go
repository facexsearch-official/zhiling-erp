package handler

import (
	"encoding/json"
	"strconv"
	"strings"
	"unicode/utf8"

	"pisa_server/internal/model"
	"pisa_server/internal/pkg/context"
	"pisa_server/internal/pkg/response"
	"pisa_server/internal/repository"

	"github.com/gin-gonic/gin"
)

type PropertyHandler struct {
	repo *repository.PropertyRepository
}

func NewPropertyHandler(repo *repository.PropertyRepository) *PropertyHandler {
	return &PropertyHandler{repo: repo}
}

func (h *PropertyHandler) ListAll(c *gin.Context) {
	list, err := h.repo.ListAll(c.Request.Context())
	if err != nil {
		response.ServerError(c, "查询货品属性失败")
		return
	}
	response.OK(c, list)
}

func (h *PropertyHandler) Create(c *gin.Context) {
	ctx := c.Request.Context()
	var body struct {
		Name   string   `json:"name"`
		Type   int8     `json:"type"`
		Values []string `json:"values"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	body.Name = strings.TrimSpace(body.Name)
	if body.Name == "" {
		response.BadRequest(c, "属性名称不能为空")
		return
	}
	if utf8.RuneCountInString(body.Name) > 20 {
		response.BadRequest(c, "属性名称不能超过20个字符")
		return
	}
	if body.Type != 1 && body.Type != 2 {
		body.Type = 1
	}
	n, err := h.repo.Count(ctx)
	if err != nil {
		response.ServerError(c, "查询货品属性失败")
		return
	}
	if n >= 6 {
		response.BadRequest(c, "最多可添加6项属性")
		return
	}
	values := body.Values
	if values == nil {
		values = []string{}
	}
	raw, _ := json.Marshal(values)
	p := model.GoodsProperty{
		TenantID: context.GetTenantID(ctx),
		Name:     body.Name,
		Type:     body.Type,
		Values:   string(raw),
		Status:   1,
	}
	if err := h.repo.Create(ctx, &p); err != nil {
		response.ServerError(c, "创建货品属性失败")
		return
	}
	response.OK(c, p)
}

func (h *PropertyHandler) Update(c *gin.Context) {
	ctx := c.Request.Context()
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	p, err := h.repo.GetByID(ctx, id)
	if err != nil {
		response.NotFound(c, "货品属性不存在")
		return
	}
	var body struct {
		Name   string   `json:"name"`
		Type   int8     `json:"type"`
		Values []string `json:"values"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	body.Name = strings.TrimSpace(body.Name)
	if body.Name == "" {
		response.BadRequest(c, "属性名称不能为空")
		return
	}
	if utf8.RuneCountInString(body.Name) > 20 {
		response.BadRequest(c, "属性名称不能超过20个字符")
		return
	}
	p.Name = body.Name
	if body.Type == 1 || body.Type == 2 {
		p.Type = body.Type
	}
	if body.Values != nil {
		raw, _ := json.Marshal(body.Values)
		p.Values = string(raw)
	}
	if err := h.repo.Save(ctx, p); err != nil {
		response.ServerError(c, "更新货品属性失败")
		return
	}
	response.OK(c, p)
}

func (h *PropertyHandler) Delete(c *gin.Context) {
	ctx := c.Request.Context()
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	if err := h.repo.Delete(ctx, id); err != nil {
		response.ServerError(c, "删除货品属性失败")
		return
	}
	response.OKMsg(c, "删除成功")
}
