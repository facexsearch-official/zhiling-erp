package model

import "time"

// Assembly 组装单
type Assembly struct {
	ID            int64          `json:"id" gorm:"primaryKey"`
	IDStr         string         `json:"id_str" gorm:"->"`
	TenantID      int64          `json:"tenant_id" gorm:"index"`
	OrderNo       string         `json:"order_no" gorm:"size:32"`
	Type          int8           `json:"type" gorm:"default:1"` // 1=组装 2=拆分
	SalesmanID    int64          `json:"salesman_id"`
	BillDate      string         `json:"bill_date" gorm:"size:10"`
	AssemblyFee   float64        `json:"assembly_fee"`
	AccountID     int64          `json:"account_id"`
	Status        int8           `json:"status" gorm:"default:1"` // 1=正常 9=作废
	Remark        string         `json:"remark" gorm:"size:500"`
	CreatedBy     int64          `json:"created_by"`
	CreatedAt     time.Time      `json:"created_at"`
	Items         []AssemblyItem `json:"items" gorm:"foreignKey:AssemblyID"`
	SalesmanName  string         `json:"salesman_name" gorm:"->"`
	AccountName   string         `json:"account_name" gorm:"->"`
	MakerName     string         `json:"maker_name" gorm:"->"`
	ComponentName string         `json:"component_name" gorm:"-"`
	ComponentQty  float64        `json:"component_qty" gorm:"-"`
}

// AssemblyItem 组装明细 kind:1=组合件 2=子件
type AssemblyItem struct {
	ID         int64     `json:"id" gorm:"primaryKey"`
	TenantID   int64     `json:"tenant_id" gorm:"index"`
	AssemblyID int64     `json:"assembly_id" gorm:"index"`
	Kind       int8      `json:"kind"`
	GoodsID    int64     `json:"goods_id"`
	Quantity   float64   `json:"quantity"`
	UnitCost   float64   `json:"unit_cost"`
	Amount     float64   `json:"amount"`
	Remark     string    `json:"remark" gorm:"size:255"`
	CreatedAt  time.Time `json:"created_at"`
	GoodsName  string    `json:"goods_name" gorm:"->"`
	GoodsCode  string    `json:"goods_code" gorm:"->"`
	UnitName   string    `json:"unit_name" gorm:"->"`
	Spec       string    `json:"spec" gorm:"->"`
	Barcode    string    `json:"barcode" gorm:"->"`
	ImageURL   string    `json:"image_url" gorm:"->"`
}

// Recipe 配方
type Recipe struct {
	ID         int64        `json:"id" gorm:"primaryKey"`
	IDStr      string       `json:"id_str" gorm:"->"`
	TenantID   int64        `json:"tenant_id" gorm:"index"`
	GoodsID    int64        `json:"goods_id"`
	Quantity   float64      `json:"quantity"`
	Remark     string       `json:"remark" gorm:"size:255"`
	CreatedAt  time.Time    `json:"created_at"`
	Items      []RecipeItem `json:"items" gorm:"foreignKey:RecipeID"`
	GoodsName  string       `json:"goods_name" gorm:"->"`
	GoodsCode  string       `json:"goods_code" gorm:"->"`
	UnitName   string       `json:"unit_name" gorm:"->"`
	Spec       string       `json:"spec" gorm:"->"`
	Barcode    string       `json:"barcode" gorm:"->"`
	ImageURL   string       `json:"image_url" gorm:"->"`
	SubSummary string       `json:"sub_summary" gorm:"-"`
}

// RecipeItem 配方子件
type RecipeItem struct {
	ID        int64     `json:"id" gorm:"primaryKey"`
	TenantID  int64     `json:"tenant_id" gorm:"index"`
	RecipeID  int64     `json:"recipe_id" gorm:"index"`
	GoodsID   int64     `json:"goods_id"`
	Quantity  float64   `json:"quantity"`
	UnitCost  float64   `json:"unit_cost"`
	Remark    string    `json:"remark" gorm:"size:255"`
	CreatedAt time.Time `json:"created_at"`
	GoodsName string    `json:"goods_name" gorm:"->"`
	GoodsCode string    `json:"goods_code" gorm:"->"`
	UnitName  string    `json:"unit_name" gorm:"->"`
	Spec      string    `json:"spec" gorm:"->"`
	Barcode   string    `json:"barcode" gorm:"->"`
	ImageURL  string    `json:"image_url" gorm:"->"`
}
