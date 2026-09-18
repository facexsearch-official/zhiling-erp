package repository

import (
	"context"
	"pisa_server/internal/model"
)

type StockLogRepository struct {
	BaseRepository
}

func NewStockLogRepository(base BaseRepository) *StockLogRepository {
	return &StockLogRepository{BaseRepository: base}
}

func (r *StockLogRepository) ListByGoods(ctx context.Context, goodsID int64, page, pageSize int) ([]model.StockLog, int64) {
	var total int64
	var list []model.StockLog
	q := r.Scoped(ctx).Table("stock_logs").Where("goods_id = ?", goodsID)
	q.Count(&total)
	q.Offset((page - 1) * pageSize).Limit(pageSize).Order("created_at DESC").Find(&list)
	return list, total
}

func (r *StockLogRepository) Create(ctx context.Context, log *model.StockLog) error {
	return r.Scoped(ctx).Table("stock_logs").Create(log).Error
}

func (r *StockLogRepository) CountByDate(ctx context.Context, date string) int64 {
	var count int64
	r.Scoped(ctx).Table("stock_logs").
		Where("DATE(created_at) = ?", date).
		Count(&count)
	return count
}
