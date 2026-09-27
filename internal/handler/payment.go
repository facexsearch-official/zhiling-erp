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

type PaymentHandler struct{ db *gorm.DB }

func NewPaymentHandler(db *gorm.DB) *PaymentHandler { return &PaymentHandler{db: db} }

type paymentCreateReq struct {
	ShopID         int64   `json:"shop_id"`
	RelatedNo      string  `json:"related_no"`
	Type           string  `json:"type"`
	SupplierID     int64   `json:"supplier_id"`
	SalesmanID     int64   `json:"salesman_id"`
	BillDate       string  `json:"bill_date"`
	Amount         float64 `json:"amount"`
	DiscountAmount float64 `json:"discount_amount"`
	AccountID      int64   `json:"account_id"`
	Attachments    string  `json:"attachments"`
	Remark         string  `json:"remark"`
}

func (h *PaymentHandler) List(c *gin.Context) {
	ctx := c.Request.Context()
	tenantID := context.GetTenantID(ctx)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	keyword := c.Query("keyword")
	dateFrom := c.Query("date_from")
	dateTo := c.Query("date_to")
	hideVoid := c.Query("hide_void") == "1"

	var total int64
	var list []model.Payment
	q := h.db.Table("payments AS p").
		Joins("LEFT JOIN suppliers s ON s.id = p.supplier_id").
		Joins("LEFT JOIN accounts a ON a.id = p.account_id").
		Joins("LEFT JOIN salesmen sm ON sm.id = p.salesman_id").
		Joins("LEFT JOIN users u ON u.id = p.created_by").
		Where("p.tenant_id = ?", tenantID)
	if keyword != "" {
		kw := "%" + keyword + "%"
		q = q.Where("s.name LIKE ? OR p.order_no LIKE ? OR p.remark LIKE ?", kw, kw, kw)
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
	q.Session(&gorm.Session{}).Count(&total)
	q.Select("p.*, s.name AS supplier_name, a.name AS account_name, sm.name AS salesman_name, u.nickname AS maker_name").
		Offset((page - 1) * pageSize).Limit(pageSize).Order("p.created_at DESC").Scan(&list)
	for i := range list {
		list[i].IDStr = strconv.FormatInt(list[i].ID, 10)
	}

	var agg struct {
		Amount   float64
		Discount float64
	}
	h.db.Table("payments AS p").Joins("LEFT JOIN suppliers s ON s.id = p.supplier_id").Where("p.tenant_id = ?", tenantID).
		Select("COALESCE(SUM(p.amount),0) AS amount, COALESCE(SUM(p.discount_amount),0) AS discount").Scan(&agg)

	response.OK(c, gin.H{
		"list": list, "total": total, "page": page, "page_size": pageSize,
		"summary": gin.H{"amount": round2o(agg.Amount), "discount": round2o(agg.Discount)},
	})
}

func (h *PaymentHandler) GetByID(c *gin.Context) {
	ctx := c.Request.Context()
	tenantID := context.GetTenantID(ctx)
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var p model.Payment
	if err := h.db.Table("payments").Where("id = ? AND tenant_id = ?", id, tenantID).First(&p).Error; err != nil {
		response.NotFound(c, "付款单不存在")
		return
	}
	p.IDStr = strconv.FormatInt(p.ID, 10)
	if p.SupplierID != 0 {
		var s model.Supplier
		if h.db.Where("id = ?", p.SupplierID).First(&s).Error == nil {
			p.SupplierName = s.Name
		}
	}
	if p.AccountID != 0 {
		var a model.Account
		if h.db.Where("id = ?", p.AccountID).First(&a).Error == nil {
			p.AccountName = a.Name
		}
	}
	if p.SalesmanID != 0 {
		var sm model.Salesman
		if h.db.Where("id = ?", p.SalesmanID).First(&sm).Error == nil {
			p.SalesmanName = sm.Name
		}
	}
	if p.CreatedBy != 0 {
		var u model.User
		if h.db.Where("id = ?", p.CreatedBy).First(&u).Error == nil {
			p.MakerName = u.Nickname
		}
	}
	response.OK(c, p)
}

func (h *PaymentHandler) Create(c *gin.Context) {
	ctx := c.Request.Context()
	tenantID := context.GetTenantID(ctx)
	userID := context.GetUserID(ctx)
	var req paymentCreateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	if req.Amount <= 0 {
		response.BadRequest(c, "请输入付款金额")
		return
	}
	if req.Type == "" {
		req.Type = "直接付款"
	}
	p := model.Payment{
		ID: snowflake.GenID(), TenantID: tenantID, ShopID: req.ShopID,
		OrderNo: nextNo(h.db, tenantID, "payments", "FKD"), RelatedNo: req.RelatedNo, Type: req.Type,
		SupplierID: req.SupplierID, SalesmanID: req.SalesmanID, BillDate: req.BillDate,
		Amount: req.Amount, DiscountAmount: req.DiscountAmount, AccountID: req.AccountID,
		Attachments: req.Attachments, Status: 1, Remark: req.Remark, CreatedBy: userID,
	}
	err := h.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Table("payments").Create(&p).Error; err != nil {
			return err
		}
		if req.AccountID != 0 {
			if err := tx.Model(&model.Account{}).Where("id = ? AND tenant_id = ?", req.AccountID, tenantID).
				UpdateColumn("balance", gorm.Expr("balance - ?", req.Amount)).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}
	p.IDStr = strconv.FormatInt(p.ID, 10)
	response.OK(c, p)
}

func (h *PaymentHandler) Update(c *gin.Context) {
	ctx := c.Request.Context()
	tenantID := context.GetTenantID(ctx)
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var p model.Payment
	if err := h.db.Table("payments").Where("id = ? AND tenant_id = ?", id, tenantID).First(&p).Error; err != nil {
		response.NotFound(c, "付款单不存在")
		return
	}
	var req paymentCreateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	if req.Type == "" {
		req.Type = "直接付款"
	}
	err := h.db.Transaction(func(tx *gorm.DB) error {
		if p.AccountID != 0 {
			if err := tx.Model(&model.Account{}).Where("id = ? AND tenant_id = ?", p.AccountID, tenantID).
				UpdateColumn("balance", gorm.Expr("balance + ?", p.Amount)).Error; err != nil {
				return err
			}
		}
		if err := tx.Table("payments").Where("id = ?", p.ID).Updates(map[string]interface{}{
			"related_no": req.RelatedNo, "type": req.Type, "supplier_id": req.SupplierID,
			"salesman_id": req.SalesmanID, "bill_date": req.BillDate, "amount": req.Amount,
			"discount_amount": req.DiscountAmount, "account_id": req.AccountID, "remark": req.Remark,
		}).Error; err != nil {
			return err
		}
		if req.AccountID != 0 {
			if err := tx.Model(&model.Account{}).Where("id = ? AND tenant_id = ?", req.AccountID, tenantID).
				UpdateColumn("balance", gorm.Expr("balance - ?", req.Amount)).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}
	response.OKMsg(c, "更新成功")
}

func (h *PaymentHandler) Void(c *gin.Context) {
	ctx := c.Request.Context()
	tenantID := context.GetTenantID(ctx)
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var p model.Payment
	if err := h.db.Table("payments").Where("id = ? AND tenant_id = ?", id, tenantID).First(&p).Error; err != nil {
		response.NotFound(c, "付款单不存在")
		return
	}
	if p.Status == 9 {
		response.OKMsg(c, "已作废")
		return
	}
	err := h.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Table("payments").Where("id = ?", p.ID).Update("status", 9).Error; err != nil {
			return err
		}
		if p.AccountID != 0 {
			if err := tx.Model(&model.Account{}).Where("id = ? AND tenant_id = ?", p.AccountID, tenantID).
				UpdateColumn("balance", gorm.Expr("balance + ?", p.Amount)).Error; err != nil {
				return err
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
