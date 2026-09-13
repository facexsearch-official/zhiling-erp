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

// List 查询销货单列表（自动路由到分片表）
func (r *SaleRepository) List(ctx context.Context, page, pageSize int) ([]model.Sale, int64) {
	var total int64
	var list []model.Sale

	q := r.ScopedShard(ctx, "sales")
	q.Count(&total)
	q.Offset((page - 1) * pageSize).Limit(pageSize).Order("created_at DESC").Find(&list)
	return list, total
}

// GetByID 根据ID查询单条
func (r *SaleRepository) GetByID(ctx context.Context, id int64) (*model.Sale, error) {
	var sale model.Sale
	err := r.ScopedShard(ctx, "sales").Where("id = ?", id).First(&sale).Error
	return &sale, err
}

// Create 新增销货单（写入分片表）
func (r *SaleRepository) Create(ctx context.Context, sale *model.Sale) error {
	tableName := r.GetShardTable(ctx, "sales")
	return r.DB.Table(tableName).Create(sale).Error
}

// Update 更新销货单
func (r *SaleRepository) Update(ctx context.Context, sale *model.Sale) error {
	tableName := r.GetShardTable(ctx, "sales")
	return r.DB.Table(tableName).Save(sale).Error
}

// Delete 删除销货单
func (r *SaleRepository) Delete(ctx context.Context, id int64) error {
	tableName := r.GetShardTable(ctx, "sales")
	return r.DB.Table(tableName).Where("id = ?", id).Delete(&model.Sale{}).Error
}

// CountByDate 统计某天的销货单数和金额
func (r *SaleRepository) CountByDate(ctx context.Context, date string) (int, float64) {
	var result struct {
		Count int
		Total float64
	}
	r.ScopedShard(ctx, "sales").
		Select("COUNT(*) as count, COALESCE(SUM(total_amount),0) as total").
		Where("bill_date = ? AND status != 4", date).
		Scan(&result)
	return result.Count, result.Total
}
