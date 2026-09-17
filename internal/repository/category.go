package repository

import (
	"context"
	"pisa_server/internal/model"
)

type CategoryRepository struct {
	BaseRepository
}

func NewCategoryRepository(base BaseRepository) *CategoryRepository {
	return &CategoryRepository{BaseRepository: base}
}

// ListAll 返回当前商户全部分类
func (r *CategoryRepository) ListAll(ctx context.Context) ([]model.GoodsCategory, error) {
	var list []model.GoodsCategory
	err := r.Scoped(ctx).Order("sort ASC, id ASC").Find(&list).Error
	return list, err
}

// Create 新建分类
func (r *CategoryRepository) Create(ctx context.Context, c *model.GoodsCategory) error {
	return r.Scoped(ctx).Create(c).Error
}
