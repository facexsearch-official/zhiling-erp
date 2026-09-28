package handler

import (
	"strconv"

	"pisa_server/internal/model"
	"pisa_server/internal/pkg/context"
	"pisa_server/internal/pkg/response"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type SalesStatHandler struct{ db *gorm.DB }

func NewSalesStatHandler(db *gorm.DB) *SalesStatHandler { return &SalesStatHandler{db: db} }

func (h *SalesStatHandler) dateFilter(q *gorm.DB, c *gin.Context) *gorm.DB {
	if v := c.Query("date_from"); v != "" {
		q = q.Where("s.bill_date >= ?", v)
	}
	if v := c.Query("date_to"); v != "" {
		q = q.Where("s.bill_date <= ?", v)
	}
	if v := c.Query("order_no"); v != "" {
		q = q.Where("s.order_no LIKE ?", "%"+v+"%")
	}
	return q
}

func margin(amount, cost float64) float64 {
	if amount == 0 {
		return 0
	}
	return round2o((amount - cost) / amount * 100)
}

func (h *SalesStatHandler) ByDoc(c *gin.Context) {
	ctx := c.Request.Context()
	tenantID := context.GetTenantID(ctx)

	type raw struct {
		ID           int64
		BillDate     string
		OrderNo      string
		CustomerName string
		ShopName     string
		Amount       float64
		Freight      float64
		RoundOff     float64
		Cost         float64
	}
	q := h.db.Table("sales AS s").
		Joins("LEFT JOIN customers cu ON cu.id = s.customer_id").
		Joins("LEFT JOIN shops sh ON sh.id = s.shop_id").
		Where("s.tenant_id = ? AND s.status <> 9", tenantID)
	q = h.dateFilter(q, c)
	if v := c.Query("customer_id"); v != "" && v != "0" {
		q = q.Where("s.customer_id = ?", v)
	}
	var rows []raw
	q.Select(`s.id, s.bill_date, s.order_no, COALESCE(cu.name,'') AS customer_name, COALESCE(sh.name,'') AS shop_name,
		s.subtotal AS amount, s.freight, s.round_off,
		COALESCE((SELECT SUM(si.quantity * g.purchase_price) FROM sale_items si JOIN goods g ON g.id = si.goods_id WHERE si.sale_id = s.id),0) AS cost`).
		Order("s.bill_date DESC, s.id DESC").Scan(&rows)

	type outRow struct {
		IDStr        string  `json:"id_str"`
		BillDate     string  `json:"bill_date"`
		OrderNo      string  `json:"order_no"`
		ShopName     string  `json:"shop_name"`
		CustomerName string  `json:"customer_name"`
		Amount       float64 `json:"amount"`
		Freight      float64 `json:"freight"`
		RoundOff     float64 `json:"round_off"`
		Cost         float64 `json:"cost"`
		Tax          float64 `json:"tax"`
		Profit       float64 `json:"profit"`
		Margin       float64 `json:"margin"`
	}
	list := make([]outRow, 0, len(rows))
	var tAmount, tFreight, tRound, tCost, tProfit float64
	for _, r := range rows {
		profit := round2o(r.Amount - r.Cost)
		list = append(list, outRow{
			IDStr: strconv.FormatInt(r.ID, 10), BillDate: r.BillDate, OrderNo: r.OrderNo,
			ShopName: r.ShopName, CustomerName: r.CustomerName, Amount: r.Amount, Freight: r.Freight,
			RoundOff: r.RoundOff, Cost: r.Cost, Tax: 0, Profit: profit, Margin: margin(r.Amount, r.Cost),
		})
		tAmount += r.Amount
		tFreight += r.Freight
		tRound += r.RoundOff
		tCost += r.Cost
		tProfit += profit
	}
	response.OK(c, gin.H{
		"list": list,
		"summary": gin.H{
			"amount": round2o(tAmount), "freight": round2o(tFreight), "round_off": round2o(tRound),
			"count": len(list), "cost": round2o(tCost), "tax": 0, "profit": round2o(tProfit),
			"margin": margin(tAmount, tCost),
		},
	})
}

func (h *SalesStatHandler) ByGoods(c *gin.Context) {
	ctx := c.Request.Context()
	tenantID := context.GetTenantID(ctx)

	type raw struct {
		GoodsID   int64
		Name      string
		Code      string
		MainUnit  string
		SpecGroup string
		Qty       int
		Amount    float64
		Cost      float64
	}
	q := h.db.Table("sale_items AS si").
		Joins("JOIN sales s ON s.id = si.sale_id").
		Joins("JOIN goods g ON g.id = si.goods_id").
		Where("s.tenant_id = ? AND s.status <> 9", tenantID)
	if v := c.Query("date_from"); v != "" {
		q = q.Where("s.bill_date >= ?", v)
	}
	if v := c.Query("date_to"); v != "" {
		q = q.Where("s.bill_date <= ?", v)
	}
	if v := c.Query("order_no"); v != "" {
		q = q.Where("s.order_no LIKE ?", "%"+v+"%")
	}
	if v := c.Query("customer_id"); v != "" && v != "0" {
		q = q.Where("s.customer_id = ?", v)
	}
	var rows []raw
	q.Select("si.goods_id, g.name, g.code, g.main_unit, g.spec_groups, SUM(si.quantity) AS qty, SUM(si.amount) AS amount, SUM(si.quantity * g.purchase_price) AS cost").
		Group("si.goods_id, g.name, g.code, g.main_unit, g.spec_groups").Order("amount DESC").Scan(&rows)

	type outRow struct {
		IDStr  string  `json:"id_str"`
		Name   string  `json:"name"`
		Spec   string  `json:"spec"`
		Code   string  `json:"code"`
		Unit   string  `json:"unit"`
		Qty    int     `json:"qty"`
		Amount float64 `json:"amount"`
		Cost   float64 `json:"cost"`
		Tax    float64 `json:"tax"`
		Profit float64 `json:"profit"`
		Margin float64 `json:"margin"`
	}
	list := make([]outRow, 0, len(rows))
	var tAmount, tCost, tProfit float64
	var tQty int
	for _, r := range rows {
		profit := round2o(r.Amount - r.Cost)
		list = append(list, outRow{
			IDStr: strconv.FormatInt(r.GoodsID, 10), Name: r.Name, Spec: orderSpecNames(r.SpecGroup),
			Code: r.Code, Unit: r.MainUnit, Qty: r.Qty, Amount: round2o(r.Amount), Cost: round2o(r.Cost),
			Tax: 0, Profit: profit, Margin: margin(r.Amount, r.Cost),
		})
		tAmount += r.Amount
		tCost += r.Cost
		tProfit += profit
		tQty += r.Qty
	}
	response.OK(c, gin.H{
		"list": list,
		"summary": gin.H{
			"amount": round2o(tAmount), "qty": tQty, "cost": round2o(tCost), "tax": 0,
			"profit": round2o(tProfit), "margin": margin(tAmount, tCost),
		},
	})
}

func (h *SalesStatHandler) ByCustomer(c *gin.Context) {
	ctx := c.Request.Context()
	tenantID := context.GetTenantID(ctx)

	type raw struct {
		CustomerID   int64
		Name         string
		CategoryName string
		Count        int
		Amount       float64
		Cost         float64
	}
	q := h.db.Table("sales AS s").
		Joins("LEFT JOIN customers cu ON cu.id = s.customer_id").
		Where("s.tenant_id = ? AND s.status <> 9", tenantID)
	q = h.dateFilter(q, c)
	if v := c.Query("customer_id"); v != "" && v != "0" {
		q = q.Where("s.customer_id = ?", v)
	}
	var rows []raw
	q.Select(`s.customer_id, COALESCE(cu.name,'未指定') AS name, '' AS category_name, COUNT(*) AS count, SUM(s.subtotal) AS amount,
		COALESCE((SELECT SUM(si.quantity * g.purchase_price) FROM sale_items si JOIN goods g ON g.id = si.goods_id WHERE si.sale_id = s.id),0) AS cost`).
		Group("s.customer_id, cu.name").Order("amount DESC").Scan(&rows)

	type outRow struct {
		IDStr        string  `json:"id_str"`
		Name         string  `json:"name"`
		CategoryName string  `json:"category_name"`
		Amount       float64 `json:"amount"`
		Count        int     `json:"count"`
		Cost         float64 `json:"cost"`
		Tax          float64 `json:"tax"`
		Profit       float64 `json:"profit"`
		Margin       float64 `json:"margin"`
	}
	list := make([]outRow, 0, len(rows))
	var tAmount, tCost, tProfit float64
	var tCount int
	for _, r := range rows {
		profit := round2o(r.Amount - r.Cost)
		list = append(list, outRow{
			IDStr: strconv.FormatInt(r.CustomerID, 10), Name: r.Name, CategoryName: r.CategoryName,
			Amount: round2o(r.Amount), Count: r.Count, Cost: round2o(r.Cost), Tax: 0, Profit: profit, Margin: margin(r.Amount, r.Cost),
		})
		tAmount += r.Amount
		tCost += r.Cost
		tProfit += profit
		tCount += r.Count
	}
	response.OK(c, gin.H{
		"list": list,
		"summary": gin.H{
			"amount": round2o(tAmount), "count": tCount, "cost": round2o(tCost), "tax": 0,
			"profit": round2o(tProfit), "margin": margin(tAmount, tCost),
		},
	})
}

var _ = model.Sale{}
