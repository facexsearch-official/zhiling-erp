package repository

import (
	"context"
	"pisa_server/internal/model"
)

type PurchaseRepository struct {
	BaseRepository
}

func NewPurchaseRepository(base BaseRepository) *PurchaseRepository {
	return &PurchaseRepository{BaseRepository: base}
}

func (r *PurchaseRepository) List(ctx context.Context, page, pageSize int) ([]model.Purchase, int64) {
	var total int64
	var list []model.Purchase
	q := r.ScopedShard(ctx, "purchases")
	q.Count(&total)
	q.Offset((page - 1) * pageSize).Limit(pageSize).Order("created_at DESC").Find(&list)
	return list, total
}

func (r *PurchaseRepository) GetByID(ctx context.Context, id int64) (*model.Purchase, error) {
	var purchase model.Purchase
	err := r.ScopedShard(ctx, "purchases").Where("id = ?", id).First(&purchase).Error
	return &purchase, err
}

func (r *PurchaseRepository) Create(ctx context.Context, purchase *model.Purchase) error {
	return r.DB.Table(r.GetShardTable(ctx, "purchases")).Create(purchase).Error
}

func (r *PurchaseRepository) Update(ctx context.Context, purchase *model.Purchase) error {
	return r.DB.Table(r.GetShardTable(ctx, "purchases")).Save(purchase).Error
}
