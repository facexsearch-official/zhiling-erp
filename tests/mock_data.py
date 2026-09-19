#!/usr/bin/env python3
"""
Mock data seeder for PISA 进销存系统 - 货品模块.

Creates realistic test data via API calls, with proper dependency ordering:
    单位 → 货品分类 → 规格 → 货品属性 → 供应商 → 货品

Usage:
    python mock_data.py [--base-url http://localhost:8080]
    python mock_data.py --cleanup
"""

import argparse
import json
import sys
import time

try:
    import requests
except ImportError:
    print("需要安装 requests 库: pip install requests")
    sys.exit(1)

BASE_URL = "http://localhost:8080"
PHONE = "admin"
PASSWORD = "123456"

# ── Colors ──────────────────────────────────────────────────────────────

G = "\033[92m"  # green
Y = "\033[93m"  # yellow
R = "\033[91m"  # red
C = "\033[96m"  # cyan
B = "\033[1m"   # bold
D = "\033[0m"   # reset

SESSION = requests.Session()
CREATED = {}  # module -> [ids]


def log(msg):
    print(f"{C}▸{D} {msg}")


def ok(msg):
    print(f"{G}  ✓ {msg}{D}")


def warn(msg):
    print(f"{Y}  ⚠ {msg}{D}")


def fail(msg):
    print(f"{R}  ✗ {msg}{D}")


def track(module, item_id):
    if item_id:
        CREATED.setdefault(module, []).append(item_id)
    return item_id


# ── API ─────────────────────────────────────────────────────────────────

def login():
    log("登录获取 Token...")
    r = SESSION.post(f"{BASE_URL}/api/auth/login", json={"phone": PHONE, "password": PASSWORD})
    if r.status_code != 200:
        fail(f"登录失败: {r.status_code} {r.text}")
        sys.exit(1)
    data = r.json()
    token = data.get("token") or data.get("data", {}).get("token")
    if not token:
        fail(f"未获取到 Token: {json.dumps(data, ensure_ascii=False)}")
        sys.exit(1)
    SESSION.headers["Authorization"] = f"Bearer {token}"
    SESSION.headers["Content-Type"] = "application/json"
    ok("登录成功，Token 已设置")
    return data


def create(module, path, payload):
    r = SESSION.post(f"{BASE_URL}{path}", json=payload)
    if r.status_code in (200, 201):
        resp = r.json()
        item_id = resp.get("id") or resp.get("data", {}).get("id")
        track(module, item_id)
        return resp.get("data") or resp
    warn(f"  创建失败 [{payload.get('name', payload.get('code', '?'))}]: {r.status_code} {r.text[:120]}")
    return None


def list_all(path):
    r = SESSION.get(f"{BASE_URL}{path}")
    if r.status_code == 200:
        d = r.json()
        return d.get("data") or []
    return []


# ── Mock Data ───────────────────────────────────────────────────────────

MOCK_UNITS = ["件", "盒", "包", "箱", "袋", "瓶", "斤", "打"]

MOCK_CATEGORIES = [
    {"name": "数码配件", "children": ["手机配件", "电脑配件", "音频设备"]},
    {"name": "日用百货", "children": ["清洁用品", "厨房用品"]},
    {"name": "食品饮料", "children": ["休闲零食", "冲调饮品"]},
]

MOCK_ATTRIBUTES = [
    {"name": "颜色", "values": ["红色", "白色", "黑色", "蓝色", "米色"]},
    {"name": "尺码", "values": ["S", "M", "L", "XL", "XXL"]},
    {"name": "容量", "values": ["250ml", "500ml", "1L", "2L"]},
    {"name": "材质", "values": ["塑料", "不锈钢", "玻璃", "陶瓷"]},
    {"name": "口味", "values": ["原味", "盐焗", "奶油", "香辣"]},
    {"name": "尺寸", "values": ["小号", "中号", "大号"]},
]

MOCK_PROPERTIES = [
    {"name": "品牌", "type": 1, "values": ["华为", "小米", "苹果", "三星"]},
    {"name": "产地", "type": 2},
    {"name": "保质期", "type": 2},
]

MOCK_SUPPLIERS = [
    {"name": "优品数码供应商", "contact": "王建国", "phone": "13800138001",
     "address": "深圳市华强北电子市场A区101", "bank_name": "工商银行深圳分行",
     "bank_account": "6222021234567890001"},
    {"name": "金鑫电子配件", "contact": "李明华", "phone": "13800138002",
     "address": "广州市天河区天河路385号", "bank_name": "建设银行广州分行",
     "bank_account": "6227001234567890002"},
    {"name": "恒达贸易有限公司", "contact": "张伟东", "phone": "13800138003",
     "address": "东莞市南城区科技路88号", "bank_name": "农业银行东莞分行",
     "bank_account": "6228481234567890003"},
]

# 每个货品覆盖不同的能力组合：
#   levels       自定义价格等级
#   units        启用多单位（含换算系数 factor，相对主单位）
#   specs        启用多规格（可多组，笛卡尔积）
#   alert        启用库存预警（min/max）
MOCK_GOODS = [
    # ① 单单位 + 单规格 + 自定义价格等级 + 库存预警
    {"name": "手机壳 iPhone15", "code": "CASE-IP15", "barcode": "6900000000011",
     "category": "手机配件", "unit": "件", "brand": "苹果", "origin": "深圳",
     "remark": "热销款", "purchase": 8, "retail": 29, "wholesale": 18,
     "levels": {"VIP": 25, "批发": 16}, "alert": {"min": 20, "max": 500}, "stock": 300},

    # ② 单单位 + 单规格 + 自定义价格等级
    {"name": "钢化膜 通用", "code": "FILM-UNI", "barcode": "6900000000012",
     "category": "手机配件", "unit": "件", "brand": "华为", "origin": "广州",
     "purchase": 3, "retail": 15, "wholesale": 8,
     "levels": {"VIP": 12}, "stock": 1000},

    # ③ 单单位 + 单规格（无价格等级）
    {"name": "数据线 Type-C", "code": "CABLE-TYP", "barcode": "6900000000013",
     "category": "手机配件", "unit": "件", "brand": "小米", "origin": "东莞",
     "purchase": 5, "retail": 19, "wholesale": 12, "stock": 800},

    # ④ 多单位（2 辅单位）+ 单规格 + 自定义价格等级
    {"name": "充电宝 20000mAh", "code": "PB20K", "barcode": "6900000000014",
     "category": "手机配件", "unit": "件", "brand": "小米", "origin": "深圳",
     "purchase": 45, "retail": 99, "wholesale": 75,
     "levels": {"VIP": 89, "批发": 70},
     "units": [{"name": "盒", "factor": 20}, {"name": "箱", "factor": 400}],
     "alert": {"min": 10, "max": 200}, "stock": 150},

    # ⑤ 多单位（1 辅单位）+ 单规格 + 自定义价格等级
    {"name": "蓝牙音箱", "code": "BTSPEAKER", "barcode": "6900000000015",
     "category": "音频设备", "unit": "件", "brand": "华为", "origin": "东莞",
     "purchase": 80, "retail": 169, "wholesale": 130,
     "levels": {"VIP": 150}, "units": [{"name": "盒", "factor": 10}], "stock": 60},

    # ⑥ 单单位 + 单规格 + 价格等级 + 预警
    {"name": "无线鼠标", "code": "WIRELESS-M", "barcode": "6900000000016",
     "category": "电脑配件", "unit": "件", "brand": "小米", "origin": "深圳",
     "purchase": 35, "retail": 79, "wholesale": 60,
     "levels": {"VIP": 72}, "alert": {"min": 15, "max": 300}, "stock": 120},

    # ⑦ 单单位 + 多规格（2 组：颜色×尺码）+ 价格等级
    {"name": "纯棉 T恤", "code": "TSHIRT-COT", "barcode": "6900000000017",
     "category": "日用百货", "unit": "件", "brand": "", "origin": "广州",
     "purchase": 18, "retail": 59, "wholesale": 40,
     "levels": {"VIP": 50, "批发": 38},
     "specs": [
         {"group": "颜色", "values": ["红色", "白色", "黑色"]},
         {"group": "尺码", "values": ["S", "M", "L", "XL"]},
     ],
     "stock": 40},

    # ⑧ 单单位 + 多规格（2 组）+ 价格等级 + 预警
    {"name": "运动跑鞋", "code": "RUNSHOE", "barcode": "6900000000018",
     "category": "日用百货", "unit": "件", "brand": "", "origin": "福建",
     "purchase": 120, "retail": 299, "wholesale": 220,
     "levels": {"VIP": 260},
     "specs": [
         {"group": "颜色", "values": ["黑色", "白色"]},
         {"group": "尺码", "values": ["S", "M", "L"]},
     ],
     "alert": {"min": 5, "max": 100}, "stock": 25},

    # ⑨ 单单位 + 多规格（1 组：容量）+ 价格等级
    {"name": "保温杯 不锈钢", "code": "THERMOS", "barcode": "6900000000019",
     "category": "厨房用品", "unit": "件", "brand": "", "origin": "潮州",
     "purchase": 25, "retail": 59, "wholesale": 42,
     "levels": {"VIP": 50},
     "specs": [{"group": "容量", "values": ["250ml", "500ml", "1L"]}],
     "stock": 80},

    # ⑩ 多单位 + 单规格
    {"name": "洗洁精 500ml", "code": "DETERGENT", "barcode": "6900000000020",
     "category": "清洁用品", "unit": "瓶", "brand": "", "origin": "佛山",
     "purchase": 6, "retail": 12.5, "wholesale": 9,
     "units": [{"name": "箱", "factor": 24}], "stock": 500},

    # ⑪ 多单位 + 多规格（1 组）+ 价格等级 + 预警
    {"name": "每日坚果 750g", "code": "NUTS750", "barcode": "6900000000021",
     "category": "休闲零食", "unit": "袋", "brand": "", "origin": "杭州",
     "purchase": 38, "retail": 69, "wholesale": 55,
     "levels": {"VIP": 62},
     "units": [{"name": "箱", "factor": 12}],
     "specs": [{"group": "口味", "values": ["原味", "盐焗"]}],
     "alert": {"min": 10, "max": 200}, "stock": 60},

    # ⑫ 多单位 + 多规格（1 组）+ 价格等级
    {"name": "速溶咖啡 100条", "code": "COFFEE100", "barcode": "6900000000022",
     "category": "冲调饮品", "unit": "盒", "brand": "", "origin": "昆明",
     "purchase": 45, "retail": 89, "wholesale": 70,
     "levels": {"VIP": 82},
     "units": [{"name": "箱", "factor": 6}],
     "specs": [{"group": "口味", "values": ["原味", "香辣", "奶油"]}],
     "stock": 90},

    # ⑬ 多单位（2 辅单位）+ 多规格（2 组）+ 价格等级 + 预警（最全组合）
    {"name": "厨房收纳盒", "code": "STORAGE-BOX", "barcode": "6900000000023",
     "category": "厨房用品", "unit": "件", "brand": "", "origin": "台州",
     "purchase": 12, "retail": 39, "wholesale": 26,
     "levels": {"VIP": 34, "批发": 24},
     "units": [{"name": "盒", "factor": 12}, {"name": "箱", "factor": 144}],
     "specs": [
         {"group": "材质", "values": ["塑料", "玻璃"]},
         {"group": "尺寸", "values": ["小号", "中号", "大号"]},
     ],
     "alert": {"min": 20, "max": 600}, "stock": 200},

    # ⑭ 单单位 + 单规格（无价格等级）
    {"name": "矿泉水 550ml", "code": "WATER550", "barcode": "6900000000024",
     "category": "冲调饮品", "unit": "瓶", "brand": "", "origin": "河源",
     "purchase": 1, "retail": 2.5, "wholesale": 1.8,
     "units": [{"name": "箱", "factor": 24}], "stock": 2000},
]


# ── Seeders（幂等：同名/同编码已存在则复用，不重复创建） ──────────────────

def _existing_by_name(path):
    return {x.get("name"): x for x in list_all(path)}


def seed_units():
    log(f"创建 单位 ({len(MOCK_UNITS)} 条)...")
    existing = _existing_by_name("/api/shop/unit/all")
    result = {}
    added = 0
    for name in MOCK_UNITS:
        if name in existing:
            result[name] = existing[name]["id"]
            continue
        item = create("unit", "/api/shop/unit", {"name": name})
        if item:
            result[name] = item["id"]
            added += 1
    ok(f"  新增 {added} 条，复用 {len(result) - added} 条 单位")
    return result


def seed_categories():
    log("创建 货品分类...")
    existing = _existing_by_name("/api/shop/category/all")
    result = {}
    added = 0
    for top in MOCK_CATEGORIES:
        parent = existing.get(top["name"])
        if not parent:
            parent = create("category", "/api/shop/category", {"name": top["name"]})
            added += 1
        if not parent:
            continue
        result[top["name"]] = parent["id"]
        for child in top["children"]:
            c = existing.get(child)
            if not c:
                c = create("category", "/api/shop/category",
                           {"name": child, "parent_id": parent["id"]})
                added += 1
            if c:
                result[child] = c["id"]
    ok(f"  新增 {added} 条，复用 {len(result) - added} 条 分类（含子分类）")
    return result


def seed_attributes():
    log(f"创建 规格 ({len(MOCK_ATTRIBUTES)} 条)...")
    existing = _existing_by_name("/api/shop/attribute/all")
    created = []
    for a in MOCK_ATTRIBUTES:
        if a["name"] in existing:
            continue
        item = create("attribute", "/api/shop/attribute", a)
        if item:
            created.append(item)
    ok(f"  成功创建 {len(created)} 条 规格（跳过已存在 {len(MOCK_ATTRIBUTES) - len(created)} 条）")
    return created


def seed_properties():
    log(f"创建 货品属性 ({len(MOCK_PROPERTIES)} 条)...")
    existing = _existing_by_name("/api/shop/property/all")
    created = []
    for p in MOCK_PROPERTIES:
        if p["name"] in existing:
            continue
        item = create("property", "/api/shop/property", p)
        if item:
            created.append(item)
    ok(f"  成功创建 {len(created)} 条 货品属性（跳过已存在 {len(MOCK_PROPERTIES) - len(created)} 条）")
    return created


def seed_suppliers():
    log(f"创建 供应商 ({len(MOCK_SUPPLIERS)} 条)...")
    existing = _existing_by_name("/api/shop/supplier/all")
    result = []
    added = 0
    for s in MOCK_SUPPLIERS:
        if s["name"] in existing:
            result.append(existing[s["name"]])
            continue
        item = create("supplier", "/api/shop/supplier", s)
        if item:
            result.append(item)
            added += 1
    ok(f"  新增 {added} 条，复用 {len(result) - added} 条 供应商")
    return result


def _price_row(purchase, retail, wholesale, levels, code="", disabled=False):
    return {
        "code": code,
        "barcode": "",
        "purchase_price": str(round(purchase, 2)),
        "retail_price": str(round(retail, 2)),
        "wholesale_price": str(round(wholesale, 2)),
        "custom": {k: str(round(v, 2)) for k, v in levels.items()},
        "disabled": disabled,
    }


def _combos(specs):
    """笛卡尔积，键使用前端一致的 ' / ' 连接。"""
    combos = [""]
    for s in specs:
        combos = [(c + (" / " if c else "") + v) for c in combos for v in s["values"]]
    return combos or [""]


def build_goods_payload(g, units_map, cats_map, suppliers):
    base_p = g["purchase"]
    base_r = g["retail"]
    base_w = g["wholesale"]
    levels = g.get("levels") or {}
    cols = list(levels.keys())
    aux_units = g.get("units") or []
    specs = g.get("specs") or []
    is_multi_unit = len(aux_units) > 0
    is_multi_spec = len(specs) > 0

    unit_defs = [{"name": g["unit"], "factor": 1.0}]
    for u in aux_units:
        unit_defs.append({"name": u["name"], "factor": float(u["factor"])})

    combos = _combos(specs)

    spec_groups = [
        {"name": s["group"], "has_image": False,
         "values": [{"name": v, "image": ""} for v in s["values"]]}
        for s in specs
    ]

    price_rows = {}
    stock_rows = {}

    if is_multi_spec:
        # 多规格：按单位 → 规格组合
        for ud in unit_defs:
            f = ud["factor"]
            price_rows[ud["name"]] = {}
            for ci, combo in enumerate(combos):
                lv = {k: v * f for k, v in levels.items()}
                price_rows[ud["name"]][combo] = _price_row(
                    base_p * f, base_r * f + ci, base_w * f, lv,
                    code=f'{g["code"]}-{ud["name"]}-{ci + 1}')
        alert = g.get("alert") or {}
        for ci, combo in enumerate(combos):
            stock_rows[combo] = {
                "stock": str(10 + ci * 3),
                "init_cost": str(base_p),
                "min_stock": str(alert.get("min", 0)),
                "safe_stock": str(alert.get("min", 0)),
                "max_stock": str(alert.get("max", 0)),
            }
    elif is_multi_unit:
        # 多单位：按单位，单组合
        for ud in unit_defs:
            f = ud["factor"]
            lv = {k: v * f for k, v in levels.items()}
            price_rows[ud["name"]] = {
                "": _price_row(base_p * f, base_r * f, base_w * f, lv,
                               code=f'{g["code"]}-{ud["name"]}')
            }
    else:
        # 单单位单规格
        price_rows = {"__simple__": {"": _price_row(base_p, base_r, base_w, levels)}}

    alert = g.get("alert") or {}
    stock = g.get("stock", 0)
    payload = {
        "name": g["name"],
        "code": g["code"],
        "barcode": g.get("barcode", ""),
        "brand": g.get("brand", ""),
        "origin": g.get("origin", ""),
        "remark": g.get("remark", ""),
        "spec": "" if is_multi_spec else "",
        "category_id": cats_map.get(g.get("category")),
        "unit_id": units_map.get(g.get("unit")),
        "purchase_price": base_p,
        "retail_price": base_r,
        "wholesale_price": base_w,
        "has_multi_unit": 1 if is_multi_unit else 0,
        "has_multi_spec": 1 if is_multi_spec else 0,
        "enable_stock_alert": 1 if alert else 0,
        "min_stock": alert.get("min", 0),
        "max_stock": alert.get("max", 0),
        "init_cost": base_p,
        "current_stock": stock,
        "sales_unit": g["unit"],
        "purchase_unit": g["unit"],
        "images": json.dumps([], ensure_ascii=False),
        "image_url": "",
        "suppliers": json.dumps([s["id"] for s in suppliers[:1]], ensure_ascii=False),
        "supplier_id": suppliers[0]["id"] if suppliers else None,
        "spec_groups": json.dumps(spec_groups, ensure_ascii=False),
        "price_columns": json.dumps(cols, ensure_ascii=False),
        "price_rows": json.dumps(price_rows, ensure_ascii=False),
        "stock_rows": json.dumps(stock_rows, ensure_ascii=False),
        "specs": [],
    }
    if is_multi_unit:
        payload["units"] = [
            {"unit_id": units_map.get(u["name"]), "unit_name": u["name"],
             "factor": u["factor"], "is_main": 0}
            for u in aux_units
        ]
    return payload


def seed_goods(units, categories, suppliers):
    log(f"创建 货品 ({len(MOCK_GOODS)} 条，覆盖单/多单位、单/多规格等组合)...")
    resp = list_all("/api/shop/goods/list?page_size=500")
    rows = resp.get("list") if isinstance(resp, dict) else resp
    existing_codes = {g.get("code") for g in (rows or [])}
    created = 0
    skipped = 0
    for g in MOCK_GOODS:
        if g["code"] in existing_codes:
            skipped += 1
            continue
        payload = build_goods_payload(g, units, categories, suppliers)
        item = create("goods", "/api/shop/goods", payload)
        if item:
            created += 1
    ok(f"  成功创建 {created} 条，跳过已存在 {skipped} 条 货品")


# ── Cleanup ─────────────────────────────────────────────────────────────

import os

STATE_FILE = os.path.join(os.path.dirname(os.path.abspath(__file__)), ".mock_created.json")

CLEANUP_ORDER = [
    ("goods", "/api/shop/goods"),
    ("attribute", "/api/shop/attribute"),
    ("category", "/api/shop/category"),
    ("unit", "/api/shop/unit"),
    ("property", "/api/shop/property"),
    ("supplier", "/api/shop/supplier"),
]


def save_state():
    with open(STATE_FILE, "w", encoding="utf-8") as f:
        json.dump(CREATED, f, ensure_ascii=False)


def cleanup():
    """按上次播种记录的 ID 精确清理，避免误删既有数据。"""
    if not os.path.exists(STATE_FILE):
        warn("未找到 .mock_created.json，无法按记录清理（请先运行播种）")
        return
    with open(STATE_FILE, encoding="utf-8") as f:
        state = json.load(f)

    log("清理 mock 数据（按记录 ID）...")
    total = 0
    for module, path in CLEANUP_ORDER:
        ids = list(state.get(module, []))
        if module == "category":
            ids = list(reversed(ids))  # 子级优先
        for item_id in ids:
            r = SESSION.delete(f"{BASE_URL}{path}/{item_id}")
            if r.status_code in (200, 204, 404):
                total += 1
    os.remove(STATE_FILE)
    ok(f"  已清理 {total} 条数据")


# ── Main ────────────────────────────────────────────────────────────────

def main():
    global BASE_URL

    parser = argparse.ArgumentParser(description="PISA 货品模块 Mock 数据")
    parser.add_argument("--base-url", default=BASE_URL, help="API base URL")
    parser.add_argument("--cleanup", action="store_true", help="删除本脚本创建的 mock 数据")
    args = parser.parse_args()
    BASE_URL = args.base_url.rstrip("/")

    print(f"\n{B}═══════════════════════════════════════════════{D}")
    print(f"{B}  PISA 进销存系统 - 货品模块 Mock Data{D}")
    print(f"{B}═══════════════════════════════════════════════{D}\n")

    login()
    print()

    if args.cleanup:
        cleanup()
        print()
        return

    units = seed_units()
    categories = seed_categories()
    seed_attributes()
    seed_properties()
    suppliers = seed_suppliers()
    seed_goods(units, categories, suppliers)
    save_state()

    print(f"\n{B}── 数据统计 ──{D}")
    for module in ("goods", "attribute", "category", "unit", "property", "supplier"):
        print(f"  {module}: {G}{len(CREATED.get(module, []))}{D} 条")
    print(f"\n{G}{B}Mock 数据创建完成 ✓{D}")
    print(f"  提示: 运行 {C}python mock_data.py --cleanup{D} 可删除本次创建的数据。\n")


if __name__ == "__main__":
    main()
