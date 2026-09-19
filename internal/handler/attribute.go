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

type AttributeHandler struct {
	repo *repository.AttributeRepository
}

func NewAttributeHandler(repo *repository.AttributeRepository) *AttributeHandler {
	return &AttributeHandler{repo: repo}
}

func (h *AttributeHandler) ListAll(c *gin.Context) {
	list, err := h.repo.ListAll(c.Request.Context())
	if err != nil {
		response.ServerError(c, "查询规格失败")
		return
	}
	response.OK(c, list)
}

func (h *AttributeHandler) List(c *gin.Context) {
	name := c.Query("name")
	content := c.Query("content")
	list, err := h.repo.List(c.Request.Context(), name, content)
	if err != nil {
		response.ServerError(c, "查询规格失败")
		return
	}
	response.OK(c, list)
}

func (h *AttributeHandler) Create(c *gin.Context) {
	ctx := c.Request.Context()
	var body struct {
		Name   string   `json:"name"`
		Values []string `json:"values"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	body.Name = strings.TrimSpace(body.Name)
	if body.Name == "" {
		response.BadRequest(c, "规格名称不能为空")
		return
	}
	if utf8.RuneCountInString(body.Name) > 64 {
		response.BadRequest(c, "规格名称不能超过64个字符")
		return
	}
	values := body.Values
	if values == nil {
		values = []string{}
	}
	raw, _ := json.Marshal(values)
	a := model.GoodsAttribute{
		TenantID: context.GetTenantID(ctx),
		Name:     body.Name,
		Values:   string(raw),
		Status:   1,
	}
	if err := h.repo.Create(ctx, &a); err != nil {
		response.ServerError(c, "创建规格失败")
		return
	}
	response.OK(c, a)
}

func (h *AttributeHandler) Update(c *gin.Context) {
	ctx := c.Request.Context()
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	a, err := h.repo.GetByID(ctx, id)
	if err != nil {
		response.NotFound(c, "规格不存在")
		return
	}
	var body struct {
		Name   string   `json:"name"`
		Values []string `json:"values"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	body.Name = strings.TrimSpace(body.Name)
	if body.Name == "" {
		response.BadRequest(c, "规格名称不能为空")
		return
	}
	if utf8.RuneCountInString(body.Name) > 64 {
		response.BadRequest(c, "规格名称不能超过64个字符")
		return
	}
	a.Name = body.Name
	if body.Values != nil {
		raw, _ := json.Marshal(body.Values)
		a.Values = string(raw)
	}
	if err := h.repo.Save(ctx, a); err != nil {
		response.ServerError(c, "更新规格失败")
		return
	}
	response.OK(c, a)
}

func (h *AttributeHandler) Delete(c *gin.Context) {
	ctx := c.Request.Context()
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	a, err := h.repo.GetByID(ctx, id)
	if err != nil {
		response.NotFound(c, "规格不存在")
		return
	}
	if h.repo.HasGoods(ctx, a.Name) {
		response.BadRequest(c, "该规格已被货品使用，无法删除")
		return
	}
	if err := h.repo.Delete(ctx, id); err != nil {
		response.ServerError(c, "删除规格失败")
		return
	}
	response.OKMsg(c, "删除成功")
}

func (h *AttributeHandler) AddValue(c *gin.Context) {
	ctx := c.Request.Context()
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var body struct {
		Value string `json:"value"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	body.Value = strings.TrimSpace(body.Value)
	if body.Value == "" {
		response.BadRequest(c, "规格值不能为空")
		return
	}
	a, err := h.repo.GetByID(ctx, id)
	if err != nil {
		response.NotFound(c, "规格不存在")
		return
	}
	var values []string
	if a.Values != "" {
		_ = json.Unmarshal([]byte(a.Values), &values)
	}
	for _, v := range values {
		if v == body.Value {
			response.OK(c, a)
			return
		}
	}
	values = append(values, body.Value)
	raw, _ := json.Marshal(values)
	a.Values = string(raw)
	if err := h.repo.Save(ctx, a); err != nil {
		response.ServerError(c, "保存规格值失败")
		return
	}
	response.OK(c, a)
}
