package model

import "time"

// Customer 客户
type Customer struct {
	ID              int64     `json:"id" gorm:"primaryKey"`
	TenantID        int64     `json:"tenant_id" gorm:"index"`
	Name            string    `json:"name" gorm:"size:128"`
	Code            string    `json:"code" gorm:"size:64"`
	Type            int8      `json:"type"` // 1=客户 2=供应商 3=双身份
	Contact         string    `json:"contact" gorm:"size:64"`
	Phone           string    `json:"phone" gorm:"size:20"`
	Address         string    `json:"address" gorm:"size:255"`
	Remark          string    `json:"remark" gorm:"size:500"`
	Balance         float64   `json:"balance"`
	TotalReceivable float64   `json:"total_receivable"`
	TotalPayable    float64   `json:"total_payable"`
	Status          int8      `json:"status" gorm:"default:1"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// Supplier 供应商
type Supplier struct {
	ID           int64     `json:"id" gorm:"primaryKey"`
	TenantID     int64     `json:"tenant_id" gorm:"index"`
	Name         string    `json:"name" gorm:"size:220"`
	Code         string    `json:"code" gorm:"size:64"`
	Contact      string    `json:"contact" gorm:"size:64"`
	Phone        string    `json:"phone" gorm:"size:20"`
	Address      string    `json:"address" gorm:"size:255"`
	BankName     string    `json:"bank_name" gorm:"size:128"`
	BankAccount  string    `json:"bank_account" gorm:"size:64"`
	Remark       string    `json:"remark" gorm:"size:500"`
	TotalPayable float64   `json:"total_payable"`
	Status       int8      `json:"status" gorm:"default:1"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
