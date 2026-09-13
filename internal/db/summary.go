package db

import (
	"fmt"
	"time"

	"gorm.io/gorm"
)

// PlatformSummary 平台汇总表（跨分片聚合结果）
type PlatformSummary struct {
	ID              int64     `gorm:"primaryKey" json:"id"`
	SummaryDate     string    `gorm:"size:10" json:"summary_date"`     // 2026-09-13
	SummaryType     string    `gorm:"size:32" json:"summary_type"`     // daily / monthly
	TotalTenants    int       `json:"total_tenants"`
	TotalSales      float64   `json:"total_sales"`
	TotalPurchases  float64   `json:"total_purchases"`
	TotalReceipts   float64   `json:"total_receipts"`
	TotalPayments   float64   `json:"total_payments"`
	SalesCount      int       `json:"sales_count"`
	PurchaseCount   int       `json:"purchase_count"`
	ReceiptCount    int       `json:"receipt_count"`
	PaymentCount    int       `json:"payment_count"`
	CreatedAt       time.Time `json:"created_at"`
}

func (PlatformSummary) TableName() string {
	return "platform_summary"
}

// CreateSummaryTable 创建汇总表
func CreateSummaryTable(db *gorm.DB) error {
	sql := `CREATE TABLE IF NOT EXISTS platform_summary (
  id BIGINT PRIMARY KEY,
  summary_date VARCHAR(10) NOT NULL,
  summary_type VARCHAR(32) NOT NULL DEFAULT 'daily',
  total_tenants INT DEFAULT 0,
  total_sales DECIMAL(12,2) DEFAULT 0,
  total_purchases DECIMAL(12,2) DEFAULT 0,
  total_receipts DECIMAL(12,2) DEFAULT 0,
  total_payments DECIMAL(12,2) DEFAULT 0,
  sales_count INT DEFAULT 0,
  purchase_count INT DEFAULT 0,
  receipt_count INT DEFAULT 0,
  payment_count INT DEFAULT 0,
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  UNIQUE KEY uk_date_type (summary_date, summary_type)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`
	return db.Exec(sql).Error
}

// AggregateDaily 聚合某天的数据（跨所有分片 UNION ALL）
func AggregateDaily(db *gorm.DB, router *ShardRouter, date string) error {
	var totalSales, totalPurchases, totalReceipts, totalPayments float64
	var salesCount, purchaseCount, receiptCount, paymentCount int

	// 遍历所有分片聚合 sales
	for i := 0; i < router.ShardNum(); i++ {
		var result struct {
			Total float64
			Count int
		}
		table := fmt.Sprintf("sales_t%02d", i)
		db.Table(table).Select("COALESCE(SUM(total_amount),0) as total, COUNT(*) as count").
			Where("bill_date = ? AND status != 4", date).Scan(&result)
		totalSales += result.Total
		salesCount += result.Count
	}

	// 聚合 purchases
	for i := 0; i < router.ShardNum(); i++ {
		var result struct {
			Total float64
			Count int
		}
		table := fmt.Sprintf("purchases_t%02d", i)
		db.Table(table).Select("COALESCE(SUM(total_amount),0) as total, COUNT(*) as count").
			Where("bill_date = ? AND status != 4", date).Scan(&result)
		totalPurchases += result.Total
		purchaseCount += result.Count
	}

	// 聚合 receipts
	for i := 0; i < router.ShardNum(); i++ {
		var result struct {
			Total float64
			Count int
		}
		table := fmt.Sprintf("receipts_t%02d", i)
		db.Table(table).Select("COALESCE(SUM(amount),0) as total, COUNT(*) as count").
			Where("bill_date = ? AND status != 3", date).Scan(&result)
		totalReceipts += result.Total
		receiptCount += result.Count
	}

	// 聚合 payments
	for i := 0; i < router.ShardNum(); i++ {
		var result struct {
			Total float64
			Count int
		}
		table := fmt.Sprintf("payments_t%02d", i)
		db.Table(table).Select("COALESCE(SUM(amount),0) as total, COUNT(*) as count").
			Where("bill_date = ? AND status != 3", date).Scan(&result)
		totalPayments += result.Total
		paymentCount += result.Count
	}

	// 统计活跃租户数（从任一分片统计有当天数据的 tenant_id 数量）
	tenantSet := make(map[int64]bool)
	for i := 0; i < router.ShardNum(); i++ {
		var tenantIDs []int64
		table := fmt.Sprintf("sales_t%02d", i)
		db.Table(table).Distinct("tenant_id").Where("bill_date = ?", date).Pluck("tenant_id", &tenantIDs)
		for _, tid := range tenantIDs {
			tenantSet[tid] = true
		}
	}

	// 写入汇总表
	summary := PlatformSummary{
		SummaryDate:    date,
		SummaryType:    "daily",
		TotalTenants:   len(tenantSet),
		TotalSales:     totalSales,
		TotalPurchases: totalPurchases,
		TotalReceipts:  totalReceipts,
		TotalPayments:  totalPayments,
		SalesCount:     salesCount,
		PurchaseCount:  purchaseCount,
		ReceiptCount:   receiptCount,
		PaymentCount:   paymentCount,
	}

	return db.Save(&summary).Error
}
