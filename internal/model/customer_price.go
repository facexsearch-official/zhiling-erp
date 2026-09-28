package model

import "time"

// CustomerPrice 客户报价（客户 × 商品 的最新报价）
type CustomerPrice struct {
	ID         int64     `json:"id" gorm:"primaryKey"`
	TenantID   int64     `json:"tenant_id" gorm:"index"`
	CustomerID int64     `json:"customer_id" gorm:"index"`
	GoodsID    int64     `json:"goods_id" gorm:"index"`
	Price      float64   `json:"price"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}
