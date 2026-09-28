package model

import "time"

// Customer 客户
type Customer struct {
	ID              int64     `json:"id" gorm:"primaryKey"`
	TenantID        int64     `json:"tenant_id" gorm:"index"`
	Name            string    `json:"name" gorm:"size:128"`
	Code            string    `json:"code" gorm:"size:64"`
	Type            int8      `json:"type"` // 1=客户 2=供应商 3=双身份
	CategoryID      *int64    `json:"category_id"`
	PriceLevel      string    `json:"price_level" gorm:"size:32"`
	SalesmanID      int64     `json:"salesman_id" gorm:"index"`
	Discount        float64   `json:"discount" gorm:"default:100"`
	Points          int       `json:"points"`
	InitDebt        float64   `json:"init_debt"`
	Contact         string    `json:"contact" gorm:"size:64"`
	Phone           string    `json:"phone" gorm:"size:20"`
	Address         string    `json:"address" gorm:"size:255"`
	AddressDetail   string    `json:"address_detail" gorm:"size:255"`
	Email           string    `json:"email" gorm:"size:128"`
	TaxNo           string    `json:"tax_no" gorm:"size:64"`
	Fax             string    `json:"fax" gorm:"size:32"`
	BankName        string    `json:"bank_name" gorm:"size:128"`
	BankAccount     string    `json:"bank_account" gorm:"size:64"`
	Birthday        string    `json:"birthday" gorm:"size:10"`
	Wechat          string    `json:"wechat" gorm:"size:64"`
	QQ              string    `json:"qq" gorm:"size:32"`
	Remark          string    `json:"remark" gorm:"size:500"`
	Balance         float64   `json:"balance"`
	TotalReceivable float64   `json:"total_receivable"`
	TotalPayable    float64   `json:"total_payable"`
	Status          int8      `json:"status" gorm:"default:1"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
	CategoryName    string    `json:"category_name" gorm:"->"`
	SalesmanName    string    `json:"salesman_name" gorm:"->"`
}

// CustomerCategory 客户分类
type CustomerCategory struct {
	ID       int64  `json:"id" gorm:"primaryKey"`
	TenantID int64  `json:"tenant_id" gorm:"index"`
	Name     string `json:"name" gorm:"size:64"`
	ParentID int64  `json:"parent_id" gorm:"index"`
	Sort     int    `json:"sort"`
}

// CustomerAddress 客户收货地址
type CustomerAddress struct {
	ID         int64  `json:"id" gorm:"primaryKey"`
	TenantID   int64  `json:"tenant_id" gorm:"index"`
	CustomerID int64  `json:"customer_id" gorm:"index"`
	Receiver   string `json:"receiver" gorm:"size:64"`
	Phone      string `json:"phone" gorm:"size:32"`
	Region     string `json:"region" gorm:"size:255"`
	Detail     string `json:"detail" gorm:"size:255"`
	IsDefault  int8   `json:"is_default"`
}

// SupplierCategory 供应商分类（与货品分类结构一致，独立表）
type SupplierCategory struct {
	ID       int64  `json:"id" gorm:"primaryKey"`
	TenantID int64  `json:"tenant_id" gorm:"index"`
	Name     string `json:"name" gorm:"size:30"`
	ParentID int64  `json:"parent_id" gorm:"index"`
	Sort     int    `json:"sort"`
}

// Supplier 供应商
type Supplier struct {
	ID           int64     `json:"id" gorm:"primaryKey"`
	IDStr        string    `json:"id_str" gorm:"->"`
	TenantID     int64     `json:"tenant_id" gorm:"index"`
	Name         string    `json:"name" gorm:"size:220"`
	Code         string    `json:"code" gorm:"size:64"`
	CategoryID   *int64    `json:"category_id" gorm:"index"`
	Contact      string    `json:"contact" gorm:"size:64"`
	Phone        string    `json:"phone" gorm:"size:20"`
	Address      string    `json:"address" gorm:"size:255"`
	BankName     string    `json:"bank_name" gorm:"size:128"`
	BankAccount  string    `json:"bank_account" gorm:"size:64"`
	Remark       string    `json:"remark" gorm:"size:500"`
	InitPayable  float64   `json:"init_payable"`
	TotalPayable float64   `json:"total_payable"`
	Status       int8      `json:"status" gorm:"default:1"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`

	// ── 辅助信息 ──
	Email    string `json:"email" gorm:"size:128"`
	Fax      string `json:"fax" gorm:"size:64"`
	Wechat   string `json:"wechat" gorm:"size:64"`
	QQ       string `json:"qq" gorm:"size:32"`
	Birthday string `json:"birthday" gorm:"size:32"`

	// ── 联系地址 ──
	Province      string `json:"province" gorm:"size:64"`
	City          string `json:"city" gorm:"size:64"`
	District      string `json:"district" gorm:"size:64"`
	AddressDetail string `json:"address_detail" gorm:"size:255"`

	// 非持久化：随详情一起返回
	Addresses []SupplierAddress `json:"addresses" gorm:"-"`
}

// SupplierAddress 供应商收货信息（一对多）
type SupplierAddress struct {
	ID             int64  `json:"id" gorm:"primaryKey"`
	TenantID       int64  `json:"tenant_id" gorm:"index"`
	SupplierID     int64  `json:"supplier_id" gorm:"index"`
	Consignee      string `json:"consignee" gorm:"size:64"`
	ConsigneePhone string `json:"consignee_phone" gorm:"size:32"`
	Province       string `json:"province" gorm:"size:64"`
	City           string `json:"city" gorm:"size:64"`
	District       string `json:"district" gorm:"size:64"`
	AddressDetail  string `json:"address_detail" gorm:"size:255"`
	IsDefault      int8   `json:"is_default" gorm:"default:0"`
	Sort           int    `json:"sort"`
}
