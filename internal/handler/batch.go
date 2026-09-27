package handler

import (
	"strconv"
	"time"

	"pisa_server/internal/model"
	"pisa_server/internal/pkg/context"
	"pisa_server/internal/pkg/response"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type BatchHandler struct{ db *gorm.DB }

func NewBatchHandler(db *gorm.DB) *BatchHandler { return &BatchHandler{db: db} }

type batchRow struct {
	ID           int64  `json:"id"`
	IDStr        string `json:"id_str"`
	BatchNo      string `json:"batch_no"`
	GoodsID      int64  `json:"goods_id"`
	GoodsName    string `json:"goods_name"`
	GoodsCode    string `json:"goods_code"`
	Spec         string `json:"spec"`
	CategoryName string `json:"category_name"`
	UnitName     string `json:"unit_name"`
	Stock        int    `json:"stock"`
}

type expiryRow struct {
	ID             int64  `json:"id"`
	IDStr          string `json:"id_str"`
	BatchNo        string `json:"batch_no"`
	GoodsID        int64  `json:"goods_id"`
	GoodsName      string `json:"goods_name"`
	GoodsCode      string `json:"goods_code"`
	Spec           string `json:"spec"`
	CategoryName   string `json:"category_name"`
	UnitName       string `json:"unit_name"`
	ProductionDate string `json:"production_date"`
	ExpiryDate     string `json:"expiry_date"`
	RemainDays     int    `json:"remain_days"`
	Stock          int    `json:"stock"`
}

func (h *BatchHandler) batches(c *gin.Context) []string {
	ctx := c.Request.Context()
	tenantID := context.GetTenantID(ctx)
	var out []string
	h.db.Model(&model.GoodsBatch{}).Where("tenant_id = ? AND batch_no <> ''", tenantID).Distinct().Pluck("batch_no", &out)
	return out
}

func (h *BatchHandler) List(c *gin.Context) {
	ctx := c.Request.Context()
	tenantID := context.GetTenantID(ctx)
	q := h.db.Table("goods_batches AS b").
		Joins("JOIN goods g ON g.id = b.goods_id").
		Joins("LEFT JOIN goods_categories gc ON gc.id = g.category_id").
		Where("b.tenant_id = ?", tenantID)
	if v := c.Query("batch_no"); v != "" && v != "0" {
		q = q.Where("b.batch_no = ?", v)
	}
	if v := c.Query("goods_id"); v != "" && v != "0" {
		q = q.Where("b.goods_id = ?", v)
	}
	if v := c.Query("supplier_id"); v != "" && v != "0" {
		q = q.Where("b.supplier_id = ?", v)
	}
	if v := c.Query("brand"); v != "" && v != "0" {
		q = q.Where("b.brand = ?", v)
	}
	if v := c.Query("category_id"); v != "" && v != "0" {
		q = q.Where("g.category_id = ?", v)
	}
	if c.Query("hide_disabled") == "1" {
		q = q.Where("g.status = 1")
	}
	var rows []struct {
		ID           int64
		BatchNo      string
		GoodsID      int64
		SpecKey      string
		Stock        int
		GoodsName    string
		GoodsCode    string
		CategoryName string
		UnitName     string
	}
	q.Select("b.id, b.batch_no, b.goods_id, b.spec_key, b.stock, g.name AS goods_name, g.code AS goods_code, COALESCE(gc.name,'') AS category_name, g.main_unit AS unit_name").
		Order("b.id DESC").Scan(&rows)

	list := make([]batchRow, 0, len(rows))
	var total int
	for _, r := range rows {
		spec := r.SpecKey
		if spec == "" {
			spec = "--"
		}
		list = append(list, batchRow{
			ID: r.ID, IDStr: strconv.FormatInt(r.ID, 10), BatchNo: r.BatchNo, GoodsID: r.GoodsID,
			GoodsName: r.GoodsName, GoodsCode: r.GoodsCode, Spec: spec, CategoryName: r.CategoryName,
			UnitName: r.UnitName, Stock: r.Stock,
		})
		total += r.Stock
	}
	response.OK(c, gin.H{"list": list, "total_stock": total, "batches": h.batches(c)})
}

func (h *BatchHandler) Expiry(c *gin.Context) {
	ctx := c.Request.Context()
	tenantID := context.GetTenantID(ctx)
	q := h.db.Table("goods_batches AS b").
		Joins("JOIN goods g ON g.id = b.goods_id").
		Joins("LEFT JOIN goods_categories gc ON gc.id = g.category_id").
		Where("b.tenant_id = ? AND b.expiry_date <> ''", tenantID)
	if v := c.Query("goods_id"); v != "" && v != "0" {
		q = q.Where("b.goods_id = ?", v)
	}
	if v := c.Query("supplier_id"); v != "" && v != "0" {
		q = q.Where("b.supplier_id = ?", v)
	}
	if v := c.Query("category_id"); v != "" && v != "0" {
		q = q.Where("g.category_id = ?", v)
	}
	if v := c.Query("days_from"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			q = q.Where("DATEDIFF(b.expiry_date, CURDATE()) >= ?", n)
		}
	}
	if v := c.Query("days_to"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			q = q.Where("DATEDIFF(b.expiry_date, CURDATE()) <= ?", n)
		}
	}
	status := c.Query("status")
	if status == "已过期" {
		q = q.Where("b.expiry_date < CURDATE()")
	} else if status == "临期" {
		q = q.Where("b.expiry_date >= CURDATE() AND DATEDIFF(b.expiry_date, CURDATE()) <= 30")
	} else if status == "正常" {
		q = q.Where("DATEDIFF(b.expiry_date, CURDATE()) > 30")
	}
	if c.Query("hide_disabled") == "1" {
		q = q.Where("g.status = 1")
	}
	var rows []struct {
		ID             int64
		BatchNo        string
		GoodsID        int64
		SpecKey        string
		Stock          int
		ProductionDate string
		ExpiryDate     string
		GoodsName      string
		GoodsCode      string
		CategoryName   string
		UnitName       string
	}
	q.Select("b.id, b.batch_no, b.goods_id, b.spec_key, b.stock, b.production_date, b.expiry_date, g.name AS goods_name, g.code AS goods_code, COALESCE(gc.name,'') AS category_name, g.main_unit AS unit_name").
		Order("b.expiry_date ASC").Scan(&rows)

	today := time.Now()
	todayStr := today.Format("2006-01-02")
	_ = todayStr
	list := make([]expiryRow, 0, len(rows))
	var total int
	for _, r := range rows {
		spec := r.SpecKey
		if spec == "" {
			spec = "--"
		}
		remain := 0
		if ed, err := time.Parse("2006-01-02", r.ExpiryDate); err == nil {
			remain = int(ed.Sub(time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, time.Local)).Hours() / 24)
		}
		list = append(list, expiryRow{
			ID: r.ID, IDStr: strconv.FormatInt(r.ID, 10), BatchNo: r.BatchNo, GoodsID: r.GoodsID,
			GoodsName: r.GoodsName, GoodsCode: r.GoodsCode, Spec: spec, CategoryName: r.CategoryName,
			UnitName: r.UnitName, ProductionDate: r.ProductionDate, ExpiryDate: r.ExpiryDate,
			RemainDays: remain, Stock: r.Stock,
		})
		total += r.Stock
	}
	response.OK(c, gin.H{"list": list, "total_stock": total})
}
