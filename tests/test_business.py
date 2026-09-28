#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
PISA 进销存 — 全功能接口测试（Python）

覆盖「客户 / 进货 / 销售 / 库存 / 资金 / 分析 / 设置」全部子功能的读取与写入。
使用测试租户（与真实数据隔离），需先运行 tests/mockdata.py 造数。

用法:
    python3 tests/apitest.py
"""
import base64
import hashlib
import hmac
import json
import os
import re
import sys
import time

try:
    import requests
except ImportError:
    sys.exit("缺少依赖 requests，请先执行: pip install requests")

BASE = "http://127.0.0.1:8080/api"
TEST_TENANT_ID = 7000000000000000001
TEST_USER_ID = 7000000000000000002
ROOT = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))

PASS = 0
FAIL = 0
FAILURES = []
REFS = {}


# ── JWT ──
def load_secret():
    with open(os.path.join(ROOT, "config.yaml"), "r", encoding="utf-8") as f:
        for line in f:
            m = re.match(r"\s*secret:\s*[\"']?([^\"'\s]+)", line)
            if m:
                return m.group(1)
    return "pisa-secret-key-change-in-production"


def b64(b):
    return base64.urlsafe_b64encode(b).rstrip(b"=")


def make_token(secret, uid, tid, role=1, hours=72):
    now = int(time.time())
    hdr = {"alg": "HS256", "typ": "JWT"}
    pl = {"user_id": uid, "tenant_id": tid, "shop_id": 0, "role": role,
          "exp": now + hours * 3600, "iat": now}
    seg = b64(json.dumps(hdr, separators=(",", ":")).encode()) + b"." + \
          b64(json.dumps(pl, separators=(",", ":")).encode())
    sig = b64(hmac.new(secret.encode(), seg, hashlib.sha256).digest())
    return (seg + b"." + sig).decode()


SESSION = requests.Session()
TOKEN = ""


def do(method, path, body=None):
    try:
        r = SESSION.request(method, BASE + path, json=body, headers={"Authorization": "Bearer " + TOKEN}, timeout=30)
        try:
            return r.status_code, r.json()
        except ValueError:
            return r.status_code, None
    except Exception:
        return 0, None


def code_of(b):
    if not b:
        return None
    c = b.get("code")
    try:
        return int(c)
    except (TypeError, ValueError):
        return None


def id_of(m):
    if not m:
        return ""
    if m.get("id_str"):
        return str(m["id_str"])
    if m.get("id") is not None:
        return str(m["id"])
    return ""


def first_id(list_path):
    _, b = do("GET", list_path)
    if not b:
        return ""
    data = b.get("data")
    if isinstance(data, dict):
        lst = data.get("list") or []
        if lst:
            return id_of(lst[0])
    if isinstance(data, list) and data:
        return id_of(data[0])
    return ""


def first_id_int(list_path):
    try:
        return int(first_id(list_path))
    except (TypeError, ValueError):
        return 0


def record(name, status, body):
    global PASS, FAIL
    if status != 200 or body is None:
        FAIL += 1
        FAILURES.append("%s — HTTP %s" % (name, status))
        print("  x %-26s HTTP %s" % (name, status))
        return False
    if code_of(body) != 0:
        FAIL += 1
        FAILURES.append("%s — code=%s %s" % (name, body.get("code"), body.get("message")))
        print("  x %-26s %s" % (name, body.get("message")))
        return False
    PASS += 1
    print("  ok %-25s" % name)
    return True


def check(name, method, path, body=None):
    if not path:
        return
    st, b = do(method, path, body)
    record(name, st, b)


def check_fail(name, method, path, body=None):
    """预期业务拒绝（code != 0，HTTP 可为 200/400）"""
    global PASS, FAIL
    _, b = do(method, path, body)
    if b is not None and code_of(b) not in (None, 0):
        PASS += 1
        print("  ok %-25s (预期拒绝)" % name)
        return
    FAIL += 1
    FAILURES.append(name + " — 预期拒绝但成功")
    print("  x %-26s 预期拒绝但成功" % name)


def check_count(name, path, minimum):
    global PASS, FAIL
    _, b = do("GET", path)
    n = 0
    if b and isinstance(b.get("data"), list):
        n = len(b["data"])
    if n >= minimum:
        PASS += 1
        print("  ok %-25s (%d 条)" % (name, n))
    else:
        FAIL += 1
        FAILURES.append("%s — 仅 %d 条 (<%d)" % (name, n, minimum))
        print("  x %-26s 仅 %d 条 (<%d)" % (name, n, minimum))


def resolve(path):
    """把 {id} 用列表首条 id 替换"""
    if "{id}" not in path:
        return path
    mapping = [
        ("/customer/stats/", "/shop/customer/list?page=1&page_size=1"),
        ("/customer/statement/", "/shop/customer/list?page=1&page_size=1"),
        ("/supplier/", "/shop/supplier/list?page=1&page_size=1"),
        ("/purchase-return/", "/shop/purchase-return/list?page=1&page_size=1"),
        ("/purchase-order/", "/shop/purchase-order/list?page=1&page_size=1"),
        ("/purchase/", "/shop/purchase/list?page=1&page_size=1"),
        ("/sale-order/", "/shop/sale-order/list?page=1&page_size=1"),
        ("/sale-return/", "/shop/sale-return/list?page=1&page_size=1"),
        ("/sale/", "/shop/sale/list?page=1&page_size=1"),
        ("/quote/", "/shop/quote/list?page=1&page_size=1"),
        ("/stock-count/", "/shop/stock-count/list?page=1&page_size=1"),
        ("/assembly/", "/shop/assembly/list?page=1&page_size=1"),
        ("/split/", "/shop/split/list?page=1&page_size=1"),
        ("/recipe/", "/shop/recipe/list?page=1&page_size=1"),
        ("/combo/", "/shop/combo/list?page=1&page_size=1"),
        ("/warehouse/", "/shop/warehouse/list?page=1&page_size=1"),
        ("/receipt/", "/shop/receipt/list?page=1&page_size=1"),
        ("/payment/", "/shop/payment/list?page=1&page_size=1"),
        ("/income/", "/shop/income/list?page=1&page_size=1"),
        ("/goods/", "/shop/goods/list?page=1&page_size=1"),
        ("/customer/", "/shop/customer/list?page=1&page_size=1"),
        ("/customer-price/", "/shop/customer-price/list?page=1&page_size=1"),
    ]
    list_path = ""
    for key, lp in mapping:
        if key in path:
            list_path = lp
            break
    if not list_path:
        return path
    i = first_id(list_path)
    if not i:
        record(path, 0, None)
        return ""
    return path.replace("{id}", i)


# ── 读取用例 ──
READ_GROUPS = [
    ("客户", [
        ("客户列表", "GET", "/shop/customer/list?page=1&page_size=20"),
        ("客户列表(关键字)", "GET", "/shop/customer/list?keyword=测试客户0001"),
        ("客户列表(分页2)", "GET", "/shop/customer/list?page=2&page_size=20"),
        ("全部客户", "GET", "/shop/customer/all"),
        ("客户详情", "GET", "/shop/customer/{id}"),
        ("客户统计", "GET", "/shop/customer/stats/{id}"),
        ("客户对账单", "GET", "/shop/customer/statement/{id}"),
        ("客户分类列表", "GET", "/shop/customer-category/list"),
        ("价格等级列表", "GET", "/shop/price-level/list"),
        ("报价管理列表", "GET", "/shop/customer-price/list?page=1&page_size=20"),
        ("按客户报价", "GET", "/shop/customer-price/by-customer/{id}"),
        ("按商品报价", "GET", "/shop/customer-price/by-goods/{id}"),
        ("报价历史", "GET", "/shop/customer-price/history/{id}"),
    ]),
    ("进货", [
        ("供应商列表", "GET", "/shop/supplier/list?page=1&page_size=20"),
        ("供应商关键字", "GET", "/shop/supplier/list?keyword=测试供应商0001"),
        ("全部供应商", "GET", "/shop/supplier/all"),
        ("供应商详情", "GET", "/shop/supplier/{id}"),
        ("供应商分类", "GET", "/shop/supplier-category/all"),
        ("进货单列表", "GET", "/shop/purchase/list?page=1&page_size=20"),
        ("进货单日期筛选", "GET", "/shop/purchase/list?date_from=2020-01-01&date_to=2030-01-01"),
        ("进货单详情", "GET", "/shop/purchase/{id}"),
        ("进货预订列表", "GET", "/shop/purchase-order/list?page=1&page_size=20"),
        ("进货预订详情", "GET", "/shop/purchase-order/{id}"),
        ("进货退货列表", "GET", "/shop/purchase-return/list?page=1&page_size=20"),
        ("进货退货详情", "GET", "/shop/purchase-return/{id}"),
    ]),
    ("销售", [
        ("销售单列表", "GET", "/shop/sale/list?page=1&page_size=20"),
        ("销售单分页3", "GET", "/shop/sale/list?page=3&page_size=20"),
        ("销售单详情", "GET", "/shop/sale/{id}"),
        ("销售预订列表", "GET", "/shop/sale-order/list?page=1&page_size=20"),
        ("销售预订详情", "GET", "/shop/sale-order/{id}"),
        ("销售退货列表", "GET", "/shop/sale-return/list?page=1&page_size=20"),
        ("销售退货详情", "GET", "/shop/sale-return/{id}"),
        ("报价单列表", "GET", "/shop/quote/list?page=1&page_size=20"),
        ("报价单详情", "GET", "/shop/quote/{id}"),
        ("提成规则列表", "GET", "/shop/commission-rule/list"),
        ("业绩提成报表", "GET", "/shop/commission/report"),
    ]),
    ("库存", [
        ("库存查询列表", "GET", "/shop/stock-query/list?page=1&page_size=20"),
        ("库存预警", "GET", "/shop/stock-query/alert"),
        ("库存流水", "GET", "/shop/stock-query/flow/{id}"),
        ("成本明细", "GET", "/shop/stock-query/cost/{id}"),
        ("批次查询", "GET", "/shop/batch/list?page=1&page_size=20"),
        ("保质期查询", "GET", "/shop/batch/expiry"),
        ("盘点单列表", "GET", "/shop/stock-count/list?page=1&page_size=20"),
        ("盘点单详情", "GET", "/shop/stock-count/{id}"),
        ("组装单列表", "GET", "/shop/assembly/list?page=1&page_size=20"),
        ("组装单详情", "GET", "/shop/assembly/{id}"),
        ("拆分单列表", "GET", "/shop/split/list?page=1&page_size=20"),
        ("拆分单详情", "GET", "/shop/split/{id}"),
        ("配方列表", "GET", "/shop/recipe/list?page=1&page_size=20"),
        ("配方详情", "GET", "/shop/recipe/{id}"),
        ("套餐列表", "GET", "/shop/combo/list?page=1&page_size=20"),
        ("套餐详情", "GET", "/shop/combo/{id}"),
        ("仓库列表", "GET", "/shop/warehouse/list?page=1&page_size=20"),
        ("全部仓库", "GET", "/shop/warehouse/all"),
        ("商品列表", "GET", "/shop/goods/list?page=1&page_size=20"),
        ("商品列表(关键字)", "GET", "/shop/goods/list?keyword=测试商品0001"),
        ("全部商品", "GET", "/shop/goods/all"),
        ("商品详情", "GET", "/shop/goods/{id}"),
        ("商品品牌", "GET", "/shop/goods/brands"),
        ("商品产地", "GET", "/shop/goods/origins"),
        ("商品分类", "GET", "/shop/category/all"),
        ("单位列表", "GET", "/shop/unit/all"),
        ("辅助属性", "GET", "/shop/attribute/all"),
        ("货品属性", "GET", "/shop/property/all"),
        ("价格列表", "GET", "/shop/price/list?page=1&page_size=20"),
    ]),
    ("资金", [
        ("账户列表", "GET", "/shop/account/list?page=1&page_size=20"),
        ("全部账户", "GET", "/shop/account/all"),
        ("转账列表", "GET", "/shop/transfer/list?page=1&page_size=20"),
        ("收款单列表", "GET", "/shop/receipt/list?page=1&page_size=20"),
        ("收款单详情", "GET", "/shop/receipt/{id}"),
        ("付款单列表", "GET", "/shop/payment/list?page=1&page_size=20"),
        ("付款单详情", "GET", "/shop/payment/{id}"),
        ("其他收入列表", "GET", "/shop/income/list?page=1&page_size=20"),
        ("其他收入详情", "GET", "/shop/income/{id}"),
        ("收支类型列表", "GET", "/shop/income-type/list"),
        ("其他支出列表", "GET", "/shop/expense/list?page=1&page_size=20"),
        ("客户对账", "GET", "/shop/reconcile/customer"),
        ("供应商对账", "GET", "/shop/reconcile/supplier"),
        ("资金流水", "GET", "/shop/fund-flow/list?page=1&page_size=20"),
    ]),
    ("分析", [
        ("销售统计(按单据)", "GET", "/shop/sales-stat/doc"),
        ("销售统计(按商品)", "GET", "/shop/sales-stat/goods"),
        ("销售统计(按客户)", "GET", "/shop/sales-stat/customer"),
        ("热销分析", "GET", "/shop/analysis/hot-sales"),
        ("员工业绩统计", "GET", "/shop/analysis/staff-perf"),
        ("进货统计(按商品)", "GET", "/shop/analysis/purchase-goods"),
        ("进货统计(按供应商)", "GET", "/shop/analysis/purchase-supplier"),
        ("库存统计", "GET", "/shop/analysis/stock-stat"),
        ("经营利润", "GET", "/shop/analysis/profit"),
        ("盘盈明细", "GET", "/shop/analysis/surplus-detail"),
        ("盘亏明细", "GET", "/shop/analysis/loss-detail"),
    ]),
    ("设置", [
        ("商户信息", "GET", "/shop/tenant-info"),
        ("门店列表", "GET", "/shop/shops"),
        ("员工列表", "GET", "/shop/staff/users"),
        ("业务员列表", "GET", "/shop/salesmen"),
        ("角色列表", "GET", "/shop/role/list"),
        ("系统设置", "GET", "/shop/system-setting"),
        ("用户偏好设置", "GET", "/shop/user-preference"),
        ("积分设置", "GET", "/shop/points-setting"),
        ("打印设置", "GET", "/shop/print-setting"),
    ]),
]


def load_refs():
    REFS["goods"] = first_id_int("/shop/goods/list?page=1&page_size=1")
    REFS["goods2"] = first_id_int("/shop/goods/list?page=2&page_size=1")
    REFS["supplier"] = first_id_int("/shop/supplier/list?page=1&page_size=1")
    REFS["customer"] = first_id_int("/shop/customer/list?page=1&page_size=1")
    REFS["account"] = first_id_int("/shop/account/list?page=1&page_size=1")
    REFS["account2"] = first_id_int("/shop/account/list?page=2&page_size=1")
    REFS["warehouse"] = first_id_int("/shop/warehouse/list?page=1&page_size=1")
    REFS["incomeType"] = first_id_int("/shop/income-type/list")
    REFS["goodsCategory"] = first_id_int("/shop/category/all")
    REFS["unit"] = first_id_int("/shop/unit/all")
    REFS["shop"] = first_id_int("/shop/shops")


def seq(name, path, create_body, update_body):
    st, b = do("POST", path, create_body)
    if not record(name + "-新增", st, b):
        return
    i = id_of((b or {}).get("data") or {})
    if update_body is not None and i:
        st, b = do("PUT", path + "/" + i, update_body)
        record(name + "-修改", st, b)
    st, b = do("DELETE", path + "/" + i)
    record(name + "-删除", st, b)


def doc_flow(name, path, create_body, actions, delete):
    st, b = do("POST", path, create_body)
    if not record(name + "-新增", st, b):
        return
    i = id_of((b or {}).get("data") or {})
    for a in actions:
        st, b = do("POST", path + "/" + i + "/" + a)
        record(name + "-" + a, st, b)
    if delete:
        st, b = do("DELETE", path + "/" + i)
        record(name + "-删除", st, b)


def doc_audit(name, path, create_body):
    st, b = do("POST", path, create_body)
    if not record(name + "-新增", st, b):
        return
    i = id_of((b or {}).get("data") or {})
    st, b = do("POST", path + "/" + i + "/audit")
    record(name + "-审核", st, b)


def customer_category_cases():
    print("\n【客户·客户分类】")
    check("客户分类-列表", "GET", "/shop/customer-category/list")
    check_count("客户分类-数据量≥100", "/shop/customer-category/list", 100)

    st, b = do("POST", "/shop/customer-category", {"name": "API测试分类"})
    if not record("客户分类-新增", st, b):
        return
    i = id_of((b or {}).get("data") or {})

    st, b = do("POST", "/shop/customer-category", {"name": "API测试子分类", "parent_id": i or "0"})
    record("客户分类-新增子分类", st, b)
    child = id_of((b or {}).get("data") or {})

    if i:
        st, b = do("PUT", "/shop/customer-category/" + i, {"name": "API测试分类改"})
        record("客户分类-修改", st, b)
    if child:
        st, b = do("DELETE", "/shop/customer-category/" + child)
        record("客户分类-删除子分类", st, b)
    if i:
        st, b = do("DELETE", "/shop/customer-category/" + i)
        record("客户分类-删除", st, b)


def write_cases():
    g, g2 = REFS["goods"], REFS["goods2"]
    sup, cus, acc, acc2 = REFS["supplier"], REFS["customer"], REFS["account"], REFS["account2"]
    wh, itype, gcat, unit, shop = REFS["warehouse"], REFS["incomeType"], REFS["goodsCategory"], REFS["unit"], REFS["shop"]

    print("\n【客户·写入】")
    customer_category_cases()
    seq("客户", "/shop/customer", {"name": "API测试客户", "phone": "13900000001"}, {"name": "API测试客户改"})
    seq("价格等级", "/shop/price-level", {"name": "API价格等级"}, {"name": "API价格等级改"})
    check("报价保存", "POST", "/shop/customer-price/save",
          {"items": [{"customer_id": cus, "goods_id": g, "price": 9.9}]})

    print("\n【进货·写入】")
    seq("供应商分类", "/shop/supplier-category", {"name": "API供应商分类"}, {"name": "API供应商分类改"})
    seq("供应商", "/shop/supplier", {"name": "API测试供应商", "phone": "13900000002"}, {"name": "API测试供应商改"})
    doc_flow("进货单", "/shop/purchase",
             {"supplier_id": sup, "warehouse_id": wh, "bill_date": "2026-09-01",
              "items": [{"goods_id": g, "quantity": 2, "unit_price": 5.0}]}, ["audit", "unaudit"], True)
    doc_flow("进货预订", "/shop/purchase-order",
             {"supplier_id": sup, "warehouse_id": wh, "order_date": "2026-09-01",
              "items": [{"goods_id": g, "quantity": 2, "unit_price": 5.0}]}, [], True)
    doc_audit("进货预订审核", "/shop/purchase-order",
              {"supplier_id": sup, "warehouse_id": wh, "order_date": "2026-09-01",
               "items": [{"goods_id": g, "quantity": 2, "unit_price": 5.0}]})
    doc_flow("进货退货", "/shop/purchase-return",
             {"supplier_id": sup, "warehouse_id": wh, "bill_date": "2026-09-01",
              "items": [{"goods_id": g, "quantity": 1, "unit_price": 5.0}]}, [], True)
    doc_audit("进货退货审核", "/shop/purchase-return",
              {"supplier_id": sup, "warehouse_id": wh, "bill_date": "2026-09-01",
               "items": [{"goods_id": g, "quantity": 1, "unit_price": 5.0}]})

    print("\n【销售·写入】")
    doc_flow("销售单", "/shop/sale",
             {"customer_id": cus, "warehouse_id": wh, "bill_date": "2026-09-01",
              "items": [{"goods_id": g, "quantity": 1, "unit_price": 12.0}]}, [], True)
    doc_flow("销售预订", "/shop/sale-order",
             {"customer_id": cus, "warehouse_id": wh, "order_date": "2026-09-01",
              "items": [{"goods_id": g, "quantity": 1, "unit_price": 12.0}]}, [], True)
    doc_flow("销售退货", "/shop/sale-return",
             {"customer_id": cus, "warehouse_id": wh, "bill_date": "2026-09-01",
              "items": [{"goods_id": g, "quantity": 1, "unit_price": 12.0}]}, [], True)
    doc_flow("报价单", "/shop/quote",
             {"customer_id": cus, "bill_date": "2026-09-01",
              "items": [{"goods_id": g, "quantity": 1, "unit_price": 12.0}]}, [], True)
    seq("提成规则", "/shop/commission-rule", {"name": "API提成规则", "method": 1, "ratio": 5},
        {"name": "API提成规则改", "method": 1, "ratio": 6})

    print("\n【库存·写入】")
    seq("商品分类", "/shop/category", {"name": "API商品分类"}, {"name": "API商品分类改"})
    seq("单位", "/shop/unit", {"name": "API单位"}, {"name": "API单位改"})
    seq("辅助属性", "/shop/attribute", {"name": "API属性", "values": ["a", "b"]},
        {"name": "API属性改", "values": ["a"]})
    check_fail("货品属性-新增(超上限)", "POST", "/shop/property", {"name": "API货品属性", "type": 1, "values": ["x"]})
    pid = first_id("/shop/property/all")
    if pid:
        check("货品属性-修改", "PUT", "/shop/property/" + pid, {"name": "API货品属性改", "type": 1, "values": ["y"]})
        check("货品属性-删除", "DELETE", "/shop/property/" + pid)
    seq("商品", "/shop/goods",
        {"name": "API测试商品", "code": "APIG0001", "purchase_price": 8, "retail_price": 12,
         "unit_id": unit, "category_id": gcat}, {"name": "API测试商品改"})
    seq("套餐", "/shop/combo", {"name": "API测试套餐", "items": [{"goods_id": g, "quantity": 1}]}, None)
    seq("仓库", "/shop/warehouse", {"name": "API测试仓库", "shop_id": shop},
        {"name": "API测试仓库改", "shop_id": shop})
    check("批量改价", "PUT", "/shop/price/batch",
          {"scope": "selected", "keys": [], "changes": [{"field": "retail_price", "mode": "percent", "percent": 1}]})
    check("价格行保存", "PUT", "/shop/price/rows",
          {"rows": [{"goods_id": g, "unit_key": "个", "spec_key": "", "purchase_price": 8,
                     "retail_price": 12, "wholesale_price": 10}]})
    doc_flow("盘点单", "/shop/stock-count",
             {"warehouse_id": wh, "bill_date": "2026-09-01",
              "items": [{"goods_id": g, "book_qty": 10, "actual_qty": 12}]}, ["void"], False)
    doc_flow("组装单", "/shop/assembly",
             {"bill_date": "2026-09-01", "assembly_fee": 1.5,
              "items": [{"kind": 1, "goods_id": g, "quantity": 1, "unit_cost": 10},
                        {"kind": 2, "goods_id": g2, "quantity": 2, "unit_cost": 5}]}, ["void"], False)
    doc_flow("拆分单", "/shop/split",
             {"bill_date": "2026-09-01",
              "items": [{"kind": 1, "goods_id": g, "quantity": 1, "unit_cost": 10},
                        {"kind": 2, "goods_id": g2, "quantity": 2, "unit_cost": 5}]}, ["void"], False)
    seq("配方", "/shop/recipe",
        {"goods_id": g, "quantity": 1, "items": [{"goods_id": g2, "quantity": 2, "unit_cost": 3}]},
        {"goods_id": g, "quantity": 2, "items": [{"goods_id": g2, "quantity": 3, "unit_cost": 3}]})

    print("\n【资金·写入】")
    seq("账户", "/shop/account", {"name": "API测试账户", "type": 1}, {"name": "API测试账户改", "type": 1})
    check("转账", "POST", "/shop/transfer",
          {"from_account_id": acc, "to_account_id": acc2, "amount": 100, "bill_date": "2026-09-01"})
    doc_flow("收款单", "/shop/receipt",
             {"customer_id": cus, "account_id": acc, "amount": 100, "bill_date": "2026-09-01"}, ["void"], False)
    doc_flow("付款单", "/shop/payment",
             {"supplier_id": sup, "account_id": acc, "amount": 100, "bill_date": "2026-09-01"}, ["void"], False)
    seq("收支类型", "/shop/income-type", {"name": "API收支类型", "direction": 1},
        {"name": "API收支类型改", "direction": 1})
    doc_flow("其他收入", "/shop/income",
             {"type_id": itype, "bill_date": "2026-09-01",
              "items": [{"account_id": acc, "amount": 50}]}, ["void"], False)
    doc_flow("其他支出", "/shop/expense",
             {"type_id": itype, "bill_date": "2026-09-01",
              "items": [{"account_id": acc, "amount": 50}]}, ["void"], False)

    print("\n【设置·写入】")
    seq("业务员", "/shop/salesmen", {"name": "API业务员", "phone": "13900000003", "shop_id": shop},
        {"name": "API业务员改", "phone": "13900000003", "shop_id": shop})
    seq("门店", "/shop/shops", {"name": "API测试门店", "phone": "13900000004"},
        {"name": "API测试门店改", "phone": "13900000004"})
    seq("角色", "/shop/role", {"name": "API测试角色", "description": "d", "permissions": "{}", "sensitive_data": "{}"},
        {"name": "API测试角色改", "description": "d", "permissions": "{}", "sensitive_data": "{}"})
    user_seq()
    check("商户信息保存", "PUT", "/shop/tenant-info", {"name": "压测商户(测试专用)", "business_mode": 2})
    check("系统设置保存", "PUT", "/shop/system-setting", {"multi_shop": 1, "batch_manage": 1})
    check("用户偏好保存", "PUT", "/shop/user-preference", {"sale_default_date": "today"})
    check("积分设置保存", "PUT", "/shop/points-setting", {"points_enabled": 1, "points_per": 100, "points_reward": 5})
    check("打印设置保存", "PUT", "/shop/print-setting",
          {"templates": {"sale": {"list": ["二等分", "A4"], "default": "二等分"}}})


def user_seq():
    st, b = do("POST", "/shop/staff/users",
               {"name": "API员工", "phone": "13900000009", "password": "123456", "role": 3})
    if not record("员工-新增", st, b):
        return
    i = id_of((b or {}).get("data") or {})
    if i:
        st, b = do("PUT", "/shop/staff/users/" + i, {"name": "API员工改", "role": 3})
        record("员工-修改", st, b)
        st, b = do("DELETE", "/shop/staff/users/" + i)
        record("员工-删除", st, b)


def main():
    global TOKEN
    TOKEN = make_token(load_secret(), TEST_USER_ID, TEST_TENANT_ID, 1)
    print("=" * 56)
    print(" PISA 进销存 — 全功能接口测试报告")
    print(" 时间: %s" % time.strftime("%Y-%m-%d %H:%M:%S"))
    print(" 租户: %d   服务: %s" % (TEST_TENANT_ID, BASE))
    print("=" * 56)

    load_refs()
    for module, cases in READ_GROUPS:
        print("\n【%s·读取】" % module)
        for name, method, path in cases:
            check(name, method, resolve(path))

    write_cases()

    print("\n" + "=" * 56)
    print(" 通过: %d   失败: %d   合计: %d" % (PASS, FAIL, PASS + FAIL))
    if FAIL:
        print(" -------- 失败明细 --------")
        for f in FAILURES:
            print("  x " + f)
        print("=" * 56)
        sys.exit(1)
    print(" 全部通过")
    print("=" * 56)


if __name__ == "__main__":
    main()
