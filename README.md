# 智领进销存 (Zhiling ERP)

> 开源的中小微企业进销存 / ERP 系统，支持多商户、多门店。

智领进销存是一套开箱即用的进销存管理系统，覆盖 商品、库存、进货、销售、资金、盘点、基础资料与权限 等完整业务闭环。后端 Go + Gin + GORM + MySQL，前端纯 HTML/CSS/JS（无框架依赖），部署简单、二开友好。

## 功能特性

- **商品管理**：商品/多规格/多单位、价格等级、辅助属性、库存预警
- **库存管理**：多门店仓库、库存查询、库存分布、成本明细、盘点、库存流水
- **进货管理**：进货单（审核/反审核/打印/作废）、进货退货
- **销售管理**：销售单、销售退货、收款状态跟踪
- **资金管理**：收款单、付款单、账户概览
- **基础资料**：客户/客户分类/价格等级、供应商/供应商分类、门店、员工与角色
- **权限与安全**：基于角色的功能权限、JWT 鉴权、按商户数据隔离
- **数据看板**：首页统计（今日/本月销售额、毛利、库存价值、近 7 日趋势、热销 Top5）

## 技术栈

- **后端**：Go + [Gin](https://github.com/gin-gonic/gin) + [GORM](https://gorm.io)
- **数据库**：MySQL 5.7+ / 8.0
- **前端**：原生 HTML / CSS / JavaScript，无框架、无构建
- **认证**：JWT
- **工具**：雪花 ID、bcrypt

## 快速开始

```bash
# 1. 准备配置
cp config.yaml.example config.yaml
# 编辑 config.yaml，填入 MySQL 连接信息与 JWT secret

# 2. 建库（如未创建）
mysql -uroot -p -e "CREATE DATABASE pisa DEFAULT CHARSET utf8mb4;"

# 3. 运行（首次启动会自动建表并初始化基础数据）
go run ./cmd/server/
```

浏览器访问 <http://localhost:8080>

## 默认账号

| 账号 | 密码 |
|------|------|
| admin | 123456 |

## 项目结构

```
cmd/server/        # 程序入口
cmd/mock/          # Mock 数据导入命令
internal/
  config/          # 配置加载
  db/              # 数据库连接、建表、种子数据、Mock
  handler/         # HTTP 处理器
  middleware/      # 认证、租户、套餐限制
  model/           # 数据模型
  pkg/             # JWT、雪花 ID、统一响应、权限
  repository/      # 数据访问层
  service/         # 业务逻辑层
web/               # 前端静态文件（SPA + 独立页面）
scripts/           # Python 工具脚本（Mock 数据等）
tests/             # Python 接口测试
sql/               # 数据库脚本
docs/              # 设计与架构文档
```

## 造数与测试

测试相关脚本均为 Python，依赖见 `tests/requirements.txt`。

```bash
pip install -r tests/requirements.txt

# 生成演示商户 Mock 数据（门店/角色/用户/客户/商品/单据等）
python3 scripts/mock_data.py

# 接口测试
python3 tests/test_category.py     # 分类接口
python3 tests/test_business.py     # 业务接口
```

## 界面截图

### 首页
![首页](docs/screenshots/home.png)

### 商品
| 商品列表 | 新增商品 |
|:---:|:---:|
| ![商品列表](docs/screenshots/goods-list.png) | ![新增商品](docs/screenshots/goods-add.png) |

### 客户
| 客户列表 | 新增客户 |
|:---:|:---:|
| ![客户列表](docs/screenshots/customer-list.png) | ![新增客户](docs/screenshots/customer-add.png) |

### 供应商
| 供应商列表 | 新增供应商 |
|:---:|:---:|
| ![供应商列表](docs/screenshots/supplier-list.png) | ![新增供应商](docs/screenshots/supplier-add.png) |

### 进货
| 进货单列表 | 新增进货单 |
|:---:|:---:|
| ![进货单列表](docs/screenshots/purchase-list.png) | ![新增进货单](docs/screenshots/purchase-add.png) |

| 进货退货列表 | 新增进货退货单 |
|:---:|:---:|
| ![进货退货列表](docs/screenshots/purchase-return-list.png) | ![新增进货退货单](docs/screenshots/purchase-return-add.png) |

### 销售
| 销售单列表 | 新增销售单 |
|:---:|:---:|
| ![销售单列表](docs/screenshots/sale-list.png) | ![新增销售单](docs/screenshots/sale-add.png) |

| 销售退货单列表 | 新增销售退货单 |
|:---:|:---:|
| ![销售退货单列表](docs/screenshots/sale-return-list.png) | ![新增销售退货单](docs/screenshots/sale-return-add.png) |

### 库存
| 库存查询 | 盘点单列表 | 新增盘点单 |
|:---:|:---:|:---:|
| ![库存查询](docs/screenshots/stock-query.png) | ![盘点单列表](docs/screenshots/stock-count-list.png) | ![新增盘点单](docs/screenshots/stock-count-add.png) |

### 资金
| 收款列表 | 新增收款单 | 付款单列表 |
|:---:|:---:|:---:|
| ![收款列表](docs/screenshots/receipt-list.png) | ![新增收款单](docs/screenshots/receipt-add.png) | ![付款单列表](docs/screenshots/payment-list.png) |

| 新增付款单 | 账号概览 |
|:---:|:---:|
| ![新增付款单](docs/screenshots/payment-add.png) | ![账号概览](docs/screenshots/account-overview.png) |

### 设置
| 门店管理 | 员工管理 | 角色权限 | 新增角色 |
|:---:|:---:|:---:|:---:|
| ![门店管理](docs/screenshots/shop-manage.png) | ![员工管理](docs/screenshots/staff-manage.png) | ![角色权限](docs/screenshots/role-list.png) | ![新增角色](docs/screenshots/role-add.png) |

| 商户信息 |
|:---:|
| ![商户信息](docs/screenshots/merchant-info.png) |

## 许可证

本项目基于 [MIT](./LICENSE) 许可证开源。

Copyright (c) 2026 facexsearch
