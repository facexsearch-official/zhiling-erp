package model

import "time"

// GoodsBatch 商品批次
type GoodsBatch struct {
	ID             int64     `json:"id" gorm:"primaryKey"`
	IDStr          string    `json:"id_str" gorm:"->"`
	TenantID       int64     `json:"tenant_id" gorm:"index"`
	GoodsID        int64     `json:"goods_id" gorm:"index"`
	SpecKey        string    `json:"spec_key" gorm:"size:255"`
	BatchNo        string    `json:"batch_no" gorm:"size:64"`
	ProductionDate string    `json:"production_date" gorm:"size:10"`
	ExpiryDate     string    `json:"expiry_date" gorm:"size:10"`
	Stock          int       `json:"stock"`
	SupplierID     int64     `json:"supplier_id"`
	Brand          string    `json:"brand" gorm:"size:128"`
	CreatedAt      time.Time `json:"created_at"`
}
