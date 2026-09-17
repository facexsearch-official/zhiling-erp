package repository

import (
	"context"
	"pisa_server/internal/model"
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

// Create 新建单位
func (r *UnitRepository) Create(ctx context.Context, u *model.Unit) error {
	return r.Scoped(ctx).Create(u).Error
}
