# PISA 进销存 — 测试脚本

六大模块（客户 / 进货 / 销售 / 库存 / 资金 / 分析）全功能接口测试与 Mock 数据造数，使用 **Python**。

## 依赖

```bash
pip install -r tests/requirements.txt   # pymysql, requests
```

## 1. 造数：`tests/mock_business.py`

为测试租户 `7000000000000000001`（与真实数据隔离）生成测试数据，**脚本幂等**（先清理再写入）。

- 交易单据：每类 **1000** 条（含明细）
- 主数据/子模块：每类 **100** 条

```bash
python3 tests/mock_business.py
```

> 指定商户（默认测试租户）：
> ```bash
> # 追加到真实商户（保留其现有数据，按商户派生 ID 段避免冲突）
> python3 tests/mock_business.py --tenant 2100945443734163456
> # 谨慎：先清空该商户业务数据再写入
> python3 tests/mock_business.py --tenant 2100945443734163456 --clean
> ```

覆盖实体：客户/客户分类/客户地址、供应商/供应商分类/供应商地址、商品/价格/库存/多单位/分类/单位/辅助属性/货品属性、价格等级、仓库、账户、业务员、收支类型、提成规则、进货单·预订·退货、销售单·预订·退货·报价、库存余额·流水、转账、收款、付款、其他收入·支出、盘点、组装、拆分、配方、套餐、批次、客户报价。

### 客户分类专项：`tests/mock_customer_category.py`

单独为「客户分类」生成数据（默认 100 条，含两级父/子分类），默认作用于测试租户。

```bash
python3 tests/mock_customer_category.py            # 幂等：缺多少补多少
python3 tests/mock_customer_category.py --count 200
python3 tests/mock_customer_category.py --clean    # 先清空再生成
python3 tests/mock_customer_category.py --tenant 2100945443734163456 --clean
```


## 2. 测试：`tests/test_business.py`

覆盖上述模块全部子功能的**读取 + 写入**（新增/修改/删除/审核/作废等），共 **205** 条用例。

```bash
python3 tests/test_business.py
```

- 直连运行中的服务 `http://127.0.0.1:8080`
- JWT 由 `config.yaml` 的 `secret` 现场签发（测试租户主账号）
- 退出码：全部通过为 0，否则为 1（便于 CI）

### 分类接口专项：`tests/test_category.py`

覆盖「商品分类 / 客户分类 / 供应商分类」的 **列表 / 新增子类(字符串 parent_id) / 新增顶级(空 parent_id) / 修改 / 列表校验 / 删除**，共 30 条用例。

```bash
python3 tests/test_category.py
# 指定服务地址 / 商户 / 用户
PISA_BASE=http://127.0.0.1:8080 PISA_TENANT_ID=2104556395360686080 python3 tests/test_category.py
```

> 默认商户为「演示商户」，用户取该商户主账号；退出码 0 表示全部通过。

## 建议流程

```bash
# 启动服务（后台）
go build -o server ./cmd/server/ && (nohup ./server > /tmp/pisa_server.log 2>&1 &)

# 造数 + 测试
python3 tests/mock_business.py
python3 tests/test_business.py
```

> 提示：`test_business.py` 的「货品属性-删除」用例会删除 1 条货品属性（该模块业务上限为 6），如需保持 100 条可重跑 `mock_business.py`。
>
> 另：`tests/mock_data.py`、`tests/test_settings.py` 为「货品」模块的历史脚本，与本脚本相互独立。
