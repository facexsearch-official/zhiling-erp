package model

import "time"

// PurchaseOrder 采购订单
type PurchaseOrder struct {
	ID          int64     `json:"id" gorm:"primaryKey"`
	TenantID    int64     `json:"tenant_id" gorm:"index"`
	ShopID      int64     `json:"shop_id"`
	OrderNo     string    `json:"order_no" gorm:"size:32"`
	SupplierID  int64     `json:"supplier_id"`
	OrderDate   string    `json:"order_date" gorm:"size:10"`
	TotalAmount float64   `json:"total_amount"`
	Status      int8      `json:"status"` // 1=草稿 2=待审核 3=已生效 4=已转换 5=已取消
	Remark      string    `json:"remark" gorm:"size:500"`
	CreatedBy   int64     `json:"created_by"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Purchase 进货单
type Purchase struct {
	ID           int64     `json:"id" gorm:"primaryKey"`
	TenantID     int64     `json:"tenant_id" gorm:"index"`
	ShopID       int64     `json:"shop_id"`
	WarehouseID  int64     `json:"warehouse_id"`
	OrderNo      string    `json:"order_no" gorm:"size:32"`
	SupplierID   int64     `json:"supplier_id"`
	BillDate     string    `json:"bill_date" gorm:"size:10"`
	TotalAmount  float64   `json:"total_amount"`
	PaidAmount   float64   `json:"paid_amount"`
	UnpaidAmount float64   `json:"unpaid_amount"`
	Status       int8      `json:"status"`
	Remark       string    `json:"remark" gorm:"size:500"`
	CreatedBy    int64     `json:"created_by"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// PurchaseReturn 采购退货
type PurchaseReturn struct {
	ID           int64     `json:"id" gorm:"primaryKey"`
	TenantID     int64     `json:"tenant_id" gorm:"index"`
	ShopID       int64     `json:"shop_id"`
	WarehouseID  int64     `json:"warehouse_id"`
	OrderNo      string    `json:"order_no" gorm:"size:32"`
	SupplierID   int64     `json:"supplier_id"`
	BillDate     string    `json:"bill_date" gorm:"size:10"`
	TotalAmount  float64   `json:"total_amount"`
	RefundAmount float64   `json:"refund_amount"`
	Status       int8      `json:"status"`
	Remark       string    `json:"remark" gorm:"size:500"`
	CreatedBy    int64     `json:"created_by"`
	CreatedAt    time.Time `json:"created_at"`
}

// SaleOrder 销售订单
type SaleOrder struct {
	ID          int64     `json:"id" gorm:"primaryKey"`
	TenantID    int64     `json:"tenant_id" gorm:"index"`
	ShopID      int64     `json:"shop_id"`
	OrderNo     string    `json:"order_no" gorm:"size:32"`
	CustomerID  int64     `json:"customer_id"`
	OrderDate   string    `json:"order_date" gorm:"size:10"`
	TotalAmount float64   `json:"total_amount"`
	Status      int8      `json:"status"`
	Remark      string    `json:"remark" gorm:"size:500"`
	CreatedBy   int64     `json:"created_by"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Sale 销货单
type Sale struct {
	ID               int64     `json:"id" gorm:"primaryKey"`
	TenantID         int64     `json:"tenant_id" gorm:"index"`
	ShopID           int64     `json:"shop_id"`
	WarehouseID      int64     `json:"warehouse_id"`
	OrderNo          string    `json:"order_no" gorm:"size:32"`
	CustomerID       int64     `json:"customer_id"`
	BillDate         string    `json:"bill_date" gorm:"size:10"`
	TotalAmount      float64   `json:"total_amount"`
	ReceivedAmount   float64   `json:"received_amount"`
	UnreceivedAmount float64   `json:"unreceived_amount"`
	Status           int8      `json:"status"`
	Remark           string    `json:"remark" gorm:"size:500"`
	CreatedBy        int64     `json:"created_by"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

// SalesReturn 销货退货
type SalesReturn struct {
	ID               int64     `json:"id" gorm:"primaryKey"`
	TenantID         int64     `json:"tenant_id" gorm:"index"`
	ShopID           int64     `json:"shop_id"`
	WarehouseID      int64     `json:"warehouse_id"`
	OrderNo          string    `json:"order_no" gorm:"size:32"`
	CustomerID       int64     `json:"customer_id"`
	BillDate         string    `json:"bill_date" gorm:"size:10"`
	TotalAmount      float64   `json:"total_amount"`
	RefundAmount     float64   `json:"refund_amount"`
	Status           int8      `json:"status"`
	Remark           string    `json:"remark" gorm:"size:500"`
	CreatedBy        int64     `json:"created_by"`
	CreatedAt        time.Time `json:"created_at"`
}
