package repository

import (
	"context"
	"pisa_server/internal/model"

	customContext "pisa_server/internal/pkg/context"
)

type GoodsRepository struct {
	BaseRepository
}

func NewGoodsRepository(base BaseRepository) *GoodsRepository {
	return &GoodsRepository{BaseRepository: base}
}

func (r *GoodsRepository) List(ctx context.Context, page, pageSize int, keyword string, categoryID int64) ([]model.Goods, int64) {
	var total int64
	var list []model.Goods
	q := r.Scoped(ctx)
	if keyword != "" {
		q = q.Where("name LIKE ? OR code LIKE ? OR barcode LIKE ?", "%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%")
	}
	if categoryID > 0 {
		q = q.Where("category_id = ?", categoryID)
	}
	q.Model(&model.Goods{}).Count(&total)
	q.Offset((page - 1) * pageSize).Limit(pageSize).Order("created_at DESC").Find(&list)
	return list, total
}

func (r *GoodsRepository) ListAll(ctx context.Context) ([]model.Goods, error) {
	var list []model.Goods
	err := r.Scoped(ctx).Where("status = 1").Order("created_at ASC").Find(&list).Error
	return list, err
}

func (r *GoodsRepository) GetByID(ctx context.Context, id int64) (*model.Goods, error) {
	var goods model.Goods
	tenantID := customContext.GetTenantID(ctx)
	err := r.DB.Where("id = ? AND tenant_id = ?", id, tenantID).First(&goods).Error
	return &goods, err
}

func (r *GoodsRepository) Create(ctx context.Context, goods *model.Goods) error {
	return r.Scoped(ctx).Create(goods).Error
}

func (r *GoodsRepository) Update(ctx context.Context, goods *model.Goods) error {
	return r.Scoped(ctx).Save(goods).Error
}

func (r *GoodsRepository) Delete(ctx context.Context, id int64) error {
	tenantID := customContext.GetTenantID(ctx)
	return r.DB.Where("id = ? AND tenant_id = ?", id, tenantID).Delete(&model.Goods{}).Error
}

func (r *GoodsRepository) ListByIDs(ctx context.Context, ids []int64) ([]model.Goods, error) {
	var list []model.Goods
	tenantID := customContext.GetTenantID(ctx)
	err := r.DB.Where("tenant_id = ? AND id IN ?", tenantID, ids).Find(&list).Error
	return list, err
}
