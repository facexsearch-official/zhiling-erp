package handler

import (
	"time"

	"pisa_server/internal/pkg/context"
	"pisa_server/internal/pkg/response"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type DashboardHandler struct {
	db *gorm.DB
}

func NewDashboardHandler(db *gorm.DB) *DashboardHandler {
	return &DashboardHandler{db: db}
}

// Summary 首页统计概览
func (h *DashboardHandler) Summary(c *gin.Context) {
	ctx := c.Request.Context()
	tenantID := context.GetTenantID(ctx)
	now := time.Now()
	today := now.Format("2006-01-02")
	monthStart := now.Format("2006-01") + "-01"
	sevenAgo := now.AddDate(0, 0, -6).Format("2006-01-02")

	var t struct {
		Amount float64
		Cnt    int64
	}
	h.db.Table("sales").Where("tenant_id = ? AND status <> 9 AND bill_date = ?", tenantID, today).
		Select("COALESCE(SUM(subtotal),0) AS amount, COUNT(*) AS cnt").Scan(&t)

	var m struct {
		Amount float64
		Cnt    int64
	}
	h.db.Table("sales").Where("tenant_id = ? AND status <> 9 AND bill_date >= ?", tenantID, monthStart).
		Select("COALESCE(SUM(subtotal),0) AS amount, COUNT(*) AS cnt").Scan(&m)

	var mc struct{ Cost float64 }
	h.db.Table("sale_items AS si").
		Joins("JOIN sales s ON s.id = si.sale_id").
		Joins("JOIN goods g ON g.id = si.goods_id").
		Where("s.tenant_id = ? AND s.status <> 9 AND s.bill_date >= ?", tenantID, monthStart).
		Select("COALESCE(SUM(si.quantity * g.purchase_price),0) AS cost").Scan(&mc)

	var sv struct{ Value float64 }
	h.db.Table("goods").Where("tenant_id = ?", tenantID).
		Select("COALESCE(SUM(current_stock * purchase_price),0) AS value").Scan(&sv)

	var cc int64
	h.db.Table("customers").Where("tenant_id = ?", tenantID).Count(&cc)

	type drow struct {
		D      string
		Amount float64
	}
	var drows []drow
	h.db.Table("sales").Where("tenant_id = ? AND status <> 9 AND bill_date >= ?", tenantID, sevenAgo).
		Select("bill_date AS d, COALESCE(SUM(subtotal),0) AS amount").Group("bill_date").Scan(&drows)
	byDate := map[string]float64{}
	for _, r := range drows {
		byDate[r.D] = r.Amount
	}
	trend := make([]gin.H, 0, 7)
	for i := 6; i >= 0; i-- {
		day := now.AddDate(0, 0, -i)
		ds := day.Format("2006-01-02")
		trend = append(trend, gin.H{"date": ds, "label": day.Format("01-02"), "amount": round2o(byDate[ds])})
	}

	type grow struct {
		Name   string
		Amount float64
	}
	var grows []grow
	h.db.Table("sale_items AS si").
		Joins("JOIN sales s ON s.id = si.sale_id").
		Joins("JOIN goods g ON g.id = si.goods_id").
		Where("s.tenant_id = ? AND s.status <> 9 AND s.bill_date >= ?", tenantID, monthStart).
		Select("g.name AS name, COALESCE(SUM(si.amount),0) AS amount").
		Group("g.id, g.name").Order("amount DESC").Limit(5).Scan(&grows)
	topGoods := make([]gin.H, 0, len(grows))
	for _, r := range grows {
		topGoods = append(topGoods, gin.H{"name": r.Name, "amount": round2o(r.Amount)})
	}

	profit := m.Amount - mc.Cost
	margin := 0.0
	if m.Amount > 0 {
		margin = round2o(profit / m.Amount * 100)
	}

	response.OK(c, gin.H{
		"today_sales":    round2o(t.Amount),
		"today_count":    t.Cnt,
		"month_sales":    round2o(m.Amount),
		"month_count":    m.Cnt,
		"month_profit":   round2o(profit),
		"month_margin":   margin,
		"stock_value":    round2o(sv.Value),
		"customer_count": cc,
		"trend":          trend,
		"top_goods":      topGoods,
	})
}
