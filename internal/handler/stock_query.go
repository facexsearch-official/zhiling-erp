package handler

import (
	"sort"
	"strconv"
	"time"

	"pisa_server/internal/model"
	"pisa_server/internal/pkg/context"
	"pisa_server/internal/pkg/response"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type StockQueryHandler struct{ db *gorm.DB }

func NewStockQueryHandler(db *gorm.DB) *StockQueryHandler { return &StockQueryHandler{db: db} }

type stockQueryRow struct {
	ID           int64   `json:"id"`
	IDStr        string  `json:"id_str"`
	Name         string  `json:"name"`
	Code         string  `json:"code"`
	Barcode      string  `json:"barcode"`
	ImageURL     string  `json:"image_url"`
	Spec         string  `json:"spec"`
	MainUnit     string  `json:"main_unit"`
	CategoryName string  `json:"category_name"`
	CurrentStock int     `json:"current_stock"`
	Cost         float64 `json:"cost"`
	StockValue   float64 `json:"stock_value"`
}

func (h *StockQueryHandler) applyFilters(q *gorm.DB, tenantID int64, c *gin.Context) *gorm.DB {
	if kw := c.Query("keyword"); kw != "" {
		like := "%" + kw + "%"
		q = q.Where("g.name LIKE ? OR g.barcode LIKE ? OR g.code LIKE ?", like, like, like)
	}
	if v := c.Query("category_id"); v != "" && v != "0" {
		q = q.Where("g.category_id = ?", v)
	}
	if v := c.Query("supplier_id"); v != "" && v != "0" {
		q = q.Where("g.supplier_id = ?", v)
	}
	if c.Query("hide_disabled") == "1" {
		q = q.Where("g.status = 1")
	}
	if c.Query("hide_zero") == "1" {
		q = q.Where("g.current_stock <> 0")
	}
	return q
}

func (h *StockQueryHandler) List(c *gin.Context) {
	ctx := c.Request.Context()
	tenantID := context.GetTenantID(ctx)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	var total int64
	base := h.applyFilters(h.db.Table("goods AS g").Where("g.tenant_id = ?", tenantID), tenantID, c)
	base.Session(&gorm.Session{}).Count(&total)

	var rows []struct {
		ID           int64
		Name         string
		Code         string
		Barcode      string
		ImageURL     string
		SpecGroups   string
		MainUnit     string
		CategoryName string
		CurrentStock int
		PurchaseP    float64
	}
	base.Select("g.id, g.name, g.code, g.barcode, g.image_url, g.spec_groups, g.main_unit, g.current_stock, g.purchase_price AS purchase_p, COALESCE(gc.name,'') AS category_name").
		Joins("LEFT JOIN goods_categories gc ON gc.id = g.category_id").
		Offset((page - 1) * pageSize).Limit(pageSize).Order("g.id DESC").Scan(&rows)

	list := make([]stockQueryRow, 0, len(rows))
	for _, r := range rows {
		list = append(list, stockQueryRow{
			ID: r.ID, IDStr: strconv.FormatInt(r.ID, 10), Name: r.Name, Code: r.Code, Barcode: r.Barcode,
			ImageURL: r.ImageURL, Spec: orderSpecNames(r.SpecGroups), MainUnit: r.MainUnit,
			CategoryName: r.CategoryName, CurrentStock: r.CurrentStock, Cost: r.PurchaseP,
			StockValue: round2o(float64(r.CurrentStock) * r.PurchaseP),
		})
	}

	var agg struct {
		Stock int
		Value float64
	}
	h.applyFilters(h.db.Table("goods AS g").Where("g.tenant_id = ?", tenantID), tenantID, c).
		Select("COALESCE(SUM(g.current_stock),0) AS stock, COALESCE(SUM(g.current_stock*g.purchase_price),0) AS value").Scan(&agg)

	response.OK(c, gin.H{
		"list": list, "total": total, "page": page, "page_size": pageSize,
		"summary": gin.H{"total_stock": agg.Stock, "total_value": round2o(agg.Value)},
	})
}

type stockMove struct {
	Date     string  `json:"date"`
	Party    string  `json:"party"`
	OrderNo  string  `json:"order_no"`
	DocType  string  `json:"doc_type"`
	In       float64 `json:"in"`
	Out      float64 `json:"out"`
	UnitCost float64 `json:"unit_cost"`
	seq      time.Time
}

func collectStockMoves(db *gorm.DB, tenantID, goodsID int64) []stockMove {
	moves := []stockMove{}

	// 进货单（增加）
	var pur []struct {
		BillDate  string
		OrderNo   string
		Party     string
		Qty       int
		Price     float64
		CreatedAt time.Time
	}
	db.Raw(`SELECT p.bill_date, p.order_no, COALESCE(s.name,'') AS party, i.quantity AS qty, i.unit_price AS price, p.created_at
		FROM purchases p JOIN purchase_items i ON i.purchase_id = p.id LEFT JOIN suppliers s ON s.id = p.supplier_id
		WHERE p.tenant_id = ? AND i.goods_id = ? AND p.status <> 4`, tenantID, goodsID).Scan(&pur)
	for _, r := range pur {
		moves = append(moves, stockMove{Date: r.BillDate, Party: r.Party, OrderNo: r.OrderNo, DocType: "进货单", In: float64(r.Qty), UnitCost: r.Price, seq: r.CreatedAt})
	}

	// 进货退货（减少）
	var pret []struct {
		BillDate  string
		OrderNo   string
		Party     string
		Qty       int
		Price     float64
		CreatedAt time.Time
	}
	db.Raw(`SELECT p.bill_date, p.order_no, COALESCE(s.name,'') AS party, i.quantity AS qty, i.unit_price AS price, p.created_at
		FROM purchase_returns p JOIN purchase_return_items i ON i.return_id = p.id LEFT JOIN suppliers s ON s.id = p.supplier_id
		WHERE p.tenant_id = ? AND i.goods_id = ?`, tenantID, goodsID).Scan(&pret)
	for _, r := range pret {
		moves = append(moves, stockMove{Date: r.BillDate, Party: r.Party, OrderNo: r.OrderNo, DocType: "进货退货单", Out: float64(r.Qty), UnitCost: r.Price, seq: r.CreatedAt})
	}

	// 销售单（减少）
	var sal []struct {
		BillDate  string
		OrderNo   string
		Party     string
		Qty       int
		Price     float64
		CreatedAt time.Time
	}
	db.Raw(`SELECT p.bill_date, p.order_no, COALESCE(c.name,'') AS party, i.quantity AS qty, i.unit_price AS price, p.created_at
		FROM sales p JOIN sale_items i ON i.sale_id = p.id LEFT JOIN customers c ON c.id = p.customer_id
		WHERE p.tenant_id = ? AND i.goods_id = ?`, tenantID, goodsID).Scan(&sal)
	for _, r := range sal {
		moves = append(moves, stockMove{Date: r.BillDate, Party: r.Party, OrderNo: r.OrderNo, DocType: "销售单", Out: float64(r.Qty), UnitCost: r.Price, seq: r.CreatedAt})
	}

	// 销售退货（增加）
	var sret []struct {
		BillDate  string
		OrderNo   string
		Party     string
		Qty       int
		Price     float64
		CreatedAt time.Time
	}
	db.Raw(`SELECT p.bill_date, p.order_no, COALESCE(c.name,'') AS party, i.quantity AS qty, i.unit_price AS price, p.created_at
		FROM sales_returns p JOIN sales_return_items i ON i.return_id = p.id LEFT JOIN customers c ON c.id = p.customer_id
		WHERE p.tenant_id = ? AND i.goods_id = ?`, tenantID, goodsID).Scan(&sret)
	for _, r := range sret {
		moves = append(moves, stockMove{Date: r.BillDate, Party: r.Party, OrderNo: r.OrderNo, DocType: "销售退货单", In: float64(r.Qty), UnitCost: r.Price, seq: r.CreatedAt})
	}

	// 组装/拆分
	var asm []struct {
		BillDate  string
		OrderNo   string
		Type      int8
		Kind      int8
		Qty       float64
		Cost      float64
		CreatedAt time.Time
	}
	db.Raw(`SELECT a.bill_date, a.order_no, a.type, i.kind, i.quantity AS qty, i.unit_cost AS cost, a.created_at
		FROM assemblies a JOIN assembly_items i ON i.assembly_id = a.id
		WHERE a.tenant_id = ? AND i.goods_id = ? AND a.status <> 9`, tenantID, goodsID).Scan(&asm)
	for _, r := range asm {
		m := stockMove{Date: r.BillDate, OrderNo: r.OrderNo, UnitCost: r.Cost, seq: r.CreatedAt}
		if r.Type == 2 { // 拆分
			m.DocType = "拆分单"
			if r.Kind == 1 {
				m.Out = r.Qty
			} else {
				m.In = r.Qty
			}
		} else { // 组装
			m.DocType = "组装单"
			if r.Kind == 1 {
				m.In = r.Qty
			} else {
				m.Out = r.Qty
			}
		}
		moves = append(moves, m)
	}

	// 盘点单
	var cnt []struct {
		BillDate  string
		OrderNo   string
		Diff      int
		Cost      float64
		CreatedAt time.Time
	}
	db.Raw(`SELECT p.bill_date, p.order_no, i.diff_qty AS diff, i.unit_cost AS cost, p.created_at
		FROM stock_counts p JOIN stock_count_items i ON i.count_id = p.id
		WHERE p.tenant_id = ? AND i.goods_id = ? AND p.status <> 9`, tenantID, goodsID).Scan(&cnt)
	for _, r := range cnt {
		if r.Diff == 0 {
			continue
		}
		m := stockMove{Date: r.BillDate, OrderNo: r.OrderNo, DocType: "盘点单", UnitCost: r.Cost, seq: r.CreatedAt}
		if r.Diff > 0 {
			m.In = float64(r.Diff)
		} else {
			m.Out = float64(-r.Diff)
		}
		moves = append(moves, m)
	}

	sort.SliceStable(moves, func(i, j int) bool {
		if moves[i].Date == moves[j].Date {
			return moves[i].seq.Before(moves[j].seq)
		}
		return moves[i].Date < moves[j].Date
	})
	return moves
}

func (h *StockQueryHandler) loadGoods(tenantID, goodsID int64) model.Goods {
	var g model.Goods
	h.db.Where("id = ? AND tenant_id = ?", goodsID, tenantID).First(&g)
	return g
}

type flowRow struct {
	Seq     int     `json:"seq"`
	Date    string  `json:"date"`
	Party   string  `json:"party"`
	OrderNo string  `json:"order_no"`
	DocType string  `json:"doc_type"`
	In      float64 `json:"in"`
	Out     float64 `json:"out"`
	Balance float64 `json:"balance"`
}

func (h *StockQueryHandler) Flow(c *gin.Context) {
	ctx := c.Request.Context()
	tenantID := context.GetTenantID(ctx)
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	g := h.loadGoods(tenantID, id)
	moves := collectStockMoves(h.db, tenantID, id)

	var net float64
	for _, m := range moves {
		net += m.In - m.Out
	}
	opening := float64(g.CurrentStock) - net

	bal := opening
	type item struct {
		flowRow
	}
	asc := make([]flowRow, 0, len(moves))
	for _, m := range moves {
		bal += m.In - m.Out
		asc = append(asc, flowRow{Date: m.Date, Party: m.Party, OrderNo: m.OrderNo, DocType: m.DocType, In: m.In, Out: m.Out, Balance: bal})
	}
	// 倒序 + 期初
	rows := make([]flowRow, 0, len(asc)+1)
	for i := len(asc) - 1; i >= 0; i-- {
		r := asc[i]
		r.Seq = len(rows) + 1
		rows = append(rows, r)
	}
	rows = append(rows, flowRow{Seq: len(rows) + 1, DocType: "期初", Balance: opening})

	response.OK(c, gin.H{"goods": goodsBrief(g), "rows": rows})
}

func goodsBrief(g model.Goods) gin.H {
	return gin.H{
		"id": g.ID, "id_str": strconv.FormatInt(g.ID, 10), "name": g.Name, "code": g.Code,
		"barcode": g.Barcode, "image_url": g.ImageURL, "spec": orderSpecNames(g.SpecGroups), "main_unit": g.MainUnit,
	}
}

type costRow struct {
	Seq       int     `json:"seq"`
	Date      string  `json:"date"`
	Party     string  `json:"party"`
	OrderNo   string  `json:"order_no"`
	DocType   string  `json:"doc_type"`
	InQty     float64 `json:"in_qty"`
	InCost    float64 `json:"in_cost"`
	InAmount  float64 `json:"in_amount"`
	OutQty    float64 `json:"out_qty"`
	OutCost   float64 `json:"out_cost"`
	OutAmount float64 `json:"out_amount"`
	BalQty    float64 `json:"bal_qty"`
	BalCost   float64 `json:"bal_cost"`
	BalAmount float64 `json:"bal_amount"`
}

func (h *StockQueryHandler) Cost(c *gin.Context) {
	ctx := c.Request.Context()
	tenantID := context.GetTenantID(ctx)
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	g := h.loadGoods(tenantID, id)
	moves := collectStockMoves(h.db, tenantID, id)

	var net float64
	for _, m := range moves {
		net += m.In - m.Out
	}
	openingQty := float64(g.CurrentStock) - net
	openingCost := g.PurchasePrice

	balQty := openingQty
	balCost := openingCost
	asc := make([]costRow, 0, len(moves))
	for _, m := range moves {
		r := costRow{Date: m.Date, Party: m.Party, OrderNo: m.OrderNo, DocType: m.DocType}
		if m.In > 0 {
			inCost := m.UnitCost
			if inCost == 0 {
				inCost = balCost
			}
			r.InQty = m.In
			r.InCost = round2o(inCost)
			r.InAmount = round2o(m.In * inCost)
			newQty := balQty + m.In
			if newQty > 0 {
				balCost = (balQty*balCost + m.In*inCost) / newQty
			}
			balQty = newQty
		}
		if m.Out > 0 {
			r.OutQty = m.Out
			r.OutCost = round2o(balCost)
			r.OutAmount = round2o(m.Out * balCost)
			balQty -= m.Out
		}
		r.BalQty = balQty
		r.BalCost = round2o(balCost)
		r.BalAmount = round2o(balQty * balCost)
		asc = append(asc, r)
	}
	rows := make([]costRow, 0, len(asc)+1)
	rows = append(rows, costRow{Seq: 1, DocType: "期初", BalQty: openingQty, BalCost: round2o(openingCost), BalAmount: round2o(openingQty * openingCost)})
	for i := 0; i < len(asc); i++ {
		r := asc[i]
		r.Seq = len(rows) + 1
		rows = append(rows, r)
	}
	response.OK(c, gin.H{"goods": goodsBrief(g), "rows": rows})
}

type alertRow struct {
	Seq         int     `json:"seq"`
	GoodsID     int64   `json:"goods_id"`
	IDStr       string  `json:"id_str"`
	Name        string  `json:"name"`
	Spec        string  `json:"spec"`
	Code        string  `json:"code"`
	Warehouse   string  `json:"warehouse"`
	Unit        string  `json:"unit"`
	MinStock    int     `json:"min_stock"`
	SafetyStock int     `json:"safety_stock"`
	MaxStock    int     `json:"max_stock"`
	Stock       int     `json:"stock"`
	Reorder     int     `json:"reorder"`
	Amount      float64 `json:"amount"`
}

func (h *StockQueryHandler) Alert(c *gin.Context) {
	ctx := c.Request.Context()
	tenantID := context.GetTenantID(ctx)
	keyword := c.Query("keyword")
	categoryID := c.Query("category_id")
	supplierID := c.Query("supplier_id")
	status := c.Query("status")
	base := c.DefaultQuery("base", "safety")

	var shopName string
	var shop model.Shop
	if h.db.Where("tenant_id = ? AND is_main = 1", tenantID).First(&shop).Error == nil {
		shopName = shop.Name
	}

	type raw struct {
		GoodsID     int64
		Name        string
		Code        string
		SpecKey     string
		MainUnit    string
		Stock       int
		MinStock    int
		SafetyStock int
		MaxStock    int
		Cost        float64
	}
	q := h.db.Table("goods_stocks AS gs").
		Joins("JOIN goods g ON g.id = gs.goods_id").
		Where("g.tenant_id = ? AND g.enable_stock_alert = 1", tenantID)
	if keyword != "" {
		kw := "%" + keyword + "%"
		q = q.Where("g.name LIKE ? OR g.code LIKE ? OR g.barcode LIKE ?", kw, kw, kw)
	}
	if categoryID != "" && categoryID != "0" {
		q = q.Where("g.category_id = ?", categoryID)
	}
	if supplierID != "" && supplierID != "0" {
		q = q.Where("g.supplier_id = ?", supplierID)
	}
	var rows []raw
	q.Select("gs.goods_id, g.name, g.code, gs.spec_key, g.main_unit, gs.stock, gs.min_stock, gs.safety_stock, gs.max_stock, g.purchase_price AS cost").Scan(&rows)

	list := make([]alertRow, 0)
	for _, r := range rows {
		baseVal := r.SafetyStock
		if base == "min" {
			baseVal = r.MinStock
		} else if base == "max" {
			baseVal = r.MaxStock
		}
		reorder := baseVal - r.Stock
		if reorder < 0 {
			reorder = 0
		}
		if reorder == 0 {
			continue
		}
		if status == "缺货" && r.Stock > 0 {
			continue
		}
		if status == "预警" && r.Stock <= 0 {
			continue
		}
		spec := r.SpecKey
		if spec == "" {
			spec = "--"
		}
		list = append(list, alertRow{
			Seq: len(list) + 1, GoodsID: r.GoodsID, IDStr: strconv.FormatInt(r.GoodsID, 10),
			Name: r.Name, Spec: spec, Code: r.Code, Warehouse: shopName, Unit: r.MainUnit,
			MinStock: r.MinStock, SafetyStock: r.SafetyStock, MaxStock: r.MaxStock, Stock: r.Stock,
			Reorder: reorder, Amount: round2o(float64(reorder) * r.Cost),
		})
	}
	response.OK(c, gin.H{"list": list, "total": len(list)})
}
