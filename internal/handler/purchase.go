package handler

import (
	"pisa_server/internal/model"
	"pisa_server/internal/pkg/response"
	"pisa_server/internal/service"
	"strconv"

	"github.com/gin-gonic/gin"
)

type PurchaseHandler struct {
	svc *service.PurchaseService
}

func NewPurchaseHandler(svc *service.PurchaseService) *PurchaseHandler {
	return &PurchaseHandler{svc: svc}
}

type purchaseCreateReq struct {
	ShopID        model.FlexInt64   `json:"shop_id"`
	WarehouseID   int64             `json:"warehouse_id"`
	SupplierID    int64             `json:"supplier_id"`
	SalesmanID    int64             `json:"salesman_id"`
	AccountID     int64             `json:"account_id"`
	BillDate      string            `json:"bill_date"`
	Discount      float64           `json:"discount"`
	Freight       float64           `json:"freight"`
	DepositOffset float64           `json:"deposit_offset"`
	PaidAmount    float64           `json:"paid_amount"`
	InvoiceStatus int8              `json:"invoice_status"`
	RelatedNo     string            `json:"related_no"`
	Attachments   string            `json:"attachments"`
	Remark        string            `json:"remark"`
	Items         []purchaseItemReq `json:"items"`
}

type purchaseItemReq struct {
	GoodsID   model.FlexInt64 `json:"goods_id"`
	Quantity  int     `json:"quantity"`
	UnitPrice float64 `json:"unit_price"`
	Remark    string  `json:"remark"`
}

func (h *PurchaseHandler) List(c *gin.Context) {
	ctx := c.Request.Context()
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	supplierID, _ := strconv.ParseInt(c.Query("supplier_id"), 10, 64)
	status, _ := strconv.ParseInt(c.Query("status"), 10, 64)
	keyword := c.Query("keyword")
	dateFrom := c.Query("date_from")
	dateTo := c.Query("date_to")

	list, total := h.svc.List(ctx, page, pageSize, supplierID, int8(status), keyword, dateFrom, dateTo)
	response.OKPage(c, list, total, page, pageSize)
}

func (h *PurchaseHandler) GetByID(c *gin.Context) {
	ctx := c.Request.Context()
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	purchase, err := h.svc.GetByID(ctx, id)
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}
	response.OK(c, purchase)
}

func (h *PurchaseHandler) Create(c *gin.Context) {
	ctx := c.Request.Context()
	var req purchaseCreateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	if req.SupplierID == 0 {
		response.BadRequest(c, "请选择供应商")
		return
	}
	if req.BillDate == "" {
		response.BadRequest(c, "请选择日期")
		return
	}
	if len(req.Items) == 0 {
		response.BadRequest(c, "请添加明细")
		return
	}

	purchase := &model.Purchase{
		ShopID:        int64(req.ShopID),
		WarehouseID:   req.WarehouseID,
		SupplierID:    req.SupplierID,
		SalesmanID:    req.SalesmanID,
		AccountID:     req.AccountID,
		BillDate:      req.BillDate,
		Discount:      req.Discount,
		Freight:       req.Freight,
		DepositOffset: req.DepositOffset,
		PaidAmount:    req.PaidAmount,
		InvoiceStatus: req.InvoiceStatus,
		RelatedNo:     req.RelatedNo,
		Attachments:   req.Attachments,
		Remark:        req.Remark,
	}

	var items []model.PurchaseItem
	for _, item := range req.Items {
		items = append(items, model.PurchaseItem{
			GoodsID:   int64(item.GoodsID),
			Quantity:  item.Quantity,
			UnitPrice: item.UnitPrice,
			Remark:    item.Remark,
		})
	}

	if err := h.svc.Create(ctx, purchase, items); err != nil {
		response.ServerError(c, err.Error())
		return
	}
	response.OK(c, purchase)
}

func (h *PurchaseHandler) Update(c *gin.Context) {
	ctx := c.Request.Context()
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	existing, err := h.svc.GetByID(ctx, id)
	if err != nil {
		response.NotFound(c, "进货单不存在")
		return
	}
	if existing.Status != 1 {
		response.BadRequest(c, "草稿状态才能编辑")
		return
	}

	var req purchaseCreateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	// 删除旧明细
	h.svc.GetByID(ctx, id) // ensure loaded
	if err := h.svc.Delete(ctx, id); err != nil {
		response.ServerError(c, err.Error())
		return
	}

	// 创建新的
	purchase := &model.Purchase{
		ShopID:        int64(req.ShopID),
		WarehouseID:   req.WarehouseID,
		SupplierID:    req.SupplierID,
		SalesmanID:    req.SalesmanID,
		AccountID:     req.AccountID,
		BillDate:      req.BillDate,
		Discount:      req.Discount,
		Freight:       req.Freight,
		DepositOffset: req.DepositOffset,
		PaidAmount:    req.PaidAmount,
		InvoiceStatus: req.InvoiceStatus,
		RelatedNo:     req.RelatedNo,
		Attachments:   req.Attachments,
		Remark:        req.Remark,
	}
	var items []model.PurchaseItem
	for _, item := range req.Items {
		items = append(items, model.PurchaseItem{
			GoodsID:   int64(item.GoodsID),
			Quantity:  item.Quantity,
			UnitPrice: item.UnitPrice,
			Remark:    item.Remark,
		})
	}
	if err := h.svc.Create(ctx, purchase, items); err != nil {
		response.ServerError(c, err.Error())
		return
	}
	response.OK(c, purchase)
}

func (h *PurchaseHandler) Delete(c *gin.Context) {
	ctx := c.Request.Context()
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	if err := h.svc.Delete(ctx, id); err != nil {
		response.ServerError(c, err.Error())
		return
	}
	response.OKMsg(c, "删除成功")
}

func (h *PurchaseHandler) Audit(c *gin.Context) {
	ctx := c.Request.Context()
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	if err := h.svc.Audit(ctx, id); err != nil {
		response.ServerError(c, err.Error())
		return
	}
	response.OKMsg(c, "审核成功")
}

func (h *PurchaseHandler) UnAudit(c *gin.Context) {
	ctx := c.Request.Context()
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	if err := h.svc.UnAudit(ctx, id); err != nil {
		response.ServerError(c, err.Error())
		return
	}
	response.OKMsg(c, "反审核成功")
}
