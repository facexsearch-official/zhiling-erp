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

/* ═══════════ 收支类型 ═══════════ */

type IncomeTypeHandler struct{ db *gorm.DB }

func NewIncomeTypeHandler(db *gorm.DB) *IncomeTypeHandler { return &IncomeTypeHandler{db: db} }

func (h *IncomeTypeHandler) List(c *gin.Context) {
	ctx := c.Request.Context()
	tenantID := context.GetTenantID(ctx)
	var cnt int64
	h.db.Model(&model.IncomeType{}).Where("tenant_id = ?", tenantID).Count(&cnt)
	if cnt == 0 {
		names := []struct {
			n string
			d int8
		}{
			{"分红", 1}, {"POS收款", 1},
			{"电费", 2}, {"房租", 2}, {"燃气费", 2}, {"水费", 2}, {"通讯费", 2}, {"物业管理费", 2}, {"员工工资", 2},
		}
		defs := make([]model.IncomeType, 0, len(names))
		for i, x := range names {
			defs = append(defs, model.IncomeType{ID: snowflake.GenID(), TenantID: tenantID, Name: x.n, Direction: x.d, Sort: i, Status: 1})
		}
		h.db.Table("income_types").Create(&defs)
	}
	var list []model.IncomeType
	q := h.db.Where("tenant_id = ?", tenantID)
	if v := c.Query("direction"); v != "" && v != "0" {
		q = q.Where("direction = ?", v)
	}
	q.Order("sort ASC, id ASC").Find(&list)
	response.OK(c, list)
}

type incomeTypeReq struct {
	Name      string `json:"name"`
	Direction int8   `json:"direction"`
}

func (h *IncomeTypeHandler) Create(c *gin.Context) {
	ctx := c.Request.Context()
	tenantID := context.GetTenantID(ctx)
	var req incomeTypeReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	if req.Name == "" {
		response.BadRequest(c, "请输入类别名称")
		return
	}
	t := model.IncomeType{ID: snowflake.GenID(), TenantID: tenantID, Name: req.Name, Direction: req.Direction, Status: 1}
	if err := h.db.Table("income_types").Create(&t).Error; err != nil {
		response.ServerError(c, err.Error())
		return
	}
	response.OK(c, t)
}

func (h *IncomeTypeHandler) Update(c *gin.Context) {
	ctx := c.Request.Context()
	tenantID := context.GetTenantID(ctx)
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var req incomeTypeReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	h.db.Table("income_types").Where("id = ? AND tenant_id = ?", id, tenantID).
		Updates(map[string]interface{}{"name": req.Name, "direction": req.Direction})
	response.OKMsg(c, "更新成功")
}

func (h *IncomeTypeHandler) Delete(c *gin.Context) {
	ctx := c.Request.Context()
	tenantID := context.GetTenantID(ctx)
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	h.db.Table("income_types").Where("id = ? AND tenant_id = ?", id, tenantID).Delete(&model.IncomeType{})
	response.OKMsg(c, "删除成功")
}

/* ═══════════ 其他收入 ═══════════ */

type IncomeHandler struct{ db *gorm.DB }

func NewIncomeHandler(db *gorm.DB) *IncomeHandler { return &IncomeHandler{db: db} }

type incomeItemReq struct {
	AccountID int64   `json:"account_id"`
	Amount    float64 `json:"amount"`
}

type incomeCreateReq struct {
	ShopID      int64           `json:"shop_id"`
	TypeID      int64           `json:"type_id"`
	TypeName    string          `json:"type_name"`
	PartyName   string          `json:"party_name"`
	SalesmanID  int64           `json:"salesman_id"`
	BillDate    string          `json:"bill_date"`
	Remark      string          `json:"remark"`
	Attachments string          `json:"attachments"`
	Items       []incomeItemReq `json:"items"`
}

func (h *IncomeHandler) List(c *gin.Context)        { h.list(c, 1) }
func (h *IncomeHandler) ListExpense(c *gin.Context) { h.list(c, 2) }

func (h *IncomeHandler) list(c *gin.Context, dir int8) {
	ctx := c.Request.Context()
	tenantID := context.GetTenantID(ctx)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	keyword := c.Query("keyword")
	dateFrom := c.Query("date_from")
	dateTo := c.Query("date_to")
	hideVoid := c.Query("hide_void") == "1"

	var total int64
	var list []model.Income
	q := h.db.Table("incomes AS i").
		Joins("LEFT JOIN accounts a ON a.id = i.account_id").
		Joins("LEFT JOIN salesmen sm ON sm.id = i.salesman_id").
		Joins("LEFT JOIN users u ON u.id = i.created_by").
		Where("i.tenant_id = ? AND i.direction = ?", tenantID, dir)
	if keyword != "" {
		kw := "%" + keyword + "%"
		q = q.Where("i.order_no LIKE ? OR i.remark LIKE ?", kw, kw)
	}
	if dateFrom != "" {
		q = q.Where("i.bill_date >= ?", dateFrom)
	}
	if dateTo != "" {
		q = q.Where("i.bill_date <= ?", dateTo)
	}
	if hideVoid {
		q = q.Where("i.status = 1")
	}
	q.Session(&gorm.Session{}).Count(&total)
	q.Select("i.*, a.name AS account_name, sm.name AS salesman_name, u.nickname AS maker_name").
		Offset((page - 1) * pageSize).Limit(pageSize).Order("i.created_at DESC").Scan(&list)
	for i := range list {
		list[i].IDStr = strconv.FormatInt(list[i].ID, 10)
	}
	response.OKPage(c, list, total, page, pageSize)
}

func (h *IncomeHandler) GetByID(c *gin.Context) {
	ctx := c.Request.Context()
	tenantID := context.GetTenantID(ctx)
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var inc model.Income
	if err := h.db.Table("incomes").Where("id = ? AND tenant_id = ?", id, tenantID).First(&inc).Error; err != nil {
		response.NotFound(c, "其他收入不存在")
		return
	}
	inc.IDStr = strconv.FormatInt(inc.ID, 10)
	if inc.SalesmanID != 0 {
		var sm model.Salesman
		if h.db.Where("id = ?", inc.SalesmanID).First(&sm).Error == nil {
			inc.SalesmanName = sm.Name
		}
	}
	if inc.CreatedBy != 0 {
		var u model.User
		if h.db.Where("id = ?", inc.CreatedBy).First(&u).Error == nil {
			inc.MakerName = u.Nickname
		}
	}
	var items []model.IncomeItem
	h.db.Table("income_items").Where("income_id = ?", inc.ID).Order("id ASC").Find(&items)
	var aids []int64
	for _, it := range items {
		aids = append(aids, it.AccountID)
	}
	am := map[int64]model.Account{}
	if len(aids) > 0 {
		var accs []model.Account
		h.db.Where("id IN ?", aids).Find(&accs)
		for _, a := range accs {
			am[a.ID] = a
		}
	}
	for i := range items {
		items[i].AccountName = am[items[i].AccountID].Name
	}
	inc.Items = items
	if inc.AccountID != 0 {
		inc.AccountName = am[inc.AccountID].Name
	}
	response.OK(c, inc)
}

func (h *IncomeHandler) Create(c *gin.Context)        { h.create(c, 1) }
func (h *IncomeHandler) CreateExpense(c *gin.Context) { h.create(c, 2) }

func (h *IncomeHandler) create(c *gin.Context, dir int8) {
	ctx := c.Request.Context()
	tenantID := context.GetTenantID(ctx)
	userID := context.GetUserID(ctx)
	var req incomeCreateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	var total float64
	items := make([]model.IncomeItem, 0)
	for _, it := range req.Items {
		if it.Amount <= 0 {
			continue
		}
		total += it.Amount
		items = append(items, model.IncomeItem{ID: snowflake.GenID(), TenantID: tenantID, AccountID: it.AccountID, Amount: it.Amount})
	}
	if total <= 0 {
		response.BadRequest(c, "请输入金额")
		return
	}
	inc := model.Income{
		ID: snowflake.GenID(), TenantID: tenantID, ShopID: req.ShopID, Direction: dir,
		OrderNo: nextNo(h.db, tenantID, "incomes", "SZD"), TypeID: req.TypeID, TypeName: req.TypeName,
		PartyName: req.PartyName, SalesmanID: req.SalesmanID, BillDate: req.BillDate,
		Amount: total, Attachments: req.Attachments, Status: 1, Remark: req.Remark, CreatedBy: userID,
	}
	if len(items) > 0 {
		inc.AccountID = items[0].AccountID
	}
	sign := "+"
	if dir == 2 {
		sign = "-"
	}
	err := h.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Table("incomes").Create(&inc).Error; err != nil {
			return err
		}
		for i := range items {
			items[i].IncomeID = inc.ID
		}
		if len(items) > 0 {
			if err := tx.Table("income_items").Create(&items).Error; err != nil {
				return err
			}
		}
		for _, it := range items {
			if it.AccountID != 0 {
				expr := gorm.Expr("balance + ?", it.Amount)
				if sign == "-" {
					expr = gorm.Expr("balance - ?", it.Amount)
				}
				if err := tx.Model(&model.Account{}).Where("id = ? AND tenant_id = ?", it.AccountID, tenantID).
					UpdateColumn("balance", expr).Error; err != nil {
					return err
				}
			}
		}
		return nil
	})
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}
	inc.IDStr = strconv.FormatInt(inc.ID, 10)
	inc.Items = items
	response.OK(c, inc)
}

func (h *IncomeHandler) Update(c *gin.Context) {
	ctx := c.Request.Context()
	tenantID := context.GetTenantID(ctx)
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var inc model.Income
	if err := h.db.Table("incomes").Where("id = ? AND tenant_id = ?", id, tenantID).First(&inc).Error; err != nil {
		response.NotFound(c, "其他收入不存在")
		return
	}
	var req incomeCreateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	var oldItems []model.IncomeItem
	h.db.Table("income_items").Where("income_id = ?", inc.ID).Find(&oldItems)
	var total float64
	newItems := make([]model.IncomeItem, 0)
	for _, it := range req.Items {
		if it.Amount <= 0 {
			continue
		}
		total += it.Amount
		newItems = append(newItems, model.IncomeItem{ID: snowflake.GenID(), TenantID: tenantID, AccountID: it.AccountID, Amount: it.Amount})
	}
	if total <= 0 {
		response.BadRequest(c, "请输入金额")
		return
	}
	primary := int64(0)
	if len(newItems) > 0 {
		primary = newItems[0].AccountID
	}
	sgn := 1.0
	if inc.Direction == 2 {
		sgn = -1.0
	}
	err := h.db.Transaction(func(tx *gorm.DB) error {
		for _, it := range oldItems {
			if it.AccountID != 0 {
				if err := tx.Model(&model.Account{}).Where("id = ? AND tenant_id = ?", it.AccountID, tenantID).
					UpdateColumn("balance", gorm.Expr("balance - ?", it.Amount*sgn)).Error; err != nil {
					return err
				}
			}
		}
		if err := tx.Table("income_items").Where("income_id = ?", inc.ID).Delete(&model.IncomeItem{}).Error; err != nil {
			return err
		}
		for i := range newItems {
			newItems[i].IncomeID = inc.ID
		}
		if len(newItems) > 0 {
			if err := tx.Table("income_items").Create(&newItems).Error; err != nil {
				return err
			}
		}
		for _, it := range newItems {
			if it.AccountID != 0 {
				if err := tx.Model(&model.Account{}).Where("id = ? AND tenant_id = ?", it.AccountID, tenantID).
					UpdateColumn("balance", gorm.Expr("balance + ?", it.Amount*sgn)).Error; err != nil {
					return err
				}
			}
		}
		return tx.Table("incomes").Where("id = ?", inc.ID).Updates(map[string]interface{}{
			"type_id": req.TypeID, "type_name": req.TypeName, "party_name": req.PartyName,
			"salesman_id": req.SalesmanID, "bill_date": req.BillDate, "amount": total,
			"account_id": primary, "remark": req.Remark,
		}).Error
	})
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}
	response.OKMsg(c, "更新成功")
}

func (h *IncomeHandler) Void(c *gin.Context) {
	ctx := c.Request.Context()
	tenantID := context.GetTenantID(ctx)
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var inc model.Income
	if err := h.db.Table("incomes").Where("id = ? AND tenant_id = ?", id, tenantID).First(&inc).Error; err != nil {
		response.NotFound(c, "其他收入不存在")
		return
	}
	if inc.Status == 9 {
		response.OKMsg(c, "已作废")
		return
	}
	var items []model.IncomeItem
	h.db.Table("income_items").Where("income_id = ?", inc.ID).Find(&items)
	sgn := 1.0
	if inc.Direction == 2 {
		sgn = -1.0
	}
	err := h.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Table("incomes").Where("id = ?", inc.ID).Update("status", 9).Error; err != nil {
			return err
		}
		for _, it := range items {
			if it.AccountID != 0 {
				if err := tx.Model(&model.Account{}).Where("id = ? AND tenant_id = ?", it.AccountID, tenantID).
					UpdateColumn("balance", gorm.Expr("balance - ?", it.Amount*sgn)).Error; err != nil {
					return err
				}
			}
		}
		return nil
	})
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}
	response.OKMsg(c, "作废成功")
}
