package repository

import (
	"context"
	"pisa_server/internal/model"
)

type PurchaseReturnItemRepository struct {
	BaseRepository
}

func NewPurchaseReturnItemRepository(base BaseRepository) *PurchaseReturnItemRepository {
	return &PurchaseReturnItemRepository{BaseRepository: base}
}

func (r *PurchaseReturnItemRepository) ListByReturnID(ctx context.Context, returnID int64) ([]model.PurchaseReturnItem, error) {
	var list []model.PurchaseReturnItem
	err := r.ScopedShard(ctx, "purchase_return_items").Where("purchase_return_id = ?", returnID).Find(&list).Error
	return list, err
}

func (r *PurchaseReturnItemRepository) BatchCreate(ctx context.Context, items []model.PurchaseReturnItem) error {
	if len(items) == 0 {
		return nil
	}
	return r.DB.Table(r.GetShardTable(ctx, "purchase_return_items")).Create(&items).Error
}

func (r *PurchaseReturnItemRepository) DeleteByReturnID(ctx context.Context, returnID int64) error {
	return r.ScopedShard(ctx, "purchase_return_items").Where("purchase_return_id = ?", returnID).Delete(&model.PurchaseReturnItem{}).Error
}
