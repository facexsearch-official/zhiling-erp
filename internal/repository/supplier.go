package repository

import (
	"context"
	"pisa_server/internal/model"

	customContext "pisa_server/internal/pkg/context"
)

type SupplierRepository struct {
	BaseRepository
}

func NewSupplierRepository(base BaseRepository) *SupplierRepository {
	return &SupplierRepository{BaseRepository: base}
}

func (r *SupplierRepository) List(ctx context.Context, page, pageSize int, keyword string) ([]model.Supplier, int64) {
	var total int64
	var list []model.Supplier
	q := r.Scoped(ctx)
	if keyword != "" {
		q = q.Where("name LIKE ? OR code LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}
	q.Model(&model.Supplier{}).Count(&total)
	q.Offset((page - 1) * pageSize).Limit(pageSize).Order("created_at DESC").Find(&list)
	return list, total
}

func (r *SupplierRepository) ListAll(ctx context.Context) ([]model.Supplier, error) {
	var list []model.Supplier
	err := r.Scoped(ctx).Where("status = 1").Order("created_at ASC").Find(&list).Error
	return list, err
}

func (r *SupplierRepository) GetByID(ctx context.Context, id int64) (*model.Supplier, error) {
	var supplier model.Supplier
	tenantID := customContext.GetTenantID(ctx)
	err := r.DB.Where("id = ? AND tenant_id = ?", id, tenantID).First(&supplier).Error
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
	return r.DB.Where("id = ? AND tenant_id = ?", id, tenantID).Delete(&model.Supplier{}).Error
}
