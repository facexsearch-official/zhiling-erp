package repository

import (
	"context"
	"pisa_server/internal/model"
)

type PurchaseReturnRepository struct {
	BaseRepository
}

func NewPurchaseReturnRepository(base BaseRepository) *PurchaseReturnRepository {
	return &PurchaseReturnRepository{BaseRepository: base}
}

func (r *PurchaseReturnRepository) List(ctx context.Context, page, pageSize int, supplierID int64, status int8, keyword string) ([]model.PurchaseReturn, int64) {
	var total int64
	var list []model.PurchaseReturn
	q := r.Scoped(ctx).Table("purchase_returns")
	if supplierID > 0 {
		q = q.Where("supplier_id = ?", supplierID)
	}
	if status > 0 {
		q = q.Where("status = ?", status)
	}
	if keyword != "" {
		q = q.Where("order_no LIKE ?", "%"+keyword+"%")
	}
	q.Model(&model.PurchaseReturn{}).Count(&total)
	q.Offset((page - 1) * pageSize).Limit(pageSize).Order("created_at DESC").Find(&list)
	return list, total
}

func (r *PurchaseReturnRepository) GetByID(ctx context.Context, id int64) (*model.PurchaseReturn, error) {
	var pr model.PurchaseReturn
	err := r.Scoped(ctx).Table("purchase_returns").Where("id = ?", id).First(&pr).Error
	return &pr, err
}

func (r *PurchaseReturnRepository) Create(ctx context.Context, pr *model.PurchaseReturn) error {
	return r.Scoped(ctx).Table("purchase_returns").Create(pr).Error
}

func (r *PurchaseReturnRepository) Update(ctx context.Context, pr *model.PurchaseReturn) error {
	return r.Scoped(ctx).Table("purchase_returns").Save(pr).Error
}

func (r *PurchaseReturnRepository) Delete(ctx context.Context, id int64) error {
	return r.Scoped(ctx).Table("purchase_returns").Where("id = ?", id).Delete(&model.PurchaseReturn{}).Error
}
