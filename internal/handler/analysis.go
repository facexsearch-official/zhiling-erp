package handler

import (
	"strconv"

	"pisa_server/internal/model"
	"pisa_server/internal/pkg/context"
	"pisa_server/internal/pkg/response"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type AnalysisHandler struct{ db *gorm.DB }

func NewAnalysisHandler(db *gorm.DB) *AnalysisHandler { return &AnalysisHandler{db: db} }

/* ─── 热销分析 ─── */
func (h *AnalysisHandler) HotSales(c *gin.Context) {
	ctx := c.Request.Context()
	tenantID := context.GetTenantID(ctx)

	dateCond := ""
	args := []interface{}{}
	if v := c.Query("date_from"); v != "" {
		dateCond += " AND s.bill_date >= ?"
		args = append(args, v)
	}
	if v := c.Query("date_to"); v != "" {
		dateCond += " AND s.bill_date <= ?"
		args = append(args, v)
	}

	type raw struct {
		GoodsID      int64
		Name         string
		SpecGroups   string
		MainUnit     string
		CategoryName string
		ImageURL     string
		CurrentStock int
		Qty          int
	}
	sub := "COALESCE((SELECT SUM(si.quantity) FROM sale_items si JOIN sales s ON s.id = si.sale_id WHERE si.goods_id = g.id AND s.status <> 9" + dateCond + "),0)"
	q := h.db.Table("goods AS g").
		Joins("LEFT JOIN goods_categories gc ON gc.id = g.category_id").
		Where("g.tenant_id = ?", tenantID)
	var rows []raw
	q.Select("g.id AS goods_id, g.name, g.spec_groups, g.main_unit, g.image_url, g.current_stock, COALESCE(gc.name,'') AS category_name, "+sub+" AS qty", args...).
		Having("qty > 0").Order("qty DESC").Scan(&rows)

	type outRow struct {
		IDStr        string `json:"id_str"`
		Name         string `json:"name"`
		Spec         string `json:"spec"`
		CategoryName string `json:"category_name"`
		ImageURL     string `json:"image_url"`
		Unit         string `json:"unit"`
		CurrentStock int    `json:"current_stock"`
		Qty          int    `json:"qty"`
	}
	list := make([]outRow, 0, len(rows))
	var total int
	for _, r := range rows {
		list = append(list, outRow{
			IDStr: strconv.FormatInt(r.GoodsID, 10), Name: r.Name, Spec: orderSpecNames(r.SpecGroups),
			CategoryName: r.CategoryName, ImageURL: r.ImageURL, Unit: r.MainUnit, CurrentStock: r.CurrentStock, Qty: r.Qty,
		})
		total += r.Qty
	}
	response.OK(c, gin.H{"list": list, "summary": gin.H{"qty": total}})
}

/* ─── 员工业绩统计 ─── */
func (h *AnalysisHandler) StaffPerf(c *gin.Context) {
	ctx := c.Request.Context()
	tenantID := context.GetTenantID(ctx)
	dateCond := ""
	args := []interface{}{tenantID}
	if v := c.Query("date_from"); v != "" {
		dateCond += " AND s.bill_date >= ?"
		args = append(args, v)
	}
	if v := c.Query("date_to"); v != "" {
		dateCond += " AND s.bill_date <= ?"
		args = append(args, v)
	}
	qtySub := "(SELECT COALESCE(SUM(si.quantity),0) FROM sale_items si WHERE si.sale_id = s.id)"
	costSub := "(SELECT COALESCE(SUM(si.quantity * g.purchase_price),0) FROM sale_items si JOIN goods g ON g.id = si.goods_id WHERE si.sale_id = s.id)"

	type raw struct {
		SalesmanID int64
		Name       string
		Qty        int
		Amount     float64
		Discount   float64
		Freight    float64
		Receivable float64
		Cost       float64
	}
	var rows []raw
	h.db.Table("sales AS s").
		Joins("LEFT JOIN salesmen sm ON sm.id = s.salesman_id").
		Where("s.tenant_id = ? AND s.status <> 9"+dateCond, args...).
		Select("s.salesman_id, COALESCE(sm.name,'未指定') AS name, COALESCE(SUM(" + qtySub + "),0) AS qty, SUM(s.subtotal) AS amount, SUM(s.subtotal - s.subtotal*s.discount/100) AS discount, SUM(s.freight) AS freight, SUM(s.total_amount) AS receivable, SUM(" + costSub + ") AS cost").
		Group("s.salesman_id, sm.name").Order("amount DESC").Scan(&rows)

	type outRow struct {
		IDStr      string  `json:"id_str"`
		Name       string  `json:"name"`
		Qty        int     `json:"qty"`
		Amount     float64 `json:"amount"`
		Discount   float64 `json:"discount"`
		DiscAmount float64 `json:"disc_amount"`
		Freight    float64 `json:"freight"`
		Receivable float64 `json:"receivable"`
		Profit     float64 `json:"profit"`
		CostMargin float64 `json:"cost_margin"`
	}
	list := make([]outRow, 0, len(rows))
	var tQty int
	var tAmount, tDiscount, tFreight, tReceivable, tProfit float64
	for _, r := range rows {
		profit := round2o(r.Amount - r.Cost)
		discAmt := round2o(r.Discount)
		cm := 0.0
		if r.Cost > 0 {
			cm = round2o(profit / r.Cost * 100)
		}
		list = append(list, outRow{
			IDStr: strconv.FormatInt(r.SalesmanID, 10), Name: r.Name, Qty: r.Qty, Amount: round2o(r.Amount),
			Discount: 0, DiscAmount: discAmt, Freight: round2o(r.Freight), Receivable: round2o(r.Receivable),
			Profit: profit, CostMargin: cm,
		})
		tQty += r.Qty
		tAmount += r.Amount
		tDiscount += r.Discount
		tFreight += r.Freight
		tReceivable += r.Receivable
		tProfit += profit
	}
	response.OK(c, gin.H{"list": list, "summary": gin.H{
		"qty": tQty, "amount": round2o(tAmount), "disc_amount": round2o(tDiscount),
		"freight": round2o(tFreight), "receivable": round2o(tReceivable), "profit": round2o(tProfit),
	}})
}

/* ─── 进货统计 ─── */
func (h *AnalysisHandler) PurchaseByGoods(c *gin.Context) {
	ctx := c.Request.Context()
	tenantID := context.GetTenantID(ctx)
	q := h.db.Table("purchase_items AS pi").
		Joins("JOIN purchases p ON p.id = pi.purchase_id").
		Joins("JOIN goods g ON g.id = pi.goods_id").
		Joins("LEFT JOIN goods_categories gc ON gc.id = g.category_id").
		Where("p.tenant_id = ? AND p.status <> 4", tenantID)
	if v := c.Query("date_from"); v != "" {
		q = q.Where("p.bill_date >= ?", v)
	}
	if v := c.Query("date_to"); v != "" {
		q = q.Where("p.bill_date <= ?", v)
	}
	if v := c.Query("category_id"); v != "" && v != "0" {
		q = q.Where("g.category_id = ?", v)
	}
	if v := c.Query("keyword"); v != "" {
		like := "%" + v + "%"
		q = q.Where("g.name LIKE ? OR g.code LIKE ? OR g.barcode LIKE ?", like, like, like)
	}
	type raw struct {
		GoodsID      int64
		Name         string
		SpecGroups   string
		Code         string
		MainUnit     string
		CategoryName string
		Qty          int
		Amount       float64
	}
	var rows []raw
	q.Select("g.id AS goods_id, g.name, g.spec_groups, g.code, g.main_unit, COALESCE(gc.name,'') AS category_name, SUM(pi.quantity) AS qty, SUM(pi.amount) AS amount").
		Group("g.id, g.name, g.spec_groups, g.code, g.main_unit, gc.name").Order("amount DESC").Scan(&rows)
	type outRow struct {
		IDStr        string  `json:"id_str"`
		Name         string  `json:"name"`
		Spec         string  `json:"spec"`
		Code         string  `json:"code"`
		CategoryName string  `json:"category_name"`
		Unit         string  `json:"unit"`
		Amount       float64 `json:"amount"`
		Qty          int     `json:"qty"`
	}
	list := make([]outRow, 0, len(rows))
	var tAmount float64
	var tQty int
	for _, r := range rows {
		list = append(list, outRow{IDStr: strconv.FormatInt(r.GoodsID, 10), Name: r.Name, Spec: orderSpecNames(r.SpecGroups), Code: r.Code, CategoryName: r.CategoryName, Unit: r.MainUnit, Amount: round2o(r.Amount), Qty: r.Qty})
		tAmount += r.Amount
		tQty += r.Qty
	}
	response.OK(c, gin.H{"list": list, "summary": gin.H{"amount": round2o(tAmount), "qty": tQty}})
}

func (h *AnalysisHandler) PurchaseBySupplier(c *gin.Context) {
	ctx := c.Request.Context()
	tenantID := context.GetTenantID(ctx)
	q := h.db.Table("purchases AS p").
		Joins("LEFT JOIN suppliers s ON s.id = p.supplier_id").
		Where("p.tenant_id = ? AND p.status <> 4", tenantID)
	if v := c.Query("date_from"); v != "" {
		q = q.Where("p.bill_date >= ?", v)
	}
	if v := c.Query("date_to"); v != "" {
		q = q.Where("p.bill_date <= ?", v)
	}
	if v := c.Query("supplier_id"); v != "" && v != "0" {
		q = q.Where("p.supplier_id = ?", v)
	}
	if v := c.Query("salesman_id"); v != "" && v != "0" {
		q = q.Where("p.salesman_id = ?", v)
	}
	type raw struct {
		SupplierID int64
		Name       string
		Amount     float64
		Cnt        int
	}
	var rows []raw
	q.Select("p.supplier_id, COALESCE(s.name,'未指定') AS name, SUM(p.total_amount) AS amount, COUNT(*) AS cnt").
		Group("p.supplier_id, s.name").Order("amount DESC").Scan(&rows)
	type outRow struct {
		IDStr  string  `json:"id_str"`
		Name   string  `json:"name"`
		Amount float64 `json:"amount"`
		Cnt    int     `json:"cnt"`
	}
	list := make([]outRow, 0, len(rows))
	var tAmount float64
	var tCnt int
	for _, r := range rows {
		list = append(list, outRow{IDStr: strconv.FormatInt(r.SupplierID, 10), Name: r.Name, Amount: round2o(r.Amount), Cnt: r.Cnt})
		tAmount += r.Amount
		tCnt += r.Cnt
	}
	response.OK(c, gin.H{"list": list, "summary": gin.H{"amount": round2o(tAmount), "cnt": tCnt}})
}

/* ─── 库存统计 ─── */
func (h *AnalysisHandler) StockStat(c *gin.Context) {
	ctx := c.Request.Context()
	tenantID := context.GetTenantID(ctx)
	dateFrom := c.Query("date_from")
	dateTo := c.Query("date_to")
	keyword := c.Query("keyword")

	var goods []struct {
		ID           int64
		Name         string
		SpecGroups   string
		Code         string
		MainUnit     string
		ImageURL     string
		CategoryName string
		CurrentStock int
		Cost         float64
	}
	q := h.db.Table("goods AS g").
		Joins("LEFT JOIN goods_categories gc ON gc.id = g.category_id").
		Where("g.tenant_id = ?", tenantID)
	if keyword != "" {
		like := "%" + keyword + "%"
		q = q.Where("g.name LIKE ? OR g.code LIKE ?", like, like)
	}
	q.Select("g.id, g.name, g.spec_groups, g.code, g.main_unit, g.image_url, COALESCE(gc.name,'') AS category_name, g.current_stock, g.purchase_price AS cost").Order("g.id DESC").Scan(&goods)

	type outRow struct {
		IDStr        string  `json:"id_str"`
		Name         string  `json:"name"`
		Spec         string  `json:"spec"`
		Code         string  `json:"code"`
		CategoryName string  `json:"category_name"`
		ImageURL     string  `json:"image_url"`
		Unit         string  `json:"unit"`
		OpenQty      float64 `json:"open_qty"`
		OpenAmt      float64 `json:"open_amt"`
		InQty        float64 `json:"in_qty"`
		InAmt        float64 `json:"in_amt"`
		OutQty       float64 `json:"out_qty"`
		OutAmt       float64 `json:"out_amt"`
		EndQty       float64 `json:"end_qty"`
		EndAmt       float64 `json:"end_amt"`
	}
	list := make([]outRow, 0, len(goods))
	for _, g := range goods {
		moves := collectStockMoves(h.db, tenantID, g.ID)
		var periodIn, periodOut, periodInAmt, periodOutAmt, netAfter float64
		for _, m := range moves {
			if m.Date > dateTo {
				netAfter += m.In - m.Out
				continue
			}
			if dateFrom != "" && m.Date < dateFrom {
				continue
			}
			uc := m.UnitCost
			if uc == 0 {
				uc = g.Cost
			}
			if m.In > 0 {
				periodIn += m.In
				periodInAmt += m.In * uc
			}
			if m.Out > 0 {
				periodOut += m.Out
				periodOutAmt += m.Out * uc
			}
		}
		endQty := float64(g.CurrentStock) - netAfter
		openQty := endQty - periodIn + periodOut
		list = append(list, outRow{
			IDStr: strconv.FormatInt(g.ID, 10), Name: g.Name, Spec: orderSpecNames(g.SpecGroups), Code: g.Code,
			CategoryName: g.CategoryName, ImageURL: g.ImageURL, Unit: g.MainUnit,
			OpenQty: openQty, OpenAmt: round2o(openQty * g.Cost),
			InQty: periodIn, InAmt: round2o(periodInAmt),
			OutQty: periodOut, OutAmt: round2o(periodOutAmt),
			EndQty: endQty, EndAmt: round2o(endQty * g.Cost),
		})
	}
	response.OK(c, gin.H{"list": list})
}

/* ─── 经营利润 ─── */
func (h *AnalysisHandler) ProfitSummary(c *gin.Context) {
	ctx := c.Request.Context()
	tenantID := context.GetTenantID(ctx)
	dateFrom := c.Query("date_from")
	dateTo := c.Query("date_to")

	sum := func(table, amtCol, dateCol, extra string, args ...interface{}) float64 {
		q := h.db.Table(table).Where("tenant_id = ?", tenantID)
		if extra != "" {
			q = q.Where(extra, args...)
		}
		if dateFrom != "" {
			q = q.Where(dateCol+" >= ?", dateFrom)
		}
		if dateTo != "" {
			q = q.Where(dateCol+" <= ?", dateTo)
		}
		var v struct{ V float64 }
		q.Select("COALESCE(SUM(" + amtCol + "),0) AS v").Scan(&v)
		return round2o(v.V)
	}

	sales := sum("sales", "subtotal", "bill_date", "status <> 9")
	cost := 0.0
	{
		q := h.db.Table("sale_items AS si").Joins("JOIN sales s ON s.id = si.sale_id").Joins("JOIN goods g ON g.id = si.goods_id").Where("s.tenant_id = ? AND s.status <> 9", tenantID)
		if dateFrom != "" {
			q = q.Where("s.bill_date >= ?", dateFrom)
		}
		if dateTo != "" {
			q = q.Where("s.bill_date <= ?", dateTo)
		}
		var v struct{ V float64 }
		q.Select("COALESCE(SUM(si.quantity * g.purchase_price),0) AS v").Scan(&v)
		cost = round2o(v.V)
	}
	salesReturn := sum("sales_returns", "total_amount", "bill_date", "")
	payDiscount := sum("payments", "discount_amount", "bill_date", "status <> 9")
	receiptDiscount := sum("receipts", "discount_amount", "bill_date", "status <> 9")
	otherIncome := sum("incomes", "amount", "bill_date", "status = 1 AND direction = 1")
	otherExpense := sum("incomes", "amount", "bill_date", "status = 1 AND direction = 2")

	surplus, loss := 0.0, 0.0
	{
		q := h.db.Table("stock_count_items AS i").
			Joins("JOIN stock_counts p ON p.id = i.count_id").
			Joins("JOIN goods g ON g.id = i.goods_id").
			Where("p.tenant_id = ? AND p.status <> 9", tenantID)
		if dateFrom != "" {
			q = q.Where("p.bill_date >= ?", dateFrom)
		}
		if dateTo != "" {
			q = q.Where("p.bill_date <= ?", dateTo)
		}
		var v struct{ Surplus, Loss float64 }
		q.Select("COALESCE(SUM(CASE WHEN i.diff_qty > 0 THEN i.diff_qty * COALESCE(NULLIF(i.unit_cost,0), g.purchase_price) ELSE 0 END),0) AS surplus, COALESCE(SUM(CASE WHEN i.diff_qty < 0 THEN -i.diff_qty * COALESCE(NULLIF(i.unit_cost,0), g.purchase_price) ELSE 0 END),0) AS loss").Scan(&v)
		surplus = round2o(v.Surplus)
		loss = round2o(v.Loss)
	}

	tax := 0.0
	incomeTotal := round2o(sales + surplus + payDiscount + otherIncome)
	expenseTotal := round2o(cost + tax + loss + salesReturn + receiptDiscount + otherExpense)
	profit := round2o(incomeTotal - expenseTotal)

	response.OK(c, gin.H{
		"income": gin.H{
			"sales": sales, "surplus": surplus, "pay_discount": payDiscount,
			"other_income": otherIncome, "total": incomeTotal,
		},
		"expense": gin.H{
			"cost": cost, "tax": tax, "loss": loss, "sales_return": salesReturn,
			"receipt_discount": receiptDiscount, "other_expense": otherExpense, "total": expenseTotal,
		},
		"profit": profit,
	})
}

/* ─── 盘盈/盘亏明细 ─── */
func (h *AnalysisHandler) SurplusDetail(c *gin.Context) { h.diffDetail(c, true) }
func (h *AnalysisHandler) LossDetail(c *gin.Context)    { h.diffDetail(c, false) }

func (h *AnalysisHandler) diffDetail(c *gin.Context, positive bool) {
	ctx := c.Request.Context()
	tenantID := context.GetTenantID(ctx)

	var shopName string
	var shop model.Shop
	if h.db.Where("tenant_id = ? AND is_main = 1", tenantID).First(&shop).Error == nil {
		shopName = shop.Name
	}
	if shopName == "" {
		shopName = "我的店铺"
	}

	cond := "i.diff_qty > 0"
	if !positive {
		cond = "i.diff_qty < 0"
	}
	q := h.db.Table("stock_count_items AS i").
		Joins("JOIN stock_counts p ON p.id = i.count_id").
		Joins("JOIN goods g ON g.id = i.goods_id").
		Joins("LEFT JOIN warehouses w ON w.id = p.warehouse_id").
		Where("p.tenant_id = ? AND p.status <> 9 AND "+cond, tenantID)
	if v := c.Query("date_from"); v != "" {
		q = q.Where("p.bill_date >= ?", v)
	}
	if v := c.Query("date_to"); v != "" {
		q = q.Where("p.bill_date <= ?", v)
	}

	type raw struct {
		ID         int64
		CountID    int64
		GoodsName  string
		Code       string
		SpecGroups string
		MainUnit   string
		ImageURL   string
		OrderNo    string
		BillDate   string
		Warehouse  string
		Qty        int
		Amount     float64
	}
	var rows []raw
	q.Select("i.id, i.count_id, g.name AS goods_name, g.code, g.spec_groups, g.main_unit, g.image_url, p.order_no, p.bill_date, COALESCE(NULLIF(w.name,''), ?) AS warehouse, ABS(i.diff_qty) AS qty, ABS(i.diff_qty) * COALESCE(NULLIF(i.unit_cost,0), g.purchase_price) AS amount", shopName).
		Order("p.bill_date DESC, i.id ASC").Scan(&rows)

	type outRow struct {
		IDStr     string  `json:"id_str"`
		CountID   string  `json:"count_id"`
		GoodsName string  `json:"goods_name"`
		Code      string  `json:"code"`
		Spec      string  `json:"spec"`
		Unit      string  `json:"unit"`
		ImageURL  string  `json:"image_url"`
		OrderNo   string  `json:"order_no"`
		BillDate  string  `json:"bill_date"`
		Warehouse string  `json:"warehouse"`
		Qty       int     `json:"qty"`
		Amount    float64 `json:"amount"`
	}
	list := make([]outRow, 0, len(rows))
	var tQty int
	var tAmount float64
	for _, r := range rows {
		list = append(list, outRow{
			IDStr: strconv.FormatInt(r.ID, 10), CountID: strconv.FormatInt(r.CountID, 10), GoodsName: r.GoodsName,
			Code: r.Code, Spec: orderSpecNames(r.SpecGroups), Unit: r.MainUnit, ImageURL: r.ImageURL,
			OrderNo: r.OrderNo, BillDate: r.BillDate, Warehouse: r.Warehouse, Qty: r.Qty, Amount: round2o(r.Amount),
		})
		tQty += r.Qty
		tAmount += r.Amount
	}
	response.OK(c, gin.H{"list": list, "summary": gin.H{"qty": tQty, "amount": round2o(tAmount)}})
}
