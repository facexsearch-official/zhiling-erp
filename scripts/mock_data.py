#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
生成/重建「演示商户」的全部测试数据（Mock）。
顺序：门店(20) → 角色(20) → 用户(200) → 客户价格等级(100) → 客户(1w)
      → 供应商(200) → 商品规格(50) → 商品单位(50) → 商品分类(200)
      → 商品(2w) → 进货单(1w) → 进货退货单(5k) → 销售单(2w) → 销售退货单(2w)
      → 盘点单(5k) → 账户(20) → 收款单(5k) / 付款单(5k)

用法:
    python3 scripts/mock_data.py
    可选环境变量: PISA_DSN / PISA_DB_HOST / PISA_DB_PORT / PISA_DB_USER / PISA_DB_PASS / PISA_DB_NAME
"""
import json
import os
import random
import re
import sys
import time
from datetime import datetime, timedelta

import pymysql

SEED = 20260101
rng = random.Random(SEED)

ID_BASE = 2200000000000000000
_id_seq = 0


def new_id():
    global _id_seq
    _id_seq += 1
    return ID_BASE + _id_seq


def ri(a, b):
    return rng.randint(a, b)


def rf(a, b):
    return round(a + rng.random() * (b - a), 2)


def pick(a):
    return a[rng.randrange(len(a))]


def rand_dt(days=720):
    return datetime.now() - timedelta(days=ri(0, days), seconds=ri(0, 86400))


def rand_date(days=365):
    return (datetime.now() - timedelta(days=ri(0, days))).strftime("%Y-%m-%d")


def order_no(prefix, i):
    return "%s%s%05d" % (prefix, datetime.now().strftime("%Y%m%d"), i)


# ─────────────── 数据库连接 ───────────────
def read_dsn():
    if os.environ.get("PISA_DSN"):
        return os.environ["PISA_DSN"]
    path = os.path.join(os.path.dirname(os.path.dirname(os.path.abspath(__file__))), "config.yaml")
    try:
        txt = open(path, encoding="utf-8").read()
        m = re.search(r'dsn:\s*"([^"]+)"', txt)
        if m:
            return m.group(1)
    except OSError:
        pass
    # 未找到 config.yaml 时使用无密码的本地默认（可通过 PISA_DSN 覆盖）
    return "root:@tcp(127.0.0.1:3306)/pisa?charset=utf8mb4&parseTime=True&loc=Local"


def parse_dsn(dsn):
    m = re.match(r"([^:]+):([^@]*)@tcp\(([^:]+):(\d+)\)/([^?]+)", dsn)
    if not m:
        raise SystemExit("无法解析 DSN: " + dsn)
    return {
        "user": m.group(1),
        "password": m.group(2),
        "host": m.group(3),
        "port": int(m.group(4)),
        "database": m.group(5),
    }


def connect():
    cfg = parse_dsn(read_dsn())
    cfg["user"] = os.environ.get("PISA_DB_USER", cfg["user"])
    cfg["password"] = os.environ.get("PISA_DB_PASS", cfg["password"])
    cfg["host"] = os.environ.get("PISA_DB_HOST", cfg["host"])
    cfg["port"] = int(os.environ.get("PISA_DB_PORT", cfg["port"]))
    cfg["database"] = os.environ.get("PISA_DB_NAME", cfg["database"])
    return pymysql.connect(charset="utf8mb4", autocommit=False, **cfg)


# ─────────────── 批量写入 ───────────────
def insert(cur, table, columns, rows, chunk=1000):
    if not rows:
        return
    collist = ",".join("`%s`" % c for c in columns)
    ph = ",".join(["%s"] * len(columns))
    sql = "INSERT INTO `%s` (%s) VALUES (%s)" % (table, collist, ph)
    for i in range(0, len(rows), chunk):
        cur.executemany(sql, rows[i:i + chunk])
    cur.connection.commit()


# ─────────────── 清空 ───────────────
WIPE_TABLES = [
    "purchase_order_items", "purchase_orders", "purchase_items", "purchases",
    "purchase_return_items", "purchase_returns",
    "sale_order_items", "sale_orders", "sale_items", "sales",
    "sales_return_items", "sales_returns", "quote_items", "quotes",
    "stock_count_items", "stock_counts", "assembly_items", "assemblies",
    "recipe_items", "recipes", "goods_batches", "transfers",
    "receipts", "payments", "income_items", "incomes", "income_types",
    "combo_items", "combos", "goods_prices", "goods_stocks", "goods_units", "goods",
    "goods_categories", "goods_attributes", "goods_properties", "units", "warehouses",
    "customer_prices", "customer_addresses", "customers", "customer_categories", "price_levels",
    "supplier_addresses", "suppliers", "supplier_categories",
    "accounts", "stock_logs", "stock_balances",
    "salesmen", "roles", "shops",
]


def wipe(cur, tenant_id):
    for t in WIPE_TABLES:
        cur.execute("DELETE FROM `%s` WHERE tenant_id=%%s" % t, (tenant_id,))
    cur.execute("SELECT id FROM users WHERE phone='admin' LIMIT 1")
    row = cur.fetchone()
    admin_id = row[0] if row else 0
    cur.execute(
        "DELETE FROM users WHERE id<>%s AND id IN "
        "(SELECT user_id FROM (SELECT user_id FROM user_tenants WHERE tenant_id=%s) x)",
        (admin_id, tenant_id))
    cur.execute("DELETE FROM user_tenants WHERE tenant_id=%s AND user_id<>%s", (tenant_id, admin_id))
    cur.connection.commit()


# ─────────────── 门店 ───────────────
def gen_shops(cur, tid):
    cities = ["北京", "上海", "广州", "深圳", "杭州", "成都", "武汉", "西安", "南京", "苏州"]
    cols = ["id", "tenant_id", "name", "address", "phone", "logo", "address_detail",
            "type", "remark", "is_main", "status", "created_at"]
    rows, ids = [], []
    for i in range(1, 21):
        c = cities[(i - 1) % len(cities)]
        sid = new_id()
        name = "总店" if i == 1 else "%s门店%02d" % (c, i)
        rows.append((sid, tid, name, "%s市%s区示范路%d号" % (c, c, i * 7),
                     "1%d%09d" % (3 + i % 6, 100000000 + i * 137), "",
                     "%d楼%d室" % (i % 20 + 1, i), 1 + i % 2, "第 %d 家门店" % i,
                     1 if i == 1 else 0, 1, rand_dt()))
        ids.append(sid)
    insert(cur, "shops", cols, rows)
    return ids


# ─────────────── 角色 ───────────────
def gen_roles(cur, tid):
    names = ["总经理", "副总经理", "店长", "收银员", "仓管员", "采购员", "销售员", "财务主管",
             "会计", "出纳", "客服", "理货员", "配送员", "运营专员", "市场专员", "督导",
             "审核员", "数据员", "店助", "临时工"]
    descs = ["拥有全部权限", "协助管理日常事务", "负责门店整体运营", "负责收银开单", "负责仓库进出库",
             "负责采购进货", "负责销售跟单", "负责财务核算", "负责账务处理", "负责现金收付",
             "负责客户接待", "负责货品整理", "负责配送", "负责线上运营", "负责市场推广",
             "负责门店巡查", "负责单据审核", "负责数据统计", "协助店长", "临时用工"]
    full = json.dumps({"*": True}, ensure_ascii=False)
    view = json.dumps({"goods.goods.view": True, "sale.sale.view": True, "purchase.purchase.view": True,
                       "stock.query.view": True, "funds.receipt.view": True, "funds.payment.view": True,
                       "customer.customer.view": True, "purchase.supplier.view": True}, ensure_ascii=False)
    cols = ["id", "tenant_id", "name", "description", "permissions", "sensitive_data",
            "is_system", "status", "created_at", "updated_at"]
    rows, ids = [], []
    now = datetime.now()
    for i in range(20):
        rid = new_id()
        rows.append((rid, tid, names[i], descs[i], full if i < 3 else view, "{}", 0, 1, rand_dt(), now))
        ids.append(rid)
    insert(cur, "roles", cols, rows)
    return ids


# ─────────────── 业务员 ───────────────
def gen_salesmen(cur, tid, shop_ids):
    cols = ["id", "tenant_id", "name", "phone", "shop_id", "status", "created_at", "updated_at"]
    rows, ids = [], []
    now = datetime.now()
    for i in range(1, 21):
        sid = new_id()
        rows.append((sid, tid, "业务员%02d" % i, "151%08d" % (10000000 + i * 311),
                     pick(shop_ids), 1, rand_dt(), now))
        ids.append(sid)
    insert(cur, "salesmen", cols, rows)
    return ids


# ─────────────── 用户 ───────────────
def gen_users(cur, tid, role_ids):
    cur.execute("SELECT password_hash FROM users WHERE phone='admin' LIMIT 1")
    row = cur.fetchone()
    pwd = row[0] if row else "$2a$10$8SaS7Kexgriw44JvRxawXOvajjjiU/MJnR.tBFSS8WaSkhJZOab.u"
    ucols = ["id", "phone", "password_hash", "nickname", "avatar", "default_tenant_id",
             "status", "last_login_at", "created_at", "updated_at"]
    lcols = ["id", "user_id", "tenant_id", "is_owner", "role", "role_id", "staff_name",
             "staff_phone", "permissions", "invited_by", "status", "joined_at"]
    urows, lrows, uids = [], [], []
    now = datetime.now()
    for i in range(1, 201):
        uid = new_id()
        phone = ("13%d" % (100000000 + i * 1237))[:11]
        urows.append((uid, phone, pwd, "员工%03d" % i, "", tid, 1, rand_dt(60), rand_dt(), now))
        uids.append(uid)
    insert(cur, "users", ucols, urows)
    for i, uid in enumerate(uids):
        lrows.append((new_id(), uid, tid, 0, 1 + i % 3, role_ids[i % len(role_ids)],
                      "员工%03d" % (i + 1), urows[i][1], "{}", None, 1, rand_dt()))
    insert(cur, "user_tenants", lcols, lrows)
    return uids


# ─────────────── 价格等级 ───────────────
def gen_levels(cur, tid):
    cols = ["id", "tenant_id", "name", "sort", "status", "created_at"]
    rows, names = [], []
    for i in range(1, 101):
        nid = new_id()
        nm = "价格等级%03d" % i
        rows.append((nid, tid, nm, i, 1, rand_dt()))
        names.append(nm)
    insert(cur, "price_levels", cols, rows)
    return names


# ─────────────── 客户分类 / 客户 ───────────────
def gen_customer_categories(cur, tid):
    names = ["普通客户", "VIP客户", "企业客户", "代理商", "经销商", "批发客户", "零售客户",
             "长期合作", "新客户", "潜在客户", "重点客户", "战略客户", "个人客户", "政府客户",
             "电商客户", "线下客户", "连锁客户", "加盟商", "合作单位", "其他"]
    cols = ["id", "tenant_id", "name", "parent_id", "sort"]
    rows, ids = [], []
    for i, n in enumerate(names):
        cid = new_id()
        rows.append((cid, tid, n, 0, i + 1))
        ids.append(cid)
    insert(cur, "customer_categories", cols, rows)
    return ids


def gen_customers(cur, tid, level_names, cat_ids, salesman_ids):
    total = 10000
    cols = ["id", "tenant_id", "name", "code", "type", "category_id", "price_level", "salesman_id",
            "discount", "init_debt", "contact", "phone", "address", "address_detail", "email",
            "tax_no", "fax", "bank_name", "bank_account", "birthday", "wechat", "qq", "remark",
            "balance", "total_receivable", "total_payable", "status", "created_at", "updated_at"]
    addrs = ["广东省深圳市南山区", "北京市朝阳区", "上海市浦东新区", "浙江省杭州市西湖区", "江苏省南京市玄武区"]
    banks = ["中国工商银行", "招商银行", "中国建设银行", "中国农业银行", "中国银行"]
    rows = []
    now = datetime.now()
    for i in range(1, total + 1):
        cid = new_id()
        rows.append((cid, tid, "客户%05d" % i, "KH%05d" % i, 1 + i % 3,
                     cat_ids[i % len(cat_ids)], level_names[i % len(level_names)],
                     salesman_ids[i % len(salesman_ids)], rf(80, 100), rf(0, 5000),
                     "联系人%d" % (i % 50 + 1), "1%d%09d" % (3 + i % 6, 200000000 + i * 97),
                     pick(addrs), "科技园%d栋%d单元" % (i % 30 + 1, i % 20 + 1),
                     "customer%05d@example.com" % i, "91%016d" % (1000000000 + i),
                     "0755-%08d" % (8000000 + i), pick(banks), "62%015d" % (600000000000 + i),
                     "19%02d-%02d-%02d" % (70 + i % 30, i % 12 + 1, i % 28 + 1),
                     "wx_%05d" % i, str(100000 + i), "客户备注 %d" % i,
                     rf(0, 20000), rf(0, 100000), rf(0, 50000), 1, rand_dt(), now))
    insert(cur, "customers", cols, rows, chunk=2000)

    # 收货地址（前 1000 客户）
    acols = ["id", "tenant_id", "customer_id", "receiver", "phone", "region", "detail", "is_default"]
    arows = []
    for i in range(1000):
        for j in range(1 + i % 2):
            arows.append((new_id(), tid, rows[i][0], rows[i][10], rows[i][11],
                          rows[i][12], rows[i][13], 1 if j == 0 else 0))
    insert(cur, "customer_addresses", acols, arows)
    return [r[0] for r in rows]


# ─────────────── 供应商分类 / 供应商 ───────────────
def gen_supplier_categories(cur, tid):
    names = ["常规供应商", "一级代理", "生产厂家", "贸易商", "批发商", "进口商", "本地供应商",
             "外省供应商", "长期合作", "临时供应商", "战略供应商", "备选供应商", "食品类",
             "日化类", "数码类", "服装类", "办公类", "五金类", "包装类", "其他"]
    cols = ["id", "tenant_id", "name", "parent_id", "sort"]
    rows, ids = [], []
    for i, n in enumerate(names):
        cid = new_id()
        rows.append((cid, tid, n, 0, i + 1))
        ids.append(cid)
    insert(cur, "supplier_categories", cols, rows)
    return ids


def gen_suppliers(cur, tid, cat_ids):
    total = 200
    cols = ["id", "tenant_id", "name", "code", "category_id", "contact", "phone", "address",
            "bank_name", "bank_account", "remark", "init_payable", "total_payable", "status",
            "created_at", "updated_at", "email", "fax", "wechat", "qq", "birthday",
            "province", "city", "district", "address_detail"]
    addrs = ["浙江省义乌市", "广东省东莞市", "江苏省常熟市", "福建省晋江市"]
    banks = ["中国银行", "交通银行", "平安银行", "兴业银行"]
    rows = []
    now = datetime.now()
    for i in range(1, total + 1):
        sid = new_id()
        rows.append((sid, tid, "供应商%04d" % i, "GYS%04d" % i, cat_ids[i % len(cat_ids)],
                     "供联%d" % (i % 30 + 1), "1%d%09d" % (3 + i % 6, 300000000 + i * 131),
                     pick(addrs), pick(banks), "62%015d" % (600000000000 + i),
                     "供应商备注 %d" % i, rf(0, 10000), rf(0, 200000), 1, rand_dt(), now,
                     "supplier%04d@example.com" % i, "0571-%08d" % (8000000 + i),
                     "swx_%04d" % i, str(200000 + i),
                     "19%02d-%02d-%02d" % (70 + i % 30, i % 12 + 1, i % 28 + 1),
                     "浙江省", "杭州市", "西湖区", "产业园%d号" % i))
    insert(cur, "suppliers", cols, rows)
    acols = ["id", "tenant_id", "supplier_id", "consignee", "consignee_phone", "province",
             "city", "district", "address_detail", "is_default", "sort"]
    arows = []
    for i, r in enumerate(rows):
        arows.append((new_id(), tid, r[0], r[5], r[6], r[21], r[22], r[23], r[24], 1, 1))
    insert(cur, "supplier_addresses", acols, arows)
    return [r[0] for r in rows]


# ─────────────── 单位 / 规格 ───────────────
def gen_units(cur, tid):
    names = ["个", "件", "箱", "包", "袋", "瓶", "盒", "套", "对", "双", "台", "只", "支", "条",
             "张", "片", "块", "卷", "桶", "罐", "斤", "公斤", "克", "吨", "米", "厘米", "升",
             "毫升", "打", "组", "打箱", "大包", "小包", "托盘", "捆", "扎", "串", "束", "车",
             "排", "提", "筐", "篓", "篮", "袋装", "瓶装", "罐装", "盒装", "散装", "整箱"]
    cols = ["id", "tenant_id", "name", "created_at"]
    rows, names_out = [], []
    for n in names:
        uid = new_id()
        rows.append((uid, tid, n, rand_dt()))
        names_out.append(n)
    insert(cur, "units", cols, rows)
    return names_out


def gen_attributes(cur, tid):
    vmap = {
        "颜色": ["红色", "蓝色", "绿色", "黑色", "白色", "黄色"],
        "尺码": ["S", "M", "L", "XL", "XXL"],
        "材质": ["棉", "涤纶", "混纺", "真皮", "帆布"],
        "口味": ["原味", "香辣", "五香", "麻辣", "香甜"],
        "规格": ["小", "中", "大", "特大"],
        "型号": ["A型", "B型", "C型", "Pro"],
        "容量": ["500ml", "1L", "1.5L", "2L"],
        "重量": ["100g", "500g", "1kg", "2kg", "5kg"],
        "功率": ["10W", "20W", "50W", "100W"],
        "版本": ["标准版", "豪华版", "旗舰版"],
        "款式": ["简约", "复古", "运动", "商务"],
        "图案": ["纯色", "条纹", "格子", "印花"],
        "香型": ["柠檬", "薰衣草", "玫瑰", "薄荷"],
        "度数": ["38度", "42度", "52度"],
        "产地": ["国产", "进口"],
        "包装": ["袋装", "盒装", "瓶装", "罐装"],
        "净含量": ["100g", "250g", "500g", "1kg"],
        "适用人群": ["男", "女", "儿童", "通用"],
        "季节": ["春季", "夏季", "秋季", "冬季"],
        "风格": ["现代", "北欧", "中式", "工业"],
    }
    base = list(vmap.keys())
    cols = ["id", "tenant_id", "name", "values", "sort", "status"]
    rows = []
    for i in range(50):
        key = base[i % len(base)]
        nm = key if i < len(base) else "%s%d" % (key, i // len(base) + 1)
        rows.append((new_id(), tid, nm, json.dumps(vmap[key], ensure_ascii=False), i + 1, 1))
    insert(cur, "goods_attributes", cols, rows)
    return len(rows)


# ─────────────── 商品分类（深层级，200） ───────────────
def gen_goods_categories(cur, tid):
    roots = ["食品饮料", "日用百货", "数码家电", "服装鞋帽", "办公用品", "母婴玩具", "美妆护肤", "家居家纺"]
    subs = ["一级", "二级", "三级", "四级", "系列", "分类", "专区", "专区"]
    cols = ["id", "tenant_id", "name", "parent_id", "sort"]
    rows, nodes = [], []
    for i, r in enumerate(roots):
        cid = new_id()
        rows.append((cid, tid, r, 0, i + 1))
        nodes.append([cid, 1])
    seq = 0
    while len(rows) < 200:
        parent = nodes[rng.randrange(len(nodes))]
        if parent[1] >= 8:
            continue
        seq += 1
        cid = new_id()
        rows.append((cid, tid, "%s%d%03d" % (pick(subs), parent[1] + 1, seq), parent[0], seq % 20 + 1))
        nodes.append([cid, parent[1] + 1])
    insert(cur, "goods_categories", cols, rows)
    return [r[0] for r in rows]


# ─────────────── 商品 ───────────────
def spec_values(vals):
    return [{"name": v, "image": ""} for v in vals]


def gen_goods(cur, tid, unit_names, cat_ids, supplier_ids, shop_ids):
    total = 20000
    adjs = ["高清", "加厚", "便携", "多功能", "家用", "商用", "大容量", "迷你", "智能", "环保",
            "轻奢", "经典", "升级款", "豪华", "简约"]
    nouns = ["水杯", "收纳盒", "数据线", "充电器", "耳机", "笔记本", "T恤", "运动鞋", "台灯", "雨伞",
             "背包", "水壶", "毛巾", "拖鞋", "拖鞋", "剃须刀", "吹风机", "电饭煲", "炒锅", "刀具",
             "洗衣液", "抽纸", "牙膏", "洗发水", "沐浴露", "面膜", "口红", "香水", "零食", "饼干",
             "坚果", "牛奶", "咖啡", "茶叶", "矿泉水", "马克笔", "文件夹", "打印纸", "订书机", "计算器"]
    brands = ["小米", "华为", "美的", "海尔", "格力", "苏泊尔", "九阳", "得力", "晨光", "南极人",
              "无印", "优衣库", "宝洁", "联合利华", "三只松鼠"]
    origins = ["广东深圳", "浙江义乌", "江苏苏州", "福建泉州", "山东青岛", "上海", "北京",
               "浙江杭州", "广东东莞", "四川成都"]
    stock_shops = shop_ids[:3]

    gcols = ["id", "tenant_id", "name", "code", "barcode", "category_id", "unit_id", "supplier_id",
             "suppliers", "image_url", "current_stock", "max_stock", "min_stock", "purchase_price",
             "retail_price", "wholesale_price", "cost_method", "status", "created_at", "updated_at",
             "remark", "spec", "brand", "origin", "has_multi_unit", "has_multi_spec", "sales_unit",
             "purchase_unit", "enable_stock_alert", "safety_stock", "has_batch", "has_shelf_life",
             "shelf_life_days", "expiry_alert", "expiry_warn_days", "has_serial", "init_cost",
             "images", "spec_groups", "price_columns", "main_unit", "retail_min", "retail_max", "total_stock"]
    grows, urows, prows, srows = [], [], [], []
    gids = []
    now = datetime.now()
    for i in range(1, total + 1):
        gid = new_id()
        gids.append(gid)
        unit = pick(unit_names)
        cat = cat_ids[rng.randrange(len(cat_ids))]
        sup = supplier_ids[rng.randrange(len(supplier_ids))]
        pp = rf(5, 500)
        multi_unit = 1 if i % 5 == 0 else 0
        multi_spec = 1 if i % 4 == 0 else 0
        spec_vals = []
        spec_groups = "[]"
        if multi_spec:
            spec_vals = ["S", "M", "L"] if i % 8 else ["红", "蓝", "绿"]
            spec_groups = json.dumps([{"name": "规格", "has_image": False, "values": spec_values(spec_vals)}],
                                     ensure_ascii=False)
        spec_keys = spec_vals if multi_spec else [""]
        spec_str = pick(spec_vals) if spec_vals else ""
        rprice = rf(pp * 1.2, pp * 2)
        wprice = rf(pp * 1.05, pp * 1.3)
        total_stock = 0
        retail_min, retail_max = rprice, rprice
        for sk in spec_keys:
            rp = rprice if sk == "" else rf(pp * 1.2, pp * 2)
            retail_min = min(retail_min, rp)
            retail_max = max(retail_max, rp)
            prows.append((new_id(), tid, gid, unit, sk, "", "SP%06d" % i, pp, rp, wprice, 0,
                          '{"等级1":0,"等级2":0}', 1))
        urows.append((new_id(), tid, gid, None, unit, 1, 1, 1))
        if multi_unit:
            u2 = pick(unit_names)
            while u2 == unit:
                u2 = pick(unit_names)
            urows.append((new_id(), tid, gid, None, u2, ri(2, 24), 0, 2))
            prows.append((new_id(), tid, gid, u2, "", "", "SP%06d-1" % i, pp, rf(pp * 1.2, pp * 2),
                          wprice, 0, "{}", 2))
        for sh in stock_shops:
            for sk in spec_keys:
                st = ri(0, 200)
                total_stock += st
                srows.append((new_id(), tid, gid, sh, sk, st, ri(5, 50), ri(5, 60), ri(100, 1000), pp))
        grows.append((gid, tid, "%s%s%d" % (pick(adjs), pick(nouns), i), "SP%06d" % i,
                      "69%011d" % (10000000 + i), cat, None, sup, json.dumps([sup]),
                      "", total_stock, ri(100, 1000), ri(5, 50), pp, rprice, wprice,
                      1 + i % 2, 1, rand_dt(), now, "商品备注%d" % i, spec_str, pick(brands),
                      pick(origins), multi_unit, multi_spec, unit, unit, i % 3 % 2, ri(5, 60),
                      i % 10 % 2, i % 7 % 2, ri(30, 730), i % 6 % 2, ri(7, 60), i % 11 % 2,
                      pp, "[]", spec_groups, '["等级1","等级2"]', unit, retail_min, retail_max, total_stock))
    # unit_id 需为 FlexInt64 字符串/数字；此处填 None 亦可，保持 main_unit 生效
    insert(cur, "goods", gcols, grows, chunk=1000)
    insert(cur, "goods_units",
           ["id", "tenant_id", "goods_id", "unit_id", "unit_name", "factor", "is_main", "sort"], urows, 2000)
    insert(cur, "goods_prices",
           ["id", "tenant_id", "goods_id", "unit_key", "spec_key", "barcode", "code",
            "purchase_price", "retail_price", "wholesale_price", "disabled", "custom", "sort"], prows, 2000)
    insert(cur, "goods_stocks",
           ["id", "tenant_id", "goods_id", "shop_id", "spec_key", "stock", "min_stock",
            "safety_stock", "max_stock", "init_cost"], srows, 2000)
    return gids


# ─────────────── 账户 ───────────────
def gen_accounts(cur, tid, shop_ids):
    names = ["现金", "微信", "支付宝", "中国银行", "工商银行", "建设银行", "农业银行", "招商银行",
             "交通银行", "平安银行", "兴业银行", "浦发银行", "中信银行", "光大银行", "民生银行",
             "广发银行", "华夏银行", "邮储银行", "备用金", "其他"]
    cols = ["id", "tenant_id", "name", "type", "bank_name", "card_no", "shop_id", "balance",
            "remark", "sort", "status", "created_at"]
    rows, ids = [], []
    for i, n in enumerate(names):
        t = 2
        if i == 0:
            t = 1
        elif i <= 2:
            t = 3
        aid = new_id()
        rows.append((aid, tid, n, t, n, "62%015d" % (600000000000 + i), shop_ids[i % len(shop_ids)],
                     rf(1000, 500000), "账户-%s" % n, i + 1, 1, rand_dt()))
        ids.append(aid)
    insert(cur, "accounts", cols, rows)
    return ids


# ─────────────── 明细工具 ───────────────
def gen_items(goods_ids, prices, n):
    items, total = [], 0.0
    for _ in range(n):
        idx = rng.randrange(len(goods_ids))
        qty = ri(1, 20)
        price = prices[idx]
        total += qty * price
        items.append((goods_ids[idx], qty, price))
    return items, total


# ─────────────── 进货单 ───────────────
def gen_purchases(cur, tid, supplier_ids, goods_ids, prices, salesman_ids, account_ids, shop_ids, total=10000):
    cols = ["id", "tenant_id", "shop_id", "warehouse_id", "order_no", "supplier_id", "salesman_id",
            "account_id", "bill_date", "subtotal", "discount", "discounted_amount", "freight",
            "deposit_offset", "total_amount", "paid_amount", "unpaid_amount", "invoice_status",
            "print_status", "related_no", "attachments", "status", "remark", "created_by",
            "created_at", "updated_at"]
    ic = ["id", "tenant_id", "purchase_id", "goods_id", "spec_key", "quantity", "unit_price",
          "amount", "remark", "created_at"]
    rows, items = [], []
    now = datetime.now()
    for i in range(1, total + 1):
        pid = new_id()
        bi, s = gen_items(goods_ids, prices, ri(1, 5))
        disc = float(ri(90, 100))
        disc_amt = round(s * disc / 100, 2)
        freight = float(ri(0, 100))
        total_amt = round(disc_amt + freight, 2)
        paid = round(total_amt * ri(0, 100) / 100, 2)
        rows.append((pid, tid, pick(shop_ids), 0, order_no("JH", i), pick(supplier_ids),
                     pick(salesman_ids), pick(account_ids), rand_date(), round(s, 2), disc, disc_amt,
                     freight, 0, total_amt, paid, round(total_amt - paid, 2), i % 2, i % 2, "",
                     "[]", ri(1, 4), "进货单备注 %d" % i, 0, rand_dt(365), now))
        for gid, qty, price in bi:
            items.append((new_id(), tid, pid, gid, "", qty, price, round(qty * price, 2), "", now))
    insert(cur, "purchases", cols, rows, 1000)
    insert(cur, "purchase_items", ic, items, 2000)


def gen_purchase_returns(cur, tid, supplier_ids, goods_ids, prices, salesman_ids, account_ids, shop_ids, total=5000):
    cols = ["id", "tenant_id", "shop_id", "warehouse_id", "order_no", "supplier_id", "salesman_id",
            "account_id", "bill_date", "deposit_offset", "paid_amount", "unpaid_amount",
            "invoice_status", "print_status", "related_no", "attachments", "total_amount",
            "refund_amount", "status", "remark", "created_by", "created_at"]
    ic = ["id", "tenant_id", "purchase_return_id", "goods_id", "spec_key", "quantity",
          "unit_price", "amount", "remark", "created_at"]
    rows, items = [], []
    now = datetime.now()
    for i in range(1, total + 1):
        rid = new_id()
        bi, s = gen_items(goods_ids, prices, ri(1, 3))
        amt = round(s, 2)
        rows.append((rid, tid, pick(shop_ids), 0, order_no("JHT", i), pick(supplier_ids),
                     pick(salesman_ids), pick(account_ids), rand_date(), 0, amt, 0, i % 2, i % 2, "",
                     "[]", -amt, amt, ri(1, 4), "进货退货备注 %d" % i, 0, rand_dt(365)))
        for gid, qty, price in bi:
            items.append((new_id(), tid, rid, gid, "", -qty, price, round(-qty * price, 2), "", now))
    insert(cur, "purchase_returns", cols, rows, 1000)
    insert(cur, "purchase_return_items", ic, items, 2000)


# ─────────────── 销售单 ───────────────
def gen_sales(cur, tid, customer_ids, goods_ids, prices, salesman_ids, account_ids, shop_ids, total=20000):
    cols = ["id", "tenant_id", "shop_id", "warehouse_id", "order_no", "related_order_no",
            "customer_id", "salesman_id", "account_id", "bill_date", "discount", "subtotal",
            "round_off", "total_amount", "received_amount", "unreceived_amount", "receive_status",
            "invoice_status", "print_status", "attachments", "status", "remark", "created_by",
            "created_at", "updated_at"]
    ic = ["id", "tenant_id", "sale_id", "goods_id", "spec_key", "quantity", "unit_price",
          "amount", "remark", "created_at"]
    rows, items = [], []
    now = datetime.now()
    for i in range(1, total + 1):
        sid = new_id()
        bi, s = gen_items(goods_ids, prices, ri(1, 4))
        disc = float(ri(90, 100))
        subtotal = round(s, 2)
        total_amt = round(subtotal * disc / 100, 2)
        round_off = round(total_amt - int(total_amt), 2)
        total_amt = float(int(total_amt))
        received = round(total_amt * ri(0, 100) / 100, 2)
        rs = 2 if received >= total_amt else (1 if received > 0 else 0)
        rows.append((sid, tid, pick(shop_ids), 0, order_no("XS", i), "", pick(customer_ids),
                     pick(salesman_ids), pick(account_ids), rand_date(), disc, subtotal, round_off,
                     total_amt, received, round(total_amt - received, 2), rs, i % 2, i % 2, "[]",
                     1 + i % 3 % 2, "销售单备注 %d" % i, 0, rand_dt(365), now))
        for gid, qty, price in bi:
            items.append((new_id(), tid, sid, gid, "", qty, price, round(qty * price, 2), "", now))
    insert(cur, "sales", cols, rows, 1000)
    insert(cur, "sale_items", ic, items, 2000)


def gen_sales_returns(cur, tid, customer_ids, goods_ids, prices, salesman_ids, account_ids, shop_ids, total=20000):
    cols = ["id", "tenant_id", "shop_id", "warehouse_id", "order_no", "customer_id", "salesman_id",
            "account_id", "bill_date", "round_off", "total_amount", "received_amount",
            "unreceived_amount", "print_status", "attachments", "status", "remark", "created_by",
            "created_at"]
    ic = ["id", "tenant_id", "return_id", "goods_id", "spec_key", "quantity", "unit_price",
          "amount", "remark", "created_at"]
    rows, items = [], []
    now = datetime.now()
    for i in range(1, total + 1):
        rid = new_id()
        bi, s = gen_items(goods_ids, prices, ri(1, 3))
        total_amt = round(s, 2)
        received = round(total_amt * ri(0, 100) / 100, 2)
        rows.append((rid, tid, pick(shop_ids), 0, order_no("XST", i), pick(customer_ids),
                     pick(salesman_ids), pick(account_ids), rand_date(), 0, total_amt, received,
                     round(total_amt - received, 2), i % 2, "[]", 1 + i % 3 % 2,
                     "销售退货备注 %d" % i, 0, rand_dt(365)))
        for gid, qty, price in bi:
            items.append((new_id(), tid, rid, gid, "", qty, price, round(qty * price, 2), "", now))
    insert(cur, "sales_returns", cols, rows, 1000)
    insert(cur, "sales_return_items", ic, items, 2000)


# ─────────────── 盘点单 ───────────────
def gen_stock_counts(cur, tid, goods_ids, salesman_ids, shop_ids, total=5000):
    cols = ["id", "tenant_id", "shop_id", "warehouse_id", "order_no", "salesman_id", "bill_date",
            "book_qty", "actual_qty", "diff_qty", "status", "attachments", "remark", "created_by",
            "created_at"]
    ic = ["id", "tenant_id", "count_id", "goods_id", "book_qty", "actual_qty", "diff_qty",
          "remark", "created_at"]
    rows, items = [], []
    now = datetime.now()
    for i in range(1, total + 1):
        cid = new_id()
        book, actual = 0, 0
        for _ in range(ri(1, 5)):
            gid = pick(goods_ids)
            bq = ri(0, 200)
            aq = max(0, bq + ri(-10, 10))
            book += bq
            actual += aq
            items.append((new_id(), tid, cid, gid, bq, aq, aq - bq, "", now))
        rows.append((cid, tid, pick(shop_ids), 0, order_no("PD", i), pick(salesman_ids),
                     rand_date(), book, actual, actual - book, 1 + i % 5 % 2, "[]",
                     "盘点备注 %d" % i, 0, rand_dt(365)))
    insert(cur, "stock_counts", cols, rows, 1000)
    insert(cur, "stock_count_items", ic, items, 2000)


# ─────────────── 收款 / 付款 ───────────────
def gen_receipts(cur, tid, customer_ids, salesman_ids, account_ids, shop_ids, total=5000):
    cols = ["id", "tenant_id", "shop_id", "order_no", "related_no", "type", "customer_id",
            "salesman_id", "bill_date", "amount", "discount_amount", "deposit_offset", "account_id",
            "attachments", "status", "remark", "created_by", "created_at"]
    rows = []
    types = ["直接收款", "预收款", "销售收款"]
    for i in range(1, total + 1):
        rows.append((new_id(), tid, pick(shop_ids), order_no("SK", i), "", pick(types),
                     pick(customer_ids), pick(salesman_ids), rand_date(), rf(100, 50000),
                     rf(0, 200), 0, pick(account_ids), "[]", 1 + i % 7 % 2,
                     "收款备注 %d" % i, 0, rand_dt(365)))
    insert(cur, "receipts", cols, rows)


def gen_payments(cur, tid, supplier_ids, salesman_ids, account_ids, shop_ids, total=5000):
    cols = ["id", "tenant_id", "shop_id", "order_no", "related_no", "type", "supplier_id",
            "salesman_id", "bill_date", "amount", "discount_amount", "account_id", "attachments",
            "status", "remark", "created_by", "created_at"]
    rows = []
    types = ["直接付款", "预付款", "采购付款"]
    for i in range(1, total + 1):
        rows.append((new_id(), tid, pick(shop_ids), order_no("FK", i), "", pick(types),
                     pick(supplier_ids), pick(salesman_ids), rand_date(), rf(100, 50000),
                     rf(0, 200), pick(account_ids), "[]", 1 + i % 7 % 2,
                     "付款备注 %d" % i, 0, rand_dt(365)))
    insert(cur, "payments", cols, rows)


# ─────────────── 主流程 ───────────────
def main():
    t0 = time.time()
    conn = connect()
    cur = conn.cursor()
    cur.execute("SELECT id FROM tenants WHERE name=%s LIMIT 1", ("演示商户",))
    row = cur.fetchone()
    if not row:
        raise SystemExit("未找到「演示商户」，请先启动一次服务初始化基础数据")
    tid = row[0]
    print("演示商户 tenant_id =", tid)

    print("清空旧测试数据 ...")
    wipe(cur, tid)

    print("门店(20) ...")
    shop_ids = gen_shops(cur, tid)
    print("角色(20) ...")
    role_ids = gen_roles(cur, tid)
    print("业务员(20) ...")
    salesman_ids = gen_salesmen(cur, tid, shop_ids)
    print("用户(200) ...")
    gen_users(cur, tid, role_ids)
    print("客户价格等级(100) ...")
    level_names = gen_levels(cur, tid)
    print("客户分类 / 客户(10000) ...")
    cat_ids = gen_customer_categories(cur, tid)
    customer_ids = gen_customers(cur, tid, level_names, cat_ids, salesman_ids)
    print("供应商(200) ...")
    sup_cat_ids = gen_supplier_categories(cur, tid)
    supplier_ids = gen_suppliers(cur, tid, sup_cat_ids)
    print("商品单位(50) / 规格(50) ...")
    unit_names = gen_units(cur, tid)
    gen_attributes(cur, tid)
    print("商品分类(200) ...")
    goods_cat_ids = gen_goods_categories(cur, tid)
    print("商品(20000) + 明细/价格/库存 ...")
    goods_ids = gen_goods(cur, tid, unit_names, goods_cat_ids, supplier_ids, shop_ids)
    print("账户(20) ...")
    account_ids = gen_accounts(cur, tid, shop_ids)

    # 商品售价用于单据
    prices = [rf(5, 500) for _ in goods_ids]

    print("进货单(10000) ...")
    gen_purchases(cur, tid, supplier_ids, goods_ids, prices, salesman_ids, account_ids, shop_ids)
    print("进货退货单(5000) ...")
    gen_purchase_returns(cur, tid, supplier_ids, goods_ids, prices, salesman_ids, account_ids, shop_ids)
    print("销售单(20000) ...")
    gen_sales(cur, tid, customer_ids, goods_ids, prices, salesman_ids, account_ids, shop_ids)
    print("销售退货单(20000) ...")
    gen_sales_returns(cur, tid, customer_ids, goods_ids, prices, salesman_ids, account_ids, shop_ids)
    print("盘点单(5000) ...")
    gen_stock_counts(cur, tid, goods_ids, salesman_ids, shop_ids)
    print("收款单(5000) / 付款单(5000) ...")
    gen_receipts(cur, tid, customer_ids, salesman_ids, account_ids, shop_ids)
    gen_payments(cur, tid, supplier_ids, salesman_ids, account_ids, shop_ids)

    cur.close()
    conn.close()
    print("完成，用时 %.1fs" % (time.time() - t0))


if __name__ == "__main__":
    main()
