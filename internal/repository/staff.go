package repository

import (
	"context"
	"time"

	"pisa_server/internal/model"

	customContext "pisa_server/internal/pkg/context"
)

type StaffRepository struct {
	BaseRepository
}

func NewStaffRepository(base BaseRepository) *StaffRepository {
	return &StaffRepository{BaseRepository: base}
}

// StaffUserRow 用户列表行（user_tenants JOIN users）
type StaffUserRow struct {
	ID        int64     `json:"id"` // user_tenant id
	UserID    int64     `json:"user_id"`
	Name      string    `json:"name"`
	Phone     string    `json:"phone"`
	Role      int8      `json:"role"`
	IsOwner   int8      `json:"is_owner"`
	Status    int8      `json:"status"`
	JoinedAt  time.Time `json:"joined_at"`
	IsCurrent bool      `json:"is_current" gorm:"-"`
}

// ListUsers 返回当前商户的用户列表
func (r *StaffRepository) ListUsers(ctx context.Context) ([]StaffUserRow, error) {
	tenantID := customContext.GetTenantID(ctx)
	var rows []StaffUserRow
	err := r.DB.Table("user_tenants AS ut").
		Select("ut.id AS id, ut.user_id AS user_id, "+
			"COALESCE(NULLIF(ut.staff_name, ''), u.nickname) AS name, "+
			"u.phone AS phone, ut.role AS role, ut.is_owner AS is_owner, "+
			"ut.status AS status, ut.joined_at AS joined_at").
		Joins("JOIN users u ON u.id = ut.user_id").
		Where("ut.tenant_id = ?", tenantID).
		Order("ut.is_owner DESC, ut.id ASC").
		Scan(&rows).Error
	return rows, err
}

func (r *StaffRepository) GetUserTenant(ctx context.Context, id int64) (*model.UserTenant, error) {
	var ut model.UserTenant
	err := r.Scoped(ctx).Where("id = ?", id).First(&ut).Error
	if err != nil {
		return nil, err
	}
	return &ut, nil
}

func (r *StaffRepository) UpdateUserTenant(ctx context.Context, ut *model.UserTenant) error {
	return r.Scoped(ctx).Save(ut).Error
}

func (r *StaffRepository) DeleteUserTenant(ctx context.Context, id int64) error {
	tenantID := customContext.GetTenantID(ctx)
	return r.DB.Where("id = ? AND tenant_id = ?", id, tenantID).Delete(&model.UserTenant{}).Error
}

func (r *StaffRepository) CountUserTenants(ctx context.Context) int64 {
	var count int64
	r.Scoped(ctx).Model(&model.UserTenant{}).Count(&count)
	return count
}

// CountUserTenantsByUser 统计某用户在全部商户中的成员记录数
func (r *StaffRepository) CountUserTenantsByUser(userID int64) int64 {
	var count int64
	r.DB.Model(&model.UserTenant{}).Where("user_id = ?", userID).Count(&count)
	return count
}

func (r *StaffRepository) FindUserByPhone(phone string) (*model.User, error) {
	var u model.User
	err := r.DB.Where("phone = ?", phone).First(&u).Error
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *StaffRepository) CreateUser(u *model.User) error {
	return r.DB.Create(u).Error
}

func (r *StaffRepository) UpdateUser(u *model.User) error {
	return r.DB.Save(u).Error
}

func (r *StaffRepository) CountUserInTenant(userID, tenantID int64) int64 {
	var count int64
	r.DB.Model(&model.UserTenant{}).Where("user_id = ? AND tenant_id = ?", userID, tenantID).Count(&count)
	return count
}

func (r *StaffRepository) CreateUserTenant(ut *model.UserTenant) error {
	return r.DB.Create(ut).Error
}

func (r *StaffRepository) GetTenant(ctx context.Context, id int64) (*model.Tenant, error) {
	var t model.Tenant
	err := r.DB.Where("id = ?", id).First(&t).Error
	if err != nil {
		return nil, err
	}
	return &t, nil
}

/* ────────── Salesmen ────────── */

func (r *StaffRepository) ListSalesmen(ctx context.Context) ([]model.Salesman, error) {
	tenantID := customContext.GetTenantID(ctx)
	var list []model.Salesman
	err := r.DB.Where("tenant_id = ?", tenantID).Order("id DESC").Find(&list).Error
	if err != nil {
		return list, err
	}
	// 填充门店名
	shops, _ := r.ListShops(ctx)
	shopMap := map[int64]string{}
	for _, s := range shops {
		shopMap[s.ID] = s.Name
	}
	for i := range list {
		list[i].ShopName = shopMap[list[i].ShopID]
	}
	return list, nil
}

func (r *StaffRepository) CreateSalesman(ctx context.Context, s *model.Salesman) error {
	return r.Scoped(ctx).Create(s).Error
}

func (r *StaffRepository) GetSalesman(ctx context.Context, id int64) (*model.Salesman, error) {
	var s model.Salesman
	err := r.Scoped(ctx).Where("id = ?", id).First(&s).Error
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *StaffRepository) UpdateSalesman(ctx context.Context, s *model.Salesman) error {
	return r.Scoped(ctx).Save(s).Error
}

func (r *StaffRepository) DeleteSalesman(ctx context.Context, id int64) error {
	tenantID := customContext.GetTenantID(ctx)
	return r.DB.Where("id = ? AND tenant_id = ?", id, tenantID).Delete(&model.Salesman{}).Error
}

/* ────────── Shops ────────── */

func (r *StaffRepository) ListShops(ctx context.Context) ([]model.Shop, error) {
	tenantID := customContext.GetTenantID(ctx)
	var list []model.Shop
	err := r.DB.Where("tenant_id = ?", tenantID).Order("is_main DESC, id ASC").Find(&list).Error
	return list, err
}

func (r *StaffRepository) GetShop(ctx context.Context, id int64) (*model.Shop, error) {
	var s model.Shop
	tenantID := customContext.GetTenantID(ctx)
	if err := r.DB.Where("id = ? AND tenant_id = ?", id, tenantID).First(&s).Error; err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *StaffRepository) CreateShop(ctx context.Context, s *model.Shop) error {
	return r.Scoped(ctx).Create(s).Error
}

func (r *StaffRepository) UpdateShop(ctx context.Context, s *model.Shop) error {
	return r.Scoped(ctx).Save(s).Error
}

func (r *StaffRepository) DeleteShop(ctx context.Context, id int64) error {
	tenantID := customContext.GetTenantID(ctx)
	return r.DB.Where("id = ? AND tenant_id = ?", id, tenantID).Delete(&model.Shop{}).Error
}

// ClearMainShop 将当前商户其它门店的主店标记清除
func (r *StaffRepository) ClearMainShop(ctx context.Context, exceptID int64) {
	tenantID := customContext.GetTenantID(ctx)
	r.DB.Model(&model.Shop{}).Where("tenant_id = ? AND id <> ?", tenantID, exceptID).Update("is_main", 0)
}
