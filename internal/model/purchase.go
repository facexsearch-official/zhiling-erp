package model

import "time"

// PurchaseOrder 采购订单（进货预订）
type PurchaseOrder struct {
	ID               int64               `json:"id" gorm:"primaryKey"`
	IDStr            string              `json:"id_str" gorm:"->"`
	TenantID         int64               `json:"tenant_id" gorm:"index"`
	ShopID           int64               `json:"shop_id"`
	WarehouseID      int64               `json:"warehouse_id"`
	OrderNo          string              `json:"order_no" gorm:"size:32"`
	SupplierID       int64               `json:"supplier_id"`
	SalesmanID       int64               `json:"salesman_id"`
	AccountID        int64               `json:"account_id"`
	OrderDate        string              `json:"order_date" gorm:"size:10"`
	Subtotal         float64             `json:"subtotal"`
	Discount         float64             `json:"discount"`
	DiscountedAmount float64             `json:"discounted_amount"`
	Freight          float64             `json:"freight"`
	DepositOffset    float64             `json:"deposit_offset"`
	TotalAmount      float64             `json:"total_amount"`
	PaidAmount       float64             `json:"paid_amount"`
	UnpaidAmount     float64             `json:"unpaid_amount"`
	InvoiceStatus    int8                `json:"invoice_status"`
	RelatedNo        string              `json:"related_no" gorm:"size:32"`
	Attachments      string              `json:"attachments" gorm:"type:text"`
	Status           int8                `json:"status"` // 1=草稿 2=待审核 3=已生效 4=已转换 5=已取消
	Remark           string              `json:"remark" gorm:"size:500"`
	CreatedBy        int64               `json:"created_by"`
	CreatedAt        time.Time           `json:"created_at"`
	UpdatedAt        time.Time           `json:"updated_at"`
	Items            []PurchaseOrderItem `json:"items" gorm:"foreignKey:OrderID"`
	SupplierName     string              `json:"supplier_name" gorm:"->"`
	SalesmanName     string              `json:"salesman_name" gorm:"->"`
	AccountName      string              `json:"account_name" gorm:"->"`
	MakerName        string              `json:"maker_name" gorm:"->"`
}

// PurchaseOrderItem 采购订单明细
type PurchaseOrderItem struct {
	ID        int64     `json:"id" gorm:"primaryKey"`
	TenantID  int64     `json:"tenant_id" gorm:"index"`
	OrderID   int64     `json:"order_id" gorm:"index"`
	GoodsID   int64     `json:"goods_id"`
	Quantity  int       `json:"quantity"`
	UnitPrice float64   `json:"unit_price"`
	Amount    float64   `json:"amount"`
	Remark    string    `json:"remark" gorm:"size:255"`
	CreatedAt time.Time `json:"created_at"`
	GoodsName string    `json:"goods_name" gorm:"->"`
	GoodsCode string    `json:"goods_code" gorm:"->"`
	UnitName  string    `json:"unit_name" gorm:"->"`
	Spec      string    `json:"spec" gorm:"->"`
	Barcode   string    `json:"barcode" gorm:"->"`
	ImageURL  string    `json:"image_url" gorm:"->"`
	Brand     string    `json:"brand" gorm:"->"`
	Origin    string    `json:"origin" gorm:"->"`
}

// Purchase 进货单
type Purchase struct {
	ID               int64          `json:"id" gorm:"primaryKey"`
	IDStr            string         `json:"id_str" gorm:"->"`
	TenantID         int64          `json:"tenant_id" gorm:"index"`
	ShopID           int64          `json:"shop_id"`
	WarehouseID      int64          `json:"warehouse_id"`
	OrderNo          string         `json:"order_no" gorm:"size:32"`
	SupplierID       int64          `json:"supplier_id"`
	SalesmanID       int64          `json:"salesman_id"`
	AccountID        int64          `json:"account_id"`
	BillDate         string         `json:"bill_date" gorm:"size:10"`
	Subtotal         float64        `json:"subtotal"`          // 明细合计
	Discount         float64        `json:"discount"`          // 折扣百分比 100=无折扣
	DiscountedAmount float64        `json:"discounted_amount"` // 折后金额
	Freight          float64        `json:"freight"`           // 运费
	DepositOffset    float64        `json:"deposit_offset"`    // 订金抵扣
	TotalAmount      float64        `json:"total_amount"`      // 本单应付
	PaidAmount       float64        `json:"paid_amount"`
	UnpaidAmount     float64        `json:"unpaid_amount"`
	InvoiceStatus    int8           `json:"invoice_status"` // 0=未开票 1=已开票
	PrintStatus      int8           `json:"print_status"`   // 0=未打印 1=已打印
	RelatedNo        string         `json:"related_no" gorm:"size:32"`
	Attachments      string         `json:"attachments" gorm:"type:text"`
	Status           int8           `json:"status"` // 1=草稿 2=待审核 3=已生效 4=已作废
	Remark           string         `json:"remark" gorm:"size:500"`
	CreatedBy        int64          `json:"created_by"`
	CreatedAt        time.Time      `json:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at"`
	Items            []PurchaseItem `json:"items" gorm:"foreignKey:PurchaseID"`
	SupplierName     string         `json:"supplier_name" gorm:"->"`
	SalesmanName     string         `json:"salesman_name" gorm:"->"`
	AccountName      string         `json:"account_name" gorm:"->"`
	MakerName        string         `json:"maker_name" gorm:"->"`
	WarehouseName    string         `json:"warehouse_name" gorm:"->"`
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
	GoodsName  string    `json:"goods_name" gorm:"->"`
	GoodsCode  string    `json:"goods_code" gorm:"->"`
	UnitName   string    `json:"unit_name" gorm:"->"`
	Spec       string    `json:"spec" gorm:"->"`
	Barcode    string    `json:"barcode" gorm:"->"`
	ImageURL   string    `json:"image_url" gorm:"->"`
	Brand      string    `json:"brand" gorm:"->"`
	Origin     string    `json:"origin" gorm:"->"`
}

// PurchaseReturn 采购退货
type PurchaseReturn struct {
	ID            int64                `json:"id" gorm:"primaryKey"`
	IDStr         string               `json:"id_str" gorm:"->"`
	TenantID      int64                `json:"tenant_id" gorm:"index"`
	ShopID        int64                `json:"shop_id"`
	WarehouseID   int64                `json:"warehouse_id"`
	OrderNo       string               `json:"order_no" gorm:"size:32"`
	SupplierID    int64                `json:"supplier_id"`
	SalesmanID    int64                `json:"salesman_id"`
	AccountID     int64                `json:"account_id"`
	BillDate      string               `json:"bill_date" gorm:"size:10"`
	DepositOffset float64              `json:"deposit_offset"`
	PaidAmount    float64              `json:"paid_amount"`
	UnpaidAmount  float64              `json:"unpaid_amount"`
	InvoiceStatus int8                 `json:"invoice_status"`
	PrintStatus   int8                 `json:"print_status"`
	RelatedNo     string               `json:"related_no" gorm:"size:32"`
	Attachments   string               `json:"attachments" gorm:"type:text"`
	TotalAmount   float64              `json:"total_amount"`
	RefundAmount  float64              `json:"refund_amount"`
	Status        int8                 `json:"status"` // 1=草稿 2=待审核 3=已生效 4=已作废
	Remark        string               `json:"remark" gorm:"size:500"`
	CreatedBy     int64                `json:"created_by"`
	CreatedAt     time.Time            `json:"created_at"`
	Items         []PurchaseReturnItem `json:"items" gorm:"foreignKey:PurchaseReturnID"`
	SupplierName  string               `json:"supplier_name" gorm:"->"`
	SalesmanName  string               `json:"salesman_name" gorm:"->"`
	AccountName   string               `json:"account_name" gorm:"->"`
	MakerName     string               `json:"maker_name" gorm:"->"`
	WarehouseName string               `json:"warehouse_name" gorm:"->"`
}

// PurchaseReturnItem 退货明细
type PurchaseReturnItem struct {
	ID               int64     `json:"id" gorm:"primaryKey"`
	TenantID         int64     `json:"tenant_id" gorm:"index"`
	PurchaseReturnID int64     `json:"purchase_return_id" gorm:"index"`
	GoodsID          int64     `json:"goods_id"`
	Quantity         int       `json:"quantity"`
	UnitPrice        float64   `json:"unit_price"`
	Amount           float64   `json:"amount"`
	Remark           string    `json:"remark" gorm:"size:255"`
	CreatedAt        time.Time `json:"created_at"`
	GoodsName        string    `json:"goods_name" gorm:"->"`
	GoodsCode        string    `json:"goods_code" gorm:"->"`
	UnitName         string    `json:"unit_name" gorm:"->"`
	Spec             string    `json:"spec" gorm:"->"`
	Barcode          string    `json:"barcode" gorm:"->"`
	ImageURL         string    `json:"image_url" gorm:"->"`
	Brand            string    `json:"brand" gorm:"->"`
	Origin           string    `json:"origin" gorm:"->"`
}

// SaleOrder 销售订单（销售预订）
type SaleOrder struct {
	ID               int64           `json:"id" gorm:"primaryKey"`
	IDStr            string          `json:"id_str" gorm:"->"`
	TenantID         int64           `json:"tenant_id" gorm:"index"`
	ShopID           int64           `json:"shop_id"`
	WarehouseID      int64           `json:"warehouse_id"`
	OrderNo          string          `json:"order_no" gorm:"size:32"`
	CustomerID       int64           `json:"customer_id"`
	SalesmanID       int64           `json:"salesman_id"`
	AccountID        int64           `json:"account_id"`
	OrderDate        string          `json:"order_date" gorm:"size:10"`
	TotalAmount      float64         `json:"total_amount"`
	ReceivedAmount   float64         `json:"received_amount"`
	UnreceivedAmount float64         `json:"unreceived_amount"`
	PrintStatus      int8            `json:"print_status"`
	Attachments      string          `json:"attachments" gorm:"type:text"`
	Status           int8            `json:"status"`
	Remark           string          `json:"remark" gorm:"size:500"`
	CreatedBy        int64           `json:"created_by"`
	CreatedAt        time.Time       `json:"created_at"`
	UpdatedAt        time.Time       `json:"updated_at"`
	Items            []SaleOrderItem `json:"items" gorm:"foreignKey:OrderID"`
	CustomerName     string          `json:"customer_name" gorm:"->"`
	SalesmanName     string          `json:"salesman_name" gorm:"->"`
	AccountName      string          `json:"account_name" gorm:"->"`
	MakerName        string          `json:"maker_name" gorm:"->"`
}

// SaleOrderItem 销售订单明细
type SaleOrderItem struct {
	ID        int64     `json:"id" gorm:"primaryKey"`
	TenantID  int64     `json:"tenant_id" gorm:"index"`
	OrderID   int64     `json:"order_id" gorm:"index"`
	GoodsID   int64     `json:"goods_id"`
	Quantity  int       `json:"quantity"`
	UnitPrice float64   `json:"unit_price"`
	Amount    float64   `json:"amount"`
	Remark    string    `json:"remark" gorm:"size:255"`
	CreatedAt time.Time `json:"created_at"`
	GoodsName string    `json:"goods_name" gorm:"->"`
	GoodsCode string    `json:"goods_code" gorm:"->"`
	UnitName  string    `json:"unit_name" gorm:"->"`
	Spec      string    `json:"spec" gorm:"->"`
	Barcode   string    `json:"barcode" gorm:"->"`
	ImageURL  string    `json:"image_url" gorm:"->"`
	Brand     string    `json:"brand" gorm:"->"`
	Origin    string    `json:"origin" gorm:"->"`
}

// Sale 销货单
type Sale struct {
	ID               int64      `json:"id" gorm:"primaryKey"`
	IDStr            string     `json:"id_str" gorm:"->"`
	TenantID         int64      `json:"tenant_id" gorm:"index"`
	ShopID           int64      `json:"shop_id"`
	WarehouseID      int64      `json:"warehouse_id"`
	OrderNo          string     `json:"order_no" gorm:"size:32"`
	RelatedOrderNo   string     `json:"related_order_no" gorm:"size:32"`
	CustomerID       int64      `json:"customer_id"`
	SalesmanID       int64      `json:"salesman_id"`
	AccountID        int64      `json:"account_id"`
	BillDate         string     `json:"bill_date" gorm:"size:10"`
	Discount         float64    `json:"discount"`
	Subtotal         float64    `json:"subtotal"`          // 折前金额
	RoundOff         float64    `json:"round_off"`         // 抹零金额
	TotalAmount      float64    `json:"total_amount"`      // 应收金额
	ReceivedAmount   float64    `json:"received_amount"`   // 实收金额
	UnreceivedAmount float64    `json:"unreceived_amount"` // 待收金额
	ReceiveStatus    int8       `json:"receive_status"`    // 0未收款 1部分收款 2已收清
	InvoiceStatus    int8       `json:"invoice_status"`
	PrintStatus      int8       `json:"print_status"`
	Attachments      string     `json:"attachments" gorm:"type:text"`
	Status           int8       `json:"status"`
	Remark           string     `json:"remark" gorm:"size:500"`
	CreatedBy        int64      `json:"created_by"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
	Items            []SaleItem `json:"items" gorm:"foreignKey:SaleID"`
	CustomerName     string     `json:"customer_name" gorm:"->"`
	SalesmanName     string     `json:"salesman_name" gorm:"->"`
	AccountName      string     `json:"account_name" gorm:"->"`
	MakerName        string     `json:"maker_name" gorm:"->"`
}

// SaleItem 销货单明细
type SaleItem struct {
	ID        int64     `json:"id" gorm:"primaryKey"`
	TenantID  int64     `json:"tenant_id" gorm:"index"`
	SaleID    int64     `json:"sale_id" gorm:"index"`
	GoodsID   int64     `json:"goods_id"`
	Quantity  int       `json:"quantity"`
	UnitPrice float64   `json:"unit_price"`
	Amount    float64   `json:"amount"`
	Remark    string    `json:"remark" gorm:"size:255"`
	CreatedAt time.Time `json:"created_at"`
	GoodsName string    `json:"goods_name" gorm:"->"`
	GoodsCode string    `json:"goods_code" gorm:"->"`
	UnitName  string    `json:"unit_name" gorm:"->"`
	Spec      string    `json:"spec" gorm:"->"`
	Barcode   string    `json:"barcode" gorm:"->"`
	ImageURL  string    `json:"image_url" gorm:"->"`
	Brand     string    `json:"brand" gorm:"->"`
	Origin    string    `json:"origin" gorm:"->"`
}

// SalesReturn 销货退货
type SalesReturn struct {
	ID               int64             `json:"id" gorm:"primaryKey"`
	IDStr            string            `json:"id_str" gorm:"->"`
	TenantID         int64             `json:"tenant_id" gorm:"index"`
	ShopID           int64             `json:"shop_id"`
	WarehouseID      int64             `json:"warehouse_id"`
	OrderNo          string            `json:"order_no" gorm:"size:32"`
	CustomerID       int64             `json:"customer_id"`
	SalesmanID       int64             `json:"salesman_id"`
	AccountID        int64             `json:"account_id"`
	BillDate         string            `json:"bill_date" gorm:"size:10"`
	RoundOff         float64           `json:"round_off"`         // 抹零金额
	TotalAmount      float64           `json:"total_amount"`      // 应收金额
	ReceivedAmount   float64           `json:"received_amount"`   // 实收金额
	UnreceivedAmount float64           `json:"unreceived_amount"` // 待收金额
	PrintStatus      int8              `json:"print_status"`
	Attachments      string            `json:"attachments" gorm:"type:text"`
	Status           int8              `json:"status"`
	Remark           string            `json:"remark" gorm:"size:500"`
	CreatedBy        int64             `json:"created_by"`
	CreatedAt        time.Time         `json:"created_at"`
	Items            []SalesReturnItem `json:"items" gorm:"foreignKey:ReturnID"`
	CustomerName     string            `json:"customer_name" gorm:"->"`
	SalesmanName     string            `json:"salesman_name" gorm:"->"`
	AccountName      string            `json:"account_name" gorm:"->"`
	MakerName        string            `json:"maker_name" gorm:"->"`
}

// SalesReturnItem 销货退货明细
type SalesReturnItem struct {
	ID        int64     `json:"id" gorm:"primaryKey"`
	TenantID  int64     `json:"tenant_id" gorm:"index"`
	ReturnID  int64     `json:"return_id" gorm:"index"`
	GoodsID   int64     `json:"goods_id"`
	Quantity  int       `json:"quantity"`
	UnitPrice float64   `json:"unit_price"`
	Amount    float64   `json:"amount"`
	Remark    string    `json:"remark" gorm:"size:255"`
	CreatedAt time.Time `json:"created_at"`
	GoodsName string    `json:"goods_name" gorm:"->"`
	GoodsCode string    `json:"goods_code" gorm:"->"`
	UnitName  string    `json:"unit_name" gorm:"->"`
	Spec      string    `json:"spec" gorm:"->"`
	Barcode   string    `json:"barcode" gorm:"->"`
	ImageURL  string    `json:"image_url" gorm:"->"`
	Brand     string    `json:"brand" gorm:"->"`
	Origin    string    `json:"origin" gorm:"->"`
}

// Quote 报价单
type Quote struct {
	ID           int64       `json:"id" gorm:"primaryKey"`
	IDStr        string      `json:"id_str" gorm:"->"`
	TenantID     int64       `json:"tenant_id" gorm:"index"`
	ShopID       int64       `json:"shop_id"`
	OrderNo      string      `json:"order_no" gorm:"size:32"`
	CustomerID   int64       `json:"customer_id"`
	SalesmanID   int64       `json:"salesman_id"`
	BillDate     string      `json:"bill_date" gorm:"size:10"`
	TotalAmount  float64     `json:"total_amount"`
	Remark       string      `json:"remark" gorm:"size:500"`
	CreatedBy    int64       `json:"created_by"`
	CreatedAt    time.Time   `json:"created_at"`
	Items        []QuoteItem `json:"items" gorm:"foreignKey:QuoteID"`
	CustomerName string      `json:"customer_name" gorm:"->"`
	SalesmanName string      `json:"salesman_name" gorm:"->"`
	MakerName    string      `json:"maker_name" gorm:"->"`
}

// QuoteItem 报价单明细
type QuoteItem struct {
	ID        int64     `json:"id" gorm:"primaryKey"`
	TenantID  int64     `json:"tenant_id" gorm:"index"`
	QuoteID   int64     `json:"quote_id" gorm:"index"`
	GoodsID   int64     `json:"goods_id"`
	Quantity  int       `json:"quantity"`
	UnitPrice float64   `json:"unit_price"`
	Amount    float64   `json:"amount"`
	Remark    string    `json:"remark" gorm:"size:255"`
	CreatedAt time.Time `json:"created_at"`
	GoodsName string    `json:"goods_name" gorm:"->"`
	GoodsCode string    `json:"goods_code" gorm:"->"`
	UnitName  string    `json:"unit_name" gorm:"->"`
	Spec      string    `json:"spec" gorm:"->"`
	Barcode   string    `json:"barcode" gorm:"->"`
	ImageURL  string    `json:"image_url" gorm:"->"`
	Brand     string    `json:"brand" gorm:"->"`
	Origin    string    `json:"origin" gorm:"->"`
}
