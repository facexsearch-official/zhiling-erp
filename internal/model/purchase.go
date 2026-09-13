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
	ID           int64           `json:"id" gorm:"primaryKey"`
	TenantID     int64           `json:"tenant_id" gorm:"index"`
	ShopID       int64           `json:"shop_id"`
	WarehouseID  int64           `json:"warehouse_id"`
	OrderNo      string          `json:"order_no" gorm:"size:32"`
	SupplierID   int64           `json:"supplier_id"`
	BillDate     string          `json:"bill_date" gorm:"size:10"`
	TotalAmount  float64         `json:"total_amount"`
	PaidAmount   float64         `json:"paid_amount"`
	UnpaidAmount float64         `json:"unpaid_amount"`
	Status       int8            `json:"status"` // 1=草稿 2=待审核 3=已生效 4=已作废
	Remark       string          `json:"remark" gorm:"size:500"`
	CreatedBy    int64           `json:"created_by"`
	CreatedAt    time.Time       `json:"created_at"`
	UpdatedAt    time.Time       `json:"updated_at"`
	Items        []PurchaseItem  `json:"items" gorm:"foreignKey:PurchaseID"`
	SupplierName string          `json:"supplier_name" gorm:"-"`
	WarehouseName string         `json:"warehouse_name" gorm:"-"`
}

// PurchaseItem 进货单明细（分录行）
type PurchaseItem struct {
	ID         int64     `json:"id" gorm:"primaryKey"`
	TenantID   int64     `json:"tenant_id" gorm:"index"`
	PurchaseID int64     `json:"purchase_id" gorm:"index"`
	GoodsID    int64     `json:"goods_id"`
	Quantity   int       `json:"quantity"`
	UnitPrice  float64   `json:"unit_price"`
	Amount     float64   `json:"amount"`
	Remark     string    `json:"remark" gorm:"size:255"`
	CreatedAt  time.Time `json:"created_at"`
	GoodsName  string    `json:"goods_name" gorm:"-"`
	GoodsCode  string    `json:"goods_code" gorm:"-"`
	UnitName   string    `json:"unit_name" gorm:"-"`
}

// PurchaseReturn 采购退货
type PurchaseReturn struct {
	ID           int64                   `json:"id" gorm:"primaryKey"`
	TenantID     int64                   `json:"tenant_id" gorm:"index"`
	ShopID       int64                   `json:"shop_id"`
	WarehouseID  int64                   `json:"warehouse_id"`
	OrderNo      string                  `json:"order_no" gorm:"size:32"`
	SupplierID   int64                   `json:"supplier_id"`
	BillDate     string                  `json:"bill_date" gorm:"size:10"`
	TotalAmount  float64                 `json:"total_amount"`
	RefundAmount float64                 `json:"refund_amount"`
	Status       int8                    `json:"status"` // 1=草稿 2=待审核 3=已生效 4=已作废
	Remark       string                  `json:"remark" gorm:"size:500"`
	CreatedBy    int64                   `json:"created_by"`
	CreatedAt    time.Time               `json:"created_at"`
	Items        []PurchaseReturnItem    `json:"items" gorm:"foreignKey:PurchaseReturnID"`
	SupplierName string                  `json:"supplier_name" gorm:"-"`
	WarehouseName string                 `json:"warehouse_name" gorm:"-"`
}

// PurchaseReturnItem 退货明细
type PurchaseReturnItem struct {
	ID                int64     `json:"id" gorm:"primaryKey"`
	TenantID          int64     `json:"tenant_id" gorm:"index"`
	PurchaseReturnID  int64     `json:"purchase_return_id" gorm:"index"`
	GoodsID           int64     `json:"goods_id"`
	Quantity          int       `json:"quantity"`
	UnitPrice         float64   `json:"unit_price"`
	Amount            float64   `json:"amount"`
	Remark            string    `json:"remark" gorm:"size:255"`
	CreatedAt         time.Time `json:"created_at"`
	GoodsName         string    `json:"goods_name" gorm:"-"`
	GoodsCode         string    `json:"goods_code" gorm:"-"`
	UnitName          string    `json:"unit_name" gorm:"-"`
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
