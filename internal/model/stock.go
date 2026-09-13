package model

import "time"

// StockBalance 库存余额
type StockBalance struct {
	ID           int64     `json:"id" gorm:"primaryKey"`
	TenantID     int64     `json:"tenant_id" gorm:"index"`
	ShopID       int64     `json:"shop_id"`
	WarehouseID  int64     `json:"warehouse_id"`
	GoodsID      int64     `json:"goods_id"`
	Quantity     int       `json:"quantity"`
	CostPrice    float64   `json:"cost_price"`
	TotalCost    float64   `json:"total_cost"`
	LastInTime   *time.Time `json:"last_in_time"`
	LastOutTime  *time.Time `json:"last_out_time"`
}

// StockLog 库存流水
type StockLog struct {
	ID          int64     `json:"id" gorm:"primaryKey"`
	TenantID    int64     `json:"tenant_id" gorm:"index"`
	ShopID      int64     `json:"shop_id"`
	WarehouseID int64     `json:"warehouse_id"`
	GoodsID     int64     `json:"goods_id"`
	Type        int8      `json:"type"`
	Quantity    int       `json:"quantity"`
	BeforeStock int       `json:"before_stock"`
	AfterStock  int       `json:"after_stock"`
	RelatedType string    `json:"related_type" gorm:"size:32"`
	RelatedID   int64     `json:"related_id"`
	RelatedNo   string    `json:"related_no" gorm:"size:32"`
	Remark      string    `json:"remark" gorm:"size:255"`
	CreatedBy   int64     `json:"created_by"`
	CreatedAt   time.Time `json:"created_at"`
}

// Account 结算账户
type Account struct {
	ID        int64     `json:"id" gorm:"primaryKey"`
	TenantID  int64     `json:"tenant_id" gorm:"index"`
	Name      string    `json:"name" gorm:"size:128"`
	Type      int8      `json:"type"` // 1=现金 2=银行 3=在线
	Balance   float64   `json:"balance"`
	Sort      int       `json:"sort"`
	Status    int8      `json:"status" gorm:"default:1"`
	CreatedAt time.Time `json:"created_at"`
}

// Receipt 收款单
type Receipt struct {
	ID        int64     `json:"id" gorm:"primaryKey"`
	TenantID  int64     `json:"tenant_id" gorm:"index"`
	ShopID    int64     `json:"shop_id"`
	OrderNo   string    `json:"order_no" gorm:"size:32"`
	CustomerID int64   `json:"customer_id"`
	BillDate  string    `json:"bill_date" gorm:"size:10"`
	Amount    float64   `json:"amount"`
	AccountID int64     `json:"account_id"`
	Status    int8      `json:"status"`
	Remark    string    `json:"remark" gorm:"size:500"`
	CreatedBy int64     `json:"created_by"`
	CreatedAt time.Time `json:"created_at"`
}

// Payment 付款单
type Payment struct {
	ID         int64     `json:"id" gorm:"primaryKey"`
	TenantID   int64     `json:"tenant_id" gorm:"index"`
	ShopID     int64     `json:"shop_id"`
	OrderNo    string    `json:"order_no" gorm:"size:32"`
	SupplierID int64     `json:"supplier_id"`
	BillDate   string    `json:"bill_date" gorm:"size:10"`
	Amount     float64   `json:"amount"`
	AccountID  int64     `json:"account_id"`
	Status     int8      `json:"status"`
	Remark     string    `json:"remark" gorm:"size:500"`
	CreatedBy  int64     `json:"created_by"`
	CreatedAt  time.Time `json:"created_at"`
}
