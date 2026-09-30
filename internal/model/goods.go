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
	ID        int64     `json:"id" gorm:"primaryKey"`
	IDStr     string    `json:"id_str" gorm:"->"`
	TenantID  int64     `json:"tenant_id" gorm:"index"`
	Name      string    `json:"name" gorm:"size:20"`
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
	ID        int64     `json:"id" gorm:"primaryKey"`
	TenantID  int64     `json:"tenant_id" gorm:"index"`
	Name      string    `json:"name" gorm:"size:128"`
	Type      int8      `json:"type"` // 1=普通 2=原料仓 3=成品仓
	Address   string    `json:"address" gorm:"size:255"`
	Keeper    string    `json:"keeper" gorm:"size:64"`
	ShopID    int64     `json:"shop_id" gorm:"index"`
	Contact   string    `json:"contact" gorm:"size:64"`
	Phone     string    `json:"phone" gorm:"size:20"`
	Remark    string    `json:"remark" gorm:"size:255"`
	Sort      int       `json:"sort"`
	Status    int8      `json:"status" gorm:"default:1"`
	CreatedAt time.Time `json:"created_at"`

	ShopName string `json:"shop_name" gorm:"->"`
}

// Goods 商品
type Goods struct {
	ID             int64     `json:"id" gorm:"primaryKey"`
	IDStr          string    `json:"id_str" gorm:"->"`
	TenantID       int64     `json:"tenant_id" gorm:"index"`
	Name           string    `json:"name" gorm:"size:255"`
	Code           string    `json:"code" gorm:"size:64"`
	Barcode        string    `json:"barcode" gorm:"size:64"`
	CategoryID     *FlexInt64 `json:"category_id"`
	CategoryIDStr  string    `json:"category_id_str" gorm:"->"`
	UnitID         *FlexInt64 `json:"unit_id"`
	UnitIDStr      string    `json:"unit_id_str" gorm:"->"`
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
	SafetyStock      int     `json:"safety_stock"`
	HasBatch         int8    `json:"has_batch" gorm:"default:0"`
	HasShelfLife     int8    `json:"has_shelf_life" gorm:"default:0"`
	ShelfLifeDays    int     `json:"shelf_life_days"`
	ExpiryAlert      int8    `json:"expiry_alert" gorm:"default:0"`
	ExpiryWarnDays   int     `json:"expiry_warn_days"`
	HasSerial        int8    `json:"has_serial" gorm:"default:0"`
	InitCost         float64 `json:"init_cost"`
	Images           string  `json:"images" gorm:"type:text"`        // JSON 数组
	SpecGroups       string  `json:"spec_groups" gorm:"type:text"`   // JSON: [{name,has_image,values:[{name,image}]}]
	PriceColumns     string  `json:"price_columns" gorm:"type:text"` // JSON: [自定义价格等级名]

	// ── 冗余聚合列（由子表重算，列表页直接读） ──
	MainUnit   string  `json:"main_unit" gorm:"size:32"`
	RetailMin  float64 `json:"retail_min"`
	RetailMax  float64 `json:"retail_max"`
	TotalStock int     `json:"total_stock"`

	// 非持久化：随详情 / 列表一起返回
	Units         []GoodsUnit        `json:"units" gorm:"-"`
	Prices        []GoodsPrice       `json:"prices" gorm:"-"`
	Stocks        []GoodsStock       `json:"stocks" gorm:"-"`
	UnitSummaries []GoodsUnitSummary `json:"units_summary" gorm:"-"`
}

// GoodsUnit 货品多单位（单位定义，价格/条码见 goods_prices）
type GoodsUnit struct {
	ID       int64   `json:"id" gorm:"primaryKey"`
	TenantID int64   `json:"tenant_id" gorm:"index"`
	GoodsID  int64   `json:"goods_id" gorm:"index"`
	UnitID   *int64  `json:"unit_id"`
	UnitName string  `json:"unit_name" gorm:"size:32"`
	Factor   float64 `json:"factor" gorm:"default:1"`
	IsMain   int8    `json:"is_main" gorm:"default:0"`
	Sort     int     `json:"sort"`
}

// GoodsPrice 货品价格明细（单位 × 规格）
type GoodsPrice struct {
	ID             int64   `json:"id" gorm:"primaryKey"`
	TenantID       int64   `json:"tenant_id" gorm:"index"`
	GoodsID        int64   `json:"goods_id" gorm:"index;uniqueIndex:uk_gp"`
	UnitKey        string  `json:"unit_key" gorm:"size:32;uniqueIndex:uk_gp"`
	SpecKey        string  `json:"spec_key" gorm:"size:255;uniqueIndex:uk_gp"`
	Barcode        string  `json:"barcode" gorm:"size:64"`
	Code           string  `json:"code" gorm:"size:64"`
	PurchasePrice  float64 `json:"purchase_price"`
	RetailPrice    float64 `json:"retail_price"`
	WholesalePrice float64 `json:"wholesale_price"`
	Disabled       int8    `json:"disabled" gorm:"default:0"`
	Custom         string  `json:"custom" gorm:"type:text"` // JSON: {"33":0,"等级1":0}
	Sort           int     `json:"sort"`
}

// GoodsStock 货品库存（规格级，主单位口径）
type GoodsStock struct {
	ID          int64   `json:"id" gorm:"primaryKey"`
	TenantID    int64   `json:"tenant_id" gorm:"index"`
	GoodsID     int64   `json:"goods_id" gorm:"index;uniqueIndex:uk_gs"`
	SpecKey     string  `json:"spec_key" gorm:"size:255;uniqueIndex:uk_gs"`
	Stock       int     `json:"stock"`
	MinStock    int     `json:"min_stock"`
	SafetyStock int     `json:"safety_stock"`
	MaxStock    int     `json:"max_stock"`
	InitCost    float64 `json:"init_cost"`
}

// GoodsUnitSummary 货品列表的单位聚合摘要（非持久化）
type GoodsUnitSummary struct {
	UnitName     string  `json:"unit_name"`
	Factor       float64 `json:"factor"`
	IsMain       int8    `json:"is_main"`
	RetailMin    float64 `json:"retail_min"`
	RetailMax    float64 `json:"retail_max"`
	WholesaleMin float64 `json:"wholesale_min"`
	WholesaleMax float64 `json:"wholesale_max"`
	Codes        string  `json:"codes"` // "33444/33445/..."
}
