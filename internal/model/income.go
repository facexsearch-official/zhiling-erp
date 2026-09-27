package model

import "time"

// IncomeType 收支类型
type IncomeType struct {
	ID        int64     `json:"id" gorm:"primaryKey"`
	TenantID  int64     `json:"tenant_id" gorm:"index"`
	Name      string    `json:"name" gorm:"size:64"`
	Direction int8      `json:"direction"` // 1=收入 2=支出
	Sort      int       `json:"sort"`
	Status    int8      `json:"status" gorm:"default:1"`
	CreatedAt time.Time `json:"created_at"`
}

// Income 其他收入
type Income struct {
	ID           int64        `json:"id" gorm:"primaryKey"`
	IDStr        string       `json:"id_str" gorm:"->"`
	TenantID     int64        `json:"tenant_id" gorm:"index"`
	ShopID       int64        `json:"shop_id"`
	OrderNo      string       `json:"order_no" gorm:"size:32"`
	Direction    int8         `json:"direction" gorm:"default:1"` // 1=收入 2=支出
	TypeID       int64        `json:"type_id"`
	TypeName     string       `json:"type_name" gorm:"size:64"`
	PartyName    string       `json:"party_name" gorm:"size:128"`
	SalesmanID   int64        `json:"salesman_id"`
	BillDate     string       `json:"bill_date" gorm:"size:10"`
	Amount       float64      `json:"amount"`
	AccountID    int64        `json:"account_id"`
	Attachments  string       `json:"attachments" gorm:"type:text"`
	Status       int8         `json:"status" gorm:"default:1"` // 1=正常 9=作废
	Remark       string       `json:"remark" gorm:"size:500"`
	CreatedBy    int64        `json:"created_by"`
	CreatedAt    time.Time    `json:"created_at"`
	Items        []IncomeItem `json:"items" gorm:"foreignKey:IncomeID"`
	SalesmanName string       `json:"salesman_name" gorm:"->"`
	MakerName    string       `json:"maker_name" gorm:"->"`
	AccountName  string       `json:"account_name" gorm:"->"`
}

// IncomeItem 其他收入收款明细
type IncomeItem struct {
	ID          int64   `json:"id" gorm:"primaryKey"`
	TenantID    int64   `json:"tenant_id" gorm:"index"`
	IncomeID    int64   `json:"income_id" gorm:"index"`
	AccountID   int64   `json:"account_id"`
	Amount      float64 `json:"amount"`
	AccountName string  `json:"account_name" gorm:"->"`
}
