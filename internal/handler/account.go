package handler

import (
	"pisa_server/internal/model"
	"pisa_server/internal/pkg/context"
	"pisa_server/internal/pkg/response"
	"pisa_server/internal/repository"
	"strconv"

	"github.com/gin-gonic/gin"
)

type AccountHandler struct {
	repo *repository.AccountRepository
}

func NewAccountHandler(repo *repository.AccountRepository) *AccountHandler {
	return &AccountHandler{repo: repo}
}

func (h *AccountHandler) List(c *gin.Context) {
	ctx := c.Request.Context()
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	keyword := c.Query("keyword")
	list, total := h.repo.List(ctx, page, pageSize, keyword)
	response.OKPage(c, list, total, page, pageSize)
}

func (h *AccountHandler) ListAll(c *gin.Context) {
	ctx := c.Request.Context()
	list, err := h.repo.ListAll(ctx)
	if err != nil {
		response.ServerError(c, "查询账户失败")
		return
	}
	response.OK(c, list)
}

func (h *AccountHandler) GetByID(c *gin.Context) {
	ctx := c.Request.Context()
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	account, err := h.repo.GetByID(ctx, id)
	if err != nil {
		response.NotFound(c, "账户不存在")
		return
	}
	response.OK(c, account)
}

func (h *AccountHandler) Create(c *gin.Context) {
	ctx := c.Request.Context()
	var account model.Account
	if err := c.ShouldBindJSON(&account); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	if account.Name == "" {
		response.BadRequest(c, "账户名称不能为空")
		return
	}
	account.Status = 1
	account.TenantID = context.GetTenantID(ctx)
	if err := h.repo.Create(ctx, &account); err != nil {
		response.ServerError(c, "创建账户失败")
		return
	}
	response.OK(c, account)
}

func (h *AccountHandler) Update(c *gin.Context) {
	ctx := c.Request.Context()
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	existing, err := h.repo.GetByID(ctx, id)
	if err != nil {
		response.NotFound(c, "账户不存在")
		return
	}
	if err := c.ShouldBindJSON(existing); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	if err := h.repo.Update(ctx, existing); err != nil {
		response.ServerError(c, "更新账户失败")
		return
	}
	response.OK(c, existing)
}

func (h *AccountHandler) Delete(c *gin.Context) {
	ctx := c.Request.Context()
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	if err := h.repo.Delete(ctx, id); err != nil {
		response.ServerError(c, "删除账户失败")
		return
	}
	response.OKMsg(c, "删除成功")
}
