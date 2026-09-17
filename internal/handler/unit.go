package handler

import (
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
	u.TenantID = context.GetTenantID(ctx)
	if err := h.repo.Create(ctx, &u); err != nil {
		response.ServerError(c, "创建单位失败")
		return
	}
	response.OK(c, u)
}
