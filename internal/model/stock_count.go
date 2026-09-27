package model

import "time"

// StockCount 盘点单
type StockCount struct {
	ID            int64            `json:"id" gorm:"primaryKey"`
	IDStr         string           `json:"id_str" gorm:"->"`
	TenantID      int64            `json:"tenant_id" gorm:"index"`
	ShopID        int64            `json:"shop_id"`
	WarehouseID   int64            `json:"warehouse_id"`
	OrderNo       string           `json:"order_no" gorm:"size:32"`
	SalesmanID    int64            `json:"salesman_id"`
	BillDate      string           `json:"bill_date" gorm:"size:10"`
	BookQty       int              `json:"book_qty"`
	ActualQty     int              `json:"actual_qty"`
	DiffQty       int              `json:"diff_qty"`
	Status        int8             `json:"status" gorm:"default:1"` // 1=正常 9=作废
	Attachments   string           `json:"attachments" gorm:"type:text"`
	Remark        string           `json:"remark" gorm:"size:500"`
	CreatedBy     int64            `json:"created_by"`
	CreatedAt     time.Time        `json:"created_at"`
	Items         []StockCountItem `json:"items" gorm:"foreignKey:CountID"`
	WarehouseName string           `json:"warehouse_name" gorm:"->"`
	SalesmanName  string           `json:"salesman_name" gorm:"->"`
	MakerName     string           `json:"maker_name" gorm:"->"`
}

// StockCountItem 盘点明细
type StockCountItem struct {
	ID        int64     `json:"id" gorm:"primaryKey"`
	TenantID  int64     `json:"tenant_id" gorm:"index"`
	CountID   int64     `json:"count_id" gorm:"index"`
	GoodsID   int64     `json:"goods_id"`
	BookQty   int       `json:"book_qty"`
	ActualQty int       `json:"actual_qty"`
	DiffQty   int       `json:"diff_qty"`
	Remark    string    `json:"remark" gorm:"size:255"`
	CreatedAt time.Time `json:"created_at"`
	GoodsName string    `json:"goods_name" gorm:"->"`
	GoodsCode string    `json:"goods_code" gorm:"->"`
	UnitName  string    `json:"unit_name" gorm:"->"`
	Spec      string    `json:"spec" gorm:"->"`
	Barcode   string    `json:"barcode" gorm:"->"`
	ImageURL  string    `json:"image_url" gorm:"->"`
}
