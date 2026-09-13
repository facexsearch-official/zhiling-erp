# PISA 进销存

智能进销存管理系统，支持多商户、分店、物理分表。

## 技术栈

- **后端**: Go + Gin + GORM
- **数据库**: MySQL (物理分片，tenant_id % 16)
- **前端**: 纯 HTML/CSS/JS SPA（无框架依赖）
- **认证**: JWT

## 快速开始

```bash
# 配置
cp config.yaml.example config.yaml
# 编辑 config.yaml 填入数据库连接信息

# 运行
go run ./cmd/server/
```

浏览器访问 `http://localhost:8080`

## 默认账号

| 账号 | 密码 |
|------|------|
| admin | 123456 |

## 项目结构

```
cmd/server/        # 入口
internal/
  config/          # 配置加载
  db/              # 数据库连接、分片路由、种子数据
  handler/         # API 处理器
  middleware/       # 认证、租户、套餐限制
  model/           # 数据模型
  pkg/             # JWT、雪花ID、统一响应
  repository/      # 数据访问层
  service/         # 业务逻辑层
web/               # 前端静态文件
sql/               # 数据库建表脚本
```

## 模块

进货 | 销售 | 库存 | 财务 | 报表 | 设置（基础资料 + 系统配置）
