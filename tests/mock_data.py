#!/usr/bin/env python3
"""
Mock data seeder for PISA 进销存系统.
Creates realistic test data for all settings modules via API calls.

Usage:
    python mock_data.py [--base-url http://localhost:8080]
"""

import argparse
import json
import random
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


def log(msg):
    print(f"{C}▸{D} {msg}")


def ok(msg):
    print(f"{G}  ✓ {msg}{D}")


def warn(msg):
    print(f"{Y}  ⚠ {msg}{D}")


def fail(msg):
    print(f"{R}  ✗ {msg}{D}")


# ── Auth ────────────────────────────────────────────────────────────────

def login(session):
    log("登录获取 Token...")
    r = session.post(f"{BASE_URL}/api/auth/login", json={
        "phone": PHONE,
        "password": PASSWORD,
    })
    if r.status_code != 200:
        fail(f"登录失败: {r.status_code} {r.text}")
        sys.exit(1)
    data = r.json()
    token = data.get("token") or data.get("data", {}).get("token")
    if not token:
        fail(f"未获取到 Token: {json.dumps(data, ensure_ascii=False)}")
        sys.exit(1)
    session.headers["Authorization"] = f"Bearer {token}"
    ok("登录成功，Token 已设置")
    return data


def switch_tenant(session):
    log("切换到演示商户...")
    r = session.post(f"{BASE_URL}/api/auth/switch-tenant", json={
        "tenant_id": 2,
    })
    if r.status_code != 200:
        warn(f"切换商户失败 (可能已经是当前商户): {r.status_code}")
    else:
        ok("切换商户成功")
    return r.json() if r.status_code == 200 else {}


# ── CRUD Helper ─────────────────────────────────────────────────────────

class ModuleSeeder:
    def __init__(self, session, module_name, api_path, items):
        self.session = session
        self.module_name = module_name
        self.api_path = api_path  # e.g. "/api/shop/supplier"
        self.items = items
        self.created_ids = []

    def seed(self):
        log(f"创建 {self.module_name} 数据 ({len(self.items)} 条)...")
        created = 0
        for item in self.items:
            r = self.session.post(f"{BASE_URL}{self.api_path}", json=item)
            if r.status_code in (200, 201):
                resp = r.json()
                item_id = resp.get("id") or resp.get("data", {}).get("id")
                if item_id:
                    self.created_ids.append(item_id)
                created += 1
            else:
                warn(f"  创建失败 [{item.get('name', item.get('code', '?'))}]: {r.status_code} {r.text[:100]}")
        ok(f"  成功创建 {created}/{len(self.items)} 条 {self.module_name}")
        return self.created_ids

    def cleanup(self):
        log(f"清理 {self.module_name} 数据 ({len(self.created_ids)} 条)...")
        deleted = 0
        for item_id in self.created_ids:
            r = self.session.delete(f"{BASE_URL}{self.api_path}/{item_id}")
            if r.status_code in (200, 204):
                deleted += 1
        ok(f"  已清理 {deleted} 条 {self.module_name}")


# ── Mock Data ───────────────────────────────────────────────────────────

MOCK_SUPPLIERS = [
    {"name": "优品数码供应商", "contact": "王建国", "phone": "13800138001", "address": "深圳市华强北电子市场A区101", "bank_name": "工商银行深圳分行", "bank_account": "6222021234567890001"},
    {"name": "金鑫电子配件", "contact": "李明华", "phone": "13800138002", "address": "广州市天河区天河路385号", "bank_name": "建设银行广州分行", "bank_account": "6227001234567890002"},
    {"name": "恒达科技有限公司", "contact": "张伟东", "phone": "13800138003", "address": "东莞市南城区科技路88号", "bank_name": "农业银行东莞分行", "bank_account": "6228481234567890003"},
    {"name": "瑞丰贸易有限公司", "contact": "陈志强", "phone": "13800138004", "address": "佛山市顺德区容桂大道168号", "bank_name": "中国银行佛山分行", "bank_account": "6216261234567890004"},
    {"name": "华强电子市场", "contact": "刘国华", "phone": "13800138005", "address": "深圳市福田区华强北路1001号", "bank_name": "招商银行深圳分行", "bank_account": "6214831234567890005"},
    {"name": "深圳市优联科技", "contact": "赵明辉", "phone": "13800138006", "address": "深圳市南山区科技园南区", "bank_name": "交通银行深圳分行", "bank_account": "6222601234567890006"},
    {"name": "广州鑫达贸易", "contact": "孙丽萍", "phone": "13800138007", "address": "广州市白云区机场路168号", "bank_name": "浦发银行广州分行", "bank_account": "6225221234567890007"},
    {"name": "东莞市宏发实业", "contact": "周大伟", "phone": "13800138008", "address": "东莞市长安镇乌沙社区", "bank_name": "民生银行东莞分行", "bank_account": "6226181234567890008"},
    {"name": "佛山市顺德供应链", "contact": "吴小明", "phone": "13800138009", "address": "佛山市顺德区陈村镇", "bank_name": "光大银行佛山分行", "bank_account": "6226681234567890009"},
    {"name": "中山市天成贸易", "contact": "郑海涛", "phone": "13800138010", "address": "中山市石岐区兴中道", "bank_name": "兴业银行中山分行", "bank_account": "6229081234567890010"},
]

MOCK_CUSTOMERS = [
    {"name": "张三五金店", "contact": "张三", "phone": "13900139001", "address": "深圳市罗湖区东门南路18号", "type": 1},
    {"name": "李四便利店", "contact": "李四", "phone": "13900139002", "address": "深圳市福田区福华路88号", "type": 1},
    {"name": "王五数码专营店", "contact": "王五", "phone": "13900139003", "address": "广州市天河区天河路228号", "type": 2},
    {"name": "赵六手机配件城", "contact": "赵六", "phone": "13900139004", "address": "广州市越秀区中山五路66号", "type": 2},
    {"name": "陈七电脑维修中心", "contact": "陈七", "phone": "13900139005", "address": "深圳市南山区深南大道9988号", "type": 1},
    {"name": "刘八电器商行", "contact": "刘八", "phone": "13900139006", "address": "东莞市东城区东城大道100号", "type": 2},
    {"name": "杨九日用百货", "contact": "杨九", "phone": "13900139007", "address": "佛山市禅城区汾江路88号", "type": 1},
    {"name": "黄十建材市场", "contact": "黄十", "phone": "13900139008", "address": "中山市火炬开发区", "type": 2},
    {"name": "周十一母婴用品", "contact": "周十一", "phone": "13900139009", "address": "深圳市宝安区西乡大道66号", "type": 1},
    {"name": "吴十二服装批发", "contact": "吴十二", "phone": "13900139010", "address": "广州市白云区嘉禾望岗", "type": 2},
    {"name": "郑十三食品商行", "contact": "郑十三", "phone": "13900139011", "address": "东莞市厚街镇家具大道", "type": 1},
    {"name": "冯十四汽车配件", "contact": "冯十四", "phone": "13900139012", "address": "佛山市南海区桂城街道", "type": 2},
    {"name": "朱十五文具办公", "contact": "朱十五", "phone": "13900139013", "address": "深圳市龙岗区布吉街道", "type": 1},
    {"name": "秦十六家居建材", "contact": "秦十六", "phone": "13900139014", "address": "中山市小榄镇民安路", "type": 2},
    {"name": "许十八宠物用品", "contact": "许十八", "phone": "13900139015", "address": "广州市番禺区市桥街", "type": 1},
]

MOCK_WAREHOUSES = [
    {"name": "主仓库（深圳）", "type": 1, "address": "深圳市龙华区大浪街道华荣路148号", "keeper": "王管理", "sort": 1},
    {"name": "分仓库（广州）", "type": 1, "address": "广州市花都区新华街工业区", "keeper": "李管理", "sort": 2},
    {"name": "东莞仓", "type": 1, "address": "东莞市长安镇沙头社区", "keeper": "张管理", "sort": 3},
    {"name": "临时仓", "type": 3, "address": "深圳市宝安区福永街道", "keeper": "赵管理", "sort": 4},
    {"name": "退货仓", "type": 2, "address": "深圳市龙华区民治街道", "keeper": "刘管理", "sort": 5},
]

MOCK_ACCOUNTS = [
    {"name": "现金账户", "type": 1, "sort": 1},
    {"name": "工商银行", "type": 2, "sort": 2},
    {"name": "建设银行", "type": 2, "sort": 3},
    {"name": "支付宝", "type": 3, "sort": 4},
    {"name": "微信支付", "type": 3, "sort": 5},
]

MOCK_GOODS = [
    {"name": "iPhone 15 Pro Max", "code": "IP15PM", "barcode": "6901234567890", "buy_price": 8500, "sell_price": 9999, "wholesale_price": 9200, "stock_quantity": 50, "alert_quantity": 10},
    {"name": "MacBook Pro 14", "code": "MBP14", "barcode": "6901234567891", "buy_price": 12000, "sell_price": 14999, "wholesale_price": 13500, "stock_quantity": 20, "alert_quantity": 5},
    {"name": "iPad Air", "code": "IPADAIR", "barcode": "6901234567892", "buy_price": 3800, "sell_price": 4799, "wholesale_price": 4300, "stock_quantity": 30, "alert_quantity": 8},
    {"name": "AirPods Pro 2", "code": "APP2", "barcode": "6901234567893", "buy_price": 1200, "sell_price": 1799, "wholesale_price": 1500, "stock_quantity": 100, "alert_quantity": 20},
    {"name": "Apple Watch Series 9", "code": "AWS9", "barcode": "6901234567894", "buy_price": 2500, "sell_price": 3299, "wholesale_price": 2900, "stock_quantity": 40, "alert_quantity": 10},
    {"name": "华为 Mate 60 Pro", "code": "HWM60P", "barcode": "6901234567895", "buy_price": 5500, "sell_price": 6999, "wholesale_price": 6200, "stock_quantity": 35, "alert_quantity": 10},
    {"name": "小米 14 Pro", "code": "XM14P", "barcode": "6901234567896", "buy_price": 3200, "sell_price": 4299, "wholesale_price": 3800, "stock_quantity": 60, "alert_quantity": 15},
    {"name": "ThinkPad X1 Carbon", "code": "TPX1C", "barcode": "6901234567897", "buy_price": 8000, "sell_price": 10999, "wholesale_price": 9500, "stock_quantity": 15, "alert_quantity": 3},
    {"name": "戴尔显示器 27寸", "code": "DELL27", "barcode": "6901234567898", "buy_price": 1800, "sell_price": 2499, "wholesale_price": 2100, "stock_quantity": 25, "alert_quantity": 5},
    {"name": "罗技鼠标 MX Master 3", "code": "LGMX3", "barcode": "6901234567899", "buy_price": 500, "sell_price": 799, "wholesale_price": 650, "stock_quantity": 80, "alert_quantity": 20},
    {"name": "机械键盘 Cherry", "code": "MKBC", "barcode": "6901234567900", "buy_price": 350, "sell_price": 599, "wholesale_price": 480, "stock_quantity": 45, "alert_quantity": 10},
    {"name": "USB-C扩展坞", "code": "USBC-HUB", "barcode": "6901234567901", "buy_price": 120, "sell_price": 249, "wholesale_price": 180, "stock_quantity": 120, "alert_quantity": 30},
    {"name": "移动硬盘 1TB", "code": "HDD1T", "barcode": "6901234567902", "buy_price": 280, "sell_price": 429, "wholesale_price": 360, "stock_quantity": 50, "alert_quantity": 10},
    {"name": "网线 CAT6 100米", "code": "CAT6-100", "barcode": "6901234567903", "buy_price": 80, "sell_price": 139, "wholesale_price": 110, "stock_quantity": 200, "alert_quantity": 50},
    {"name": "电源适配器 65W", "code": "PD65W", "barcode": "6901234567904", "buy_price": 60, "sell_price": 129, "wholesale_price": 95, "stock_quantity": 150, "alert_quantity": 30},
    {"name": "手机壳 iPhone15", "code": "CASE-IP15", "barcode": "6901234567905", "buy_price": 8, "sell_price": 29, "wholesale_price": 18, "stock_quantity": 500, "alert_quantity": 100},
    {"name": "钢化膜 通用", "code": "FILM-UNI", "barcode": "6901234567906", "buy_price": 3, "sell_price": 15, "wholesale_price": 8, "stock_quantity": 1000, "alert_quantity": 200},
    {"name": "数据线 Type-C", "code": "CABLE-TYP", "barcode": "6901234567907", "buy_price": 5, "sell_price": 19, "wholesale_price": 12, "stock_quantity": 800, "alert_quantity": 150},
    {"name": "充电宝 20000mAh", "code": "PB20K", "barcode": "6901234567908", "buy_price": 45, "sell_price": 99, "wholesale_price": 75, "stock_quantity": 100, "alert_quantity": 20},
    {"name": "蓝牙音箱", "code": "BTSPEAKER", "barcode": "6901234567909", "buy_price": 80, "sell_price": 169, "wholesale_price": 130, "stock_quantity": 60, "alert_quantity": 15},
]


# ── Main ────────────────────────────────────────────────────────────────

def main():
    global BASE_URL

    parser = argparse.ArgumentParser(description="PISA Mock Data Seeder")
    parser.add_argument("--base-url", default=BASE_URL, help="API base URL")
    parser.add_argument("--cleanup", action="store_true", help="Cleanup seeded data instead of creating")
    args = parser.parse_args()

    BASE_URL = args.base_url.rstrip("/")

    session = requests.Session()
    session.headers["Content-Type"] = "application/json"

    print(f"\n{B}═══════════════════════════════════════════════{D}")
    print(f"{B}  PISA 进销存系统 - Mock Data Seeder{D}")
    print(f"{B}═══════════════════════════════════════════════{D}\n")

    # Auth
    login(session)
    switch_tenant(session)
    print()

    seeders = [
        ModuleSeeder(session, "供应商", "/api/shop/supplier", MOCK_SUPPLIERS),
        ModuleSeeder(session, "客户", "/api/shop/customer", MOCK_CUSTOMERS),
        ModuleSeeder(session, "仓库", "/api/shop/warehouse", MOCK_WAREHOUSES),
        ModuleSeeder(session, "结算账户", "/api/shop/account", MOCK_ACCOUNTS),
        ModuleSeeder(session, "货品", "/api/shop/goods", MOCK_GOODS),
    ]

    if args.cleanup:
        print(f"\n{B}── 清理模式 ──{D}")
        for s in seeders:
            s.cleanup()
        print(f"\n{G}{B}清理完成 ✓{D}\n")
    else:
        print(f"{B}── 创建数据 ──{D}")
        for s in seeders:
            s.seed()

        # Print summary
        print(f"\n{B}── 数据统计 ──{D}")
        for s in seeders:
            count = len(s.created_ids)
            print(f"  {s.module_name}: {G}{count}{D} 条")

        print(f"\n{G}{B}Mock 数据创建完成 ✓{D}")
        print(f"  供应商: {len(MOCK_SUPPLIERS)} 条")
        print(f"  客户:   {len(MOCK_CUSTOMERS)} 条")
        print(f"  仓库:   {len(MOCK_WAREHOUSES)} 条")
        print(f"  账户:   {len(MOCK_ACCOUNTS)} 条")
        print(f"  货品:   {len(MOCK_GOODS)} 条")
        print(f"  合计:   {len(MOCK_SUPPLIERS) + len(MOCK_CUSTOMERS) + len(MOCK_WAREHOUSES) + len(MOCK_ACCOUNTS) + len(MOCK_GOODS)} 条")
        print(f"\n  提示: 运行 {C}python mock_data.py --cleanup{D} 可清理所有 mock 数据\n")


if __name__ == "__main__":
    main()
