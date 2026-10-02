package repository

import (
	"context"
	"strconv"

	"pisa_server/internal/model"

	customContext "pisa_server/internal/pkg/context"

	"gorm.io/gorm"
)

type PurchaseReturnRepository struct {
	BaseRepository
}

func NewPurchaseReturnRepository(base BaseRepository) *PurchaseReturnRepository {
	return &PurchaseReturnRepository{BaseRepository: base}
}

func (r *PurchaseReturnRepository) List(ctx context.Context, page, pageSize int, supplierID int64, status int8, keyword string) ([]model.PurchaseReturn, int64) {
	tenantID := customContext.GetTenantID(ctx)
	var total int64
	var list []model.PurchaseReturn
	q := r.DB.Table("purchase_returns AS p").
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
	q.Session(&gorm.Session{}).Count(&total)
	q.Select("p.*, s.name AS supplier_name, COALESCE(sm.name, su.nickname) AS salesman_name, a.name AS account_name, u.nickname AS maker_name").
		Offset((page - 1) * pageSize).Limit(pageSize).Order("p.created_at DESC").Scan(&list)
	for i := range list {
		list[i].IDStr = strconv.FormatInt(list[i].ID, 10)
	}
	return list, total
}

func (r *PurchaseReturnRepository) GetByID(ctx context.Context, id int64) (*model.PurchaseReturn, error) {
	var pr model.PurchaseReturn
	err := r.Scoped(ctx).Table("purchase_returns").Where("id = ?", id).First(&pr).Error
	if err != nil {
		return &pr, err
	}
	pr.IDStr = strconv.FormatInt(pr.ID, 10)
	return &pr, nil
}

func (r *PurchaseReturnRepository) Create(ctx context.Context, pr *model.PurchaseReturn) error {
	return r.Scoped(ctx).Table("purchase_returns").Create(pr).Error
}

func (r *PurchaseReturnRepository) Update(ctx context.Context, pr *model.PurchaseReturn) error {
	return r.Scoped(ctx).Table("purchase_returns").Save(pr).Error
}

func (r *PurchaseReturnRepository) Delete(ctx context.Context, id int64) error {
	return r.Scoped(ctx).Table("purchase_returns").Where("id = ?", id).Delete(&model.PurchaseReturn{}).Error
}

// FillNames 填充单条退货单的供应商/业务员/结算账户/制单人名称
func (r *PurchaseReturnRepository) FillNames(ctx context.Context, pr *model.PurchaseReturn) {
	tenantID := customContext.GetTenantID(ctx)
	if pr.SupplierID != 0 {
		var s model.Supplier
		if r.DB.Where("id = ? AND tenant_id = ?", pr.SupplierID, tenantID).First(&s).Error == nil {
			pr.SupplierName = s.Name
		}
	}
	if pr.SalesmanID != 0 {
		var sm model.Salesman
		if r.DB.Where("id = ? AND tenant_id = ?", pr.SalesmanID, tenantID).First(&sm).Error == nil {
			pr.SalesmanName = sm.Name
		}
	}
	if pr.AccountID != 0 {
		var a model.Account
		if r.DB.Where("id = ? AND tenant_id = ?", pr.AccountID, tenantID).First(&a).Error == nil {
			pr.AccountName = a.Name
		}
	}
	if pr.CreatedBy != 0 {
		var u model.User
		if r.DB.Where("id = ?", pr.CreatedBy).First(&u).Error == nil {
			pr.MakerName = u.Nickname
		}
	}
}

// FillItemDetails 填充退货明细的货品信息
func (r *PurchaseReturnRepository) FillItemDetails(ctx context.Context, items []model.PurchaseReturnItem) {
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
