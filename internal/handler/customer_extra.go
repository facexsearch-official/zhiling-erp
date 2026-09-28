package handler

import (
	"sort"
	"strconv"

	"pisa_server/internal/model"
	"pisa_server/internal/pkg/context"
	"pisa_server/internal/pkg/response"
	"pisa_server/internal/pkg/snowflake"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

/* ═══════════ 客户分类 ═══════════ */

type CustomerCategoryHandler struct{ db *gorm.DB }

func NewCustomerCategoryHandler(db *gorm.DB) *CustomerCategoryHandler {
	return &CustomerCategoryHandler{db: db}
}

func (h *CustomerCategoryHandler) List(c *gin.Context) {
	ctx := c.Request.Context()
	tenantID := context.GetTenantID(ctx)
	var list []model.CustomerCategory
	h.db.Where("tenant_id = ?", tenantID).Order("sort ASC, id ASC").Find(&list)
	// 以字符串返回 id/parent_id，避免前端 JS 大整数精度丢失导致分类树错乱
	out := make([]gin.H, 0, len(list))
	for _, cc := range list {
		out = append(out, gin.H{
			"id":        strconv.FormatInt(cc.ID, 10),
			"parent_id": strconv.FormatInt(cc.ParentID, 10),
			"name":      cc.Name,
			"sort":      cc.Sort,
		})
	}
	response.OK(c, out)
}

type ccReq struct {
	Name     string      `json:"name"`
	ParentID interface{} `json:"parent_id"`
}

func toInt64(v interface{}) int64 {
	switch t := v.(type) {
	case float64:
		return int64(t)
	case int64:
		return t
	case string:
		n, _ := strconv.ParseInt(t, 10, 64)
		return n
	}
	return 0
}

func (h *CustomerCategoryHandler) Create(c *gin.Context) {
	ctx := c.Request.Context()
	tenantID := context.GetTenantID(ctx)
	var req ccReq
	if err := c.ShouldBindJSON(&req); err != nil || req.Name == "" {
		response.BadRequest(c, "请输入分类名称")
		return
	}
	cc := model.CustomerCategory{ID: snowflake.GenID(), TenantID: tenantID, Name: req.Name, ParentID: toInt64(req.ParentID)}
	h.db.Table("customer_categories").Create(&cc)
	response.OK(c, cc)
}

func (h *CustomerCategoryHandler) Update(c *gin.Context) {
	ctx := c.Request.Context()
	tenantID := context.GetTenantID(ctx)
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var req ccReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	h.db.Table("customer_categories").Where("id = ? AND tenant_id = ?", id, tenantID).
		Updates(map[string]interface{}{"name": req.Name, "parent_id": toInt64(req.ParentID)})
	response.OKMsg(c, "更新成功")
}

func (h *CustomerCategoryHandler) Delete(c *gin.Context) {
	ctx := c.Request.Context()
	tenantID := context.GetTenantID(ctx)
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	h.db.Table("customer_categories").Where("id = ? AND tenant_id = ?", id, tenantID).Delete(&model.CustomerCategory{})
	response.OKMsg(c, "删除成功")
}

/* ═══════════ 客户详情统计 ═══════════ */

func (h *CustomerHandler) Stats(c *gin.Context) {
	ctx := c.Request.Context()
	tenantID := context.GetTenantID(ctx)
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	var saleAgg struct {
		Cnt      int
		Amount   float64
		LastDate string
	}
	h.db.Table("sales").Where("tenant_id = ? AND customer_id = ? AND status <> 9", tenantID, id).
		Select("COUNT(*) AS cnt, COALESCE(SUM(subtotal),0) AS amount, COALESCE(MAX(bill_date),'') AS last_date").Scan(&saleAgg)

	var recAgg struct{ LastDate string }
	h.db.Table("receipts").Where("tenant_id = ? AND customer_id = ? AND status = 1", tenantID, id).
		Select("COALESCE(MAX(bill_date),'') AS last_date").Scan(&recAgg)
	lastReceipt := recAgg.LastDate

	response.OK(c, gin.H{
		"total_amount": round2o(saleAgg.Amount), "total_count": saleAgg.Cnt,
		"last_sale_date": saleAgg.LastDate, "last_receipt_date": lastReceipt,
	})
}

/* ═══════════ 客户对账单 ═══════════ */

type stmtRow struct {
	BillDate   string  `json:"bill_date"`
	OrderNo    string  `json:"order_no"`
	DocType    string  `json:"doc_type"`
	Goods      string  `json:"goods"`
	Spec       string  `json:"spec"`
	Unit       string  `json:"unit"`
	Qty        float64 `json:"qty"`
	Price      float64 `json:"price"`
	Amount     float64 `json:"amount"`
	Freight    float64 `json:"freight"`
	Discount   float64 `json:"discount"`
	Receivable float64 `json:"receivable"`
	Received   float64 `json:"received"`
}

func (h *CustomerHandler) Statement(c *gin.Context) {
	ctx := c.Request.Context()
	tenantID := context.GetTenantID(ctx)
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	dateFrom := c.Query("date_from")
	dateTo := c.Query("date_to")

	dateSQL := func(alias string) (string, []interface{}) {
		s := ""
		var a []interface{}
		if dateFrom != "" {
			s += " AND " + alias + ".bill_date >= ?"
			a = append(a, dateFrom)
		}
		if dateTo != "" {
			s += " AND " + alias + ".bill_date <= ?"
			a = append(a, dateTo)
		}
		return s, a
	}

	// 期初欠款
	var openSales, openReceipts float64
	{
		s := ""
		a := []interface{}{tenantID, id}
		if dateFrom != "" {
			s = " AND bill_date < ?"
			a = append(a, dateFrom)
		}
		var v struct{ V float64 }
		h.db.Raw("SELECT COALESCE(SUM(total_amount),0) AS v FROM sales WHERE tenant_id=? AND customer_id=? AND status<>9"+s, a...).Scan(&v)
		openSales = v.V
		var v2 struct{ V float64 }
		h.db.Raw("SELECT COALESCE(SUM(amount),0) AS v FROM receipts WHERE tenant_id=? AND customer_id=? AND status=1"+s, a...).Scan(&v2)
		openReceipts = v2.V
	}
	opening := round2o(openSales - openReceipts)

	var rows []stmtRow
	// 销售明细
	{
		s, a := dateSQL("s")
		args := append([]interface{}{tenantID, id}, a...)
		var tmp []struct {
			BillDate   string
			OrderNo    string
			Goods      string
			SpecGroups string
			Unit       string
			Qty        int
			Price      float64
			Amount     float64
		}
		h.db.Raw(`SELECT s.bill_date, s.order_no, g.name AS goods, g.spec_groups, g.main_unit AS unit, si.quantity AS qty, si.unit_price AS price, si.amount AS amount
			FROM sale_items si JOIN sales s ON s.id = si.sale_id JOIN goods g ON g.id = si.goods_id
			WHERE s.tenant_id = ? AND s.customer_id = ? AND s.status <> 9`+s, args...).Scan(&tmp)
		for _, t := range tmp {
			rows = append(rows, stmtRow{BillDate: t.BillDate, OrderNo: t.OrderNo, DocType: "销售单", Goods: t.Goods, Spec: orderSpecNames(t.SpecGroups), Unit: t.Unit, Qty: float64(t.Qty), Price: t.Price, Amount: t.Amount, Receivable: t.Amount})
		}
	}
	// 销售退货
	{
		s, a := dateSQL("r")
		args := append([]interface{}{tenantID, id}, a...)
		var tmp []struct {
			BillDate string
			OrderNo  string
			Amount   float64
		}
		h.db.Raw("SELECT r.bill_date, r.order_no, r.total_amount AS amount FROM sales_returns r WHERE r.tenant_id = ? AND r.customer_id = ?"+s, args...).Scan(&tmp)
		for _, t := range tmp {
			rows = append(rows, stmtRow{BillDate: t.BillDate, OrderNo: t.OrderNo, DocType: "销售退货单", Receivable: -t.Amount})
		}
	}
	// 收款单
	{
		s, a := dateSQL("r")
		args := append([]interface{}{tenantID, id}, a...)
		var tmp []struct {
			BillDate string
			OrderNo  string
			Discount float64
			Amount   float64
		}
		h.db.Raw("SELECT r.bill_date, r.order_no, r.discount_amount AS discount, r.amount AS amount FROM receipts r WHERE r.tenant_id = ? AND r.customer_id = ? AND r.status = 1"+s, args...).Scan(&tmp)
		for _, t := range tmp {
			rows = append(rows, stmtRow{BillDate: t.BillDate, OrderNo: t.OrderNo, DocType: "收款单", Discount: t.Discount, Received: t.Amount})
		}
	}
	sort.SliceStable(rows, func(i, j int) bool { return rows[i].BillDate < rows[j].BillDate })

	var tQty, tAmount, tFreight, tDiscount, tReceivable, tReceived float64
	for _, r := range rows {
		tQty += r.Qty
		tAmount += r.Amount
		tFreight += r.Freight
		tDiscount += r.Discount
		tReceivable += r.Receivable
		tReceived += r.Received
	}

	response.OK(c, gin.H{
		"opening": opening,
		"rows":    rows,
		"summary": gin.H{
			"qty": tQty, "amount": round2o(tAmount), "freight": round2o(tFreight),
			"discount": round2o(tDiscount), "receivable": round2o(tReceivable), "received": round2o(tReceived),
		},
	})
}
