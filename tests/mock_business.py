#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
PISA 进销存 — Mock 数据造数脚本（Python）

为「客户 / 进货 / 销售 / 库存 / 资金 / 分析」六大模块的全部子模块生成测试数据：
  - 交易单据：每类 1000 条（含明细）
  - 主数据/子模块：每类 100 条
测试租户与真实数据隔离（tenant_id=7000000000000000001），脚本幂等（先清理再写入）。

用法:
    python3 tests/mockdata.py
"""
import datetime
import random
import sys

try:
    import pymysql
except ImportError:
    sys.exit("缺少依赖 pymysql，请先执行: pip install pymysql")

# ── 配置 ──
import os
DB = dict(host=os.environ.get("PISA_DB_HOST", "127.0.0.1"), port=int(os.environ.get("PISA_DB_PORT", "3306")),
          user=os.environ.get("PISA_DB_USER", "root"), password=os.environ.get("PISA_DB_PASS", ""),
          database=os.environ.get("PISA_DB_NAME", "pisa"), charset="utf8mb4", autocommit=False)
TEST_TENANT_ID = 7000000000000000001
TEST_USER_ID = 7000000000000000002
N = 1000   # 交易单据
M = 100    # 主数据/子模块最小记录数

# 运行期租户（默认测试租户；可由 --tenant 覆盖）
TENANT_ID = TEST_TENANT_ID

random.seed(20260927)
_idc = 7100000000000000000


def set_id_base(tenant_id):
    """按商户派生独立 ID 段，避免不同商户之间主键冲突。"""
    global _idc
    _idc = 7100000000000000000 + (tenant_id % 1000000) * 100000000


def nid():
    global _idc
    _idc += 1
    return _idc


def money():
    return round((random.randint(1000, 91000)) / 100.0, 2)


def qty():
    return random.randint(1, 50)


def phone():
    return "13%09d" % random.randint(0, 999999999)


def date_ago(days=365):
    return (datetime.date.today() - datetime.timedelta(days=random.randint(0, days))).isoformat()


def dt_ago(days=365):
    return datetime.datetime.now() - datetime.timedelta(days=random.randint(0, days))


def now():
    return datetime.datetime.now()


def insert_many(cur, table, cols, rows):
    if not rows:
        return
    sql = "INSERT INTO `%s` (%s) VALUES (%s)" % (
        table,
        ",".join("`%s`" % c for c in cols),
        ",".join(["%s"] * len(cols)),
    )
    cur.executemany(sql, rows)


CLEAN_TABLES = [
    # 子表优先（外键）
    "purchase_order_items", "purchase_orders", "purchase_items", "purchases",
    "purchase_return_items", "purchase_returns", "sale_order_items", "sale_orders",
    "sale_items", "sales", "sales_return_items", "sales_returns", "quote_items", "quotes",
    "income_items", "incomes", "stock_count_items", "stock_counts",
    "assembly_items", "assemblies", "recipe_items", "recipes", "combo_items", "combos",
    # 主数据
    "customer_addresses", "customers", "customer_categories", "customer_prices",
    "supplier_addresses", "suppliers", "supplier_categories",
    "goods_prices", "goods_stocks", "goods_units", "goods_batches", "goods", "goods_categories",
    "goods_attributes", "goods_properties", "price_levels", "units",
    "accounts", "salesmen", "income_types", "commission_rules", "shops", "warehouses",
    "stock_balances", "stock_logs", "transfers", "receipts", "payments",
    "tenant_settings", "user_preferences", "roles",
]


def seed_tenant(cur):
    cur.execute("DELETE FROM tenants WHERE id=%s", (TENANT_ID,))
    cur.execute("DELETE FROM users WHERE id=%s", (TEST_USER_ID,))
    cur.execute("DELETE FROM user_tenants WHERE user_id=%s AND tenant_id=%s", (TEST_USER_ID, TENANT_ID))
    cur.execute(
        "INSERT INTO tenants (id,name,type,contact_name,contact_phone,status,created_at,updated_at) "
        "VALUES (%s,%s,%s,%s,%s,%s,%s,%s)",
        (TENANT_ID, "压测商户(测试专用)", 2, "测试", "13800000000", 1, now(), now()),
    )
    cur.execute(
        "INSERT INTO users (id,phone,password_hash,nickname,status,created_at,updated_at) "
        "VALUES (%s,%s,%s,%s,%s,%s,%s)",
        (TEST_USER_ID, "testseed", "x", "压测管理员", 1, now(), now()),
    )
    cur.execute(
        "INSERT INTO user_tenants (user_id,tenant_id,is_owner,role,staff_name,staff_phone,status,joined_at) "
        "VALUES (%s,%s,%s,%s,%s,%s,%s,%s)",
        (TEST_USER_ID, TENANT_ID, 1, 1, "压测管理员", "testseed", 1, now()),
    )


def clean(cur):
    total = 0
    for t in CLEAN_TABLES:
        total += cur.execute("DELETE FROM `%s` WHERE tenant_id=%%s" % t, (TENANT_ID,))
    print("  cleaned %d rows from test tenant" % total)


# ── 主数据 ──
def seed_customer_categories(cur):
    rows = []
    names = ["普通客户", "VIP客户", "批发客户", "零售客户", "企业客户"]
    for i in range(M):
        nm = names[i % len(names)] if i < len(names) else "%s%02d" % (names[i % len(names)], i // len(names) + 1)
        rows.append((nid(), TENANT_ID, nm, 0, i))
    insert_many(cur, "customer_categories", ["id", "tenant_id", "name", "parent_id", "sort"], rows)
    return [r[0] for r in rows]


def seed_customers(cur, cats):
    rows = []
    for i in range(N):
        rows.append((nid(), TENANT_ID, "测试客户%04d" % (i + 1), "KH%05d" % (i + 1), 1,
                     "联系人%d" % (i + 1), phone(), "广东省深圳市南山区", 1, now(), now(),
                     random.choice(cats), "零售价", 100, random.randint(0, 500), random.randint(0, 2000)))
    insert_many(cur, "customers",
                ["id", "tenant_id", "name", "code", "type", "contact", "phone", "address", "status",
                 "created_at", "updated_at", "category_id", "price_level", "discount", "points", "init_debt"], rows)
    return [r[0] for r in rows]


def seed_customer_addresses(cur, custs):
    rows = []
    for i in range(M):
        rows.append((nid(), TENANT_ID, random.choice(custs), "收货人%03d" % (i + 1), phone(),
                     "广东省/深圳市/南山区", "科技园路1号", 1 if i % 5 == 0 else 0))
    insert_many(cur, "customer_addresses",
                ["id", "tenant_id", "customer_id", "receiver", "phone", "region", "detail", "is_default"], rows)


def seed_supplier_categories(cur):
    rows = []
    names = ["食品供应商", "日用品供应商", "电子产品供应商", "服装供应商"]
    for i in range(M):
        nm = names[i % len(names)] if i < len(names) else "%s%02d" % (names[i % len(names)], i // len(names) + 1)
        rows.append((nid(), TENANT_ID, nm, 0, i))
    insert_many(cur, "supplier_categories", ["id", "tenant_id", "name", "parent_id", "sort"], rows)
    return [r[0] for r in rows]


def seed_suppliers(cur, cats):
    rows = []
    for i in range(N):
        rows.append((nid(), TENANT_ID, "测试供应商%04d" % (i + 1), "GYS%05d" % (i + 1),
                     "供联系人%d" % (i + 1), phone(), "浙江省杭州市", random.randint(0, 3000), 1, now(), now(),
                     random.choice(cats)))
    insert_many(cur, "suppliers",
                ["id", "tenant_id", "name", "code", "contact", "phone", "address", "init_payable",
                 "status", "created_at", "updated_at", "category_id"], rows)
    return [r[0] for r in rows]


def seed_supplier_addresses(cur, sups):
    rows = []
    for i in range(M):
        rows.append((nid(), TENANT_ID, random.choice(sups), "供联系人%03d" % (i + 1), phone(),
                     "浙江省", "杭州市", "西湖区", "文三路1号", 1 if i % 5 == 0 else 0, i))
    insert_many(cur, "supplier_addresses",
                ["id", "tenant_id", "supplier_id", "consignee", "consignee_phone", "province", "city",
                 "district", "address_detail", "is_default", "sort"], rows)


def seed_units(cur):
    rows = []
    names = ["个", "箱", "件", "瓶", "包", "盒", "kg", "米"]
    for i in range(M):
        nm = names[i % len(names)] if i < len(names) else "%s%d" % (names[i % len(names)], i // len(names) + 1)
        rows.append((nid(), TENANT_ID, nm, now()))
    insert_many(cur, "units", ["id", "tenant_id", "name", "created_at"], rows)
    return [r[0] for r in rows]


def seed_goods_categories(cur):
    rows = []
    names = ["食品饮料", "日用百货", "数码家电", "服装鞋帽", "办公用品", "生鲜果蔬"]
    for i in range(M):
        nm = names[i % len(names)] if i < len(names) else "%s%02d" % (names[i % len(names)], i // len(names) + 1)
        rows.append((nid(), TENANT_ID, nm, 0, i))
    insert_many(cur, "goods_categories", ["id", "tenant_id", "name", "parent_id", "sort"], rows)
    return [r[0] for r in rows]


def seed_goods(cur, cats, units, sups):
    rows = []
    for i in range(N):
        pp = money()
        rows.append((nid(), TENANT_ID, "测试商品%04d" % (i + 1), "SP%05d" % (i + 1), "69%011d" % (i + 1),
                     random.choice(cats), random.choice(units), random.choice(sups),
                     random.randint(0, 500), 1000, 10, pp, round(pp * 1.3, 2), round(pp * 1.15, 2),
                     1, 1, now(), now(), "标准", "测试品牌", "个"))
    insert_many(cur, "goods",
                ["id", "tenant_id", "name", "code", "barcode", "category_id", "unit_id", "supplier_id",
                 "current_stock", "max_stock", "min_stock", "purchase_price", "retail_price", "wholesale_price",
                 "cost_method", "status", "created_at", "updated_at", "spec", "brand", "main_unit"], rows)
    return [r[0] for r in rows]


def seed_goods_prices(cur, goods):
    rows = []
    for g in goods:
        pp = money()
        rows.append((nid(), TENANT_ID, g, "个", "", pp, round(pp * 1.3, 2), round(pp * 1.15, 2), 0))
    insert_many(cur, "goods_prices",
                ["id", "tenant_id", "goods_id", "unit_key", "spec_key", "purchase_price", "retail_price",
                 "wholesale_price", "sort"], rows)


def seed_goods_stocks(cur, goods):
    rows = [(nid(), TENANT_ID, g, "", random.randint(0, 500), 10, 20, 1000) for g in goods]
    insert_many(cur, "goods_stocks",
                ["id", "tenant_id", "goods_id", "spec_key", "stock", "min_stock", "safety_stock", "max_stock"], rows)


def seed_goods_units(cur, goods, units):
    rows = []
    for i in range(M):
        rows.append((nid(), TENANT_ID, random.choice(goods), random.choice(units), "个", 1,
                     1 if i % 3 == 0 else 0, i))
    insert_many(cur, "goods_units",
                ["id", "tenant_id", "goods_id", "unit_id", "unit_name", "factor", "is_main", "sort"], rows)


def seed_goods_attributes(cur):
    rows = [(nid(), TENANT_ID, "辅助属性%03d" % (i + 1), '["红色","蓝色","绿色"]', i, 1) for i in range(M)]
    insert_many(cur, "goods_attributes", ["id", "tenant_id", "name", "values", "sort", "status"], rows)


def seed_goods_properties(cur):
    rows = [(nid(), TENANT_ID, "货品属性%03d" % (i + 1), 1 + i % 2, '["A","B"]', i, 1, now()) for i in range(M)]
    insert_many(cur, "goods_properties", ["id", "tenant_id", "name", "type", "values", "sort", "status", "created_at"], rows)


def seed_price_levels(cur):
    rows = [(nid(), TENANT_ID, "价格等级%03d" % (i + 1), i, 1, now()) for i in range(M)]
    insert_many(cur, "price_levels", ["id", "tenant_id", "name", "sort", "status", "created_at"], rows)


def seed_shops(cur):
    rows = []
    for i in range(3):
        rows.append((nid(), TENANT_ID, "测试门店%d" % (i + 1), "深圳市", "", 1 if i == 0 else 0, 1, now(), 1))
    insert_many(cur, "shops", ["id", "tenant_id", "name", "address", "phone", "is_main", "status", "created_at", "type"], rows)
    return [r[0] for r in rows]


def seed_warehouses(cur, shops):
    rows = [(nid(), TENANT_ID, "测试仓库%03d" % (i + 1), 1 + i % 3, random.choice(shops), 1, now()) for i in range(M)]
    insert_many(cur, "warehouses", ["id", "tenant_id", "name", "type", "shop_id", "status", "created_at"], rows)
    return [r[0] for r in rows]


def seed_accounts(cur):
    rows = []
    names = ["现金", "微信", "支付宝", "工商银行", "建设银行", "招商银行"]
    types = [1, 3, 3, 2, 2, 2]
    for i in range(M):
        nm = names[i % len(names)] if i < len(names) else "%s%02d" % (names[i % len(names)], i // len(names) + 1)
        rows.append((nid(), TENANT_ID, nm, types[i % len(types)], round(money() * 100, 2), 1, now()))
    insert_many(cur, "accounts", ["id", "tenant_id", "name", "type", "balance", "status", "created_at"], rows)
    return [r[0] for r in rows]


def seed_salesmen(cur, shops):
    rows = [(nid(), TENANT_ID, "业务员%03d" % (i + 1), phone(), random.choice(shops), 1, now(), now()) for i in range(M)]
    insert_many(cur, "salesmen", ["id", "tenant_id", "name", "phone", "shop_id", "status", "created_at", "updated_at"], rows)
    return [r[0] for r in rows]


def seed_income_types(cur):
    base = [("销售返利", 1), ("利息收入", 1), ("运费支出", 2), ("办公费用", 2), ("其他收入", 1)]
    rows = []
    for i in range(M):
        n, d = base[i % len(base)]
        nm = n if i < len(base) else "%s%02d" % (n, i // len(base) + 1)
        rows.append((nid(), TENANT_ID, nm, d, i, 1, now()))
    insert_many(cur, "income_types", ["id", "tenant_id", "name", "direction", "sort", "status", "created_at"], rows)
    return [r[0] for r in rows]


def seed_commission_rules(cur):
    rows = [(nid(), TENANT_ID, "提成规则%03d" % (i + 1), random.randint(1, 3), 1, random.randint(1, 10), 1, now())
            for i in range(M)]
    insert_many(cur, "commission_rules",
                ["id", "tenant_id", "name", "method", "calc_type", "ratio", "status", "created_at"], rows)


# ── 交易单据 ──
def items_for(goods, cnt):
    out = []
    for _ in range(cnt):
        g = random.choice(goods)
        q = qty()
        p = money()
        out.append((g, q, p, round(q * p, 2)))
    return out


def seed_purchase_orders(cur, sups, salesmen, accts, goods, whs):
    orders, items = [], []
    for i in range(N):
        oid = nid()
        its = items_for(goods, 3)
        sub = round(sum(x[3] for x in its), 2)
        for g, q, p, a in its:
            items.append((nid(), TENANT_ID, oid, g, q, p, a, now()))
        orders.append((oid, TENANT_ID, 0, random.choice(whs), "CGYD%06d" % (i + 1), random.choice(sups),
                       random.choice(salesmen), random.choice(accts), date_ago(), sub, 100, sub, sub, sub,
                       random.randint(1, 3), dt_ago(), now()))
    insert_many(cur, "purchase_orders",
                ["id", "tenant_id", "shop_id", "warehouse_id", "order_no", "supplier_id", "salesman_id",
                 "account_id", "order_date", "subtotal", "discount", "discounted_amount", "total_amount",
                 "unpaid_amount", "status", "created_at", "updated_at"], orders)
    insert_many(cur, "purchase_order_items",
                ["id", "tenant_id", "order_id", "goods_id", "quantity", "unit_price", "amount", "created_at"], items)


def seed_purchases(cur, sups, salesmen, accts, goods, whs):
    orders, items = [], []
    for i in range(N):
        oid = nid()
        its = items_for(goods, 3)
        sub = round(sum(x[3] for x in its), 2)
        for g, q, p, a in its:
            items.append((nid(), TENANT_ID, oid, g, q, p, a, now()))
        paid = sub if random.randint(0, 1) == 0 else 0.0
        orders.append((oid, TENANT_ID, 0, random.choice(whs), "CG%06d" % (i + 1), random.choice(sups),
                       random.choice(salesmen), random.choice(accts), date_ago(), sub, 100, sub, sub, paid,
                       round(sub - paid, 2), 3, dt_ago(), now()))
    insert_many(cur, "purchases",
                ["id", "tenant_id", "shop_id", "warehouse_id", "order_no", "supplier_id", "salesman_id",
                 "account_id", "bill_date", "subtotal", "discount", "discounted_amount", "total_amount",
                 "paid_amount", "unpaid_amount", "status", "created_at", "updated_at"], orders)
    insert_many(cur, "purchase_items",
                ["id", "tenant_id", "purchase_id", "goods_id", "quantity", "unit_price", "amount", "created_at"], items)


def seed_purchase_returns(cur, sups, salesmen, accts, goods, whs):
    orders, items = [], []
    for i in range(N):
        oid = nid()
        its = items_for(goods, 2)
        sub = round(sum(x[3] for x in its), 2)
        for g, q, p, a in its:
            items.append((nid(), TENANT_ID, oid, g, q, p, a, now()))
        orders.append((oid, TENANT_ID, 0, random.choice(whs), "CGTH%06d" % (i + 1), random.choice(sups),
                       random.choice(salesmen), random.choice(accts), date_ago(), sub, sub, 3, dt_ago()))
    insert_many(cur, "purchase_returns",
                ["id", "tenant_id", "shop_id", "warehouse_id", "order_no", "supplier_id", "salesman_id",
                 "account_id", "bill_date", "total_amount", "refund_amount", "status", "created_at"], orders)
    insert_many(cur, "purchase_return_items",
                ["id", "tenant_id", "purchase_return_id", "goods_id", "quantity", "unit_price", "amount", "created_at"], items)


def seed_sale_orders(cur, custs, salesmen, accts, goods, whs):
    orders, items = [], []
    for i in range(N):
        oid = nid()
        its = items_for(goods, 3)
        sub = round(sum(x[3] for x in its), 2)
        for g, q, p, a in its:
            items.append((nid(), TENANT_ID, oid, g, q, p, a, now()))
        orders.append((oid, TENANT_ID, 0, random.choice(whs), "XSYD%06d" % (i + 1), random.choice(custs),
                       random.choice(salesmen), random.choice(accts), date_ago(), sub, sub, 0,
                       random.randint(1, 3), dt_ago(), now()))
    insert_many(cur, "sale_orders",
                ["id", "tenant_id", "shop_id", "warehouse_id", "order_no", "customer_id", "salesman_id",
                 "account_id", "order_date", "total_amount", "received_amount", "unreceived_amount",
                 "status", "created_at", "updated_at"], orders)
    insert_many(cur, "sale_order_items",
                ["id", "tenant_id", "order_id", "goods_id", "quantity", "unit_price", "amount", "created_at"], items)


def seed_sales(cur, custs, salesmen, accts, goods, whs):
    orders, items = [], []
    for i in range(N):
        oid = nid()
        its = items_for(goods, 3)
        sub = round(sum(x[3] for x in its), 2)
        for g, q, p, a in its:
            items.append((nid(), TENANT_ID, oid, g, q, p, a, now()))
        recv = sub if random.randint(0, 1) == 0 else 0.0
        rs = 2 if recv else 0
        orders.append((oid, TENANT_ID, 0, random.choice(whs), "XS%06d" % (i + 1), random.choice(custs),
                       random.choice(salesmen), random.choice(accts), date_ago(), 100, sub, sub, recv,
                       round(sub - recv, 2), rs, 1, dt_ago(), now()))
    insert_many(cur, "sales",
                ["id", "tenant_id", "shop_id", "warehouse_id", "order_no", "customer_id", "salesman_id",
                 "account_id", "bill_date", "discount", "subtotal", "total_amount", "received_amount",
                 "unreceived_amount", "receive_status", "status", "created_at", "updated_at"], orders)
    insert_many(cur, "sale_items",
                ["id", "tenant_id", "sale_id", "goods_id", "quantity", "unit_price", "amount", "created_at"], items)


def seed_sales_returns(cur, custs, salesmen, accts, goods, whs):
    orders, items = [], []
    for i in range(N):
        oid = nid()
        its = items_for(goods, 2)
        sub = round(sum(x[3] for x in its), 2)
        for g, q, p, a in its:
            items.append((nid(), TENANT_ID, oid, g, q, p, a, now()))
        orders.append((oid, TENANT_ID, 0, random.choice(whs), "XSTH%06d" % (i + 1), random.choice(custs),
                       random.choice(salesmen), random.choice(accts), date_ago(), sub, sub, 1, dt_ago()))
    insert_many(cur, "sales_returns",
                ["id", "tenant_id", "shop_id", "warehouse_id", "order_no", "customer_id", "salesman_id",
                 "account_id", "bill_date", "total_amount", "received_amount", "status", "created_at"], orders)
    insert_many(cur, "sales_return_items",
                ["id", "tenant_id", "return_id", "goods_id", "quantity", "unit_price", "amount", "created_at"], items)


def seed_quotes(cur, custs, salesmen, goods):
    orders, items = [], []
    for i in range(N):
        oid = nid()
        its = items_for(goods, 3)
        sub = round(sum(x[3] for x in its), 2)
        for g, q, p, a in its:
            items.append((nid(), TENANT_ID, oid, g, q, p, a, now()))
        orders.append((oid, TENANT_ID, 0, "BJ%06d" % (i + 1), random.choice(custs),
                       random.choice(salesmen), date_ago(), sub, dt_ago()))
    insert_many(cur, "quotes",
                ["id", "tenant_id", "shop_id", "order_no", "customer_id", "salesman_id", "bill_date",
                 "total_amount", "created_at"], orders)
    insert_many(cur, "quote_items",
                ["id", "tenant_id", "quote_id", "goods_id", "quantity", "unit_price", "amount", "created_at"], items)


def seed_stock_balances(cur, goods, whs):
    rows = []
    wid = whs[0]
    for g in goods[:N]:
        q = random.randint(20, 520)
        cp = money()
        rows.append((nid(), TENANT_ID, 0, wid, g, q, cp, round(q * cp, 2)))
    insert_many(cur, "stock_balances",
                ["id", "tenant_id", "shop_id", "warehouse_id", "goods_id", "quantity", "cost_price", "total_cost"], rows)


def seed_stock_logs(cur, goods, whs):
    rows = []
    for i in range(N):
        before = random.randint(0, 500)
        q = qty()
        rows.append((nid(), TENANT_ID, 0, random.choice(whs), random.choice(goods), random.choice([1, 2, 3, 4]),
                     q, before, before + q, "purchase", "CG%06d" % (i + 1), "压测流水", dt_ago()))
    insert_many(cur, "stock_logs",
                ["id", "tenant_id", "shop_id", "warehouse_id", "goods_id", "type", "quantity", "before_stock",
                 "after_stock", "related_type", "related_no", "remark", "created_at"], rows)


def seed_transfers(cur, accts):
    rows = []
    for i in range(N):
        a = random.choice(accts)
        b = random.choice(accts)
        while b == a:
            b = random.choice(accts)
        rows.append((nid(), TENANT_ID, "ZZ%06d" % (i + 1), date_ago(), a, b, money(), "压测转账", dt_ago()))
    insert_many(cur, "transfers",
                ["id", "tenant_id", "order_no", "bill_date", "from_account_id", "to_account_id", "amount",
                 "remark", "created_at"], rows)


def seed_receipts(cur, custs, accts, salesmen):
    rows = [(nid(), TENANT_ID, 0, "SK%06d" % (i + 1), "销售收款", random.choice(custs), random.choice(salesmen),
             date_ago(), money(), random.choice(accts), 1, dt_ago()) for i in range(N)]
    insert_many(cur, "receipts",
                ["id", "tenant_id", "shop_id", "order_no", "type", "customer_id", "salesman_id", "bill_date",
                 "amount", "account_id", "status", "created_at"], rows)


def seed_payments(cur, sups, accts, salesmen):
    rows = [(nid(), TENANT_ID, 0, "FK%06d" % (i + 1), "进货付款", random.choice(sups), random.choice(salesmen),
             date_ago(), money(), random.choice(accts), 1, dt_ago()) for i in range(N)]
    insert_many(cur, "payments",
                ["id", "tenant_id", "shop_id", "order_no", "type", "supplier_id", "salesman_id", "bill_date",
                 "amount", "account_id", "status", "created_at"], rows)


def seed_incomes(cur, itypes, accts, salesmen):
    rows, items = [], []
    for i in range(N):
        iid = nid()
        direction = 1 if random.randint(0, 1) else 2
        amt = money()
        acc = random.choice(accts)
        rows.append((iid, TENANT_ID, 0, "QT%06d" % (i + 1), direction, random.choice(itypes), date_ago(),
                     amt, acc, random.choice(salesmen), 1, dt_ago()))
        items.append((nid(), TENANT_ID, iid, acc, amt))
    insert_many(cur, "incomes",
                ["id", "tenant_id", "shop_id", "order_no", "direction", "type_id", "bill_date", "amount",
                 "account_id", "salesman_id", "status", "created_at"], rows)
    insert_many(cur, "income_items", ["id", "tenant_id", "income_id", "account_id", "amount"], items)


def seed_stock_counts(cur, goods, salesmen, whs):
    rows, items = [], []
    for i in range(N):
        cid = nid()
        book = random.randint(0, 500)
        actual = book + random.randint(-10, 10)
        for _ in range(2):
            bq = random.randint(0, 200)
            aq = bq + random.randint(-5, 5)
            items.append((nid(), TENANT_ID, cid, random.choice(goods), bq, aq, aq - bq))
        rows.append((cid, TENANT_ID, 0, random.choice(whs), "PD%06d" % (i + 1), random.choice(salesmen),
                     date_ago(), book, actual, actual - book, 1, dt_ago()))
    insert_many(cur, "stock_counts",
                ["id", "tenant_id", "shop_id", "warehouse_id", "order_no", "salesman_id", "bill_date",
                 "book_qty", "actual_qty", "diff_qty", "status", "created_at"], rows)
    insert_many(cur, "stock_count_items",
                ["id", "tenant_id", "count_id", "goods_id", "book_qty", "actual_qty", "diff_qty"], items)


def seed_assemblies(cur, goods, salesmen, accts):
    rows, items = [], []
    for i in range(N):
        aid = nid()
        typ = 1 if random.randint(0, 1) else 2
        prefix = "ZZD" if typ == 1 else "CFD"
        for k in range(3):
            kind = 1 if k == 0 else 2
            uc = money()
            q = qty()
            items.append((nid(), TENANT_ID, aid, kind, random.choice(goods), q, uc, round(q * uc, 2)))
        rows.append((aid, TENANT_ID, "%s%06d" % (prefix, i + 1), typ, random.choice(salesmen), date_ago(),
                     money(), random.choice(accts), 1, dt_ago()))
    insert_many(cur, "assemblies",
                ["id", "tenant_id", "order_no", "type", "salesman_id", "bill_date", "assembly_fee",
                 "account_id", "status", "created_at"], rows)
    insert_many(cur, "assembly_items",
                ["id", "tenant_id", "assembly_id", "kind", "goods_id", "quantity", "unit_cost", "amount"], items)


def seed_recipes(cur, goods):
    rows, items = [], []
    for i in range(N):
        rid = nid()
        for _ in range(3):
            items.append((nid(), TENANT_ID, rid, random.choice(goods), qty(), money()))
        rows.append((rid, TENANT_ID, random.choice(goods), qty(), "压测配方", dt_ago()))
    insert_many(cur, "recipes", ["id", "tenant_id", "goods_id", "quantity", "remark", "created_at"], rows)
    insert_many(cur, "recipe_items",
                ["id", "tenant_id", "recipe_id", "goods_id", "quantity", "unit_cost"], items)


def seed_combos(cur, goods, units, cats):
    rows, items = [], []
    for i in range(N):
        cid = nid()
        tp = tr = 0.0
        for _ in range(3):
            pp = money()
            tp += pp
            tr += pp * 1.3
            items.append((nid(), TENANT_ID, cid, random.choice(goods), qty(), pp, round(pp * 1.3, 2), round(pp * 1.15, 2)))
        rows.append((cid, TENANT_ID, "测试套餐%04d" % (i + 1), random.choice(units), "个", random.choice(cats),
                     round(tp, 2), round(tr, 2), round(tr, 2), round(tr * 0.9, 2), random.randint(0, 100), 1, dt_ago(), now()))
    insert_many(cur, "combos",
                ["id", "tenant_id", "name", "unit_id", "unit_name", "category_id", "total_purchase", "total_retail",
                 "retail_price", "wholesale_price", "sold_count", "status", "created_at", "updated_at"], rows)
    insert_many(cur, "combo_items",
                ["id", "tenant_id", "combo_id", "goods_id", "quantity", "purchase_price", "retail_price",
                 "wholesale_price"], items)


def seed_batches(cur, goods, sups):
    rows = []
    for i in range(N):
        prod = datetime.date.today() - datetime.timedelta(days=random.randint(0, 360))
        exp = prod + datetime.timedelta(days=365)
        rows.append((nid(), TENANT_ID, random.choice(goods), "PC%06d" % (i + 1), prod.isoformat(),
                     exp.isoformat(), random.randint(0, 300), random.choice(sups), "测试品牌", dt_ago()))
    insert_many(cur, "goods_batches",
                ["id", "tenant_id", "goods_id", "batch_no", "production_date", "expiry_date", "stock",
                 "supplier_id", "brand", "created_at"], rows)


def seed_customer_prices(cur, custs, goods):
    rows = [(nid(), TENANT_ID, random.choice(custs), random.choice(goods), money(), now(), now()) for _ in range(N)]
    insert_many(cur, "customer_prices",
                ["id", "tenant_id", "customer_id", "goods_id", "price", "created_at", "updated_at"], rows)


COUNTS = ["customers", "customer_categories", "customer_addresses", "suppliers", "supplier_categories",
          "supplier_addresses", "goods", "goods_prices", "goods_stocks", "goods_units", "goods_attributes",
          "goods_properties", "price_levels", "accounts", "salesmen", "income_types", "commission_rules",
          "shops", "warehouses", "purchase_orders", "purchase_order_items", "purchases", "purchase_items",
          "purchase_returns", "sale_orders", "sales", "sales_returns", "quotes", "stock_balances", "stock_logs",
          "transfers", "receipts", "payments", "incomes", "stock_counts", "assemblies", "recipes", "combos",
          "goods_batches", "customer_prices"]


def main():
    global TENANT_ID
    import argparse
    ap = argparse.ArgumentParser(description="PISA 六大模块 Mock 数据")
    ap.add_argument("--tenant", type=int, default=TEST_TENANT_ID, help="目标商户 ID（默认测试租户）")
    ap.add_argument("--clean", action="store_true", help="先清空该商户的业务数据再写入（谨慎：会删除现有数据）")
    args = ap.parse_args()
    TENANT_ID = args.tenant
    set_id_base(TENANT_ID)

    conn = pymysql.connect(**DB)
    cur = conn.cursor()

    is_test = (TENANT_ID == TEST_TENANT_ID)
    if is_test:
        seed_tenant(cur)
        clean(cur)
    else:
        cur.execute("SELECT COUNT(*) FROM tenants WHERE id=%s", (TENANT_ID,))
        if cur.fetchone()[0] == 0:
            sys.exit("商户 %d 不存在" % TENANT_ID)
        if args.clean:
            print("!! --clean：将清空商户 %d 的现有业务数据" % TENANT_ID)
            clean(cur)
        else:
            print("追加模式：保留商户 %d 现有数据，仅新增 mock 数据" % TENANT_ID)
    conn.commit()
    print("connected; seeding tenant %d" % TENANT_ID)

    cust_cats = seed_customer_categories(cur)
    custs = seed_customers(cur, cust_cats)
    sup_cats = seed_supplier_categories(cur)
    sups = seed_suppliers(cur, sup_cats)
    units = seed_units(cur)
    goods_cats = seed_goods_categories(cur)
    goods = seed_goods(cur, goods_cats, units, sups)
    seed_goods_prices(cur, goods)
    seed_goods_stocks(cur, goods)
    seed_goods_units(cur, goods, units)
    seed_goods_attributes(cur)
    seed_goods_properties(cur)
    seed_customer_addresses(cur, custs)
    seed_supplier_addresses(cur, sups)
    shops = seed_shops(cur)
    whs = seed_warehouses(cur, shops)
    accts = seed_accounts(cur)
    salesmen = seed_salesmen(cur, shops)
    itypes = seed_income_types(cur)
    seed_commission_rules(cur)
    seed_price_levels(cur)

    seed_purchase_orders(cur, sups, salesmen, accts, goods, whs)
    seed_purchases(cur, sups, salesmen, accts, goods, whs)
    seed_purchase_returns(cur, sups, salesmen, accts, goods, whs)
    seed_sale_orders(cur, custs, salesmen, accts, goods, whs)
    seed_sales(cur, custs, salesmen, accts, goods, whs)
    seed_sales_returns(cur, custs, salesmen, accts, goods, whs)
    seed_quotes(cur, custs, salesmen, goods)
    seed_stock_balances(cur, goods, whs)
    seed_stock_logs(cur, goods, whs)
    seed_transfers(cur, accts)
    seed_receipts(cur, custs, accts, salesmen)
    seed_payments(cur, sups, accts, salesmen)
    seed_incomes(cur, itypes, accts, salesmen)
    seed_stock_counts(cur, goods, salesmen, whs)
    seed_assemblies(cur, goods, salesmen, accts)
    seed_recipes(cur, goods)
    seed_combos(cur, goods, units, goods_cats)
    seed_batches(cur, goods, sups)
    seed_customer_prices(cur, custs, goods)
    conn.commit()

    print("\n=== mock data seeded (tenant %d) ===" % TENANT_ID)
    for t in COUNTS:
        cur.execute("SELECT COUNT(*) FROM `%s` WHERE tenant_id=%%s" % t, (TENANT_ID,))
        print("  %-24s %d" % (t, cur.fetchone()[0]))
    cur.close()
    conn.close()


if __name__ == "__main__":
    main()
