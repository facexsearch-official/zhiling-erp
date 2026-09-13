package handler

import (
	"pisa_server/internal/model"
	"pisa_server/internal/pkg/context"
	"pisa_server/internal/pkg/response"
	"pisa_server/internal/repository"
	"strconv"

	"github.com/gin-gonic/gin"
)

type CustomerHandler struct {
	repo *repository.CustomerRepository
}

func NewCustomerHandler(repo *repository.CustomerRepository) *CustomerHandler {
	return &CustomerHandler{repo: repo}
}

func (h *CustomerHandler) List(c *gin.Context) {
	ctx := c.Request.Context()
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	keyword := c.Query("keyword")
	list, total := h.repo.List(ctx, page, pageSize, keyword)
	response.OKPage(c, list, total, page, pageSize)
}

func (h *CustomerHandler) ListAll(c *gin.Context) {
	ctx := c.Request.Context()
	list, err := h.repo.ListAll(ctx)
	if err != nil {
		response.ServerError(c, "查询客户失败")
		return
	}
	response.OK(c, list)
}

func (h *CustomerHandler) GetByID(c *gin.Context) {
	ctx := c.Request.Context()
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	customer, err := h.repo.GetByID(ctx, id)
	if err != nil {
		response.NotFound(c, "客户不存在")
		return
	}
	response.OK(c, customer)
}

func (h *CustomerHandler) Create(c *gin.Context) {
	ctx := c.Request.Context()
	var customer model.Customer
	if err := c.ShouldBindJSON(&customer); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	if customer.Name == "" {
		response.BadRequest(c, "客户名称不能为空")
		return
	}
	customer.Status = 1
	customer.TenantID = context.GetTenantID(ctx)
	if err := h.repo.Create(ctx, &customer); err != nil {
		response.ServerError(c, "创建客户失败")
		return
	}
	response.OK(c, customer)
}

func (h *CustomerHandler) Update(c *gin.Context) {
	ctx := c.Request.Context()
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	existing, err := h.repo.GetByID(ctx, id)
	if err != nil {
		response.NotFound(c, "客户不存在")
		return
	}
	if err := c.ShouldBindJSON(existing); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	if err := h.repo.Update(ctx, existing); err != nil {
		response.ServerError(c, "更新客户失败")
		return
	}
	response.OK(c, existing)
}

func (h *CustomerHandler) Delete(c *gin.Context) {
	ctx := c.Request.Context()
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	if err := h.repo.Delete(ctx, id); err != nil {
		response.ServerError(c, "删除客户失败")
		return
	}
	response.OKMsg(c, "删除成功")
}
