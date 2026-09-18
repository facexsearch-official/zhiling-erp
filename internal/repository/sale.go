package repository

import (
	"context"
	"pisa_server/internal/model"
)

type SaleRepository struct {
	BaseRepository
}

func NewSaleRepository(base BaseRepository) *SaleRepository {
	return &SaleRepository{BaseRepository: base}
}

func (r *SaleRepository) List(ctx context.Context, page, pageSize int) ([]model.Sale, int64) {
	var total int64
	var list []model.Sale
	q := r.Scoped(ctx).Table("sales")
	q.Count(&total)
	q.Offset((page - 1) * pageSize).Limit(pageSize).Order("created_at DESC").Find(&list)
	return list, total
}

func (r *SaleRepository) GetByID(ctx context.Context, id int64) (*model.Sale, error) {
	var sale model.Sale
	err := r.Scoped(ctx).Table("sales").Where("id = ?", id).First(&sale).Error
	return &sale, err
}

func (r *SaleRepository) Create(ctx context.Context, sale *model.Sale) error {
	return r.Scoped(ctx).Table("sales").Create(sale).Error
}

func (r *SaleRepository) Update(ctx context.Context, sale *model.Sale) error {
	return r.Scoped(ctx).Table("sales").Save(sale).Error
}

func (r *SaleRepository) Delete(ctx context.Context, id int64) error {
	return r.Scoped(ctx).Table("sales").Where("id = ?", id).Delete(&model.Sale{}).Error
}

func (r *SaleRepository) CountByDate(ctx context.Context, date string) (int, float64) {
	var result struct {
		Count int
		Total float64
	}
	r.Scoped(ctx).Table("sales").
		Select("COUNT(*) as count, COALESCE(SUM(total_amount),0) as total").
		Where("bill_date = ? AND status != 4", date).
		Scan(&result)
	return result.Count, result.Total
}
