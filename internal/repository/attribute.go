package repository

import (
	"context"

	"pisa_server/internal/model"
)

// AttributeRepository 货品规格（辅助属性）字典
type AttributeRepository struct {
	BaseRepository
}

func NewAttributeRepository(base BaseRepository) *AttributeRepository {
	return &AttributeRepository{BaseRepository: base}
}

func (r *AttributeRepository) ListAll(ctx context.Context) ([]model.GoodsAttribute, error) {
	var list []model.GoodsAttribute
	err := r.Scoped(ctx).Where("status = 1").Order("sort ASC, id ASC").Find(&list).Error
	return list, err
}

func (r *AttributeRepository) GetByID(ctx context.Context, id int64) (*model.GoodsAttribute, error) {
	var a model.GoodsAttribute
	err := r.Scoped(ctx).Where("id = ?", id).First(&a).Error
	return &a, err
}

func (r *AttributeRepository) Create(ctx context.Context, a *model.GoodsAttribute) error {
	return r.Scoped(ctx).Create(a).Error
}

func (r *AttributeRepository) Save(ctx context.Context, a *model.GoodsAttribute) error {
	return r.Scoped(ctx).Save(a).Error
}
