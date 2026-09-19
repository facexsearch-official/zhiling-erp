package repository

import (
	"context"

	"pisa_server/internal/model"

	customContext "pisa_server/internal/pkg/context"
)

// PropertyRepository 货品属性（品牌/产地/材质等）
type PropertyRepository struct {
	BaseRepository
}

func NewPropertyRepository(base BaseRepository) *PropertyRepository {
	return &PropertyRepository{BaseRepository: base}
}

// ListAll 返回当前商户全部货品属性
func (r *PropertyRepository) ListAll(ctx context.Context) ([]model.GoodsProperty, error) {
	var list []model.GoodsProperty
	err := r.Scoped(ctx).Where("status = 1").Order("sort ASC, id ASC").Find(&list).Error
	return list, err
}

// Count 统计当前商户货品属性数量
func (r *PropertyRepository) Count(ctx context.Context) (int64, error) {
	var n int64
	err := r.Scoped(ctx).Model(&model.GoodsProperty{}).Where("status = 1").Count(&n).Error
	return n, err
}

// GetByID 查询单个货品属性
func (r *PropertyRepository) GetByID(ctx context.Context, id int64) (*model.GoodsProperty, error) {
	var p model.GoodsProperty
	err := r.Scoped(ctx).Where("id = ?", id).First(&p).Error
	return &p, err
}

// Create 新建货品属性
func (r *PropertyRepository) Create(ctx context.Context, p *model.GoodsProperty) error {
	return r.Scoped(ctx).Create(p).Error
}

// Save 更新货品属性
func (r *PropertyRepository) Save(ctx context.Context, p *model.GoodsProperty) error {
	return r.Scoped(ctx).Save(p).Error
}

// Delete 删除货品属性
func (r *PropertyRepository) Delete(ctx context.Context, id int64) error {
	tenantID := customContext.GetTenantID(ctx)
	return r.DB.Where("id = ? AND tenant_id = ?", id, tenantID).Delete(&model.GoodsProperty{}).Error
}
