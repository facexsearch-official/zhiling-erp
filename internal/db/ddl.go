package db

import (
	"fmt"
	"strings"

	"gorm.io/gorm"
)

// ShardedTables 需要物理分片的表及其 DDL
var ShardedTables = map[string]string{
	"sales": `CREATE TABLE IF NOT EXISTS %s (
  id BIGINT NOT NULL,
  tenant_id BIGINT NOT NULL,
  shop_id BIGINT,
  warehouse_id BIGINT,
  order_no VARCHAR(32) NOT NULL,
  customer_id BIGINT NOT NULL,
  bill_date DATE NOT NULL,
  total_amount DECIMAL(10,2) DEFAULT 0,
  received_amount DECIMAL(10,2) DEFAULT 0,
  unreceived_amount DECIMAL(10,2) DEFAULT 0,
  status TINYINT DEFAULT 1,
  remark VARCHAR(500),
  created_by BIGINT,
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uk_order_no (tenant_id, order_no),
  INDEX idx_tenant_date (tenant_id, bill_date),
  INDEX idx_tenant_customer (tenant_id, customer_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,

	"sales_returns": `CREATE TABLE IF NOT EXISTS %s (
  id BIGINT NOT NULL,
  tenant_id BIGINT NOT NULL,
  shop_id BIGINT,
  warehouse_id BIGINT,
  order_no VARCHAR(32) NOT NULL,
  customer_id BIGINT NOT NULL,
  bill_date DATE NOT NULL,
  total_amount DECIMAL(10,2) DEFAULT 0,
  refund_amount DECIMAL(10,2) DEFAULT 0,
  status TINYINT DEFAULT 1,
  remark VARCHAR(500),
  created_by BIGINT,
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uk_order_no (tenant_id, order_no),
  INDEX idx_tenant_date (tenant_id, bill_date)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,

	"purchases": `CREATE TABLE IF NOT EXISTS %s (
  id BIGINT NOT NULL,
  tenant_id BIGINT NOT NULL,
  shop_id BIGINT,
  warehouse_id BIGINT,
  order_no VARCHAR(32) NOT NULL,
  supplier_id BIGINT NOT NULL,
  bill_date DATE NOT NULL,
  total_amount DECIMAL(10,2) DEFAULT 0,
  paid_amount DECIMAL(10,2) DEFAULT 0,
  unpaid_amount DECIMAL(10,2) DEFAULT 0,
  status TINYINT DEFAULT 1,
  remark VARCHAR(500),
  created_by BIGINT,
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uk_order_no (tenant_id, order_no),
  INDEX idx_tenant_date (tenant_id, bill_date),
  INDEX idx_tenant_supplier (tenant_id, supplier_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,

	"purchase_returns": `CREATE TABLE IF NOT EXISTS %s (
  id BIGINT NOT NULL,
  tenant_id BIGINT NOT NULL,
  shop_id BIGINT,
  warehouse_id BIGINT,
  order_no VARCHAR(32) NOT NULL,
  supplier_id BIGINT NOT NULL,
  bill_date DATE NOT NULL,
  total_amount DECIMAL(10,2) DEFAULT 0,
  refund_amount DECIMAL(10,2) DEFAULT 0,
  status TINYINT DEFAULT 1,
  remark VARCHAR(500),
  created_by BIGINT,
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uk_order_no (tenant_id, order_no),
  INDEX idx_tenant_date (tenant_id, bill_date)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,

	"sales_orders": `CREATE TABLE IF NOT EXISTS %s (
  id BIGINT NOT NULL,
  tenant_id BIGINT NOT NULL,
  shop_id BIGINT,
  order_no VARCHAR(32) NOT NULL,
  customer_id BIGINT NOT NULL,
  order_date DATE NOT NULL,
  total_amount DECIMAL(10,2) DEFAULT 0,
  status TINYINT DEFAULT 1,
  remark VARCHAR(500),
  created_by BIGINT,
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uk_order_no (tenant_id, order_no),
  INDEX idx_tenant_date (tenant_id, order_date)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,

	"purchase_orders": `CREATE TABLE IF NOT EXISTS %s (
  id BIGINT NOT NULL,
  tenant_id BIGINT NOT NULL,
  shop_id BIGINT,
  order_no VARCHAR(32) NOT NULL,
  supplier_id BIGINT NOT NULL,
  order_date DATE NOT NULL,
  total_amount DECIMAL(10,2) DEFAULT 0,
  status TINYINT DEFAULT 1,
  remark VARCHAR(500),
  created_by BIGINT,
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uk_order_no (tenant_id, order_no),
  INDEX idx_tenant_date (tenant_id, order_date)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,

	"receipts": `CREATE TABLE IF NOT EXISTS %s (
  id BIGINT NOT NULL,
  tenant_id BIGINT NOT NULL,
  shop_id BIGINT,
  order_no VARCHAR(32) NOT NULL,
  customer_id BIGINT,
  bill_date DATE NOT NULL,
  amount DECIMAL(10,2) DEFAULT 0,
  account_id BIGINT,
  status TINYINT DEFAULT 1,
  remark VARCHAR(500),
  created_by BIGINT,
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uk_order_no (tenant_id, order_no),
  INDEX idx_tenant_date (tenant_id, bill_date)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,

	"payments": `CREATE TABLE IF NOT EXISTS %s (
  id BIGINT NOT NULL,
  tenant_id BIGINT NOT NULL,
  shop_id BIGINT,
  order_no VARCHAR(32) NOT NULL,
  supplier_id BIGINT,
  bill_date DATE NOT NULL,
  amount DECIMAL(10,2) DEFAULT 0,
  account_id BIGINT,
  status TINYINT DEFAULT 1,
  remark VARCHAR(500),
  created_by BIGINT,
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uk_order_no (tenant_id, order_no),
  INDEX idx_tenant_date (tenant_id, bill_date)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,

	"stock_logs": `CREATE TABLE IF NOT EXISTS %s (
  id BIGINT NOT NULL,
  tenant_id BIGINT NOT NULL,
  shop_id BIGINT,
  warehouse_id BIGINT,
  goods_id BIGINT NOT NULL,
  type TINYINT NOT NULL,
  quantity INT NOT NULL,
  before_stock INT DEFAULT 0,
  after_stock INT DEFAULT 0,
  related_type VARCHAR(32),
  related_id BIGINT,
  related_no VARCHAR(32),
  remark VARCHAR(255),
  created_by BIGINT,
  created_at DATETIME NOT NULL,
  PRIMARY KEY (id, created_at),
  INDEX idx_tenant_goods (tenant_id, goods_id, created_at),
  INDEX idx_tenant_created (tenant_id, created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,

	"operation_logs": `CREATE TABLE IF NOT EXISTS %s (
  id BIGINT NOT NULL,
  tenant_id BIGINT NOT NULL,
  user_id BIGINT,
  action VARCHAR(64),
  target_type VARCHAR(32),
  target_id BIGINT,
  detail TEXT,
  ip VARCHAR(45),
  created_at DATETIME NOT NULL,
  PRIMARY KEY (id, created_at),
  INDEX idx_tenant_created (tenant_id, created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
}

// CreateShardedTables 为所有分片创建表
func CreateShardedTables(db *gorm.DB, shardNum int) error {
	for baseTable, ddl := range ShardedTables {
		for i := 0; i < shardNum; i++ {
			tableName := fmt.Sprintf("%s_t%02d", baseTable, i)
			sql := fmt.Sprintf(ddl, tableName)
			if err := db.Exec(sql).Error; err != nil {
				return fmt.Errorf("create table %s: %w", tableName, err)
			}
		}
	}
	return nil
}

// GenerateShardedDDL 生成所有分片建表 SQL（用于导出）
func GenerateShardedDDL(shardNum int) string {
	var sb strings.Builder
	for baseTable, ddl := range ShardedTables {
		for i := 0; i < shardNum; i++ {
			tableName := fmt.Sprintf("%s_t%02d", baseTable, i)
			sql := fmt.Sprintf(ddl, tableName)
			sb.WriteString(fmt.Sprintf("-- %s\n%s;\n\n", tableName, sql))
		}
	}
	return sb.String()
}
