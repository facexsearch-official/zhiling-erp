package model

import "time"

// Device 登录设备台账（网页端“设备码”= 本地持久化 UUID + 指纹，用于统计每日新增设备）
type Device struct {
	ID          int64     `json:"id" gorm:"primaryKey"`
	TenantID    int64     `json:"tenant_id" gorm:"index;uniqueIndex:uk_device"`
	DeviceID    string    `json:"device_id" gorm:"size:128;uniqueIndex:uk_device"`
	Fingerprint string    `json:"fingerprint" gorm:"size:128"`
	UserID      int64     `json:"user_id" gorm:"index"`
	Phone       string    `json:"phone" gorm:"size:20"`
	IP          string    `json:"ip" gorm:"size:64"`
	UserAgent   string    `json:"user_agent" gorm:"size:255"`
	Platform    string    `json:"platform" gorm:"size:64"`
	FirstSeen   string    `json:"first_seen" gorm:"size:10;index"` // 首次出现日期 YYYY-MM-DD
	LastSeen    time.Time `json:"last_seen"`
	LoginCount  int       `json:"login_count"`
	CreatedAt   time.Time `json:"created_at"`
}
