-- PISA 进销存系统 - 主库建表脚本 (mysql_main)

CREATE DATABASE IF NOT EXISTS pisa_main DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
USE pisa_main;

-- 套餐
CREATE TABLE IF NOT EXISTS plans (
  id BIGINT PRIMARY KEY,
  code VARCHAR(32) NOT NULL UNIQUE,
  name VARCHAR(64) NOT NULL,
  price DECIMAL(10,2) DEFAULT 0,
  max_goods INT DEFAULT 0 COMMENT '0=不限',
  max_staff INT DEFAULT 0,
  max_shops INT DEFAULT 0,
  max_orders INT DEFAULT 0,
  sort INT DEFAULT 0,
  status TINYINT DEFAULT 1,
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- 商户
CREATE TABLE IF NOT EXISTS tenants (
  id BIGINT PRIMARY KEY,
  name VARCHAR(128) NOT NULL,
  type TINYINT DEFAULT 1 COMMENT '1=个体户 2=有限公司 3=合伙企业',
  contact_name VARCHAR(64),
  contact_phone VARCHAR(20),
  owner_user_id BIGINT NOT NULL,
  plan_id BIGINT DEFAULT 1,
  plan_expires_at DATETIME,
  max_goods INT DEFAULT 100,
  max_staff INT DEFAULT 1,
  max_shops INT DEFAULT 1,
  max_orders INT DEFAULT 100,
  status TINYINT DEFAULT 1 COMMENT '1=正常 0=冻结 -1=注销',
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  INDEX idx_owner (owner_user_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- 用户
CREATE TABLE IF NOT EXISTS users (
  id BIGINT PRIMARY KEY,
  phone VARCHAR(20) NOT NULL UNIQUE,
  password_hash VARCHAR(255) NOT NULL,
  nickname VARCHAR(64),
  avatar VARCHAR(255),
  default_tenant_id BIGINT,
  status TINYINT DEFAULT 1,
  last_login_at DATETIME,
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- 用户-商户关系
CREATE TABLE IF NOT EXISTS user_tenants (
  id BIGINT PRIMARY KEY,
  user_id BIGINT NOT NULL,
  tenant_id BIGINT NOT NULL,
  is_owner TINYINT DEFAULT 0,
  role TINYINT DEFAULT 3 COMMENT '1=超级管理员 2=管理员 3=操作员',
  staff_name VARCHAR(64),
  staff_phone VARCHAR(20),
  permissions TEXT,
  invited_by BIGINT,
  status TINYINT DEFAULT 1,
  joined_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  UNIQUE KEY uk_user_tenant (user_id, tenant_id),
  INDEX idx_tenant (tenant_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- 门店
CREATE TABLE IF NOT EXISTS shops (
  id BIGINT PRIMARY KEY,
  tenant_id BIGINT NOT NULL,
  name VARCHAR(128) NOT NULL,
  address VARCHAR(255),
  phone VARCHAR(20),
  logo VARCHAR(255),
  is_main TINYINT DEFAULT 0,
  status TINYINT DEFAULT 1,
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  INDEX idx_tenant (tenant_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- 邀请记录
CREATE TABLE IF NOT EXISTS tenant_invitations (
  id BIGINT PRIMARY KEY,
  tenant_id BIGINT NOT NULL,
  phone VARCHAR(20) NOT NULL,
  role TINYINT DEFAULT 3,
  permissions TEXT,
  invited_by BIGINT NOT NULL,
  token VARCHAR(128),
  status TINYINT DEFAULT 1 COMMENT '1=待接受 2=已接受 3=已过期 4=已取消',
  expires_at DATETIME,
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  INDEX idx_tenant (tenant_id),
  INDEX idx_token (token)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- 初始化套餐数据
INSERT INTO plans (id, code, name, price, max_goods, max_staff, max_shops, max_orders, sort) VALUES
(1, 'free', '免费版', 0, 100, 1, 1, 100, 1),
(2, 'basic', '开单版', 398, 500, 3, 1, 0, 2),
(3, 'single', '单店版', 998, 2000, 10, 1, 0, 3),
(4, 'multi', '多店版', 1998, 0, 0, 0, 0, 4)
ON DUPLICATE KEY UPDATE name=VALUES(name);

-- 初始化管理员账号: admin / 123456
INSERT INTO users (id, phone, password_hash, nickname, default_tenant_id, status) VALUES
(1, 'admin', '$2a$10$welxxcSaxV160nMGUPkTYuyHfUrTD6lP428ii7PSOwqHn7pfQPmLq', '管理员', 1, 1)
ON DUPLICATE KEY UPDATE nickname=VALUES(nickname);

-- 初始化默认商户
INSERT INTO tenants (id, name, type, contact_name, contact_phone, owner_user_id, plan_id, max_goods, max_staff, max_shops, max_orders, status) VALUES
(1, '演示商户', 1, '管理员', 'admin', 1, 3, 2000, 10, 1, 0, 1)
ON DUPLICATE KEY UPDATE name=VALUES(name);

-- 管理员加入商户（主账号）
INSERT INTO user_tenants (id, user_id, tenant_id, is_owner, role, staff_name, staff_phone, status) VALUES
(1, 1, 1, 1, 1, '管理员', 'admin', 1)
ON DUPLICATE KEY UPDATE staff_name=VALUES(staff_name);

-- 初始化默认门店
INSERT INTO shops (id, tenant_id, name, address, is_main, status) VALUES
(1, 1, '总店', '演示地址', 1, 1)
ON DUPLICATE KEY UPDATE name=VALUES(name);
