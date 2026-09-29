package handler

import (
	"pisa_server/internal/model"
	"pisa_server/internal/pkg/context"
	"pisa_server/internal/pkg/response"
	"pisa_server/internal/pkg/snowflake"
	"pisa_server/internal/repository"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type CustomerHandler struct {
	repo *repository.CustomerRepository
	db   *gorm.DB
}

func NewCustomerHandler(repo *repository.CustomerRepository, db *gorm.DB) *CustomerHandler {
	return &CustomerHandler{repo: repo, db: db}
}

func fillCustomerCatStr(list []model.Customer) {
	for i := range list {
		list[i].IDStr = strconv.FormatInt(list[i].ID, 10)
		if list[i].CategoryID != nil {
			list[i].CategoryIDStr = strconv.FormatInt(int64(*list[i].CategoryID), 10)
		}
	}
}

// saveAddresses 覆盖保存某客户的收货地址（跳过空行）
func (h *CustomerHandler) saveAddresses(tenantID, customerID int64, list []model.CustomerAddress) {
	h.db.Where("customer_id = ?", customerID).Delete(&model.CustomerAddress{})
	for i := range list {
		a := list[i]
		if strings.TrimSpace(a.Receiver) == "" && strings.TrimSpace(a.Phone) == "" &&
			strings.TrimSpace(a.Region) == "" && strings.TrimSpace(a.Detail) == "" {
			continue
		}
		a.ID = snowflake.GenID()
		a.TenantID = tenantID
		a.CustomerID = customerID
		h.db.Create(&a)
	}
}

func (h *CustomerHandler) List(c *gin.Context) {
	ctx := c.Request.Context()
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	keyword := c.Query("keyword")
	categoryID, _ := strconv.ParseInt(c.DefaultQuery("category_id", "0"), 10, 64)
	hideDisabled := c.Query("hide_disabled") == "1"
	hideZero := c.Query("hide_zero") == "1"
	list, total := h.repo.List(ctx, page, pageSize, keyword, categoryID, hideDisabled, hideZero)
	fillCustomerCatStr(list)
	response.OKPage(c, list, total, page, pageSize)
}

func (h *CustomerHandler) ListAll(c *gin.Context) {
	ctx := c.Request.Context()
	list, err := h.repo.ListAll(ctx)
	if err != nil {
		response.ServerError(c, "查询客户失败")
		return
	}
	fillCustomerCatStr(list)
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
	h.db.Where("customer_id = ?", id).Order("is_default DESC, id ASC").Find(&customer.Addresses)
	customer.IDStr = strconv.FormatInt(customer.ID, 10)
	if customer.CategoryID != nil {
		customer.CategoryIDStr = strconv.FormatInt(int64(*customer.CategoryID), 10)
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
	h.saveAddresses(customer.TenantID, customer.ID, customer.Addresses)
	customer.IDStr = strconv.FormatInt(customer.ID, 10)
	if customer.CategoryID != nil {
		customer.CategoryIDStr = strconv.FormatInt(int64(*customer.CategoryID), 10)
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
	h.saveAddresses(existing.TenantID, existing.ID, existing.Addresses)
	existing.IDStr = strconv.FormatInt(existing.ID, 10)
	if existing.CategoryID != nil {
		existing.CategoryIDStr = strconv.FormatInt(int64(*existing.CategoryID), 10)
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
