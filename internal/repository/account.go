package repository

import (
	"context"
	"pisa_server/internal/model"

	customContext "pisa_server/internal/pkg/context"
)

type AccountRepository struct {
	BaseRepository
}

func NewAccountRepository(base BaseRepository) *AccountRepository {
	return &AccountRepository{BaseRepository: base}
}

func (r *AccountRepository) List(ctx context.Context, page, pageSize int, keyword string) ([]model.Account, int64) {
	var total int64
	var list []model.Account
	q := r.Scoped(ctx)
	if keyword != "" {
		q = q.Where("name LIKE ?", "%"+keyword+"%")
	}
	q.Model(&model.Account{}).Count(&total)
	q.Offset((page - 1) * pageSize).Limit(pageSize).Order("sort ASC, created_at ASC").Find(&list)
	r.fillShopNames(list)
	return list, total
}

func (r *AccountRepository) fillShopNames(list []model.Account) {
	ids := make([]int64, 0, len(list))
	for _, a := range list {
		if a.ShopID > 0 {
			ids = append(ids, a.ShopID)
		}
	}
	if len(ids) == 0 {
		return
	}
	type row struct {
		ID   int64
		Name string
	}
	var rows []row
	r.DB.Table("shops").Select("id, name").Where("id IN ?", ids).Scan(&rows)
	m := make(map[int64]string, len(rows))
	for _, x := range rows {
		m[x.ID] = x.Name
	}
	for i := range list {
		list[i].ShopName = m[list[i].ShopID]
	}
}

func (r *AccountRepository) ListAll(ctx context.Context) ([]model.Account, error) {
	var list []model.Account
	err := r.Scoped(ctx).Where("status = 1").Order("sort ASC").Find(&list).Error
	return list, err
}

func (r *AccountRepository) GetByID(ctx context.Context, id int64) (*model.Account, error) {
	var account model.Account
	tenantID := customContext.GetTenantID(ctx)
	err := r.DB.Where("id = ? AND tenant_id = ?", id, tenantID).First(&account).Error
	return &account, err
}

func (r *AccountRepository) Create(ctx context.Context, account *model.Account) error {
	return r.Scoped(ctx).Create(account).Error
}

func (r *AccountRepository) Update(ctx context.Context, account *model.Account) error {
	return r.Scoped(ctx).Save(account).Error
}

func (r *AccountRepository) Delete(ctx context.Context, id int64) error {
	tenantID := customContext.GetTenantID(ctx)
	return r.DB.Where("id = ? AND tenant_id = ?", id, tenantID).Delete(&model.Account{}).Error
}
