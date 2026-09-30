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

type ReceiptHandler struct{ db *gorm.DB }

func NewReceiptHandler(db *gorm.DB) *ReceiptHandler { return &ReceiptHandler{db: db} }

// recalcReceivable 重算客户应收欠款 = 销售总额(未作废) - 已收总额(未作废)
func (h *ReceiptHandler) recalcReceivable(tx *gorm.DB, tenantID, customerID int64) {
	if customerID == 0 {
		return
	}
	var saleTotal, receiptTotal float64
	tx.Raw("SELECT COALESCE(SUM(total_amount),0) FROM sales WHERE tenant_id=? AND customer_id=? AND status<>9", tenantID, customerID).Scan(&saleTotal)
	tx.Raw("SELECT COALESCE(SUM(amount),0) FROM receipts WHERE tenant_id=? AND customer_id=? AND status=1", tenantID, customerID).Scan(&receiptTotal)
	recv := saleTotal - receiptTotal
	if recv < 0 {
		recv = 0
	}
	tx.Model(&model.Customer{}).Where("id = ? AND tenant_id = ?", customerID, tenantID).
		UpdateColumn("total_receivable", recv)
}

type receiptCreateReq struct {
	ShopID         int64   `json:"shop_id"`
	RelatedNo      string  `json:"related_no"`
	Type           string  `json:"type"`
	CustomerID     model.FlexInt64 `json:"customer_id"`
	SalesmanID     int64   `json:"salesman_id"`
	BillDate       string  `json:"bill_date"`
	Amount         float64 `json:"amount"`
	DiscountAmount float64 `json:"discount_amount"`
	DepositOffset  float64 `json:"deposit_offset"`
	AccountID      int64   `json:"account_id"`
	Attachments    string  `json:"attachments"`
	Remark         string  `json:"remark"`
}

func (h *ReceiptHandler) List(c *gin.Context) {
	ctx := c.Request.Context()
	tenantID := context.GetTenantID(ctx)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	keyword := c.Query("keyword")
	dateFrom := c.Query("date_from")
	dateTo := c.Query("date_to")
	hideVoid := c.Query("hide_void") == "1"

	var total int64
	var list []model.Receipt
	q := h.db.Table("receipts AS r").
		Joins("LEFT JOIN customers cu ON cu.id = r.customer_id").
		Joins("LEFT JOIN accounts a ON a.id = r.account_id").
		Joins("LEFT JOIN salesmen sm ON sm.id = r.salesman_id").
		Joins("LEFT JOIN users u ON u.id = r.created_by").
		Where("r.tenant_id = ?", tenantID)
	if keyword != "" {
		kw := "%" + keyword + "%"
		q = q.Where("cu.name LIKE ? OR r.order_no LIKE ? OR r.remark LIKE ?", kw, kw, kw)
	}
	if dateFrom != "" {
		q = q.Where("r.bill_date >= ?", dateFrom)
	}
	if dateTo != "" {
		q = q.Where("r.bill_date <= ?", dateTo)
	}
	if hideVoid {
		q = q.Where("r.status = 1")
	}
	q.Session(&gorm.Session{}).Count(&total)
	q.Select("r.*, cu.name AS customer_name, a.name AS account_name, sm.name AS salesman_name, u.nickname AS maker_name").
		Offset((page - 1) * pageSize).Limit(pageSize).Order("r.created_at DESC").Scan(&list)
	for i := range list {
		list[i].IDStr = strconv.FormatInt(list[i].ID, 10)
	}
	response.OKPage(c, list, total, page, pageSize)
}

func (h *ReceiptHandler) GetByID(c *gin.Context) {
	ctx := c.Request.Context()
	tenantID := context.GetTenantID(ctx)
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var r model.Receipt
	if err := h.db.Table("receipts").Where("id = ? AND tenant_id = ?", id, tenantID).First(&r).Error; err != nil {
		response.NotFound(c, "收款单不存在")
		return
	}
	r.IDStr = strconv.FormatInt(r.ID, 10)
	if r.CustomerID != 0 {
		var cu model.Customer
		if h.db.Where("id = ?", r.CustomerID).First(&cu).Error == nil {
			r.CustomerName = cu.Name
		}
	}
	if r.AccountID != 0 {
		var a model.Account
		if h.db.Where("id = ?", r.AccountID).First(&a).Error == nil {
			r.AccountName = a.Name
		}
	}
	if r.SalesmanID != 0 {
		var sm model.Salesman
		if h.db.Where("id = ?", r.SalesmanID).First(&sm).Error == nil {
			r.SalesmanName = sm.Name
		}
	}
	if r.CreatedBy != 0 {
		var u model.User
		if h.db.Where("id = ?", r.CreatedBy).First(&u).Error == nil {
			r.MakerName = u.Nickname
		}
	}
	response.OK(c, r)
}

func (h *ReceiptHandler) Create(c *gin.Context) {
	ctx := c.Request.Context()
	tenantID := context.GetTenantID(ctx)
	userID := context.GetUserID(ctx)
	var req receiptCreateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	if req.Amount <= 0 {
		response.BadRequest(c, "请输入收款金额")
		return
	}
	if req.Type == "" {
		req.Type = "直接收款"
	}
	r := model.Receipt{
		ID: snowflake.GenID(), TenantID: tenantID, ShopID: req.ShopID,
		OrderNo: nextNo(h.db, tenantID, "receipts", "SKD"), RelatedNo: req.RelatedNo, Type: req.Type,
		CustomerID: int64(req.CustomerID), SalesmanID: req.SalesmanID, BillDate: req.BillDate,
		Amount: req.Amount, DiscountAmount: req.DiscountAmount, DepositOffset: req.DepositOffset,
		AccountID: req.AccountID, Attachments: req.Attachments, Status: 1, Remark: req.Remark, CreatedBy: userID,
	}
	err := h.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Table("receipts").Create(&r).Error; err != nil {
			return err
		}
		if req.AccountID != 0 {
			if err := tx.Model(&model.Account{}).Where("id = ? AND tenant_id = ?", req.AccountID, tenantID).
				UpdateColumn("balance", gorm.Expr("balance + ?", req.Amount)).Error; err != nil {
				return err
			}
		}
		h.recalcReceivable(tx, tenantID, int64(req.CustomerID))
		return nil
	})
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}
	r.IDStr = strconv.FormatInt(r.ID, 10)
	response.OK(c, r)
}

func (h *ReceiptHandler) Update(c *gin.Context) {
	ctx := c.Request.Context()
	tenantID := context.GetTenantID(ctx)
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var r model.Receipt
	if err := h.db.Table("receipts").Where("id = ? AND tenant_id = ?", id, tenantID).First(&r).Error; err != nil {
		response.NotFound(c, "收款单不存在")
		return
	}
	var req receiptCreateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	if req.Type == "" {
		req.Type = "直接收款"
	}
	err := h.db.Transaction(func(tx *gorm.DB) error {
		if r.AccountID != 0 {
			if err := tx.Model(&model.Account{}).Where("id = ? AND tenant_id = ?", r.AccountID, tenantID).
				UpdateColumn("balance", gorm.Expr("balance - ?", r.Amount)).Error; err != nil {
				return err
			}
		}
		if err := tx.Table("receipts").Where("id = ?", r.ID).Updates(map[string]interface{}{
			"related_no": req.RelatedNo, "type": req.Type, "customer_id": int64(req.CustomerID),
			"salesman_id": req.SalesmanID, "bill_date": req.BillDate, "amount": req.Amount,
			"discount_amount": req.DiscountAmount, "deposit_offset": req.DepositOffset,
			"account_id": req.AccountID, "remark": req.Remark,
		}).Error; err != nil {
			return err
		}
		if req.AccountID != 0 {
			if err := tx.Model(&model.Account{}).Where("id = ? AND tenant_id = ?", req.AccountID, tenantID).
				UpdateColumn("balance", gorm.Expr("balance + ?", req.Amount)).Error; err != nil {
				return err
			}
		}
		h.recalcReceivable(tx, tenantID, r.CustomerID)
		h.recalcReceivable(tx, tenantID, int64(req.CustomerID))
		return nil
	})
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}
	response.OKMsg(c, "更新成功")
}

func (h *ReceiptHandler) Void(c *gin.Context) {
	ctx := c.Request.Context()
	tenantID := context.GetTenantID(ctx)
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var r model.Receipt
	if err := h.db.Table("receipts").Where("id = ? AND tenant_id = ?", id, tenantID).First(&r).Error; err != nil {
		response.NotFound(c, "收款单不存在")
		return
	}
	if r.Status == 9 {
		response.OKMsg(c, "已作废")
		return
	}
	err := h.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Table("receipts").Where("id = ?", r.ID).Update("status", 9).Error; err != nil {
			return err
		}
		if r.AccountID != 0 {
			if err := tx.Model(&model.Account{}).Where("id = ? AND tenant_id = ?", r.AccountID, tenantID).
				UpdateColumn("balance", gorm.Expr("balance - ?", r.Amount)).Error; err != nil {
				return err
			}
		}
		h.recalcReceivable(tx, tenantID, r.CustomerID)
		return nil
	})
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}
	response.OKMsg(c, "作废成功")
}
