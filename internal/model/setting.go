package model

import "time"

// TenantSetting 商户系统设置（键值 JSON）
type TenantSetting struct {
	ID        int64     `json:"id" gorm:"primaryKey"`
	TenantID  int64     `json:"tenant_id" gorm:"uniqueIndex"`
	Data      string    `json:"data" gorm:"type:text"`
	UpdatedAt time.Time `json:"updated_at"`
}

// UserPreference 用户偏好设置（按账号生效，键值 JSON）
type UserPreference struct {
	ID        int64     `json:"id" gorm:"primaryKey"`
	UserID    int64     `json:"user_id" gorm:"uniqueIndex:uk_user_tenant_pref"`
	TenantID  int64     `json:"tenant_id" gorm:"uniqueIndex:uk_user_tenant_pref"`
	Data      string    `json:"data" gorm:"type:text"`
	UpdatedAt time.Time `json:"updated_at"`
}
