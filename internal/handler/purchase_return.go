package handler

import (
	"fmt"
	"pisa_server/internal/model"
	"pisa_server/internal/pkg/context"
	"pisa_server/internal/pkg/response"
	"pisa_server/internal/pkg/snowflake"
	"pisa_server/internal/repository"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type PurchaseReturnHandler struct {
	returnRepo *repository.PurchaseReturnRepository
	itemRepo   *repository.PurchaseReturnItemRepository
	db         *gorm.DB
}

func NewPurchaseReturnHandler(
	returnRepo *repository.PurchaseReturnRepository,
	itemRepo *repository.PurchaseReturnItemRepository,
	dbConn *gorm.DB,
) *PurchaseReturnHandler {
	return &PurchaseReturnHandler{
		returnRepo: returnRepo,
		itemRepo:   itemRepo,
		db:         dbConn,
	}
}

type returnCreateReq struct {
	ShopID        int64           `json:"shop_id"`
	WarehouseID   int64           `json:"warehouse_id"`
	SupplierID    int64           `json:"supplier_id"`
	SalesmanID    int64           `json:"salesman_id"`
	AccountID     int64           `json:"account_id"`
	BillDate      string          `json:"bill_date"`
	Discount      float64         `json:"discount"`
	Freight       float64         `json:"freight"`
	DepositOffset float64         `json:"deposit_offset"`
	PaidAmount    float64         `json:"paid_amount"`
	InvoiceStatus int8            `json:"invoice_status"`
	RelatedNo     string          `json:"related_no"`
	Attachments   string          `json:"attachments"`
	Remark        string          `json:"remark"`
	Items         []returnItemReq `json:"items"`
}

type returnItemReq struct {
	GoodsID   int64   `json:"goods_id"`
	Quantity  int     `json:"quantity"`
	UnitPrice float64 `json:"unit_price"`
	Remark    string  `json:"remark"`
}

func (h *PurchaseReturnHandler) List(c *gin.Context) {
	ctx := c.Request.Context()
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	supplierID, _ := strconv.ParseInt(c.Query("supplier_id"), 10, 64)
	status, _ := strconv.ParseInt(c.Query("status"), 10, 64)
	keyword := c.Query("keyword")
	list, total := h.returnRepo.List(ctx, page, pageSize, supplierID, int8(status), keyword)
	response.OKPage(c, list, total, page, pageSize)
}

func (h *PurchaseReturnHandler) GetByID(c *gin.Context) {
	ctx := c.Request.Context()
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	pr, err := h.returnRepo.GetByID(ctx, id)
	if err != nil {
		response.NotFound(c, "退货单不存在")
		return
	}
	h.returnRepo.FillNames(ctx, pr)
	items, _ := h.itemRepo.ListByReturnID(ctx, id)
	h.returnRepo.FillItemDetails(ctx, items)
	pr.Items = items
	response.OK(c, pr)
}

func (h *PurchaseReturnHandler) Create(c *gin.Context) {
	ctx := c.Request.Context()
	tenantID := context.GetTenantID(ctx)
	userID := context.GetUserID(ctx)

	var req returnCreateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	if req.SupplierID == 0 || req.BillDate == "" || len(req.Items) == 0 {
		response.BadRequest(c, "请填写必要信息")
		return
	}

	today := time.Now().Format("20060102")
	var count int64
	h.db.Table("purchase_returns").Where("tenant_id = ? AND order_no LIKE ?", tenantID, "TH"+today+"%").Count(&count)
	pr := model.PurchaseReturn{
		TenantID:      tenantID,
		ShopID:        req.ShopID,
		WarehouseID:   req.WarehouseID,
		OrderNo:       fmt.Sprintf("TH%s%04d", today, count+1),
		SupplierID:    req.SupplierID,
		SalesmanID:    req.SalesmanID,
		AccountID:     req.AccountID,
		BillDate:      req.BillDate,
		DepositOffset: req.DepositOffset,
		PaidAmount:    req.PaidAmount,
		InvoiceStatus: req.InvoiceStatus,
		RelatedNo:     req.RelatedNo,
		Attachments:   req.Attachments,
		Status:        1,
		Remark:        req.Remark,
		CreatedBy:     userID,
	}

	var total float64
	var items []model.PurchaseReturnItem
	for _, item := range req.Items {
		amt := float64(item.Quantity) * item.UnitPrice
		items = append(items, model.PurchaseReturnItem{
			TenantID:  tenantID,
			GoodsID:   item.GoodsID,
			Quantity:  item.Quantity,
			UnitPrice: item.UnitPrice,
			Amount:    amt,
			Remark:    item.Remark,
		})
		total += amt
	}
	pr.TotalAmount = total
	pr.RefundAmount = total
	pr.UnpaidAmount = total - pr.PaidAmount

	if err := h.returnRepo.Create(ctx, &pr); err != nil {
		response.ServerError(c, err.Error())
		return
	}
	for i := range items {
		items[i].PurchaseReturnID = pr.ID
	}
	h.itemRepo.BatchCreate(ctx, items)
	response.OK(c, pr)
}

func (h *PurchaseReturnHandler) Delete(c *gin.Context) {
	ctx := c.Request.Context()
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	pr, err := h.returnRepo.GetByID(ctx, id)
	if err != nil {
		response.NotFound(c, "退货单不存在")
		return
	}
	if pr.Status == 3 {
		response.BadRequest(c, "已生效的退货单不能删除")
		return
	}
	h.itemRepo.DeleteByReturnID(ctx, id)
	h.returnRepo.Delete(ctx, id)
	response.OKMsg(c, "删除成功")
}

func (h *PurchaseReturnHandler) Audit(c *gin.Context) {
	ctx := c.Request.Context()
	tenantID := context.GetTenantID(ctx)
	shopID := context.GetShopID(ctx)

	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	pr, err := h.returnRepo.GetByID(ctx, id)
	if err != nil {
		response.NotFound(c, "退货单不存在")
		return
	}
	if pr.Status != 1 {
		response.BadRequest(c, "当前状态不允许审核")
		return
	}

	items, err := h.itemRepo.ListByReturnID(ctx, id)
	if err != nil || len(items) == 0 {
		response.BadRequest(c, "无明细数据")
		return
	}

	tx := h.db.Begin()
	for _, item := range items {
		var balance model.StockBalance
		result := tx.Where("tenant_id = ? AND shop_id = ? AND goods_id = ? AND warehouse_id = ?",
			tenantID, shopID, item.GoodsID, pr.WarehouseID).First(&balance)
		if result.Error != nil {
			tx.Rollback()
			response.ServerError(c, "库存记录不存在")
			return
		}
		beforeStock := balance.Quantity
		balance.Quantity -= item.Quantity
		if balance.Quantity < 0 {
			tx.Rollback()
			response.ServerError(c, "库存不足")
			return
		}
		balance.TotalCost = float64(balance.Quantity) * balance.CostPrice
		tx.Save(&balance)

		tx.Table("stock_logs").Create(map[string]interface{}{
			"id": snowflake.GenID(), "tenant_id": tenantID, "shop_id": shopID,
			"warehouse_id": pr.WarehouseID, "goods_id": item.GoodsID,
			"type": 2, "quantity": item.Quantity, "before_stock": beforeStock,
			"after_stock": balance.Quantity, "related_type": "purchase_return",
			"related_id": pr.ID, "related_no": pr.OrderNo, "created_at": time.Now(),
		})
	}

	pr.Status = 3
	h.returnRepo.Update(ctx, pr)
	tx.Commit()
	response.OKMsg(c, "审核成功")
}
