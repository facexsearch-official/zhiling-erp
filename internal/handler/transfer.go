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

type TransferHandler struct{ db *gorm.DB }

func NewTransferHandler(db *gorm.DB) *TransferHandler { return &TransferHandler{db: db} }

type transferCreateReq struct {
	BillDate      string  `json:"bill_date"`
	FromAccountID int64   `json:"from_account_id"`
	ToAccountID   int64   `json:"to_account_id"`
	Amount        float64 `json:"amount"`
	Remark        string  `json:"remark"`
}

func (h *TransferHandler) List(c *gin.Context) {
	ctx := c.Request.Context()
	tenantID := context.GetTenantID(ctx)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	keyword := c.Query("keyword")
	dateFrom := c.Query("date_from")
	dateTo := c.Query("date_to")

	var total int64
	var list []model.Transfer
	q := h.db.Table("transfers AS t").
		Joins("LEFT JOIN accounts fa ON fa.id = t.from_account_id").
		Joins("LEFT JOIN accounts ta ON ta.id = t.to_account_id").
		Joins("LEFT JOIN users u ON u.id = t.created_by").
		Where("t.tenant_id = ?", tenantID)
	if keyword != "" {
		kw := "%" + keyword + "%"
		q = q.Where("t.order_no LIKE ? OR t.remark LIKE ?", kw, kw)
	}
	if dateFrom != "" {
		q = q.Where("t.bill_date >= ?", dateFrom)
	}
	if dateTo != "" {
		q = q.Where("t.bill_date <= ?", dateTo)
	}
	q.Session(&gorm.Session{}).Count(&total)
	q.Select("t.*, fa.name AS from_name, ta.name AS to_name, u.nickname AS maker_name").
		Offset((page - 1) * pageSize).Limit(pageSize).Order("t.created_at DESC").Scan(&list)
	for i := range list {
		list[i].IDStr = strconv.FormatInt(list[i].ID, 10)
	}
	response.OKPage(c, list, total, page, pageSize)
}

func (h *TransferHandler) Create(c *gin.Context) {
	ctx := c.Request.Context()
	tenantID := context.GetTenantID(ctx)
	userID := context.GetUserID(ctx)
	var req transferCreateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	if req.FromAccountID == 0 || req.ToAccountID == 0 {
		response.BadRequest(c, "请选择转账账户")
		return
	}
	if req.FromAccountID == req.ToAccountID {
		response.BadRequest(c, "转出与转入账户不能相同")
		return
	}
	if req.Amount <= 0 {
		response.BadRequest(c, "请输入转账金额")
		return
	}
	t := model.Transfer{
		ID: snowflake.GenID(), TenantID: tenantID,
		OrderNo: nextNo(h.db, tenantID, "transfers", "ZZ"), BillDate: req.BillDate,
		FromAccountID: req.FromAccountID, ToAccountID: req.ToAccountID, Amount: req.Amount,
		Remark: req.Remark, CreatedBy: userID,
	}
	err := h.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&model.Account{}).Where("id = ? AND tenant_id = ?", req.FromAccountID, tenantID).
			UpdateColumn("balance", gorm.Expr("balance - ?", req.Amount)).Error; err != nil {
			return err
		}
		if err := tx.Model(&model.Account{}).Where("id = ? AND tenant_id = ?", req.ToAccountID, tenantID).
			UpdateColumn("balance", gorm.Expr("balance + ?", req.Amount)).Error; err != nil {
			return err
		}
		return tx.Table("transfers").Create(&t).Error
	})
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}
	t.IDStr = strconv.FormatInt(t.ID, 10)
	response.OK(c, t)
}
