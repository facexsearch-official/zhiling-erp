package repository

import (
	"context"
	"pisa_server/internal/model"
)

type PurchaseItemRepository struct {
	BaseRepository
}

func NewPurchaseItemRepository(base BaseRepository) *PurchaseItemRepository {
	return &PurchaseItemRepository{BaseRepository: base}
}

func (r *PurchaseItemRepository) ListByPurchaseID(ctx context.Context, purchaseID int64) ([]model.PurchaseItem, error) {
	var list []model.PurchaseItem
	err := r.Scoped(ctx).Table("purchase_items").Where("purchase_id = ?", purchaseID).Find(&list).Error
	return list, err
}

func (r *PurchaseItemRepository) BatchCreate(ctx context.Context, items []model.PurchaseItem) error {
	if len(items) == 0 {
		return nil
	}
	return r.Scoped(ctx).Table("purchase_items").Create(&items).Error
}

func (r *PurchaseItemRepository) DeleteByPurchaseID(ctx context.Context, purchaseID int64) error {
	return r.Scoped(ctx).Table("purchase_items").Where("purchase_id = ?", purchaseID).Delete(&model.PurchaseItem{}).Error
}
