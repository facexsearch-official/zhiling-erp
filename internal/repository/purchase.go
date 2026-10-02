package repository

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"

	"pisa_server/internal/model"

	customContext "pisa_server/internal/pkg/context"

	"gorm.io/gorm"
)

type PurchaseRepository struct {
	BaseRepository
}

func NewPurchaseRepository(base BaseRepository) *PurchaseRepository {
	return &PurchaseRepository{BaseRepository: base}
}

func (r *PurchaseRepository) List(ctx context.Context, page, pageSize int, supplierID int64, status int8, keyword string, dateFrom, dateTo string) ([]model.Purchase, int64) {
	tenantID := customContext.GetTenantID(ctx)
	var total int64
	var list []model.Purchase
	q := r.DB.Table("purchases AS p").
		Joins("LEFT JOIN suppliers s ON s.id = p.supplier_id").
		Joins("LEFT JOIN salesmen sm ON sm.id = p.salesman_id").
		Joins("LEFT JOIN users su ON su.id = p.salesman_id").
		Joins("LEFT JOIN accounts a ON a.id = p.account_id").
		Joins("LEFT JOIN users u ON u.id = p.created_by").
		Where("p.tenant_id = ?", tenantID)
	if supplierID > 0 {
		q = q.Where("p.supplier_id = ?", supplierID)
	}
	if status > 0 {
		q = q.Where("p.status = ?", status)
	}
	if keyword != "" {
		kw := "%" + keyword + "%"
		q = q.Where("p.order_no LIKE ? OR s.name LIKE ?", kw, kw)
	}
	if dateFrom != "" {
		q = q.Where("p.bill_date >= ?", dateFrom)
	}
	if dateTo != "" {
		q = q.Where("p.bill_date <= ?", dateTo)
	}
	q.Session(&gorm.Session{}).Count(&total)
	q.Select("p.*, s.name AS supplier_name, COALESCE(sm.name, su.nickname) AS salesman_name, a.name AS account_name, u.nickname AS maker_name").
		Offset((page - 1) * pageSize).Limit(pageSize).Order("p.created_at DESC").Scan(&list)
	for i := range list {
		list[i].IDStr = strconv.FormatInt(list[i].ID, 10)
	}
	return list, total
}

func (r *PurchaseRepository) GetByID(ctx context.Context, id int64) (*model.Purchase, error) {
	var purchase model.Purchase
	err := r.Scoped(ctx).Table("purchases").Where("id = ?", id).First(&purchase).Error
	return &purchase, err
}

func (r *PurchaseRepository) Create(ctx context.Context, purchase *model.Purchase) error {
	return r.Scoped(ctx).Table("purchases").Create(purchase).Error
}

func (r *PurchaseRepository) Update(ctx context.Context, purchase *model.Purchase) error {
	return r.Scoped(ctx).Table("purchases").Save(purchase).Error
}

func (r *PurchaseRepository) Delete(ctx context.Context, id int64) error {
	return r.Scoped(ctx).Table("purchases").Where("id = ?", id).Delete(&model.Purchase{}).Error
}

func (r *PurchaseRepository) FillSupplierNames(ctx context.Context, purchases []model.Purchase, supplierRepo *SupplierRepository) {
	if len(purchases) == 0 {
		return
	}
	ids := make([]int64, 0)
	seen := make(map[int64]bool)
	for _, p := range purchases {
		if !seen[p.SupplierID] {
			ids = append(ids, p.SupplierID)
			seen[p.SupplierID] = true
		}
	}
	suppliers, _ := supplierRepo.ListAll(ctx)
	supplierMap := make(map[int64]string)
	for _, s := range suppliers {
		supplierMap[s.ID] = s.Name
	}
	for i := range purchases {
		purchases[i].SupplierName = supplierMap[purchases[i].SupplierID]
	}
}

// FillNames 填充单条进货单的供应商/业务员/结算账户名称
func (r *PurchaseRepository) FillNames(ctx context.Context, p *model.Purchase) {
	tenantID := customContext.GetTenantID(ctx)
	if p.SupplierID != 0 {
		var s model.Supplier
		if r.DB.Where("id = ? AND tenant_id = ?", p.SupplierID, tenantID).First(&s).Error == nil {
			p.SupplierName = s.Name
		}
	}
	if p.SalesmanID != 0 {
		var sm model.Salesman
		if r.DB.Where("id = ? AND tenant_id = ?", p.SalesmanID, tenantID).First(&sm).Error == nil {
			p.SalesmanName = sm.Name
		}
	}
	if p.AccountID != 0 {
		var a model.Account
		if r.DB.Where("id = ? AND tenant_id = ?", p.AccountID, tenantID).First(&a).Error == nil {
			p.AccountName = a.Name
		}
	}
}

// FillItemDetails 填充明细的货品信息（名称/编号/单位/规格/条码/图片）
func (r *PurchaseRepository) FillItemDetails(ctx context.Context, items []model.PurchaseItem) {
	if len(items) == 0 {
		return
	}
	ids := make([]int64, 0, len(items))
	for _, it := range items {
		ids = append(ids, it.GoodsID)
	}
	var goods []model.Goods
	r.DB.Where("tenant_id = ? AND id IN ?", customContext.GetTenantID(ctx), ids).Find(&goods)
	m := map[int64]model.Goods{}
	for _, g := range goods {
		m[g.ID] = g
	}
	for i := range items {
		items[i].GoodsIDStr = strconv.FormatInt(items[i].GoodsID, 10)
		g := m[items[i].GoodsID]
		items[i].GoodsName = g.Name
		items[i].GoodsCode = g.Code
		items[i].UnitName = g.MainUnit
		items[i].Spec = specNamesOf(g.SpecGroups)
		items[i].Barcode = g.Barcode
		items[i].ImageURL = g.ImageURL
		items[i].Brand = g.Brand
		items[i].Origin = g.Origin
	}
}

func specNamesOf(specGroups string) string {
	if specGroups == "" {
		return ""
	}
	var gs []struct {
		Name string `json:"name"`
	}
	if err := json.Unmarshal([]byte(specGroups), &gs); err != nil {
		return ""
	}
	names := make([]string, 0, len(gs))
	for _, g := range gs {
		if g.Name != "" {
			names = append(names, g.Name)
		}
	}
	return strings.Join(names, "/")
}
