package handler

import (
	"strconv"

	"pisa_server/internal/model"
	"pisa_server/internal/pkg/context"
	"pisa_server/internal/pkg/response"
	"pisa_server/internal/pkg/snowflake"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type StockCountHandler struct {
	db *gorm.DB
}

func NewStockCountHandler(dbConn *gorm.DB) *StockCountHandler {
	return &StockCountHandler{db: dbConn}
}

type stockCountItemReq struct {
	GoodsID   model.FlexInt64 `json:"goods_id"`
	BookQty   int    `json:"book_qty"`
	ActualQty int    `json:"actual_qty"`
	DiffQty   int    `json:"diff_qty"`
	Remark    string `json:"remark"`
}

type stockCountCreateReq struct {
	ShopID      int64               `json:"shop_id"`
	WarehouseID int64               `json:"warehouse_id"`
	SalesmanID  int64               `json:"salesman_id"`
	BillDate    string              `json:"bill_date"`
	Attachments string              `json:"attachments"`
	Remark      string              `json:"remark"`
	Items       []stockCountItemReq `json:"items"`
}

func (h *StockCountHandler) List(c *gin.Context) {
	ctx := c.Request.Context()
	tenantID := context.GetTenantID(ctx)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	keyword := c.Query("keyword")
	dateFrom := c.Query("date_from")
	dateTo := c.Query("date_to")
	hideVoid := c.Query("hide_void") == "1"
	hideZero := c.Query("hide_zero") == "1"

	var total int64
	var list []model.StockCount
	q := h.db.Table("stock_counts AS p").
		Joins("LEFT JOIN warehouses w ON w.id = p.warehouse_id").
		Joins("LEFT JOIN salesmen sm ON sm.id = p.salesman_id").
		Joins("LEFT JOIN users su ON su.id = p.salesman_id").
		Joins("LEFT JOIN users u ON u.id = p.created_by").
		Where("p.tenant_id = ?", tenantID)
	if keyword != "" {
		kw := "%" + keyword + "%"
		q = q.Where("p.order_no LIKE ? OR p.remark LIKE ?", kw, kw)
	}
	if dateFrom != "" {
		q = q.Where("p.bill_date >= ?", dateFrom)
	}
	if dateTo != "" {
		q = q.Where("p.bill_date <= ?", dateTo)
	}
	if hideVoid {
		q = q.Where("p.status = 1")
	}
	if hideZero {
		q = q.Where("p.diff_qty <> 0")
	}
	q.Session(&gorm.Session{}).Count(&total)
	q.Select("p.*, w.name AS warehouse_name, COALESCE(sm.name, su.nickname) AS salesman_name, u.nickname AS maker_name").
		Offset((page - 1) * pageSize).Limit(pageSize).Order("p.created_at DESC").Scan(&list)
	for i := range list {
		list[i].IDStr = strconv.FormatInt(list[i].ID, 10)
	}
	response.OKPage(c, list, total, page, pageSize)
}

func (h *StockCountHandler) GetByID(c *gin.Context) {
	ctx := c.Request.Context()
	tenantID := context.GetTenantID(ctx)
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var sc model.StockCount
	if err := h.db.Table("stock_counts").Where("id = ? AND tenant_id = ?", id, tenantID).First(&sc).Error; err != nil {
		response.NotFound(c, "盘点单不存在")
		return
	}
	sc.IDStr = strconv.FormatInt(sc.ID, 10)
	if sc.WarehouseID != 0 {
		var w model.Warehouse
		if h.db.Where("id = ?", sc.WarehouseID).First(&w).Error == nil {
			sc.WarehouseName = w.Name
		}
	}
	if sc.SalesmanID != 0 {
		var sm model.Salesman
		if h.db.Where("id = ?", sc.SalesmanID).First(&sm).Error == nil {
			sc.SalesmanName = sm.Name
		} else {
			var su model.User
			if h.db.Where("id = ?", sc.SalesmanID).First(&su).Error == nil {
				sc.SalesmanName = su.Nickname
			}
		}
	}
	if sc.CreatedBy != 0 {
		var u model.User
		if h.db.Where("id = ?", sc.CreatedBy).First(&u).Error == nil {
			sc.MakerName = u.Nickname
		}
	}
	var items []model.StockCountItem
	h.db.Table("stock_count_items").Where("count_id = ?", sc.ID).Order("id ASC").Find(&items)
	var ids []int64
	for _, it := range items {
		ids = append(ids, it.GoodsID)
	}
	gm := map[int64]model.Goods{}
	if len(ids) > 0 {
		var goods []model.Goods
		h.db.Where("tenant_id = ? AND id IN ?", tenantID, ids).Find(&goods)
		for _, g := range goods {
			gm[g.ID] = g
		}
	}
	for i := range items {
		g := gm[items[i].GoodsID]
		items[i].GoodsName = g.Name
		items[i].GoodsCode = g.Code
		items[i].UnitName = g.MainUnit
		items[i].Spec = orderSpecNames(g.SpecGroups)
		items[i].Barcode = g.Barcode
		items[i].ImageURL = g.ImageURL
	}
	sc.Items = items
	response.OK(c, sc)
}

func (h *StockCountHandler) Create(c *gin.Context) {
	ctx := c.Request.Context()
	tenantID := context.GetTenantID(ctx)
	userID := context.GetUserID(ctx)
	var req stockCountCreateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	if len(req.Items) == 0 {
		response.BadRequest(c, "请添加盘点明细")
		return
	}
	var tb, ta, td int
	items := make([]model.StockCountItem, 0, len(req.Items))
	for _, it := range req.Items {
		if it.GoodsID == 0 {
			continue
		}
		diff := it.ActualQty - it.BookQty
		tb += it.BookQty
		ta += it.ActualQty
		td += diff
		items = append(items, model.StockCountItem{
			ID: snowflake.GenID(), TenantID: tenantID, GoodsID: int64(it.GoodsID),
			BookQty: it.BookQty, ActualQty: it.ActualQty, DiffQty: diff, Remark: it.Remark,
		})
	}
	sc := model.StockCount{
		ID: snowflake.GenID(), TenantID: tenantID, ShopID: req.ShopID, WarehouseID: req.WarehouseID,
		OrderNo: nextNo(h.db, tenantID, "stock_counts", "PDD"), SalesmanID: req.SalesmanID,
		BillDate: req.BillDate, BookQty: tb, ActualQty: ta, DiffQty: td,
		Attachments: req.Attachments, Status: 1, Remark: req.Remark, CreatedBy: userID,
	}
	if err := h.db.Table("stock_counts").Create(&sc).Error; err != nil {
		response.ServerError(c, err.Error())
		return
	}
	for i := range items {
		items[i].CountID = sc.ID
	}
	if len(items) > 0 {
		h.db.Table("stock_count_items").Create(&items)
	}
	sc.IDStr = strconv.FormatInt(sc.ID, 10)
	sc.Items = items
	response.OK(c, sc)
}

func (h *StockCountHandler) Void(c *gin.Context) {
	ctx := c.Request.Context()
	tenantID := context.GetTenantID(ctx)
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	h.db.Table("stock_counts").Where("id = ? AND tenant_id = ?", id, tenantID).Update("status", 9)
	response.OKMsg(c, "作废成功")
}
