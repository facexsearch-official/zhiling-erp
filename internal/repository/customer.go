package repository

import (
	"context"
	"pisa_server/internal/model"

	customContext "pisa_server/internal/pkg/context"
)

type CustomerRepository struct {
	BaseRepository
}

func NewCustomerRepository(base BaseRepository) *CustomerRepository {
	return &CustomerRepository{BaseRepository: base}
}

func (r *CustomerRepository) List(ctx context.Context, page, pageSize int, keyword string) ([]model.Customer, int64) {
	var total int64
	var list []model.Customer
	q := r.Scoped(ctx)
	if keyword != "" {
		q = q.Where("name LIKE ? OR code LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}
	q.Model(&model.Customer{}).Count(&total)
	q.Offset((page - 1) * pageSize).Limit(pageSize).Order("created_at DESC").Find(&list)
	return list, total
}

func (r *CustomerRepository) ListAll(ctx context.Context) ([]model.Customer, error) {
	var list []model.Customer
	err := r.Scoped(ctx).Where("status = 1").Order("created_at ASC").Find(&list).Error
	return list, err
}

func (r *CustomerRepository) GetByID(ctx context.Context, id int64) (*model.Customer, error) {
	var customer model.Customer
	tenantID := customContext.GetTenantID(ctx)
	err := r.DB.Where("id = ? AND tenant_id = ?", id, tenantID).First(&customer).Error
	return &customer, err
}

func (r *CustomerRepository) Create(ctx context.Context, customer *model.Customer) error {
	return r.Scoped(ctx).Create(customer).Error
}

func (r *CustomerRepository) Update(ctx context.Context, customer *model.Customer) error {
	return r.Scoped(ctx).Save(customer).Error
}

func (r *CustomerRepository) Delete(ctx context.Context, id int64) error {
	tenantID := customContext.GetTenantID(ctx)
	return r.DB.Where("id = ? AND tenant_id = ?", id, tenantID).Delete(&model.Customer{}).Error
}
