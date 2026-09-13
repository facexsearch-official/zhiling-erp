package model

import "time"

// GoodsCategory 商品分类
type GoodsCategory struct {
	ID       int64  `json:"id" gorm:"primaryKey"`
	TenantID int64  `json:"tenant_id" gorm:"index"`
	Name     string `json:"name" gorm:"size:128"`
	ParentID int64  `json:"parent_id"`
	Sort     int    `json:"sort"`
}

// Unit 单位
type Unit struct {
	ID       int64     `json:"id" gorm:"primaryKey"`
	TenantID int64     `json:"tenant_id" gorm:"index"`
	Name     string    `json:"name" gorm:"size:32"`
	CreatedAt time.Time `json:"created_at"`
}

// GoodsAttribute 辅助属性
type GoodsAttribute struct {
	ID       int64  `json:"id" gorm:"primaryKey"`
	TenantID int64  `json:"tenant_id" gorm:"index"`
	Name     string `json:"name" gorm:"size:64"`
	Values   string `json:"values" gorm:"type:text"`
	Sort     int    `json:"sort"`
	Status   int8   `json:"status" gorm:"default:1"`
}

// Warehouse 仓库
type Warehouse struct {
	ID       int64     `json:"id" gorm:"primaryKey"`
	TenantID int64     `json:"tenant_id" gorm:"index"`
	Name     string    `json:"name" gorm:"size:128"`
	Type     int8      `json:"type"` // 1=普通 2=原料仓 3=成品仓
	Address  string    `json:"address" gorm:"size:255"`
	Keeper   string    `json:"keeper" gorm:"size:64"`
	Sort     int       `json:"sort"`
	Status   int8      `json:"status" gorm:"default:1"`
	CreatedAt time.Time `json:"created_at"`
}

// Goods 商品
type Goods struct {
	ID             int64     `json:"id" gorm:"primaryKey"`
	TenantID       int64     `json:"tenant_id" gorm:"index"`
	Name           string    `json:"name" gorm:"size:255"`
	Code           string    `json:"code" gorm:"size:64"`
	Barcode        string    `json:"barcode" gorm:"size:64"`
	CategoryID     *int64    `json:"category_id"`
	UnitID         *int64    `json:"unit_id"`
	SupplierID     *int64    `json:"supplier_id"`
	ImageURL       string    `json:"image_url" gorm:"size:255"`
	CurrentStock   int       `json:"current_stock"`
	MaxStock       int       `json:"max_stock"`
	MinStock       int       `json:"min_stock"`
	PurchasePrice  float64   `json:"purchase_price"`
	RetailPrice    float64   `json:"retail_price"`
	WholesalePrice float64   `json:"wholesale_price"`
	CostMethod     int8      `json:"cost_method"` // 1=加权平均 2=FIFO
	Status         int8      `json:"status" gorm:"default:1"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}
