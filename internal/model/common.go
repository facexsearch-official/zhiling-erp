package model

import "time"

// Common
type BaseModel struct {
	ID        int64     `json:"id" gorm:"primaryKey"`
	TenantID  int64     `json:"tenant_id" gorm:"index"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Pagination
type PageRequest struct {
	Page     int    `json:"page" form:"page"`
	PageSize int    `json:"page_size" form:"page_size"`
	Keyword  string `json:"keyword" form:"keyword"`
}

func (p *PageRequest) GetPage() int {
	if p.Page < 1 { return 1 }
	return p.Page
}
func (p *PageRequest) GetPageSize() int {
	if p.PageSize < 1 || p.PageSize > 100 { return 20 }
	return p.PageSize
}
