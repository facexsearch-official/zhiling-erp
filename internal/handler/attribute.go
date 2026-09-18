package handler

import (
	"encoding/json"
	"strconv"
	"strings"

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

// ListAll 返回当前商户所有规格（辅助属性）
func (h *AttributeHandler) ListAll(c *gin.Context) {
	list, err := h.repo.ListAll(c.Request.Context())
	if err != nil {
		response.ServerError(c, "查询规格失败")
		return
	}
	response.OK(c, list)
}

// Create 新增规格（可同时带入初始规格值）
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

// AddValue 为指定规格追加一个规格值
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
