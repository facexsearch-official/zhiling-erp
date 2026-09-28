package model

import "time"

// Plan 套餐
type Plan struct {
	ID        int64     `json:"id" gorm:"primaryKey"`
	Code      string    `json:"code" gorm:"uniqueIndex;size:32"`
	Name      string    `json:"name" gorm:"size:64"`
	Price     float64   `json:"price"`
	MaxGoods  int       `json:"max_goods"`
	MaxStaff  int       `json:"max_staff"`
	MaxShops  int       `json:"max_shops"`
	MaxOrders int       `json:"max_orders"`
	Sort      int       `json:"sort"`
	Status    int8      `json:"status" gorm:"default:1"`
	CreatedAt time.Time `json:"created_at"`
}

// Tenant 商户
type Tenant struct {
	ID            int64      `json:"id" gorm:"primaryKey"`
	Name          string     `json:"name" gorm:"size:128"`
	Type          int8       `json:"type"` // 1=个体户 2=有限公司 3=合伙企业
	ContactName   string     `json:"contact_name" gorm:"size:64"`
	ContactPhone  string     `json:"contact_phone" gorm:"size:20"`
	Logo          string     `json:"logo" gorm:"size:255"`
	BusinessMode  int8       `json:"business_mode"` // 1=批发 2=零售 3=批零兼营批发为主 4=批零兼营零售为主
	Industry      string     `json:"industry" gorm:"size:64"`
	Address       string     `json:"address" gorm:"size:255"`
	AddressDetail string     `json:"address_detail" gorm:"size:255"`
	OwnerUserID   int64      `json:"owner_user_id"`
	PlanID        int64      `json:"plan_id"`
	PlanExpiresAt *time.Time `json:"plan_expires_at"`
	MaxGoods      int        `json:"max_goods"`
	MaxStaff      int        `json:"max_staff"`
	MaxShops      int        `json:"max_shops"`
	MaxOrders     int        `json:"max_orders"`
	Status        int8       `json:"status" gorm:"default:1"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

// User 用户
type User struct {
	ID              int64      `json:"id" gorm:"primaryKey"`
	Phone           string     `json:"phone" gorm:"uniqueIndex;size:20"`
	PasswordHash    string     `json:"-" gorm:"size:255"`
	Nickname        string     `json:"nickname" gorm:"size:64"`
	Avatar          string     `json:"avatar" gorm:"size:255"`
	DefaultTenantID *int64     `json:"default_tenant_id"`
	Status          int8       `json:"status" gorm:"default:1"`
	LastLoginAt     *time.Time `json:"last_login_at"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

// UserTenant 用户-商户关系
type UserTenant struct {
	ID          int64     `json:"id" gorm:"primaryKey"`
	UserID      int64     `json:"user_id" gorm:"uniqueIndex:uk_user_tenant"`
	TenantID    int64     `json:"tenant_id" gorm:"uniqueIndex:uk_user_tenant"`
	IsOwner     int8      `json:"is_owner"`
	Role        int8      `json:"role"` // 1=超级管理员 2=管理员 3=操作员
	RoleID      int64     `json:"role_id" gorm:"index"`
	StaffName   string    `json:"staff_name" gorm:"size:64"`
	StaffPhone  string    `json:"staff_phone" gorm:"size:20"`
	Permissions string    `json:"permissions" gorm:"type:text"`
	InvitedBy   *int64    `json:"invited_by"`
	Status      int8      `json:"status" gorm:"default:1"`
	JoinedAt    time.Time `json:"joined_at"`
}

// Shop 门店
type Shop struct {
	ID            int64     `json:"id" gorm:"primaryKey"`
	TenantID      int64     `json:"tenant_id" gorm:"index"`
	Name          string    `json:"name" gorm:"size:128"`
	Address       string    `json:"address" gorm:"size:255"`
	Phone         string    `json:"phone" gorm:"size:20"`
	Logo          string    `json:"logo" gorm:"size:255"`
	AddressDetail string    `json:"address_detail" gorm:"size:255"`
	Type          int8      `json:"type"` // 1=直营 2=加盟
	Remark        string    `json:"remark" gorm:"size:255"`
	IsMain        int8      `json:"is_main"`
	Status        int8      `json:"status" gorm:"default:1"`
	CreatedAt     time.Time `json:"created_at"`
}

// Salesman 业务员（不可登录系统，仅用于统计销售业绩）
type Salesman struct {
	ID        int64     `json:"id" gorm:"primaryKey"`
	TenantID  int64     `json:"tenant_id" gorm:"index"`
	Name      string    `json:"name" gorm:"size:64"`
	Phone     string    `json:"phone" gorm:"size:20"`
	ShopID    int64     `json:"shop_id"`
	Status    int8      `json:"status" gorm:"default:1"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	ShopName string `json:"shop_name" gorm:"-"`
}
