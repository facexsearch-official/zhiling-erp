package model

import "time"

// PriceLevel 价格等级
type PriceLevel struct {
	ID        int64     `json:"id" gorm:"primaryKey"`
	TenantID  int64     `json:"tenant_id" gorm:"index"`
	Name      string    `json:"name" gorm:"size:32"`
	Sort      int       `json:"sort"`
	Status    int8      `json:"status" gorm:"default:1"`
	CreatedAt time.Time `json:"created_at"`
}
