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

func (r *PurchaseRepository) List(ctx context.Context, page, pageSize int, supplierID int64, status int8, keyword string, dateFrom, dateTo string) ([]model.Purchase, int64) {
	var total int64
	var list []model.Purchase
	q := r.Scoped(ctx).Table("purchases")
	if supplierID > 0 {
		q = q.Where("supplier_id = ?", supplierID)
	}
	if status > 0 {
		q = q.Where("status = ?", status)
	}
	if keyword != "" {
		q = q.Where("order_no LIKE ?", "%"+keyword+"%")
	}
	if dateFrom != "" {
		q = q.Where("bill_date >= ?", dateFrom)
	}
	if dateTo != "" {
		q = q.Where("bill_date <= ?", dateTo)
	}
	q.Model(&model.Purchase{}).Count(&total)
	q.Offset((page - 1) * pageSize).Limit(pageSize).Order("created_at DESC").Find(&list)
	return list, total
}

func (r *PurchaseRepository) GetByID(ctx context.Context, id int64) (*model.Purchase, error) {
	var purchase model.Purchase
	err := r.Scoped(ctx).Table("purchases").Where("id = ?", id).First(&purchase).Error
	return &purchase, err
}

func (r *PurchaseRepository) Create(ctx context.Context, purchase *model.Purchase) error {
	return r.Scoped(ctx).Table("purchases").Create(purchase).Error
}

func (r *PurchaseRepository) Update(ctx context.Context, purchase *model.Purchase) error {
	return r.Scoped(ctx).Table("purchases").Save(purchase).Error
}

func (r *PurchaseRepository) Delete(ctx context.Context, id int64) error {
	return r.Scoped(ctx).Table("purchases").Where("id = ?", id).Delete(&model.Purchase{}).Error
}

func (r *PurchaseRepository) FillSupplierNames(ctx context.Context, purchases []model.Purchase, supplierRepo *SupplierRepository) {
	if len(purchases) == 0 {
		return
	}
	ids := make([]int64, 0)
	seen := make(map[int64]bool)
	for _, p := range purchases {
		if !seen[p.SupplierID] {
			ids = append(ids, p.SupplierID)
			seen[p.SupplierID] = true
		}
	}
	suppliers, _ := supplierRepo.ListAll(ctx)
	supplierMap := make(map[int64]string)
	for _, s := range suppliers {
		supplierMap[s.ID] = s.Name
	}
	for i := range purchases {
		purchases[i].SupplierName = supplierMap[purchases[i].SupplierID]
	}
}
