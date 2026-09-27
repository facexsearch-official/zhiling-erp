package handler

import (
	"encoding/json"
	"math"
	"strconv"
	"strings"

	"pisa_server/internal/model"
	"pisa_server/internal/pkg/context"
	"pisa_server/internal/pkg/response"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type CommissionHandler struct {
	db *gorm.DB
}

func NewCommissionHandler(dbConn *gorm.DB) *CommissionHandler {
	return &CommissionHandler{db: dbConn}
}

func round2c(v float64) float64 { return math.Round(v*100) / 100 }

/* ── 规则 CRUD ── */

func (h *CommissionHandler) ListRules(c *gin.Context) {
	ctx := c.Request.Context()
	tenantID := context.GetTenantID(ctx)
	var list []model.CommissionRule
	h.db.Where("tenant_id = ?", tenantID).Order("id DESC").Find(&list)
	for i := range list {
		list[i].IDStr = strconv.FormatInt(list[i].ID, 10)
	}
	response.OK(c, list)
}

func (h *CommissionHandler) CreateRule(c *gin.Context) {
	ctx := c.Request.Context()
	tenantID := context.GetTenantID(ctx)
	var rule model.CommissionRule
	if err := c.ShouldBindJSON(&rule); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	rule.Name = strings.TrimSpace(rule.Name)
	if rule.Name == "" {
		response.BadRequest(c, "请输入规则名称")
		return
	}
	rule.ID = 0
	rule.TenantID = tenantID
	rule.Status = 1
	if err := h.db.Create(&rule).Error; err != nil {
		response.ServerError(c, err.Error())
		return
	}
	rule.IDStr = strconv.FormatInt(rule.ID, 10)
	response.OK(c, rule)
}

func (h *CommissionHandler) UpdateRule(c *gin.Context) {
	ctx := c.Request.Context()
	tenantID := context.GetTenantID(ctx)
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var existing model.CommissionRule
	if err := h.db.Where("id = ? AND tenant_id = ?", id, tenantID).First(&existing).Error; err != nil {
		response.NotFound(c, "规则不存在")
		return
	}
	if err := c.ShouldBindJSON(&existing); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	if err := h.db.Save(&existing).Error; err != nil {
		response.ServerError(c, err.Error())
		return
	}
	existing.IDStr = strconv.FormatInt(existing.ID, 10)
	response.OK(c, existing)
}

func (h *CommissionHandler) DeleteRule(c *gin.Context) {
	ctx := c.Request.Context()
	tenantID := context.GetTenantID(ctx)
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	h.db.Where("id = ? AND tenant_id = ?", id, tenantID).Delete(&model.CommissionRule{})
	response.OKMsg(c, "删除成功")
}

/* ── 业绩提成报表 ── */

type commissionRow struct {
	SalesmanID int64   `json:"salesman_id"`
	Name       string  `json:"name"`
	Phone      string  `json:"phone"`
	Count      int64   `json:"count"`
	Receivable float64 `json:"receivable"`
	Received   float64 `json:"received"`
	Profit     float64 `json:"profit"`
	Deposit    float64 `json:"deposit"`
	RuleName   string  `json:"rule_name"`
	Commission float64 `json:"commission"`
}

type tier struct {
	Min   float64 `json:"min"`
	Max   float64 `json:"max"`
	Ratio float64 `json:"ratio"`
}

func (h *CommissionHandler) Report(c *gin.Context) {
	ctx := c.Request.Context()
	tenantID := context.GetTenantID(ctx)
	dateFrom := c.Query("date_from")
	dateTo := c.Query("date_to")

	var rows []commissionRow
	q := h.db.Table("sales AS s").
		Select("s.salesman_id AS salesman_id, COUNT(*) AS count, COALESCE(SUM(s.total_amount),0) AS receivable, COALESCE(SUM(s.received_amount),0) AS received").
		Where("s.tenant_id = ?", tenantID)
	if dateFrom != "" {
		q = q.Where("s.bill_date >= ?", dateFrom)
	}
	if dateTo != "" {
		q = q.Where("s.bill_date <= ?", dateTo)
	}
	q.Group("s.salesman_id").Scan(&rows)

	// 毛利（销售额 - 成本）
	type profitRow struct {
		SalesmanID int64   `gorm:"column:salesman_id"`
		Profit     float64 `gorm:"column:profit"`
	}
	var profits []profitRow
	pq := h.db.Table("sale_items AS si").
		Select("s.salesman_id AS salesman_id, COALESCE(SUM(si.amount - si.quantity*g.purchase_price),0) AS profit").
		Joins("JOIN sales s ON s.id = si.sale_id").
		Joins("LEFT JOIN goods g ON g.id = si.goods_id").
		Where("s.tenant_id = ?", tenantID)
	if dateFrom != "" {
		pq = pq.Where("s.bill_date >= ?", dateFrom)
	}
	if dateTo != "" {
		pq = pq.Where("s.bill_date <= ?", dateTo)
	}
	pq.Group("s.salesman_id").Scan(&profits)
	profitMap := map[int64]float64{}
	for _, p := range profits {
		profitMap[p.SalesmanID] = round2c(p.Profit)
	}

	// 业务员姓名/电话
	smMap := map[int64]model.Salesman{}
	var salesmen []model.Salesman
	h.db.Where("tenant_id = ?", tenantID).Find(&salesmen)
	for _, sm := range salesmen {
		smMap[sm.ID] = sm
	}

	// 规则
	var rules []model.CommissionRule
	h.db.Where("tenant_id = ? AND status = 1", tenantID).Order("id DESC").Find(&rules)

	var totalReceivable, totalReceived, totalProfit, totalCommission float64
	for i := range rows {
		sm := smMap[rows[i].SalesmanID]
		rows[i].Name = sm.Name
		rows[i].Phone = sm.Phone
		rows[i].Profit = profitMap[rows[i].SalesmanID]
		rows[i].RuleName = "--"
		// 匹配规则
		for _, r := range rules {
			if !ruleApplies(r.StaffIDs, rows[i].SalesmanID) {
				continue
			}
			rows[i].RuleName = r.Name
			rows[i].Commission = calcCommission(r, rows[i].Receivable, rows[i].Received, rows[i].Profit)
			break
		}
		totalReceivable += rows[i].Receivable
		totalReceived += rows[i].Received
		totalProfit += rows[i].Profit
		totalCommission += rows[i].Commission
	}

	response.OK(c, gin.H{
		"list":             rows,
		"total_receivable": round2c(totalReceivable),
		"total_received":   round2c(totalReceived),
		"total_profit":     round2c(totalProfit),
		"total_deposit":    0,
		"total_commission": round2c(totalCommission),
	})
}

func ruleApplies(staffIDs string, salesmanID int64) bool {
	if strings.TrimSpace(staffIDs) == "" {
		return true
	}
	var ids []int64
	if err := json.Unmarshal([]byte(staffIDs), &ids); err != nil {
		return true
	}
	if len(ids) == 0 {
		return true
	}
	for _, id := range ids {
		if id == salesmanID {
			return true
		}
	}
	return false
}

func calcCommission(r model.CommissionRule, receivable, received, profit float64) float64 {
	var base float64
	switch r.Method {
	case 2:
		base = received
	case 3:
		base = profit
	default:
		base = receivable
	}
	var commission float64
	if r.CalcType == 2 && r.Tiers != "" {
		var tiers []tier
		if json.Unmarshal([]byte(r.Tiers), &tiers) == nil {
			for _, t := range tiers {
				if base >= t.Min && (t.Max == 0 || base <= t.Max) {
					commission = base * t.Ratio / 100
					break
				}
			}
		}
	} else {
		commission = base * r.Ratio / 100
	}
	if r.BonusEnabled == 1 && r.BonusTarget > 0 && base >= r.BonusTarget {
		commission += r.BonusAmount
	}
	return round2c(commission)
}
