package repository

import (
	"context"
	"pisa_server/internal/model"

	customContext "pisa_server/internal/pkg/context"
)

type WarehouseRepository struct {
	BaseRepository
}

func NewWarehouseRepository(base BaseRepository) *WarehouseRepository {
	return &WarehouseRepository{BaseRepository: base}
}

func (r *WarehouseRepository) List(ctx context.Context, page, pageSize int, keyword string) ([]model.Warehouse, int64) {
	tenantID := customContext.GetTenantID(ctx)
	var total int64
	var list []model.Warehouse
	countQ := r.DB.Table("warehouses w").Where("w.tenant_id = ?", tenantID)
	if keyword != "" {
		countQ = countQ.Where("w.name LIKE ?", "%"+keyword+"%")
	}
	countQ.Count(&total)
	q := r.DB.Table("warehouses w").
		Select("w.*, COALESCE(s.name, '') AS shop_name").
		Joins("LEFT JOIN shops s ON s.id = w.shop_id").
		Where("w.tenant_id = ?", tenantID)
	if keyword != "" {
		q = q.Where("w.name LIKE ?", "%"+keyword+"%")
	}
	q.Order("w.sort ASC, w.created_at ASC").Offset((page - 1) * pageSize).Limit(pageSize).Scan(&list)
	return list, total
}

func (r *WarehouseRepository) ListAll(ctx context.Context) ([]model.Warehouse, error) {
	var list []model.Warehouse
	err := r.Scoped(ctx).Where("status = 1").Order("sort ASC, created_at ASC").Find(&list).Error
	return list, err
}

func (r *WarehouseRepository) GetByID(ctx context.Context, id int64) (*model.Warehouse, error) {
	var warehouse model.Warehouse
	tenantID := customContext.GetTenantID(ctx)
	err := r.DB.Table("warehouses w").
		Select("w.*, COALESCE(s.name, '') AS shop_name").
		Joins("LEFT JOIN shops s ON s.id = w.shop_id").
		Where("w.id = ? AND w.tenant_id = ?", id, tenantID).
		Scan(&warehouse).Error
	return &warehouse, err
}

func (r *WarehouseRepository) Create(ctx context.Context, warehouse *model.Warehouse) error {
	return r.Scoped(ctx).Create(warehouse).Error
}

func (r *WarehouseRepository) Update(ctx context.Context, warehouse *model.Warehouse) error {
	return r.Scoped(ctx).Save(warehouse).Error
}

func (r *WarehouseRepository) Delete(ctx context.Context, id int64) error {
	tenantID := customContext.GetTenantID(ctx)
	return r.DB.Where("id = ? AND tenant_id = ?", id, tenantID).Delete(&model.Warehouse{}).Error
}
