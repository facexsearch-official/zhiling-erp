-- PISA 进销存系统 - 业务库建表脚本 (mysql_shop)

CREATE DATABASE IF NOT EXISTS pisa_shop DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
USE pisa_shop;

-- 商品分类
CREATE TABLE IF NOT EXISTS goods_categories (
  id BIGINT PRIMARY KEY,
  tenant_id BIGINT NOT NULL,
  name VARCHAR(128) NOT NULL,
  parent_id BIGINT DEFAULT 0,
  sort INT DEFAULT 0,
  INDEX idx_tenant (tenant_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- 单位
CREATE TABLE IF NOT EXISTS units (
  id BIGINT PRIMARY KEY,
  tenant_id BIGINT NOT NULL,
  name VARCHAR(32) NOT NULL,
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  INDEX idx_tenant (tenant_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- 辅助属性
CREATE TABLE IF NOT EXISTS goods_attributes (
  id BIGINT PRIMARY KEY,
  tenant_id BIGINT NOT NULL,
  name VARCHAR(64) NOT NULL,
  values TEXT,
  sort INT DEFAULT 0,
  status TINYINT DEFAULT 1,
  INDEX idx_tenant (tenant_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- 仓库
CREATE TABLE IF NOT EXISTS warehouses (
  id BIGINT PRIMARY KEY,
  tenant_id BIGINT NOT NULL,
  name VARCHAR(128) NOT NULL,
  type TINYINT DEFAULT 1 COMMENT '1=普通 2=原料仓 3=成品仓',
  address VARCHAR(255),
  keeper VARCHAR(64),
  sort INT DEFAULT 0,
  status TINYINT DEFAULT 1,
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  INDEX idx_tenant (tenant_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- 商品
CREATE TABLE IF NOT EXISTS goods (
  id BIGINT PRIMARY KEY,
  tenant_id BIGINT NOT NULL,
  name VARCHAR(255) NOT NULL,
  code VARCHAR(64),
  barcode VARCHAR(64),
  category_id BIGINT,
  unit_id BIGINT,
  supplier_id BIGINT,
  image_url VARCHAR(255),
  current_stock INT DEFAULT 0,
  max_stock INT DEFAULT 0,
  min_stock INT DEFAULT 0,
  purchase_price DECIMAL(10,2) DEFAULT 0,
  retail_price DECIMAL(10,2) DEFAULT 0,
  wholesale_price DECIMAL(10,2) DEFAULT 0,
  cost_method TINYINT DEFAULT 1 COMMENT '1=加权平均 2=FIFO',
  status TINYINT DEFAULT 1,
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  INDEX idx_tenant (tenant_id),
  INDEX idx_code (tenant_id, code)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- 客户
CREATE TABLE IF NOT EXISTS customers (
  id BIGINT PRIMARY KEY,
  tenant_id BIGINT NOT NULL,
  name VARCHAR(128) NOT NULL,
  code VARCHAR(64),
  type TINYINT DEFAULT 1 COMMENT '1=客户 2=供应商 3=双身份',
  contact VARCHAR(64),
  phone VARCHAR(20),
  address VARCHAR(255),
  remark VARCHAR(500),
  balance DECIMAL(10,2) DEFAULT 0,
  total_receivable DECIMAL(10,2) DEFAULT 0,
  total_payable DECIMAL(10,2) DEFAULT 0,
  status TINYINT DEFAULT 1,
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  INDEX idx_tenant (tenant_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- 供应商
CREATE TABLE IF NOT EXISTS suppliers (
  id BIGINT PRIMARY KEY,
  tenant_id BIGINT NOT NULL,
  name VARCHAR(128) NOT NULL,
  code VARCHAR(64),
  contact VARCHAR(64),
  phone VARCHAR(20),
  address VARCHAR(255),
  bank_name VARCHAR(128),
  bank_account VARCHAR(64),
  remark VARCHAR(500),
  total_payable DECIMAL(10,2) DEFAULT 0,
  status TINYINT DEFAULT 1,
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  INDEX idx_tenant (tenant_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- 采购订单
CREATE TABLE IF NOT EXISTS purchase_orders (
  id BIGINT PRIMARY KEY,
  tenant_id BIGINT NOT NULL,
  shop_id BIGINT,
  order_no VARCHAR(32) NOT NULL,
  supplier_id BIGINT NOT NULL,
  order_date VARCHAR(10),
  total_amount DECIMAL(10,2) DEFAULT 0,
  status TINYINT DEFAULT 1,
  remark VARCHAR(500),
  created_by BIGINT,
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  INDEX idx_tenant (tenant_id),
  UNIQUE KEY uk_order_no (tenant_id, order_no)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- 进货单
CREATE TABLE IF NOT EXISTS purchases (
  id BIGINT PRIMARY KEY,
  tenant_id BIGINT NOT NULL,
  shop_id BIGINT,
  warehouse_id BIGINT,
  order_no VARCHAR(32) NOT NULL,
  supplier_id BIGINT NOT NULL,
  bill_date VARCHAR(10),
  total_amount DECIMAL(10,2) DEFAULT 0,
  paid_amount DECIMAL(10,2) DEFAULT 0,
  unpaid_amount DECIMAL(10,2) DEFAULT 0,
  status TINYINT DEFAULT 1,
  remark VARCHAR(500),
  created_by BIGINT,
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  INDEX idx_tenant (tenant_id),
  UNIQUE KEY uk_order_no (tenant_id, order_no)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- 采购退货
CREATE TABLE IF NOT EXISTS purchase_returns (
  id BIGINT PRIMARY KEY,
  tenant_id BIGINT NOT NULL,
  shop_id BIGINT,
  warehouse_id BIGINT,
  order_no VARCHAR(32) NOT NULL,
  supplier_id BIGINT NOT NULL,
  bill_date VARCHAR(10),
  total_amount DECIMAL(10,2) DEFAULT 0,
  refund_amount DECIMAL(10,2) DEFAULT 0,
  status TINYINT DEFAULT 1,
  remark VARCHAR(500),
  created_by BIGINT,
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  INDEX idx_tenant (tenant_id),
  UNIQUE KEY uk_order_no (tenant_id, order_no)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- 销售订单
CREATE TABLE IF NOT EXISTS sales_orders (
  id BIGINT PRIMARY KEY,
  tenant_id BIGINT NOT NULL,
  shop_id BIGINT,
  order_no VARCHAR(32) NOT NULL,
  customer_id BIGINT NOT NULL,
  order_date VARCHAR(10),
  total_amount DECIMAL(10,2) DEFAULT 0,
  status TINYINT DEFAULT 1,
  remark VARCHAR(500),
  created_by BIGINT,
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  INDEX idx_tenant (tenant_id),
  UNIQUE KEY uk_order_no (tenant_id, order_no)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- 销货单
CREATE TABLE IF NOT EXISTS sales (
  id BIGINT PRIMARY KEY,
  tenant_id BIGINT NOT NULL,
  shop_id BIGINT,
  warehouse_id BIGINT,
  order_no VARCHAR(32) NOT NULL,
  customer_id BIGINT NOT NULL,
  bill_date VARCHAR(10),
  total_amount DECIMAL(10,2) DEFAULT 0,
  received_amount DECIMAL(10,2) DEFAULT 0,
  unreceived_amount DECIMAL(10,2) DEFAULT 0,
  status TINYINT DEFAULT 1,
  remark VARCHAR(500),
  created_by BIGINT,
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  INDEX idx_tenant (tenant_id),
  UNIQUE KEY uk_order_no (tenant_id, order_no)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- 销货退货
CREATE TABLE IF NOT EXISTS sales_returns (
  id BIGINT PRIMARY KEY,
  tenant_id BIGINT NOT NULL,
  shop_id BIGINT,
  warehouse_id BIGINT,
  order_no VARCHAR(32) NOT NULL,
  customer_id BIGINT NOT NULL,
  bill_date VARCHAR(10),
  total_amount DECIMAL(10,2) DEFAULT 0,
  refund_amount DECIMAL(10,2) DEFAULT 0,
  status TINYINT DEFAULT 1,
  remark VARCHAR(500),
  created_by BIGINT,
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  INDEX idx_tenant (tenant_id),
  UNIQUE KEY uk_order_no (tenant_id, order_no)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- 库存余额
CREATE TABLE IF NOT EXISTS stock_balance (
  id BIGINT PRIMARY KEY,
  tenant_id BIGINT NOT NULL,
  shop_id BIGINT,
  warehouse_id BIGINT NOT NULL,
  goods_id BIGINT NOT NULL,
  quantity INT DEFAULT 0,
  cost_price DECIMAL(10,2) DEFAULT 0,
  total_cost DECIMAL(12,2) DEFAULT 0,
  last_in_time DATETIME,
  last_out_time DATETIME,
  UNIQUE KEY uk_stock (tenant_id, warehouse_id, goods_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- 库存流水
CREATE TABLE IF NOT EXISTS stock_logs (
  id BIGINT PRIMARY KEY,
  tenant_id BIGINT NOT NULL,
  shop_id BIGINT,
  warehouse_id BIGINT,
  goods_id BIGINT NOT NULL,
  type TINYINT NOT NULL COMMENT '1=进货入库 2=销货出库 3=退货入库 4=退货出库 5=调拨 6=其他入库 7=其他出库 8=盘盈 9=盘亏',
  quantity INT NOT NULL,
  before_stock INT DEFAULT 0,
  after_stock INT DEFAULT 0,
  related_type VARCHAR(32),
  related_id BIGINT,
  related_no VARCHAR(32),
  remark VARCHAR(255),
  created_by BIGINT,
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  INDEX idx_tenant_goods (tenant_id, goods_id),
  INDEX idx_created (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- 结算账户
CREATE TABLE IF NOT EXISTS accounts (
  id BIGINT PRIMARY KEY,
  tenant_id BIGINT NOT NULL,
  name VARCHAR(128) NOT NULL,
  type TINYINT DEFAULT 1 COMMENT '1=现金 2=银行 3=在线',
  balance DECIMAL(12,2) DEFAULT 0,
  sort INT DEFAULT 0,
  status TINYINT DEFAULT 1,
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  INDEX idx_tenant (tenant_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- 收款单
CREATE TABLE IF NOT EXISTS receipts (
  id BIGINT PRIMARY KEY,
  tenant_id BIGINT NOT NULL,
  shop_id BIGINT,
  order_no VARCHAR(32) NOT NULL,
  customer_id BIGINT,
  bill_date VARCHAR(10),
  amount DECIMAL(10,2) DEFAULT 0,
  account_id BIGINT,
  status TINYINT DEFAULT 1,
  remark VARCHAR(500),
  created_by BIGINT,
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  INDEX idx_tenant (tenant_id),
  UNIQUE KEY uk_order_no (tenant_id, order_no)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- 付款单
CREATE TABLE IF NOT EXISTS payments (
  id BIGINT PRIMARY KEY,
  tenant_id BIGINT NOT NULL,
  shop_id BIGINT,
  order_no VARCHAR(32) NOT NULL,
  supplier_id BIGINT,
  bill_date VARCHAR(10),
  amount DECIMAL(10,2) DEFAULT 0,
  account_id BIGINT,
  status TINYINT DEFAULT 1,
  remark VARCHAR(500),
  created_by BIGINT,
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  INDEX idx_tenant (tenant_id),
  UNIQUE KEY uk_order_no (tenant_id, order_no)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- 系统设置
CREATE TABLE IF NOT EXISTS settings (
  id BIGINT PRIMARY KEY,
  tenant_id BIGINT NOT NULL,
  config_key VARCHAR(128) NOT NULL,
  config_value TEXT,
  UNIQUE KEY uk_tenant_key (tenant_id, config_key)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- 操作日志
CREATE TABLE IF NOT EXISTS operation_logs (
  id BIGINT PRIMARY KEY,
  tenant_id BIGINT NOT NULL,
  user_id BIGINT,
  action VARCHAR(64),
  target_type VARCHAR(32),
  target_id BIGINT,
  detail TEXT,
  ip VARCHAR(45),
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  INDEX idx_tenant (tenant_id),
  INDEX idx_created (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
