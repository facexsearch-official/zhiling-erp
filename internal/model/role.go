package model

import "time"

// Role 角色（自定义角色权限）
type Role struct {
	ID            int64     `json:"id" gorm:"primaryKey"`
	TenantID      int64     `json:"tenant_id" gorm:"index"`
	Name          string    `json:"name" gorm:"size:64"`
	Description   string    `json:"description" gorm:"size:255"`
	Permissions   string    `json:"permissions" gorm:"type:text"`
	SensitiveData string    `json:"sensitive_data" gorm:"type:text"`
	IsSystem      int8      `json:"is_system" gorm:"default:0"`
	Status        int8      `json:"status" gorm:"default:1"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`

	EmployeeCount int64 `json:"employee_count" gorm:"-"`
}
