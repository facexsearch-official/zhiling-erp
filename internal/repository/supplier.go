package repository

import (
	"context"
	"strconv"

	"pisa_server/internal/model"

	customContext "pisa_server/internal/pkg/context"

	"gorm.io/gorm"
)

func fillSupplierCategoryStr(s *model.Supplier) {
	if s.CategoryID != nil {
		s.CategoryIDStr = strconv.FormatInt(int64(*s.CategoryID), 10)
	}
}

type SupplierRepository struct {
	BaseRepository
}

func NewSupplierRepository(base BaseRepository) *SupplierRepository {
	return &SupplierRepository{BaseRepository: base}
}

func (r *SupplierRepository) List(ctx context.Context, page, pageSize int, keyword string, categoryID int64, hideDisabled, hideZero bool) ([]model.Supplier, int64) {
	var total int64
	var list []model.Supplier
	q := r.Scoped(ctx)
	if keyword != "" {
		q = q.Where("name LIKE ? OR code LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}
	if hideDisabled {
		q = q.Where("status = 1")
	}
	if hideZero {
		q = q.Where("total_payable <> 0")
	}
	if categoryID == -1 {
		// 未分类：无分类或分类为 0
		q = q.Where("category_id IS NULL OR category_id = 0")
	} else if categoryID > 0 {
		sub := r.DB.Model(&model.SupplierCategory{}).Select("id").Where("parent_id = ?", categoryID)
		if tid := customContext.GetTenantID(ctx); tid > 0 {
			sub = sub.Where("tenant_id = ?", tid)
		}
		q = q.Where("category_id = ? OR category_id IN (?)", categoryID, sub)
	}
	q.Model(&model.Supplier{}).Count(&total)
	q.Offset((page - 1) * pageSize).Limit(pageSize).Order("created_at DESC").Find(&list)
	for i := range list {
		fillSupplierCategoryStr(&list[i])
	}
	return list, total
}

func (r *SupplierRepository) ListAll(ctx context.Context) ([]model.Supplier, error) {
	var list []model.Supplier
	err := r.Scoped(ctx).Where("status = 1").Order("created_at ASC").Find(&list).Error
	for i := range list {
		fillSupplierCategoryStr(&list[i])
	}
	return list, err
}

// ExistsByName 判断同一商户下是否已存在同名供应商（excludeID>0 时排除自身）
func (r *SupplierRepository) ExistsByName(ctx context.Context, name string, excludeID int64) bool {
	tenantID := customContext.GetTenantID(ctx)
	q := r.DB.Model(&model.Supplier{}).Where("tenant_id = ? AND name = ?", tenantID, name)
	if excludeID > 0 {
		q = q.Where("id <> ?", excludeID)
	}
	var cnt int64
	q.Count(&cnt)
	return cnt > 0
}

func (r *SupplierRepository) GetByID(ctx context.Context, id int64) (*model.Supplier, error) {
	var supplier model.Supplier
	tenantID := customContext.GetTenantID(ctx)
	err := r.DB.Where("id = ? AND tenant_id = ?", id, tenantID).First(&supplier).Error
	if err == nil {
		fillSupplierCategoryStr(&supplier)
	}
	return &supplier, err
}

func (r *SupplierRepository) Create(ctx context.Context, supplier *model.Supplier) error {
	return r.Scoped(ctx).Create(supplier).Error
}

func (r *SupplierRepository) Update(ctx context.Context, supplier *model.Supplier) error {
	return r.Scoped(ctx).Save(supplier).Error
}

func (r *SupplierRepository) Delete(ctx context.Context, id int64) error {
	tenantID := customContext.GetTenantID(ctx)
	return r.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("supplier_id = ? AND tenant_id = ?", id, tenantID).Delete(&model.SupplierAddress{}).Error; err != nil {
			return err
		}
		return tx.Where("id = ? AND tenant_id = ?", id, tenantID).Delete(&model.Supplier{}).Error
	})
}

// Addresses 查询供应商收货地址
func (r *SupplierRepository) Addresses(ctx context.Context, supplierID int64) ([]model.SupplierAddress, error) {
	var list []model.SupplierAddress
	err := r.Scoped(ctx).Where("supplier_id = ?", supplierID).Order("is_default DESC, sort ASC, id ASC").Find(&list).Error
	return list, err
}

// ReplaceAddresses 覆盖写入供应商收货地址
func (r *SupplierRepository) ReplaceAddresses(ctx context.Context, supplierID int64, addrs []model.SupplierAddress) error {
	tenantID := customContext.GetTenantID(ctx)
	return r.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("supplier_id = ? AND tenant_id = ?", supplierID, tenantID).Delete(&model.SupplierAddress{}).Error; err != nil {
			return err
		}
		if len(addrs) == 0 {
			return nil
		}
		for i := range addrs {
			addrs[i].ID = 0
			addrs[i].SupplierID = supplierID
			addrs[i].TenantID = tenantID
			addrs[i].Sort = i
		}
		return tx.Create(&addrs).Error
	})
}
