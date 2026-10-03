package model

import "time"

// StockBalance 库存余额
type StockBalance struct {
	ID          int64      `json:"id" gorm:"primaryKey"`
	TenantID    int64      `json:"tenant_id" gorm:"index"`
	ShopID      int64      `json:"shop_id"`
	WarehouseID int64      `json:"warehouse_id"`
	GoodsID     int64      `json:"goods_id"`
	Quantity    int        `json:"quantity"`
	CostPrice   float64    `json:"cost_price"`
	TotalCost   float64    `json:"total_cost"`
	LastInTime  *time.Time `json:"last_in_time"`
	LastOutTime *time.Time `json:"last_out_time"`
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
	BankName  string    `json:"bank_name" gorm:"size:128"`
	CardNo    string    `json:"card_no" gorm:"size:64"`
	ShopID    int64     `json:"shop_id" gorm:"index"`
	ShopName  string    `json:"shop_name" gorm:"-"`
	Balance   float64   `json:"balance"`
	Remark    string    `json:"remark" gorm:"size:255"`
	Sort      int       `json:"sort"`
	Status    int8      `json:"status" gorm:"default:1"`
	CreatedAt time.Time `json:"created_at"`
}

// Transfer 转账单
type Transfer struct {
	ID            int64     `json:"id" gorm:"primaryKey"`
	IDStr         string    `json:"id_str" gorm:"->"`
	TenantID      int64     `json:"tenant_id" gorm:"index"`
	OrderNo       string    `json:"order_no" gorm:"size:32"`
	BillDate      string    `json:"bill_date" gorm:"size:10"`
	FromAccountID int64     `json:"from_account_id"`
	ToAccountID   int64     `json:"to_account_id"`
	Amount        float64   `json:"amount"`
	Remark        string    `json:"remark" gorm:"size:255"`
	CreatedBy     int64     `json:"created_by"`
	CreatedAt     time.Time `json:"created_at"`
	FromName      string    `json:"from_name" gorm:"->"`
	ToName        string    `json:"to_name" gorm:"->"`
	MakerName     string    `json:"maker_name" gorm:"->"`
}

// Receipt 收款单
type Receipt struct {
	ID             int64     `json:"id" gorm:"primaryKey"`
	IDStr          string    `json:"id_str" gorm:"->"`
	TenantID       int64     `json:"tenant_id" gorm:"index"`
	ShopID         int64     `json:"shop_id"`
	ShopIDStr      string    `json:"shop_id_str" gorm:"-"`
	ShopName       string    `json:"shop_name" gorm:"->"`
	OrderNo        string    `json:"order_no" gorm:"size:32"`
	RelatedNo      string    `json:"related_no" gorm:"size:32"`
	Type           string    `json:"type" gorm:"size:32"`
	CustomerID     int64     `json:"customer_id"`
	SalesmanID     int64     `json:"salesman_id"`
	BillDate       string    `json:"bill_date" gorm:"size:10"`
	Amount         float64   `json:"amount"`
	DiscountAmount float64   `json:"discount_amount"`
	DepositOffset  float64   `json:"deposit_offset"`
	AccountID      int64     `json:"account_id"`
	Attachments    string    `json:"attachments" gorm:"type:text"`
	Status         int8      `json:"status" gorm:"default:1"` // 1=正常 9=作废
	Remark         string    `json:"remark" gorm:"size:500"`
	CreatedBy      int64     `json:"created_by"`
	CreatedAt      time.Time `json:"created_at"`
	CustomerName   string    `json:"customer_name" gorm:"->"`
	AccountName    string    `json:"account_name" gorm:"->"`
	SalesmanName   string    `json:"salesman_name" gorm:"->"`
	MakerName      string    `json:"maker_name" gorm:"->"`
}

// Payment 付款单
type Payment struct {
	ID             int64     `json:"id" gorm:"primaryKey"`
	IDStr          string    `json:"id_str" gorm:"->"`
	TenantID       int64     `json:"tenant_id" gorm:"index"`
	ShopID         int64     `json:"shop_id"`
	ShopIDStr      string    `json:"shop_id_str" gorm:"-"`
	ShopName       string    `json:"shop_name" gorm:"->"`
	OrderNo        string    `json:"order_no" gorm:"size:32"`
	RelatedNo      string    `json:"related_no" gorm:"size:32"`
	Type           string    `json:"type" gorm:"size:32"`
	SupplierID     int64     `json:"supplier_id"`
	SalesmanID     int64     `json:"salesman_id"`
	BillDate       string    `json:"bill_date" gorm:"size:10"`
	Amount         float64   `json:"amount"`
	DiscountAmount float64   `json:"discount_amount"`
	AccountID      int64     `json:"account_id"`
	Attachments    string    `json:"attachments" gorm:"type:text"`
	Status         int8      `json:"status" gorm:"default:1"` // 1=正常 9=作废
	Remark         string    `json:"remark" gorm:"size:500"`
	CreatedBy      int64     `json:"created_by"`
	CreatedAt      time.Time `json:"created_at"`
	SupplierName   string    `json:"supplier_name" gorm:"->"`
	AccountName    string    `json:"account_name" gorm:"->"`
	SalesmanName   string    `json:"salesman_name" gorm:"->"`
	MakerName      string    `json:"maker_name" gorm:"->"`
}
