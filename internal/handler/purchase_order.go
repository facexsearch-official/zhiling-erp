package handler

import (
	"fmt"
	"pisa_server/internal/model"
	"pisa_server/internal/pkg/context"
	"pisa_server/internal/pkg/response"
	"pisa_server/internal/pkg/snowflake"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type PurchaseOrderHandler struct {
	db *gorm.DB
}

func NewPurchaseOrderHandler(dbConn *gorm.DB) *PurchaseOrderHandler {
	return &PurchaseOrderHandler{db: dbConn}
}

type orderCreateReq struct {
	ShopID      int64   `json:"shop_id"`
	SupplierID  int64   `json:"supplier_id"`
	OrderDate   string  `json:"order_date"`
	TotalAmount float64 `json:"total_amount"`
	Remark      string  `json:"remark"`
}

func (h *PurchaseOrderHandler) List(c *gin.Context) {
	ctx := c.Request.Context()
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	keyword := c.Query("keyword")

	tenantID := context.GetTenantID(ctx)
	var total int64
	var list []model.PurchaseOrder
	q := h.db.Table("purchase_orders").Where("tenant_id = ?", tenantID)
	if keyword != "" {
		q = q.Where("order_no LIKE ?", "%"+keyword+"%")
	}
	q.Model(&model.PurchaseOrder{}).Count(&total)
	q.Offset((page - 1) * pageSize).Limit(pageSize).Order("created_at DESC").Find(&list)
	response.OKPage(c, list, total, page, pageSize)
}

func (h *PurchaseOrderHandler) Create(c *gin.Context) {
	ctx := c.Request.Context()
	tenantID := context.GetTenantID(ctx)
	userID := context.GetUserID(ctx)

	var req orderCreateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	if req.SupplierID == 0 {
		response.BadRequest(c, "请选择供应商")
		return
	}

	today := time.Now().Format("20060102")

	var count int64
	h.db.Table("purchase_orders").Where("tenant_id = ? AND order_no LIKE ?", tenantID, "CGDD"+today+"%").Count(&count)

	order := model.PurchaseOrder{
		ID:          snowflake.GenID(),
		TenantID:    tenantID,
		ShopID:      req.ShopID,
		OrderNo:     fmt.Sprintf("CGDD%s%04d", today, count+1),
		SupplierID:  req.SupplierID,
		OrderDate:   req.OrderDate,
		TotalAmount: req.TotalAmount,
		Status:      1,
		Remark:      req.Remark,
		CreatedBy:   userID,
	}

	if err := h.db.Table("purchase_orders").Create(&order).Error; err != nil {
		response.ServerError(c, err.Error())
		return
	}
	response.OK(c, order)
}

func (h *PurchaseOrderHandler) Delete(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	tenantID := context.GetTenantID(c.Request.Context())

	var order model.PurchaseOrder
	if err := h.db.Table("purchase_orders").Where("id = ? AND tenant_id = ?", id, tenantID).First(&order).Error; err != nil {
		response.NotFound(c, "采购订单不存在")
		return
	}
	if order.Status != 1 {
		response.BadRequest(c, "草稿状态才能删除")
		return
	}
	h.db.Table("purchase_orders").Delete(&order)
	response.OKMsg(c, "删除成功")
}

func (h *PurchaseOrderHandler) Audit(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	tenantID := context.GetTenantID(c.Request.Context())

	var order model.PurchaseOrder
	if err := h.db.Table("purchase_orders").Where("id = ? AND tenant_id = ?", id, tenantID).First(&order).Error; err != nil {
		response.NotFound(c, "采购订单不存在")
		return
	}
	if order.Status != 1 && order.Status != 2 {
		response.BadRequest(c, "当前状态不允许审核")
		return
	}
	order.Status = 3
	h.db.Table("purchase_orders").Save(&order)
	response.OKMsg(c, "审核成功")
}
