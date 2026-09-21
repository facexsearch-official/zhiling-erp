package handler

import (
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"pisa_server/internal/model"
	"pisa_server/internal/pkg/context"
	"pisa_server/internal/pkg/response"
	"pisa_server/internal/pkg/snowflake"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type PurchaseOrderHandler struct {
	db *gorm.DB
}

func NewPurchaseOrderHandler(dbConn *gorm.DB) *PurchaseOrderHandler {
	return &PurchaseOrderHandler{db: dbConn}
}

func round2o(v float64) float64 { return math.Round(v*100) / 100 }

type orderItemReq struct {
	GoodsID   int64   `json:"goods_id"`
	Quantity  int     `json:"quantity"`
	UnitPrice float64 `json:"unit_price"`
	Remark    string  `json:"remark"`
}

type orderCreateReq struct {
	ShopID        int64          `json:"shop_id"`
	WarehouseID   int64          `json:"warehouse_id"`
	SupplierID    int64          `json:"supplier_id"`
	SalesmanID    int64          `json:"salesman_id"`
	AccountID     int64          `json:"account_id"`
	OrderDate     string         `json:"order_date"`
	Discount      float64        `json:"discount"`
	Freight       float64        `json:"freight"`
	DepositOffset float64        `json:"deposit_offset"`
	PaidAmount    float64        `json:"paid_amount"`
	InvoiceStatus int8           `json:"invoice_status"`
	RelatedNo     string         `json:"related_no"`
	Attachments   string         `json:"attachments"`
	Remark        string         `json:"remark"`
	Items         []orderItemReq `json:"items"`
}

func (h *PurchaseOrderHandler) List(c *gin.Context) {
	ctx := c.Request.Context()
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	keyword := c.Query("keyword")
	dateFrom := c.Query("date_from")
	dateTo := c.Query("date_to")
	tenantID := context.GetTenantID(ctx)

	var total int64
	var list []model.PurchaseOrder
	q := h.db.Table("purchase_orders AS o").
		Joins("LEFT JOIN suppliers s ON s.id = o.supplier_id").
		Joins("LEFT JOIN salesmen sm ON sm.id = o.salesman_id").
		Joins("LEFT JOIN accounts a ON a.id = o.account_id").
		Joins("LEFT JOIN users u ON u.id = o.created_by").
		Where("o.tenant_id = ?", tenantID)
	if keyword != "" {
		kw := "%" + keyword + "%"
		q = q.Where("o.order_no LIKE ? OR s.name LIKE ?", kw, kw)
	}
	if dateFrom != "" {
		q = q.Where("o.order_date >= ?", dateFrom)
	}
	if dateTo != "" {
		q = q.Where("o.order_date <= ?", dateTo)
	}
	q.Session(&gorm.Session{}).Count(&total)
	q.Select("o.*, s.name AS supplier_name, sm.name AS salesman_name, a.name AS account_name, u.nickname AS maker_name").
		Offset((page - 1) * pageSize).Limit(pageSize).Order("o.created_at DESC").Scan(&list)
	for i := range list {
		list[i].IDStr = strconv.FormatInt(list[i].ID, 10)
	}
	response.OKPage(c, list, total, page, pageSize)
}

func (h *PurchaseOrderHandler) GetByID(c *gin.Context) {
	ctx := c.Request.Context()
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	tenantID := context.GetTenantID(ctx)
	var o model.PurchaseOrder
	if err := h.db.Table("purchase_orders").Where("id = ? AND tenant_id = ?", id, tenantID).First(&o).Error; err != nil {
		response.NotFound(c, "进货预订不存在")
		return
	}
	o.IDStr = strconv.FormatInt(o.ID, 10)
	h.fillNames(tenantID, &o)
	var items []model.PurchaseOrderItem
	h.db.Table("purchase_order_items").Where("order_id = ?", o.ID).Order("id ASC").Find(&items)
	h.fillItemDetails(tenantID, items)
	o.Items = items
	response.OK(c, o)
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
	if len(req.Items) == 0 {
		response.BadRequest(c, "请添加明细")
		return
	}

	today := time.Now().Format("20060102")
	var count int64
	h.db.Table("purchase_orders").Where("tenant_id = ? AND order_no LIKE ?", tenantID, "CGDD"+today+"%").Count(&count)

	o := model.PurchaseOrder{
		ID:            snowflake.GenID(),
		TenantID:      tenantID,
		ShopID:        req.ShopID,
		WarehouseID:   req.WarehouseID,
		OrderNo:       fmt.Sprintf("CGDD%s%04d", today, count+1),
		SupplierID:    req.SupplierID,
		SalesmanID:    req.SalesmanID,
		AccountID:     req.AccountID,
		OrderDate:     req.OrderDate,
		Discount:      req.Discount,
		Freight:       req.Freight,
		DepositOffset: req.DepositOffset,
		PaidAmount:    req.PaidAmount,
		InvoiceStatus: req.InvoiceStatus,
		RelatedNo:     req.RelatedNo,
		Attachments:   req.Attachments,
		Status:        1,
		Remark:        req.Remark,
		CreatedBy:     userID,
	}
	var subtotal float64
	items := make([]model.PurchaseOrderItem, 0, len(req.Items))
	for _, it := range req.Items {
		amt := round2o(float64(it.Quantity) * it.UnitPrice)
		subtotal += amt
		items = append(items, model.PurchaseOrderItem{
			ID: snowflake.GenID(), TenantID: tenantID, GoodsID: it.GoodsID,
			Quantity: it.Quantity, UnitPrice: it.UnitPrice, Amount: amt, Remark: it.Remark,
		})
	}
	if o.Discount <= 0 {
		o.Discount = 100
	}
	o.Subtotal = round2o(subtotal)
	o.DiscountedAmount = round2o(subtotal * o.Discount / 100)
	o.TotalAmount = round2o(o.DiscountedAmount + o.Freight)
	o.UnpaidAmount = round2o(o.TotalAmount - o.PaidAmount)

	if err := h.db.Table("purchase_orders").Create(&o).Error; err != nil {
		response.ServerError(c, err.Error())
		return
	}
	for i := range items {
		items[i].OrderID = o.ID
	}
	if err := h.db.Table("purchase_order_items").Create(&items).Error; err != nil {
		response.ServerError(c, err.Error())
		return
	}
	o.Items = items
	o.IDStr = strconv.FormatInt(o.ID, 10)
	response.OK(c, o)
}

func (h *PurchaseOrderHandler) Delete(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	tenantID := context.GetTenantID(c.Request.Context())

	var order model.PurchaseOrder
	if err := h.db.Table("purchase_orders").Where("id = ? AND tenant_id = ?", id, tenantID).First(&order).Error; err != nil {
		response.NotFound(c, "进货预订不存在")
		return
	}
	if order.Status != 1 {
		response.BadRequest(c, "草稿状态才能删除")
		return
	}
	h.db.Table("purchase_order_items").Where("order_id = ?", id).Delete(&model.PurchaseOrderItem{})
	h.db.Table("purchase_orders").Delete(&order)
	response.OKMsg(c, "删除成功")
}

func (h *PurchaseOrderHandler) Audit(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	tenantID := context.GetTenantID(c.Request.Context())

	var order model.PurchaseOrder
	if err := h.db.Table("purchase_orders").Where("id = ? AND tenant_id = ?", id, tenantID).First(&order).Error; err != nil {
		response.NotFound(c, "进货预订不存在")
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

/* ── helpers ── */

func (h *PurchaseOrderHandler) fillNames(tenantID int64, o *model.PurchaseOrder) {
	if o.SupplierID != 0 {
		var s model.Supplier
		if h.db.Where("id = ? AND tenant_id = ?", o.SupplierID, tenantID).First(&s).Error == nil {
			o.SupplierName = s.Name
		}
	}
	if o.SalesmanID != 0 {
		var sm model.Salesman
		if h.db.Where("id = ? AND tenant_id = ?", o.SalesmanID, tenantID).First(&sm).Error == nil {
			o.SalesmanName = sm.Name
		}
	}
	if o.AccountID != 0 {
		var a model.Account
		if h.db.Where("id = ? AND tenant_id = ?", o.AccountID, tenantID).First(&a).Error == nil {
			o.AccountName = a.Name
		}
	}
}

func (h *PurchaseOrderHandler) fillItemDetails(tenantID int64, items []model.PurchaseOrderItem) {
	if len(items) == 0 {
		return
	}
	ids := make([]int64, 0, len(items))
	for _, it := range items {
		ids = append(ids, it.GoodsID)
	}
	var goods []model.Goods
	h.db.Where("tenant_id = ? AND id IN ?", tenantID, ids).Find(&goods)
	m := map[int64]model.Goods{}
	for _, g := range goods {
		m[g.ID] = g
	}
	for i := range items {
		g := m[items[i].GoodsID]
		items[i].GoodsName = g.Name
		items[i].GoodsCode = g.Code
		items[i].UnitName = g.MainUnit
		items[i].Spec = orderSpecNames(g.SpecGroups)
		items[i].Barcode = g.Barcode
		items[i].ImageURL = g.ImageURL
		items[i].Brand = g.Brand
		items[i].Origin = g.Origin
	}
}

func orderSpecNames(specGroups string) string {
	if specGroups == "" {
		return ""
	}
	var gs []struct {
		Name string `json:"name"`
	}
	if err := json.Unmarshal([]byte(specGroups), &gs); err != nil {
		return ""
	}
	names := make([]string, 0, len(gs))
	for _, g := range gs {
		if g.Name != "" {
			names = append(names, g.Name)
		}
	}
	return strings.Join(names, "/")
}
