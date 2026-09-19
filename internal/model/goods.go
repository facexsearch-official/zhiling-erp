package model

import "time"

// GoodsCategory 商品分类
type GoodsCategory struct {
	ID       int64  `json:"id" gorm:"primaryKey"`
	TenantID int64  `json:"tenant_id" gorm:"index"`
	Name     string `json:"name" gorm:"size:30"`
	ParentID int64  `json:"parent_id" gorm:"index"`
	Sort     int    `json:"sort"`
}

// Unit 单位
type Unit struct {
	ID       int64     `json:"id" gorm:"primaryKey"`
	TenantID int64     `json:"tenant_id" gorm:"index"`
	Name     string    `json:"name" gorm:"size:20"`
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

// GoodsProperty 货品属性（品牌/产地/材质等）
type GoodsProperty struct {
	ID        int64     `json:"id" gorm:"primaryKey"`
	TenantID  int64     `json:"tenant_id" gorm:"index"`
	Name      string    `json:"name" gorm:"size:64"`
	Type      int8      `json:"type" gorm:"default:1"` // 1=选择型 2=输入型
	Values    string    `json:"values" gorm:"type:text"`
	Sort      int       `json:"sort"`
	Status    int8      `json:"status" gorm:"default:1"`
	CreatedAt time.Time `json:"created_at"`
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
	Suppliers      string    `json:"suppliers" gorm:"type:text"` // JSON 数组，存储多个供应商 ID
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

	// ── 扩展字段 ──
	Remark           string  `json:"remark" gorm:"size:500"`
	Spec             string  `json:"spec" gorm:"size:128"`
	Brand            string  `json:"brand" gorm:"size:128"`
	Origin           string  `json:"origin" gorm:"size:255"`
	HasMultiUnit     int8    `json:"has_multi_unit" gorm:"default:0"`
	HasMultiSpec     int8    `json:"has_multi_spec" gorm:"default:0"`
	SalesUnit        string  `json:"sales_unit" gorm:"size:32"`    // 开单默认销售单位
	PurchaseUnit     string  `json:"purchase_unit" gorm:"size:32"` // 开单默认进货单位
	EnableStockAlert int8    `json:"enable_stock_alert" gorm:"default:0"`
	InitCost         float64 `json:"init_cost"`
	Images           string  `json:"images" gorm:"type:text"`      // JSON 数组
	SpecGroups       string  `json:"spec_groups" gorm:"type:text"` // JSON: [{name,has_image,values:[{name,image}]}]
	PriceRows        string  `json:"price_rows" gorm:"type:text"`    // JSON: {unitName:[{code,barcode,purchase_price,retail_price,wholesale_price,disabled}]}
	StockRows        string  `json:"stock_rows" gorm:"type:text"`    // JSON: {specKey:{stock,min_stock,safe_stock,max_stock,init_cost}}
	PriceColumns     string  `json:"price_columns" gorm:"type:text"` // JSON: [自定义价格等级名]

	// 非持久化：随详情一起返回
	Units []GoodsUnit `json:"units" gorm:"-"`
	Specs []GoodsSpec `json:"specs" gorm:"-"`
}

// GoodsUnit 货品多单位
type GoodsUnit struct {
	ID             int64   `json:"id" gorm:"primaryKey"`
	TenantID       int64   `json:"tenant_id" gorm:"index"`
	GoodsID        int64   `json:"goods_id" gorm:"index"`
	UnitID         *int64  `json:"unit_id"`
	UnitName       string  `json:"unit_name" gorm:"size:32"`
	Factor         float64 `json:"factor" gorm:"default:1"`
	Barcode        string  `json:"barcode" gorm:"size:64"`
	PurchasePrice  float64 `json:"purchase_price"`
	RetailPrice    float64 `json:"retail_price"`
	WholesalePrice float64 `json:"wholesale_price"`
	IsMain         int8    `json:"is_main" gorm:"default:0"`
	Sort           int     `json:"sort"`
}

// GoodsSpec 货品多规格（扁平 SKU 列表）
type GoodsSpec struct {
	ID             int64   `json:"id" gorm:"primaryKey"`
	TenantID       int64   `json:"tenant_id" gorm:"index"`
	GoodsID        int64   `json:"goods_id" gorm:"index"`
	Name           string  `json:"name" gorm:"size:128"`
	Code           string  `json:"code" gorm:"size:64"`
	Barcode        string  `json:"barcode" gorm:"size:64"`
	PurchasePrice  float64 `json:"purchase_price"`
	RetailPrice    float64 `json:"retail_price"`
	WholesalePrice float64 `json:"wholesale_price"`
	Stock          int     `json:"stock"`
	Sort           int     `json:"sort"`
}
