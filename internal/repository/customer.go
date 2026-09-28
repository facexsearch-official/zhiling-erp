package repository

import (
	"context"
	"pisa_server/internal/model"

	"gorm.io/gorm"

	customContext "pisa_server/internal/pkg/context"
)

type CustomerRepository struct {
	BaseRepository
}

func NewCustomerRepository(base BaseRepository) *CustomerRepository {
	return &CustomerRepository{BaseRepository: base}
}

func (r *CustomerRepository) List(ctx context.Context, page, pageSize int, keyword string, categoryID int64, hideDisabled, hideZero bool) ([]model.Customer, int64) {
	var total int64
	var list []model.Customer
	q := r.DB.Table("customers AS c").
		Joins("LEFT JOIN customer_categories cc ON cc.id = c.category_id").
		Joins("LEFT JOIN salesmen s ON s.id = c.salesman_id").
		Where("c.tenant_id = ?", customContext.GetTenantID(ctx))
	if keyword != "" {
		kw := "%" + keyword + "%"
		q = q.Where("c.name LIKE ? OR c.contact LIKE ? OR c.phone LIKE ? OR c.remark LIKE ? OR c.email LIKE ?", kw, kw, kw, kw, kw)
	}
	if categoryID == -1 {
		q = q.Where("c.category_id IS NULL")
	} else if categoryID > 0 {
		q = q.Where("c.category_id = ?", categoryID)
	}
	if hideDisabled {
		q = q.Where("c.status = 1")
	}
	if hideZero {
		q = q.Where("c.total_receivable <> 0")
	}
	q.Session(&gorm.Session{}).Count(&total)
	q.Select("c.*, COALESCE(cc.name,'未分类') AS category_name, COALESCE(s.name,'') AS salesman_name").
		Offset((page - 1) * pageSize).Limit(pageSize).Order("c.created_at DESC").Scan(&list)
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
