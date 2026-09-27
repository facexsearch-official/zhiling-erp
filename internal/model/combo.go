package model

import "time"

// Combo 套餐
type Combo struct {
	ID             int64       `json:"id" gorm:"primaryKey"`
	IDStr          string      `json:"id_str" gorm:"->"`
	TenantID       int64       `json:"tenant_id" gorm:"index"`
	Name           string      `json:"name" gorm:"size:255"`
	Barcode        string      `json:"barcode" gorm:"size:64"`
	UnitID         *int64      `json:"unit_id"`
	UnitName       string      `json:"unit_name" gorm:"size:32"`
	CategoryID     *int64      `json:"category_id"`
	ImageURL       string      `json:"image_url" gorm:"size:255"`
	TotalPurchase  float64     `json:"total_purchase"`
	TotalRetail    float64     `json:"total_retail"`
	TotalWholesale float64     `json:"total_wholesale"`
	RetailPrice    float64     `json:"retail_price"`
	WholesalePrice float64     `json:"wholesale_price"`
	PriceColumns   string      `json:"price_columns" gorm:"type:text"`
	CustomPrices   string      `json:"custom_prices" gorm:"type:text"`
	SoldCount      int         `json:"sold_count"`
	Status         int8        `json:"status" gorm:"default:1"`
	Remark         string      `json:"remark" gorm:"size:500"`
	CreatedAt      time.Time   `json:"created_at"`
	UpdatedAt      time.Time   `json:"updated_at"`
	Items          []ComboItem `json:"items" gorm:"foreignKey:ComboID"`
	CategoryName   string      `json:"category_name" gorm:"->"`
	Summary        string      `json:"summary" gorm:"-"`
}

// ComboItem 套餐组成明细
type ComboItem struct {
	ID             int64     `json:"id" gorm:"primaryKey"`
	TenantID       int64     `json:"tenant_id" gorm:"index"`
	ComboID        int64     `json:"combo_id" gorm:"index"`
	GoodsID        int64     `json:"goods_id"`
	Quantity       int       `json:"quantity"`
	PurchasePrice  float64   `json:"purchase_price"`
	RetailPrice    float64   `json:"retail_price"`
	WholesalePrice float64   `json:"wholesale_price"`
	Remark         string    `json:"remark" gorm:"size:255"`
	CreatedAt      time.Time `json:"created_at"`
	GoodsName      string    `json:"goods_name" gorm:"->"`
	GoodsCode      string    `json:"goods_code" gorm:"->"`
	UnitName       string    `json:"unit_name" gorm:"->"`
	Spec           string    `json:"spec" gorm:"->"`
	Barcode        string    `json:"barcode" gorm:"->"`
	ImageURL       string    `json:"image_url" gorm:"->"`
}
