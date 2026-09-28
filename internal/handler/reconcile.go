package handler

import (
	"sort"
	"strconv"

	"pisa_server/internal/model"
	"pisa_server/internal/pkg/context"
	"pisa_server/internal/pkg/response"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

/* ═══════════ 客户/供应商对账 ═══════════ */

type ReconcileHandler struct{ db *gorm.DB }

func NewReconcileHandler(db *gorm.DB) *ReconcileHandler { return &ReconcileHandler{db: db} }

type reconcileRow struct {
	Seq          int     `json:"seq"`
	PartyID      int64   `json:"party_id"`
	IDStr        string  `json:"id_str"`
	Name         string  `json:"name"`
	CategoryName string  `json:"category_name"`
	Opening      float64 `json:"opening"`
	Period       float64 `json:"period"`
	Total        float64 `json:"total"`
}

func (h *ReconcileHandler) Customer(c *gin.Context) { h.reconcile(c, "customer") }
func (h *ReconcileHandler) Supplier(c *gin.Context) { h.reconcile(c, "supplier") }

func (h *ReconcileHandler) reconcile(c *gin.Context, kind string) {
	ctx := c.Request.Context()
	tenantID := context.GetTenantID(ctx)
	dateFrom := c.Query("date_from")
	dateTo := c.Query("date_to")
	hideZero := c.Query("hide_zero") == "1"

	table, docTable, docAmtCol, payTable, payAmtCol := "customers", "sales", "total_amount", "receipts", "amount"
	if kind == "supplier" {
		table, docTable, docAmtCol, payTable, payAmtCol = "suppliers", "purchases", "total_amount", "payments", "amount"
	}

	type party struct {
		ID           int64
		Name         string
		CategoryName string
	}
	var parties []party
	h.db.Table(table+" AS p").Where("p.tenant_id = ? AND p.status = 1", tenantID).
		Select("p.id, p.name, '' AS category_name").Order("p.id ASC").Scan(&parties)

	// aggregate: total doc, total pay, opening doc, opening pay, period doc, period pay
	type agg struct {
		PartyID int64
		Amt     float64
	}
	sumMap := func(query string, args ...interface{}) map[int64]float64 {
		var rows []agg
		h.db.Raw(query, args...).Scan(&rows)
		m := map[int64]float64{}
		for _, r := range rows {
			m[r.PartyID] = r.Amt
		}
		return m
	}
	partyCol := "customer_id"
	if kind == "supplier" {
		partyCol = "supplier_id"
	}
	totalDoc := sumMap("SELECT "+partyCol+" AS party_id, COALESCE(SUM("+docAmtCol+"),0) AS amt FROM "+docTable+" WHERE tenant_id = ? GROUP BY "+partyCol, tenantID)
	totalPay := sumMap("SELECT "+partyCol+" AS party_id, COALESCE(SUM("+payAmtCol+"),0) AS amt FROM "+payTable+" WHERE tenant_id = ? GROUP BY "+partyCol, tenantID)

	var openingDoc, openingPay, periodDoc, periodPay map[int64]float64
	if dateFrom != "" {
		openingDoc = sumMap("SELECT "+partyCol+" AS party_id, COALESCE(SUM("+docAmtCol+"),0) AS amt FROM "+docTable+" WHERE tenant_id = ? AND bill_date < ? GROUP BY "+partyCol, tenantID, dateFrom)
		openingPay = sumMap("SELECT "+partyCol+" AS party_id, COALESCE(SUM("+payAmtCol+"),0) AS amt FROM "+payTable+" WHERE tenant_id = ? AND bill_date < ? GROUP BY "+partyCol, tenantID, dateFrom)
	} else {
		openingDoc = map[int64]float64{}
		openingPay = map[int64]float64{}
	}
	if dateFrom != "" && dateTo != "" {
		periodDoc = sumMap("SELECT "+partyCol+" AS party_id, COALESCE(SUM("+docAmtCol+"),0) AS amt FROM "+docTable+" WHERE tenant_id = ? AND bill_date >= ? AND bill_date <= ? GROUP BY "+partyCol, tenantID, dateFrom, dateTo)
		periodPay = sumMap("SELECT "+partyCol+" AS party_id, COALESCE(SUM("+payAmtCol+"),0) AS amt FROM "+payTable+" WHERE tenant_id = ? AND bill_date >= ? AND bill_date <= ? GROUP BY "+partyCol, tenantID, dateFrom, dateTo)
	} else {
		periodDoc = totalDoc
		periodPay = totalPay
	}

	list := make([]reconcileRow, 0)
	var sumOpen, sumPeriod, sumTotal float64
	for _, p := range parties {
		opening := round2o(openingDoc[p.ID] - openingPay[p.ID])
		period := round2o(periodDoc[p.ID] - periodPay[p.ID])
		total := round2o(totalDoc[p.ID] - totalPay[p.ID])
		if hideZero && opening == 0 && period == 0 && total == 0 {
			continue
		}
		list = append(list, reconcileRow{
			Seq: len(list) + 1, PartyID: p.ID, IDStr: strconv.FormatInt(p.ID, 10), Name: p.Name,
			CategoryName: p.CategoryName, Opening: opening, Period: period, Total: total,
		})
		sumOpen += opening
		sumPeriod += period
		sumTotal += total
	}
	response.OK(c, gin.H{
		"list":    list,
		"summary": gin.H{"opening": round2o(sumOpen), "period": round2o(sumPeriod), "total": round2o(sumTotal)},
	})
}

/* ═══════════ 资金流水 ═══════════ */

type FundFlowHandler struct{ db *gorm.DB }

func NewFundFlowHandler(db *gorm.DB) *FundFlowHandler { return &FundFlowHandler{db: db} }

type fundFlowRow struct {
	BillDate string  `json:"bill_date"`
	OrderNo  string  `json:"order_no"`
	Party    string  `json:"party"`
	Item     string  `json:"item"`
	Account  string  `json:"account"`
	Shop     string  `json:"shop"`
	Income   float64 `json:"income"`
	Expense  float64 `json:"expense"`
	Remark   string  `json:"remark"`
}

func (h *FundFlowHandler) List(c *gin.Context) {
	ctx := c.Request.Context()
	tenantID := context.GetTenantID(ctx)
	dateFrom := c.Query("date_from")
	dateTo := c.Query("date_to")
	orderNo := c.Query("order_no")
	accountID := c.Query("account_id")

	var shopName string
	var shop model.Shop
	if h.db.Where("tenant_id = ? AND is_main = 1", tenantID).First(&shop).Error == nil {
		shopName = shop.Name
	}
	if shopName == "" {
		shopName = "我的店铺"
	}

	where := ""
	args := []interface{}{tenantID}
	if dateFrom != "" {
		where += " AND t.bill_date >= ?"
		args = append(args, dateFrom)
	}
	if dateTo != "" {
		where += " AND t.bill_date <= ?"
		args = append(args, dateTo)
	}
	if orderNo != "" {
		where += " AND t.order_no LIKE ?"
		args = append(args, "%"+orderNo+"%")
	}
	if accountID != "" && accountID != "0" {
		where += " AND t.account_id = ?"
		args = append(args, accountID)
	}

	var flows []fundFlowRow
	raw := func(sql string, a ...interface{}) {
		var rows []fundFlowRow
		h.db.Raw(sql, a...).Scan(&rows)
		flows = append(flows, rows...)
	}

	raw(`SELECT t.bill_date, t.order_no, COALESCE(cu.name,'') AS party, COALESCE(NULLIF(t.type,''),'直接收款') AS item, COALESCE(a.name,'') AS account, ? AS shop, t.amount AS income, 0 AS expense, COALESCE(t.remark,'') AS remark
		FROM receipts t LEFT JOIN customers cu ON cu.id = t.customer_id LEFT JOIN accounts a ON a.id = t.account_id
		WHERE t.tenant_id = ? AND t.status = 1`+where, append([]interface{}{shopName}, args...)...)
	raw(`SELECT t.bill_date, t.order_no, COALESCE(su.name,'') AS party, COALESCE(NULLIF(t.type,''),'直接付款') AS item, COALESCE(a.name,'') AS account, ? AS shop, 0 AS income, t.amount AS expense, COALESCE(t.remark,'') AS remark
		FROM payments t LEFT JOIN suppliers su ON su.id = t.supplier_id LEFT JOIN accounts a ON a.id = t.account_id
		WHERE t.tenant_id = ? AND t.status = 1`+where, append([]interface{}{shopName}, args...)...)
	raw(`SELECT t.bill_date, t.order_no, COALESCE(t.party_name,'') AS party, COALESCE(t.type_name,'') AS item, COALESCE(a.name,'') AS account, ? AS shop, t.amount AS income, 0 AS expense, COALESCE(t.remark,'') AS remark
		FROM incomes t LEFT JOIN accounts a ON a.id = t.account_id
		WHERE t.tenant_id = ? AND t.status = 1 AND t.direction = 1`+where, append([]interface{}{shopName}, args...)...)
	raw(`SELECT t.bill_date, t.order_no, COALESCE(t.party_name,'') AS party, COALESCE(t.type_name,'') AS item, COALESCE(a.name,'') AS account, ? AS shop, 0 AS income, t.amount AS expense, COALESCE(t.remark,'') AS remark
		FROM incomes t LEFT JOIN accounts a ON a.id = t.account_id
		WHERE t.tenant_id = ? AND t.status = 1 AND t.direction = 2`+where, append([]interface{}{shopName}, args...)...)

	sort.SliceStable(flows, func(i, j int) bool { return flows[i].BillDate > flows[j].BillDate })

	var totalIncome, totalExpense float64
	incByItem := map[string]float64{}
	expByItem := map[string]float64{}
	for _, f := range flows {
		totalIncome += f.Income
		totalExpense += f.Expense
		if f.Income > 0 {
			incByItem[f.Item] += f.Income
		}
		if f.Expense > 0 {
			expByItem[f.Item] += f.Expense
		}
	}
	type itemAmt struct {
		Name   string  `json:"name"`
		Amount float64 `json:"amount"`
	}
	incList := make([]itemAmt, 0)
	for k, v := range incByItem {
		incList = append(incList, itemAmt{Name: k, Amount: round2o(v)})
	}
	expList := make([]itemAmt, 0)
	for k, v := range expByItem {
		expList = append(expList, itemAmt{Name: k, Amount: round2o(v)})
	}
	sort.Slice(incList, func(i, j int) bool { return incList[i].Amount > incList[j].Amount })
	sort.Slice(expList, func(i, j int) bool { return expList[i].Amount > expList[j].Amount })

	response.OK(c, gin.H{
		"list": flows,
		"summary": gin.H{
			"total_income": round2o(totalIncome), "total_expense": round2o(totalExpense),
			"income_items": incList, "expense_items": expList,
		},
	})
}
