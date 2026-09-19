package repository

import (
	"context"
	"pisa_server/internal/model"

	customContext "pisa_server/internal/pkg/context"
)

type UnitRepository struct {
	BaseRepository
}

func NewUnitRepository(base BaseRepository) *UnitRepository {
	return &UnitRepository{BaseRepository: base}
}

// ListAll 返回当前商户全部单位
func (r *UnitRepository) ListAll(ctx context.Context) ([]model.Unit, error) {
	var list []model.Unit
	err := r.Scoped(ctx).Order("id ASC").Find(&list).Error
	return list, err
}

// GetByID 查询单个单位
func (r *UnitRepository) GetByID(ctx context.Context, id int64) (*model.Unit, error) {
	var u model.Unit
	err := r.Scoped(ctx).Where("id = ?", id).First(&u).Error
	return &u, err
}

// Create 新建单位
func (r *UnitRepository) Create(ctx context.Context, u *model.Unit) error {
	return r.Scoped(ctx).Create(u).Error
}

// Save 更新单位
func (r *UnitRepository) Save(ctx context.Context, u *model.Unit) error {
	return r.Scoped(ctx).Save(u).Error
}

// Delete 删除单位
func (r *UnitRepository) Delete(ctx context.Context, id int64) error {
	tenantID := customContext.GetTenantID(ctx)
	return r.DB.Where("id = ? AND tenant_id = ?", id, tenantID).Delete(&model.Unit{}).Error
}

// HasGoods 判断单位是否被货品引用（主单位或多单位）
func (r *UnitRepository) HasGoods(ctx context.Context, unitID int64) bool {
	var n int64
	r.Scoped(ctx).Model(&model.Goods{}).Where("unit_id = ?", unitID).Count(&n)
	if n > 0 {
		return true
	}
	r.Scoped(ctx).Model(&model.GoodsUnit{}).Where("unit_id = ?", unitID).Count(&n)
	return n > 0
}
