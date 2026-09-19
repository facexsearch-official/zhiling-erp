package handler

import (
	"strconv"

	"pisa_server/internal/pkg/response"
	"pisa_server/internal/repository"

	"github.com/gin-gonic/gin"
)

type PriceHandler struct {
	repo *repository.PriceRepository
}

func NewPriceHandler(repo *repository.PriceRepository) *PriceHandler {
	return &PriceHandler{repo: repo}
}

// List 价格管理列表（按货品分页，展开多单位/多规格）
func (h *PriceHandler) List(c *gin.Context) {
	ctx := c.Request.Context()
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "50"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 500 {
		pageSize = 50
	}
	f := repository.PriceFilters{
		Keyword:      c.Query("keyword"),
		Spec:         c.Query("spec"),
		Brand:        c.Query("brand"),
		HideDisabled: c.Query("hide_disabled") == "1",
	}
	rows, total, cols := h.repo.List(ctx, f, page, pageSize)
	response.OK(c, gin.H{"list": rows, "total": total, "columns": cols, "page": page, "page_size": pageSize})
}

// Batch 批量改价
func (h *PriceHandler) Batch(c *gin.Context) {
	ctx := c.Request.Context()
	var body struct {
		Scope   string                   `json:"scope"`
		Keys    []string                 `json:"keys"`
		Filters repository.PriceFilters  `json:"filters"`
		Changes []repository.PriceChange `json:"changes"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	if body.Scope != "selected" {
		body.Scope = "all"
	}
	if len(body.Changes) == 0 {
		response.BadRequest(c, "请填写改价内容")
		return
	}
	n, err := h.repo.BatchAdjust(ctx, body.Scope, body.Keys, body.Filters, body.Changes)
	if err != nil {
		response.ServerError(c, "批量改价失败")
		return
	}
	response.OK(c, gin.H{"count": n})
}

// SaveRows 保存内联编辑的价格
func (h *PriceHandler) SaveRows(c *gin.Context) {
	ctx := c.Request.Context()
	var body struct {
		Rows []repository.PriceRowUpdate `json:"rows"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	if len(body.Rows) == 0 {
		response.BadRequest(c, "没有需要保存的数据")
		return
	}
	if err := h.repo.UpdateRows(ctx, body.Rows); err != nil {
		response.ServerError(c, "保存价格失败")
		return
	}
	response.OKMsg(c, "保存成功")
}
