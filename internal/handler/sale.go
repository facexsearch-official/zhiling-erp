package handler

import (
	"fmt"
	"strconv"
	"time"

	"pisa_server/internal/model"
	"pisa_server/internal/pkg/context"
	"pisa_server/internal/pkg/response"
	"pisa_server/internal/pkg/snowflake"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type SalesHandler struct {
	db *gorm.DB
}

func NewSalesHandler(dbConn *gorm.DB) *SalesHandler {
	return &SalesHandler{db: dbConn}
}

type saleItemReq struct {
	GoodsID   int64   `json:"goods_id"`
	Quantity  int     `json:"quantity"`
	UnitPrice float64 `json:"unit_price"`
	Remark    string  `json:"remark"`
}

type saleCreateReq struct {
	ShopID           int64         `json:"shop_id"`
	WarehouseID      int64         `json:"warehouse_id"`
	CustomerID       int64         `json:"customer_id"`
	SalesmanID       int64         `json:"salesman_id"`
	AccountID        int64         `json:"account_id"`
	BillDate         string        `json:"bill_date"`
	OrderDate        string        `json:"order_date"`
	RelatedOrderNo   string        `json:"related_order_no"`
	Discount         float64       `json:"discount"`
	Subtotal         float64       `json:"subtotal"`
	RoundOff         float64       `json:"round_off"`
	TotalAmount      float64       `json:"total_amount"`
	ReceivedAmount   float64       `json:"received_amount"`
	UnreceivedAmount float64       `json:"unreceived_amount"`
	InvoiceStatus    int8          `json:"invoice_status"`
	PrintStatus      int8          `json:"print_status"`
	Attachments      string        `json:"attachments"`
	Remark           string        `json:"remark"`
	Items            []saleItemReq `json:"items"`
}

/* ── helpers ── */

func (h *SalesHandler) names(tenantID, customerID, salesmanID, accountID, createdBy int64) (string, string, string, string) {
	var cn, sn, an, mn string
	if customerID != 0 {
		var c model.Customer
		if h.db.Where("id = ? AND tenant_id = ?", customerID, tenantID).First(&c).Error == nil {
			cn = c.Name
		}
	}
	if salesmanID != 0 {
		var sm model.Salesman
		if h.db.Where("id = ? AND tenant_id = ?", salesmanID, tenantID).First(&sm).Error == nil {
			sn = sm.Name
		}
	}
	if accountID != 0 {
		var a model.Account
		if h.db.Where("id = ? AND tenant_id = ?", accountID, tenantID).First(&a).Error == nil {
			an = a.Name
		}
	}
	if createdBy != 0 {
		var u model.User
		if h.db.Where("id = ?", createdBy).First(&u).Error == nil {
			mn = u.Nickname
		}
	}
	return cn, sn, an, mn
}

func (h *SalesHandler) goodsMap(tenantID int64, ids []int64) map[int64]model.Goods {
	m := map[int64]model.Goods{}
	if len(ids) == 0 {
		return m
	}
	var goods []model.Goods
	h.db.Where("tenant_id = ? AND id IN ?", tenantID, ids).Find(&goods)
	for _, g := range goods {
		m[g.ID] = g
	}
	return m
}

func nextNo(db *gorm.DB, tenantID int64, table, prefix string) string {
	today := time.Now().Format("20060102")
	var count int64
	db.Table(table).Where("tenant_id = ? AND order_no LIKE ?", tenantID, prefix+today+"%").Count(&count)
	return fmt.Sprintf("%s%s%04d", prefix, today, count+1)
}

func calcItems(req []saleItemReq) ([]model.SaleItem, float64) {
	var total float64
	items := make([]model.SaleItem, 0, len(req))
	for _, it := range req {
		amt := round2o(float64(it.Quantity) * it.UnitPrice)
		total += amt
		items = append(items, model.SaleItem{ID: snowflake.GenID(), GoodsID: it.GoodsID, Quantity: it.Quantity, UnitPrice: it.UnitPrice, Amount: amt, Remark: it.Remark})
	}
	return items, round2o(total)
}

/* ── 销售单 ── */

func (h *SalesHandler) ListSales(c *gin.Context) {
	ctx := c.Request.Context()
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	keyword := c.Query("keyword")
	dateFrom := c.Query("date_from")
	dateTo := c.Query("date_to")
	tenantID := context.GetTenantID(ctx)

	var total int64
	var list []model.Sale
	q := h.db.Table("sales AS p").
		Joins("LEFT JOIN customers c ON c.id = p.customer_id").
		Joins("LEFT JOIN salesmen sm ON sm.id = p.salesman_id").
		Joins("LEFT JOIN accounts a ON a.id = p.account_id").
		Joins("LEFT JOIN users u ON u.id = p.created_by").
		Where("p.tenant_id = ?", tenantID)
	if keyword != "" {
		kw := "%" + keyword + "%"
		q = q.Where("p.order_no LIKE ? OR c.name LIKE ?", kw, kw)
	}
	if dateFrom != "" {
		q = q.Where("p.bill_date >= ?", dateFrom)
	}
	if dateTo != "" {
		q = q.Where("p.bill_date <= ?", dateTo)
	}
	q.Session(&gorm.Session{}).Count(&total)
	q.Select("p.*, c.name AS customer_name, sm.name AS salesman_name, a.name AS account_name, u.nickname AS maker_name").
		Offset((page - 1) * pageSize).Limit(pageSize).Order("p.created_at DESC").Scan(&list)
	for i := range list {
		list[i].IDStr = strconv.FormatInt(list[i].ID, 10)
	}
	response.OKPage(c, list, total, page, pageSize)
}

func (h *SalesHandler) GetSale(c *gin.Context) {
	ctx := c.Request.Context()
	tenantID := context.GetTenantID(ctx)
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var s model.Sale
	if err := h.db.Table("sales").Where("id = ? AND tenant_id = ?", id, tenantID).First(&s).Error; err != nil {
		response.NotFound(c, "销货单不存在")
		return
	}
	s.IDStr = strconv.FormatInt(s.ID, 10)
	s.CustomerName, s.SalesmanName, s.AccountName, s.MakerName = h.names(tenantID, s.CustomerID, s.SalesmanID, s.AccountID, s.CreatedBy)
	var items []model.SaleItem
	h.db.Table("sale_items").Where("sale_id = ?", s.ID).Order("id ASC").Find(&items)
	gm := h.goodsMap(tenantID, itemGoodsIDs(items))
	for i := range items {
		g := gm[items[i].GoodsID]
		items[i].GoodsName = g.Name
		items[i].GoodsCode = g.Code
		items[i].UnitName = g.MainUnit
		items[i].Spec = orderSpecNames(g.SpecGroups)
		items[i].Barcode = g.Barcode
		items[i].ImageURL = g.ImageURL
		items[i].Brand = g.Brand
		items[i].Origin = g.Origin
	}
	s.Items = items
	response.OK(c, s)
}

func (h *SalesHandler) CreateSale(c *gin.Context) {
	ctx := c.Request.Context()
	tenantID := context.GetTenantID(ctx)
	userID := context.GetUserID(ctx)
	var req saleCreateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	if req.CustomerID == 0 || len(req.Items) == 0 {
		response.BadRequest(c, "请选择客户并添加明细")
		return
	}
	items, subtotal := calcItems(req.Items)
	discount := req.Discount
	if discount <= 0 {
		discount = 100
	}
	s := model.Sale{
		ID: snowflake.GenID(), TenantID: tenantID, ShopID: req.ShopID, WarehouseID: req.WarehouseID,
		OrderNo: nextNo(h.db, tenantID, "sales", "XH"), RelatedOrderNo: req.RelatedOrderNo,
		CustomerID: req.CustomerID, SalesmanID: req.SalesmanID, AccountID: req.AccountID,
		BillDate: req.BillDate, Discount: discount, Subtotal: subtotal,
		RoundOff: req.RoundOff, InvoiceStatus: req.InvoiceStatus, PrintStatus: req.PrintStatus,
		Attachments: req.Attachments, Status: 1, Remark: req.Remark, CreatedBy: userID,
	}
	s.TotalAmount = round2o(subtotal*discount/100 - s.RoundOff)
	s.ReceivedAmount = req.ReceivedAmount
	s.UnreceivedAmount = round2o(s.TotalAmount - s.ReceivedAmount)
	if s.ReceivedAmount <= 0 {
		s.ReceiveStatus = 0
	} else if s.UnreceivedAmount <= 0 {
		s.ReceiveStatus = 2
	} else {
		s.ReceiveStatus = 1
	}
	if err := h.db.Table("sales").Create(&s).Error; err != nil {
		response.ServerError(c, err.Error())
		return
	}
	for i := range items {
		items[i].TenantID = tenantID
		items[i].SaleID = s.ID
	}
	if len(items) > 0 {
		h.db.Table("sale_items").Create(&items)
	}
	s.IDStr = strconv.FormatInt(s.ID, 10)
	s.Items = items
	response.OK(c, s)
}

func (h *SalesHandler) DeleteSale(c *gin.Context) {
	ctx := c.Request.Context()
	tenantID := context.GetTenantID(ctx)
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	h.db.Table("sale_items").Where("sale_id = ?", id).Delete(&model.SaleItem{})
	h.db.Table("sales").Where("id = ? AND tenant_id = ?", id, tenantID).Delete(&model.Sale{})
	response.OKMsg(c, "删除成功")
}

/* ── 销售预订 ── */

func (h *SalesHandler) ListSaleOrders(c *gin.Context) {
	ctx := c.Request.Context()
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	keyword := c.Query("keyword")
	dateFrom := c.Query("date_from")
	dateTo := c.Query("date_to")
	tenantID := context.GetTenantID(ctx)

	var total int64
	var list []model.SaleOrder
	q := h.db.Table("sale_orders AS p").
		Joins("LEFT JOIN customers c ON c.id = p.customer_id").
		Joins("LEFT JOIN salesmen sm ON sm.id = p.salesman_id").
		Joins("LEFT JOIN accounts a ON a.id = p.account_id").
		Joins("LEFT JOIN users u ON u.id = p.created_by").
		Where("p.tenant_id = ?", tenantID)
	if keyword != "" {
		kw := "%" + keyword + "%"
		q = q.Where("p.order_no LIKE ? OR c.name LIKE ?", kw, kw)
	}
	if dateFrom != "" {
		q = q.Where("p.order_date >= ?", dateFrom)
	}
	if dateTo != "" {
		q = q.Where("p.order_date <= ?", dateTo)
	}
	q.Session(&gorm.Session{}).Count(&total)
	q.Select("p.*, c.name AS customer_name, sm.name AS salesman_name, a.name AS account_name, u.nickname AS maker_name").
		Offset((page - 1) * pageSize).Limit(pageSize).Order("p.created_at DESC").Scan(&list)
	for i := range list {
		list[i].IDStr = strconv.FormatInt(list[i].ID, 10)
	}
	response.OKPage(c, list, total, page, pageSize)
}

func (h *SalesHandler) GetSaleOrder(c *gin.Context) {
	ctx := c.Request.Context()
	tenantID := context.GetTenantID(ctx)
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var s model.SaleOrder
	if err := h.db.Table("sale_orders").Where("id = ? AND tenant_id = ?", id, tenantID).First(&s).Error; err != nil {
		response.NotFound(c, "销售预订不存在")
		return
	}
	s.IDStr = strconv.FormatInt(s.ID, 10)
	s.CustomerName, s.SalesmanName, s.AccountName, s.MakerName = h.names(tenantID, s.CustomerID, s.SalesmanID, s.AccountID, s.CreatedBy)
	var items []model.SaleOrderItem
	h.db.Table("sale_order_items").Where("order_id = ?", s.ID).Order("id ASC").Find(&items)
	gm := h.goodsMap(tenantID, orderItemGoodsIDs(items))
	for i := range items {
		g := gm[items[i].GoodsID]
		items[i].GoodsName = g.Name
		items[i].GoodsCode = g.Code
		items[i].UnitName = g.MainUnit
		items[i].Spec = orderSpecNames(g.SpecGroups)
		items[i].Barcode = g.Barcode
		items[i].ImageURL = g.ImageURL
		items[i].Brand = g.Brand
		items[i].Origin = g.Origin
	}
	s.Items = items
	response.OK(c, s)
}

func (h *SalesHandler) CreateSaleOrder(c *gin.Context) {
	ctx := c.Request.Context()
	tenantID := context.GetTenantID(ctx)
	userID := context.GetUserID(ctx)
	var req saleCreateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	if req.CustomerID == 0 || len(req.Items) == 0 {
		response.BadRequest(c, "请选择客户并添加明细")
		return
	}
	var total float64
	items := make([]model.SaleOrderItem, 0, len(req.Items))
	for _, it := range req.Items {
		amt := round2o(float64(it.Quantity) * it.UnitPrice)
		total += amt
		items = append(items, model.SaleOrderItem{ID: snowflake.GenID(), GoodsID: it.GoodsID, Quantity: it.Quantity, UnitPrice: it.UnitPrice, Amount: amt, Remark: it.Remark})
	}
	total = round2o(total)
	date := req.OrderDate
	if date == "" {
		date = req.BillDate
	}
	s := model.SaleOrder{
		ID: snowflake.GenID(), TenantID: tenantID, ShopID: req.ShopID, WarehouseID: req.WarehouseID,
		OrderNo: nextNo(h.db, tenantID, "sale_orders", "XSDD"), CustomerID: req.CustomerID,
		SalesmanID: req.SalesmanID, AccountID: req.AccountID, OrderDate: date,
		TotalAmount: total, ReceivedAmount: req.ReceivedAmount,
		PrintStatus: req.PrintStatus, Attachments: req.Attachments, Status: 1, Remark: req.Remark, CreatedBy: userID,
	}
	s.UnreceivedAmount = round2o(s.TotalAmount - s.ReceivedAmount)
	if err := h.db.Table("sale_orders").Create(&s).Error; err != nil {
		response.ServerError(c, err.Error())
		return
	}
	for i := range items {
		items[i].TenantID = tenantID
		items[i].OrderID = s.ID
	}
	if len(items) > 0 {
		h.db.Table("sale_order_items").Create(&items)
	}
	s.IDStr = strconv.FormatInt(s.ID, 10)
	response.OK(c, s)
}

func (h *SalesHandler) DeleteSaleOrder(c *gin.Context) {
	ctx := c.Request.Context()
	tenantID := context.GetTenantID(ctx)
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	h.db.Table("sale_order_items").Where("order_id = ?", id).Delete(&model.SaleOrderItem{})
	h.db.Table("sale_orders").Where("id = ? AND tenant_id = ?", id, tenantID).Delete(&model.SaleOrder{})
	response.OKMsg(c, "删除成功")
}

/* ── 销售退货 ── */

func (h *SalesHandler) ListSalesReturns(c *gin.Context) {
	ctx := c.Request.Context()
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	keyword := c.Query("keyword")
	dateFrom := c.Query("date_from")
	dateTo := c.Query("date_to")
	tenantID := context.GetTenantID(ctx)

	var total int64
	var list []model.SalesReturn
	q := h.db.Table("sales_returns AS p").
		Joins("LEFT JOIN customers c ON c.id = p.customer_id").
		Joins("LEFT JOIN salesmen sm ON sm.id = p.salesman_id").
		Joins("LEFT JOIN accounts a ON a.id = p.account_id").
		Joins("LEFT JOIN users u ON u.id = p.created_by").
		Where("p.tenant_id = ?", tenantID)
	if keyword != "" {
		kw := "%" + keyword + "%"
		q = q.Where("p.order_no LIKE ? OR c.name LIKE ?", kw, kw)
	}
	if dateFrom != "" {
		q = q.Where("p.bill_date >= ?", dateFrom)
	}
	if dateTo != "" {
		q = q.Where("p.bill_date <= ?", dateTo)
	}
	q.Session(&gorm.Session{}).Count(&total)
	q.Select("p.*, c.name AS customer_name, sm.name AS salesman_name, a.name AS account_name, u.nickname AS maker_name").
		Offset((page - 1) * pageSize).Limit(pageSize).Order("p.created_at DESC").Scan(&list)
	for i := range list {
		list[i].IDStr = strconv.FormatInt(list[i].ID, 10)
	}
	response.OKPage(c, list, total, page, pageSize)
}

func (h *SalesHandler) GetSalesReturn(c *gin.Context) {
	ctx := c.Request.Context()
	tenantID := context.GetTenantID(ctx)
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var s model.SalesReturn
	if err := h.db.Table("sales_returns").Where("id = ? AND tenant_id = ?", id, tenantID).First(&s).Error; err != nil {
		response.NotFound(c, "销售退货单不存在")
		return
	}
	s.IDStr = strconv.FormatInt(s.ID, 10)
	s.CustomerName, s.SalesmanName, s.AccountName, s.MakerName = h.names(tenantID, s.CustomerID, s.SalesmanID, s.AccountID, s.CreatedBy)
	var items []model.SalesReturnItem
	h.db.Table("sales_return_items").Where("return_id = ?", s.ID).Order("id ASC").Find(&items)
	gm := h.goodsMap(tenantID, returnItemGoodsIDs(items))
	for i := range items {
		g := gm[items[i].GoodsID]
		items[i].GoodsName = g.Name
		items[i].GoodsCode = g.Code
		items[i].UnitName = g.MainUnit
		items[i].Spec = orderSpecNames(g.SpecGroups)
		items[i].Barcode = g.Barcode
		items[i].ImageURL = g.ImageURL
		items[i].Brand = g.Brand
		items[i].Origin = g.Origin
	}
	s.Items = items
	response.OK(c, s)
}

func (h *SalesHandler) CreateSalesReturn(c *gin.Context) {
	ctx := c.Request.Context()
	tenantID := context.GetTenantID(ctx)
	userID := context.GetUserID(ctx)
	var req saleCreateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	if req.CustomerID == 0 || len(req.Items) == 0 {
		response.BadRequest(c, "请选择客户并添加明细")
		return
	}
	var total float64
	items := make([]model.SalesReturnItem, 0, len(req.Items))
	for _, it := range req.Items {
		amt := round2o(float64(it.Quantity) * it.UnitPrice)
		total += amt
		items = append(items, model.SalesReturnItem{ID: snowflake.GenID(), GoodsID: it.GoodsID, Quantity: it.Quantity, UnitPrice: it.UnitPrice, Amount: amt, Remark: it.Remark})
	}
	total = round2o(total)
	s := model.SalesReturn{
		ID: snowflake.GenID(), TenantID: tenantID, ShopID: req.ShopID, WarehouseID: req.WarehouseID,
		OrderNo: nextNo(h.db, tenantID, "sales_returns", "XSTH"), CustomerID: req.CustomerID,
		SalesmanID: req.SalesmanID, AccountID: req.AccountID, BillDate: req.BillDate,
		RoundOff: req.RoundOff, TotalAmount: round2o(total - req.RoundOff), ReceivedAmount: req.ReceivedAmount,
		PrintStatus: req.PrintStatus, Attachments: req.Attachments, Status: 1, Remark: req.Remark, CreatedBy: userID,
	}
	s.UnreceivedAmount = round2o(s.TotalAmount - s.ReceivedAmount)
	if err := h.db.Table("sales_returns").Create(&s).Error; err != nil {
		response.ServerError(c, err.Error())
		return
	}
	for i := range items {
		items[i].TenantID = tenantID
		items[i].ReturnID = s.ID
	}
	if len(items) > 0 {
		h.db.Table("sales_return_items").Create(&items)
	}
	s.IDStr = strconv.FormatInt(s.ID, 10)
	response.OK(c, s)
}

func (h *SalesHandler) DeleteSalesReturn(c *gin.Context) {
	ctx := c.Request.Context()
	tenantID := context.GetTenantID(ctx)
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	h.db.Table("sales_return_items").Where("return_id = ?", id).Delete(&model.SalesReturnItem{})
	h.db.Table("sales_returns").Where("id = ? AND tenant_id = ?", id, tenantID).Delete(&model.SalesReturn{})
	response.OKMsg(c, "删除成功")
}

/* ── 报价 ── */

func (h *SalesHandler) ListQuotes(c *gin.Context) {
	ctx := c.Request.Context()
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	keyword := c.Query("keyword")
	dateFrom := c.Query("date_from")
	dateTo := c.Query("date_to")
	tenantID := context.GetTenantID(ctx)

	var total int64
	var list []model.Quote
	q := h.db.Table("quotes AS p").
		Joins("LEFT JOIN customers c ON c.id = p.customer_id").
		Joins("LEFT JOIN salesmen sm ON sm.id = p.salesman_id").
		Joins("LEFT JOIN users u ON u.id = p.created_by").
		Where("p.tenant_id = ?", tenantID)
	if keyword != "" {
		kw := "%" + keyword + "%"
		q = q.Where("p.order_no LIKE ? OR c.name LIKE ?", kw, kw)
	}
	if dateFrom != "" {
		q = q.Where("p.bill_date >= ?", dateFrom)
	}
	if dateTo != "" {
		q = q.Where("p.bill_date <= ?", dateTo)
	}
	q.Session(&gorm.Session{}).Count(&total)
	q.Select("p.*, c.name AS customer_name, sm.name AS salesman_name, u.nickname AS maker_name").
		Offset((page - 1) * pageSize).Limit(pageSize).Order("p.created_at DESC").Scan(&list)
	for i := range list {
		list[i].IDStr = strconv.FormatInt(list[i].ID, 10)
	}
	response.OKPage(c, list, total, page, pageSize)
}

func (h *SalesHandler) GetQuote(c *gin.Context) {
	ctx := c.Request.Context()
	tenantID := context.GetTenantID(ctx)
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var q model.Quote
	if err := h.db.Table("quotes").Where("id = ? AND tenant_id = ?", id, tenantID).First(&q).Error; err != nil {
		response.NotFound(c, "报价单不存在")
		return
	}
	q.IDStr = strconv.FormatInt(q.ID, 10)
	q.CustomerName, q.SalesmanName, _, q.MakerName = h.names(tenantID, q.CustomerID, q.SalesmanID, 0, q.CreatedBy)
	var items []model.QuoteItem
	h.db.Table("quote_items").Where("quote_id = ?", q.ID).Order("id ASC").Find(&items)
	gm := h.goodsMap(tenantID, quoteItemGoodsIDs(items))
	for i := range items {
		g := gm[items[i].GoodsID]
		items[i].GoodsName = g.Name
		items[i].GoodsCode = g.Code
		items[i].UnitName = g.MainUnit
		items[i].Spec = orderSpecNames(g.SpecGroups)
		items[i].Barcode = g.Barcode
		items[i].ImageURL = g.ImageURL
		items[i].Brand = g.Brand
		items[i].Origin = g.Origin
	}
	q.Items = items
	response.OK(c, q)
}

func (h *SalesHandler) CreateQuote(c *gin.Context) {
	ctx := c.Request.Context()
	tenantID := context.GetTenantID(ctx)
	userID := context.GetUserID(ctx)
	var req saleCreateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	if req.CustomerID == 0 || len(req.Items) == 0 {
		response.BadRequest(c, "请选择客户并添加明细")
		return
	}
	items, total := calcItems(req.Items)
	q := model.Quote{
		ID: snowflake.GenID(), TenantID: tenantID, ShopID: req.ShopID,
		OrderNo: nextNo(h.db, tenantID, "quotes", "BJ"), CustomerID: req.CustomerID,
		SalesmanID: req.SalesmanID, BillDate: req.BillDate, TotalAmount: total,
		Remark: req.Remark, CreatedBy: userID,
	}
	if err := h.db.Table("quotes").Create(&q).Error; err != nil {
		response.ServerError(c, err.Error())
		return
	}
	qitems := make([]model.QuoteItem, 0, len(items))
	for _, it := range items {
		qitems = append(qitems, model.QuoteItem{ID: snowflake.GenID(), TenantID: tenantID, QuoteID: q.ID, GoodsID: it.GoodsID, Quantity: it.Quantity, UnitPrice: it.UnitPrice, Amount: it.Amount, Remark: it.Remark})
	}
	if len(qitems) > 0 {
		h.db.Table("quote_items").Create(&qitems)
	}
	q.IDStr = strconv.FormatInt(q.ID, 10)
	response.OK(c, q)
}

func (h *SalesHandler) DeleteQuote(c *gin.Context) {
	ctx := c.Request.Context()
	tenantID := context.GetTenantID(ctx)
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	h.db.Table("quote_items").Where("quote_id = ?", id).Delete(&model.QuoteItem{})
	h.db.Table("quotes").Where("id = ? AND tenant_id = ?", id, tenantID).Delete(&model.Quote{})
	response.OKMsg(c, "删除成功")
}

/* ── item id helpers ── */

func itemGoodsIDs(items []model.SaleItem) []int64 {
	ids := make([]int64, 0, len(items))
	for _, it := range items {
		ids = append(ids, it.GoodsID)
	}
	return ids
}
func orderItemGoodsIDs(items []model.SaleOrderItem) []int64 {
	ids := make([]int64, 0, len(items))
	for _, it := range items {
		ids = append(ids, it.GoodsID)
	}
	return ids
}
func returnItemGoodsIDs(items []model.SalesReturnItem) []int64 {
	ids := make([]int64, 0, len(items))
	for _, it := range items {
		ids = append(ids, it.GoodsID)
	}
	return ids
}
func quoteItemGoodsIDs(items []model.QuoteItem) []int64 {
	ids := make([]int64, 0, len(items))
	for _, it := range items {
		ids = append(ids, it.GoodsID)
	}
	return ids
}
