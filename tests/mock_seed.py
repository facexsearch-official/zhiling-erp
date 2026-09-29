#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
PISA 进销存 — 初始化 Mock 造数脚本

流程:
  1) 读取已初始化的 admin 账号与主商户（演示商户）
  2) 为 admin 创建多个商户（多商户）
  3) 为主商户 mock: 客户分类 / 客户 / 规格管理 / 单位管理 / 商品分类 / 商品信息
用法:
  python3 tests/mock_seed.py
"""
import datetime
import random
import sys

try:
    import pymysql
except ImportError:
    sys.exit("缺少依赖 pymysql，请先执行: pip install pymysql")

DB = dict(host="127.0.0.1", port=3306, user="root", password="12345678",
          database="pisa", charset="utf8mb4", autocommit=False)

random.seed(20260928)
_idc = 8200000000000000000


def nid():
    global _idc
    _idc += 1
    return _idc


def now():
    return datetime.datetime.now()


def phone():
    return "13%09d" % random.randint(0, 999999999)


def money():
    return round(random.randint(1000, 91000) / 100.0, 2)


def insert_many(cur, table, cols, rows):
    if not rows:
        return
    sql = "INSERT INTO `%s` (%s) VALUES (%s)" % (
        table, ",".join("`%s`" % c for c in cols), ",".join(["%s"] * len(cols)))
    cur.executemany(sql, rows)


EXTRA_TENANTS = [
    ("张三五金店", "张三", "13800000001"),
    ("李四超市", "李四", "13800000002"),
    ("王五批发行", "王五", "13800000003"),
]


def get_primary(cur):
    cur.execute("SELECT id,name FROM tenants WHERE name=%s LIMIT 1", ("演示商户",))
    row = cur.fetchone()
    if not row:
        sys.exit("未找到主商户（演示商户），请确认服务已初始化 seed")
    tenant_id = row[0]
    cur.execute("SELECT id,phone,nickname FROM users WHERE phone=%s LIMIT 1", ("admin",))
    u = cur.fetchone()
    if not u:
        sys.exit("未找到 admin 账号")
    return tenant_id, row[1], u[0], u[2]


def create_tenants(cur, admin_id, admin_name):
    """为 admin 创建多个商户（含主门店 + 归属关系）。"""
    created = []
    for name, contact, cphone in EXTRA_TENANTS:
        cur.execute("SELECT id FROM tenants WHERE name=%s LIMIT 1", (name,))
        if cur.fetchone():
            continue
        tid = nid()
        cur.execute(
            "INSERT INTO tenants (id,name,type,contact_name,contact_phone,owner_user_id,plan_id,"
            "max_goods,max_staff,max_shops,max_orders,status,created_at,updated_at) "
            "VALUES (%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s)",
            (tid, name, 1, contact, cphone, admin_id, 3, 2000, 10, 1, 0, 1, now(), now()))
        cur.execute(
            "INSERT INTO shops (id,tenant_id,name,address,phone,is_main,status,created_at,type) "
            "VALUES (%s,%s,%s,%s,%s,%s,%s,%s,%s)",
            (nid(), tid, name, "默认地址", cphone, 1, 1, now(), 1))
        cur.execute(
            "INSERT INTO user_tenants (id,user_id,tenant_id,is_owner,role,staff_name,staff_phone,status,joined_at) "
            "VALUES (%s,%s,%s,%s,%s,%s,%s,%s,%s)",
            (nid(), admin_id, tid, 1, 1, admin_name, "admin", 1, now()))
        created.append((tid, name))
    return created


def clean_business(cur, tenant_id):
    tables = ["customer_addresses", "customers", "customer_categories", "customer_prices",
              "goods_prices", "goods_stocks", "goods_units", "goods_batches", "goods",
              "goods_categories", "goods_attributes", "goods_properties", "units"]
    for t in tables:
        cur.execute("DELETE FROM `%s` WHERE tenant_id=%%s" % t, (tenant_id,))


def seed_customer_categories(cur, tid):
    roots = []
    # 顶层分类
    top = [("普通客户", ["普通客户A", "普通客户B"]),
           ("VIP客户", ["黄金VIP", "铂金VIP"]),
           ("批发客户", []),
           ("零售客户", []),
           ("企业客户", [])]
    for i, (name, kids) in enumerate(top):
        pid = nid()
        roots.append(pid)
        cur.execute("INSERT INTO customer_categories (id,tenant_id,name,parent_id,sort) VALUES (%s,%s,%s,%s,%s)",
                    (pid, tid, name, 0, i))
        for j, kn in enumerate(kids):
            cur.execute("INSERT INTO customer_categories (id,tenant_id,name,parent_id,sort) VALUES (%s,%s,%s,%s,%s)",
                        (nid(), tid, kn, pid, j))
    return roots


def seed_customers(cur, tid, cats, count=200):
    rows = []
    for i in range(count):
        rows.append((nid(), tid, "客户%04d" % (i + 1), "KH%05d" % (i + 1), 1,
                     "联系人%d" % (i + 1), phone(), "广东省深圳市南山区", 1, now(), now(),
                     random.choice(cats), "零售价", 100, random.randint(0, 500), random.randint(0, 2000)))
    insert_many(cur, "customers",
                ["id", "tenant_id", "name", "code", "type", "contact", "phone", "address", "status",
                 "created_at", "updated_at", "category_id", "price_level", "discount", "points", "init_debt"], rows)


def seed_units(cur, tid):
    names = ["个", "箱", "件", "瓶", "包", "盒", "kg", "米", "袋", "桶", "套", "打"]
    ids = []
    rows = []
    for i, nm in enumerate(names):
        uid = nid()
        ids.append(uid)
        rows.append((uid, tid, nm, now()))
    insert_many(cur, "units", ["id", "tenant_id", "name", "created_at"], rows)
    return ids


def seed_specs(cur, tid):
    """规格管理 -> goods_attributes"""
    specs = [("颜色", ["红色", "蓝色", "黑色", "白色"]),
             ("尺码", ["S", "M", "L", "XL", "XXL"]),
             ("容量", ["250ml", "500ml", "1L", "2L"]),
             ("口味", ["原味", "香辣", "甜味"]),
             ("材质", ["纯棉", "涤纶", "真皮"])]
    rows = [(nid(), tid, name, __import__("json").dumps(vals, ensure_ascii=False), i, 1)
            for i, (name, vals) in enumerate(specs)]
    insert_many(cur, "goods_attributes", ["id", "tenant_id", "name", "values", "sort", "status"], rows)


def seed_goods_categories(cur, tid):
    roots = []
    top = [("食品饮料", ["休闲零食", "冲调饮品"]),
           ("日用百货", ["清洁用品", "厨房用品"]),
           ("数码家电", ["手机配件", "电脑配件"]),
           ("服装鞋帽", ["男装", "女装"]),
           ("办公用品", [])]
    for i, (name, kids) in enumerate(top):
        pid = nid()
        roots.append(pid)
        cur.execute("INSERT INTO goods_categories (id,tenant_id,name,parent_id,sort) VALUES (%s,%s,%s,%s,%s)",
                    (pid, tid, name, 0, i))
        for j, kn in enumerate(kids):
            cur.execute("INSERT INTO goods_categories (id,tenant_id,name,parent_id,sort) VALUES (%s,%s,%s,%s,%s)",
                        (nid(), tid, kn, pid, j))
    return roots


def seed_goods(cur, tid, cats, units, count=300):
    rows = []
    for i in range(count):
        pp = money()
        rows.append((nid(), tid, "测试商品%04d" % (i + 1), "SP%05d" % (i + 1), "69%011d" % (i + 1),
                     random.choice(cats), random.choice(units),
                     random.randint(0, 500), 1000, 10, pp, round(pp * 1.3, 2), round(pp * 1.15, 2),
                     1, 1, now(), now(), "标准", "测试品牌", "个"))
    insert_many(cur, "goods",
                ["id", "tenant_id", "name", "code", "barcode", "category_id", "unit_id",
                 "current_stock", "max_stock", "min_stock", "purchase_price", "retail_price", "wholesale_price",
                 "cost_method", "status", "created_at", "updated_at", "spec", "brand", "main_unit"], rows)
    return [r[0] for r in rows]


def seed_goods_prices(cur, tid, goods):
    rows = []
    for g in goods:
        pp = money()
        rows.append((nid(), tid, g, "个", "", pp, round(pp * 1.3, 2), round(pp * 1.15, 2), 0))
    insert_many(cur, "goods_prices",
                ["id", "tenant_id", "goods_id", "unit_key", "spec_key", "purchase_price", "retail_price",
                 "wholesale_price", "sort"], rows)


def seed_goods_stocks(cur, tid, goods):
    rows = [(nid(), tid, g, "", random.randint(0, 500), 10, 20, 1000) for g in goods]
    insert_many(cur, "goods_stocks",
                ["id", "tenant_id", "goods_id", "spec_key", "stock", "min_stock", "safety_stock", "max_stock"], rows)


def main():
    conn = pymysql.connect(**DB)
    cur = conn.cursor()

    tenant_id, tenant_name, admin_id, admin_name = get_primary(cur)
    print("主商户: %s (%d)  账号: %s (%d)" % (tenant_name, tenant_id, admin_name, admin_id))

    created = create_tenants(cur, admin_id, admin_name)
    print("新增商户 %d 个: %s" % (len(created), ", ".join(n for _, n in created) if created else "（已存在）"))

    clean_business(cur, tenant_id)

    cust_cats = seed_customer_categories(cur, tenant_id)
    seed_customers(cur, tenant_id, cust_cats, 200)
    units = seed_units(cur, tenant_id)
    seed_specs(cur, tenant_id)
    goods_cats = seed_goods_categories(cur, tenant_id)
    goods = seed_goods(cur, tenant_id, goods_cats, units, 300)
    seed_goods_prices(cur, tenant_id, goods)
    seed_goods_stocks(cur, tenant_id, goods)

    conn.commit()

    print("\n=== 造数完成 (tenant %d) ===" % tenant_id)
    for t in ["customer_categories", "customers", "units", "goods_attributes",
              "goods_categories", "goods", "goods_prices", "goods_stocks"]:
        cur.execute("SELECT COUNT(*) FROM `%s` WHERE tenant_id=%%s" % t, (tenant_id,))
        print("  %-22s %d" % (t, cur.fetchone()[0]))

    cur.close()
    conn.close()


if __name__ == "__main__":
    main()
