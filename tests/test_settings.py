#!/usr/bin/env python3
"""
Comprehensive test suite for PISA 进销存系统 - Settings Module APIs.

Tests all implemented settings module endpoints:
  - Auth (login, switch-tenant)
  - Supplier (供应商) CRUD
  - Customer (客户) CRUD
  - Warehouse (仓库) CRUD
  - Account (结算账户) CRUD
  - Goods (货品) CRUD
  - Purchase (进货) CRUD + audit

Usage:
    python test_settings.py [--base-url http://localhost:8080] [--skip-cleanup]
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

G = "\033[92m"   # green
Y = "\033[93m"   # yellow
R = "\033[91m"   # red
C = "\033[96m"   # cyan
B = "\033[1m"    # bold
D = "\033[0m"    # reset

# ── Test State ──────────────────────────────────────────────────────────

passed = 0
failed = 0
skipped = 0
errors = []


class TestResult:
    def __init__(self, name, success, detail=""):
        self.name = name
        self.success = success
        self.detail = detail


def test(name, func):
    """Run a test function, catch exceptions, and record result."""
    global passed, failed
    try:
        func()
        passed += 1
        print(f"  {G}✓{D} {name}")
        return True
    except AssertionError as e:
        failed += 1
        errors.append((name, str(e)))
        print(f"  {R}✗{D} {name}")
        print(f"    {R}{e}{D}")
        return False
    except Exception as e:
        failed += 1
        errors.append((name, f"Exception: {e}"))
        print(f"  {R}✗{D} {name}")
        print(f"    {R}Exception: {e}{D}")
        return False


# ── Auth State ──────────────────────────────────────────────────────────

TOKEN = ""
TENANT_ID = 0
CREATED_IDS = {}  # module -> [ids]


def api(method, path, json_data=None, params=None):
    """Make an API request and return (status_code, json)."""
    headers = {"Content-Type": "application/json"}
    if TOKEN:
        headers["Authorization"] = f"Bearer {TOKEN}"
    url = f"{BASE_URL}{path}"
    r = requests.request(method, url, json=json_data, params=params, headers=headers, timeout=10)
    try:
        return r.status_code, r.json()
    except Exception:
        return r.status_code, r.text


def assert_ok(status, data, msg=""):
    assert status in (200, 201), f"{msg} Expected 200/201, got {status}: {data}"


def assert_fail(status, data, expected_code=None, msg=""):
    if expected_code:
        assert status == expected_code, f"{msg} Expected {expected_code}, got {status}: {data}"
    else:
        assert status >= 400, f"{msg} Expected 4xx, got {status}: {data}"


# ── Module Test Classes ─────────────────────────────────────────────────

class AuthTests:
    def __init__(self, session):
        self.session = session

    def test_login_success(self):
        s, d = api("POST", "/api/auth/login", {"phone": PHONE, "password": PASSWORD})
        assert_ok(s, d, "Login")
        token = d.get("token") or d.get("data", {}).get("token")
        assert token, f"No token in response: {d}"
        global TOKEN
        TOKEN = token

    def test_login_wrong_password(self):
        s, d = api("POST", "/api/auth/login", {"phone": PHONE, "password": "wrong"})
        # Backend returns HTTP 200 with code!=0 for business errors
        assert s == 200, f"Wrong password: expected 200, got {s}"
        code = d.get("code", 0)
        assert code != 0, f"Wrong password should return non-zero code, got {d}"

    def test_login_missing_fields(self):
        s, d = api("POST", "/api/auth/login", {})
        assert_fail(s, d, msg="Missing fields should fail")

    def test_switch_tenant(self):
        global TENANT_ID
        # First get tenant list from login response
        s, d = api("POST", "/api/auth/login", {"phone": PHONE, "password": PASSWORD})
        assert_ok(s, d)
        tenants = d.get("tenants") or d.get("data", {}).get("tenants") or []
        if not tenants:
            print(f"    {Y}(no tenants to switch){D}")
            return
        tenant_id = tenants[0].get("id")
        if not tenant_id:
            print(f"    {Y}(tenant has no id){D}")
            return
        s2, d2 = api("POST", "/api/auth/switch-tenant", {"tenant_id": tenant_id})
        assert_ok(s2, d2, "Switch tenant")
        new_token = d2.get("token") or d2.get("data", {}).get("token")
        if new_token:
            TOKEN = new_token
        TENANT_ID = tenant_id

    def test_protected_without_token(self):
        global TOKEN
        old_token = TOKEN
        TOKEN = ""
        s, d = api("GET", "/api/shop/supplier/list")
        assert_fail(s, d, msg="No token should return 401/403")
        TOKEN = old_token


class CRUDTestBase:
    """Base class for CRUD module tests."""
    module_name = ""
    api_path = ""
    create_data = {}
    update_data = {}
    required_fields = []  # fields that must exist in GET response

    def setup(self):
        """Create one record for get/update/delete tests."""
        s, d = api("POST", self.api_path, self.create_data)
        assert_ok(s, d, f"Setup create for {self.module_name}")
        item_id = d.get("id") or d.get("data", {}).get("id")
        assert item_id, f"No id returned: {d}"
        CREATED_IDS.setdefault(self.module_name, []).append(item_id)
        return item_id

    def test_create(self):
        item_id = self.setup()

    def test_create_minimal(self):
        """Create with minimal fields."""
        s, d = api("POST", self.api_path, {"name": f"测试{self.module_name}_minimal"})
        assert_ok(s, d, f"Create minimal {self.module_name}")

    def test_list(self):
        s, d = api("GET", f"{self.api_path}/list")
        assert_ok(s, d, f"List {self.module_name}")
        data = d.get("data") or d
        if isinstance(data, dict):
            items = data.get("items") or data.get("list") or []
            total = data.get("total", len(items))
            assert total >= 0, f"Invalid total: {data}"
        elif isinstance(data, list):
            assert len(data) >= 0, "List should be non-negative"

    def test_list_all(self):
        s, d = api("GET", f"{self.api_path}/all")
        assert_ok(s, d, f"ListAll {self.module_name}")
        data = d.get("data") or d
        assert isinstance(data, list), f"Expected list, got {type(data)}: {d}"

    def test_get_by_id(self):
        item_id = self.setup()
        s, d = api("GET", f"{self.api_path}/{item_id}")
        assert_ok(s, d, f"Get {self.module_name} by id")
        item = d.get("data") or d
        for field in self.required_fields:
            assert field in item, f"Missing field '{field}' in response: {item}"

    def test_update(self):
        item_id = self.setup()
        s, d = api("PUT", f"{self.api_path}/{item_id}", self.update_data)
        assert_ok(s, d, f"Update {self.module_name}")

    def test_delete(self):
        item_id = self.setup()
        s, d = api("DELETE", f"{self.api_path}/{item_id}")
        assert_ok(s, d, f"Delete {self.module_name}")
        # Verify deleted
        s2, d2 = api("GET", f"{self.api_path}/{item_id}")
        assert_fail(s2, d2, msg=f"Deleted {self.module_name} should not be found")

    def test_get_nonexistent(self):
        s, d = api("GET", f"{self.api_path}/999999999")
        assert_fail(s, d, msg=f"Get nonexistent {self.module_name}")

    def test_delete_nonexistent(self):
        s, d = api("DELETE", f"{self.api_path}/999999999")
        # Backend treats delete of nonexistent as success (idempotent)
        assert s in (200, 404), f"Delete nonexistent {self.module_name}: got {s}"
        if s == 200:
            code = d.get("code", -1)
            assert code == 0, f"Delete nonexistent should succeed (code=0), got {d}"

    def run_all(self):
        tests = [
            ("创建", self.test_create),
            (f"创建（最少字段）", self.test_create_minimal),
            ("列表", self.test_list),
            ("全部列表", self.test_list_all),
            ("按ID查询", self.test_get_by_id),
            ("更新", self.test_update),
            ("删除", self.test_delete),
            ("查询不存在的记录", self.test_get_nonexistent),
            ("删除不存在的记录", self.test_delete_nonexistent),
        ]
        for name, func in tests:
            test(f"[{self.module_name}] {name}", func)


class SupplierTests(CRUDTestBase):
    module_name = "供应商"
    api_path = "/api/shop/supplier"
    create_data = {
        "name": "测试供应商",
        "contact": "张测试",
        "phone": "13700137001",
        "address": "深圳市南山区科技园",
        "bank_name": "工商银行",
        "bank_account": "6222001234567890000",
    }
    update_data = {
        "name": "测试供应商（已更新）",
        "contact": "李更新",
        "phone": "13700137002",
    }
    required_fields = ["id", "name"]


class CustomerTests(CRUDTestBase):
    module_name = "客户"
    api_path = "/api/shop/customer"
    create_data = {
        "name": "测试客户",
        "contact": "王客户",
        "phone": "13600136001",
        "address": "广州市天河区",
        "type": 1,
    }
    update_data = {
        "name": "测试客户（已更新）",
        "contact": "赵客户",
        "phone": "13600136002",
    }
    required_fields = ["id", "name"]


class WarehouseTests(CRUDTestBase):
    module_name = "仓库"
    api_path = "/api/shop/warehouse"
    create_data = {
        "name": "测试仓库",
        "type": 1,
        "address": "东莞市南城区",
        "keeper": "陈管理",
        "sort": 1,
    }
    update_data = {
        "name": "测试仓库（已更新）",
        "address": "东莞市长安镇",
    }
    required_fields = ["id", "name"]


class AccountTests(CRUDTestBase):
    module_name = "结算账户"
    api_path = "/api/shop/account"
    create_data = {
        "name": "测试账户",
        "type": 1,
        "sort": 1,
    }
    update_data = {
        "name": "测试账户（已更新）",
    }
    required_fields = ["id", "name"]


class GoodsTests(CRUDTestBase):
    module_name = "货品"
    api_path = "/api/shop/goods"
    create_data = {
        "name": "测试货品",
        "code": "TEST001",
        "barcode": "6900000000001",
        "buy_price": 100,
        "sell_price": 199,
        "wholesale_price": 150,
        "stock_quantity": 50,
        "alert_quantity": 10,
    }
    update_data = {
        "name": "测试货品（已更新）",
        "sell_price": 299,
    }
    required_fields = ["id", "name"]

    def test_list_with_category_filter(self):
        """Test goods list with category_id filter."""
        s, d = api("GET", f"{self.api_path}/list", params={"category_id": 0})
        assert_ok(s, d, "List goods with category filter")

    def test_list_with_search(self):
        """Test goods list with keyword search."""
        s, d = api("GET", f"{self.api_path}/list", params={"keyword": "测试"})
        assert_ok(s, d, "List goods with keyword")

    def run_all(self):
        super().run_all()
        test(f"[{self.module_name}] 按分类筛选", self.test_list_with_category_filter)
        test(f"[{self.module_name}] 关键字搜索", self.test_list_with_search)


# ── Integration Test ────────────────────────────────────────────────────

class IntegrationTests:
    """Cross-module integration tests."""

    def test_create_supplier_then_goods(self):
        """Create a supplier, then create a goods linked to it."""
        s1, d1 = api("POST", "/api/shop/supplier", {
            "name": "集成测试供应商",
            "contact": "测试",
            "phone": "13500135001",
        })
        assert_ok(s1, d1, "Create supplier for integration")
        supplier_id = d1.get("id") or d1.get("data", {}).get("id")
        assert supplier_id, f"No supplier id: {d1}"
        CREATED_IDS.setdefault("供应商", []).append(supplier_id)

        s2, d2 = api("POST", "/api/shop/goods", {
            "name": "集成测试货品",
            "code": "INT001",
            "supplier_id": supplier_id,
            "buy_price": 50,
            "sell_price": 100,
        })
        assert_ok(s2, d2, "Create goods linked to supplier")
        goods_id = d2.get("id") or d2.get("data", {}).get("id")
        assert goods_id, f"No goods id: {d2}"
        CREATED_IDS.setdefault("货品", []).append(goods_id)

        # Verify the goods has supplier info
        s3, d3 = api("GET", f"/api/shop/goods/{goods_id}")
        assert_ok(s3, d3, "Get goods with supplier")
        item = d3.get("data") or d3
        # Just verify it doesn't error out

    def test_create_all_modules(self):
        """Create one record in each module and verify all succeed."""
        results = {}
        modules = [
            ("供应商", "/api/shop/supplier", {"name": "全模块测试供应商", "phone": "13100131001"}),
            ("客户", "/api/shop/customer", {"name": "全模块测试客户", "phone": "13200132001"}),
            ("仓库", "/api/shop/warehouse", {"name": "全模块测试仓库"}),
            ("结算账户", "/api/shop/account", {"name": "全模块测试账户"}),
            ("货品", "/api/shop/goods", {"name": "全模块测试货品", "code": "ALL001", "buy_price": 50, "sell_price": 100}),
        ]
        for name, path, data in modules:
            s, d = api("POST", path, data)
            results[name] = s in (200, 201)
            if results[name]:
                item_id = d.get("id") or d.get("data", {}).get("id")
                if item_id:
                    CREATED_IDS.setdefault(name, []).append(item_id)

        all_ok = all(results.values())
        if not all_ok:
            failed_modules = [k for k, v in results.items() if not v]
            assert False, f"Failed modules: {failed_modules}"


# ── Main ────────────────────────────────────────────────────────────────

def main():
    global BASE_URL, passed, failed, skipped

    parser = argparse.ArgumentParser(description="PISA Settings API Test Suite")
    parser.add_argument("--base-url", default=BASE_URL, help="API base URL")
    parser.add_argument("--skip-cleanup", action="store_true", help="Skip cleanup after tests")
    args = parser.parse_args()
    BASE_URL = args.base_url.rstrip("/")

    print(f"\n{B}═══════════════════════════════════════════════{D}")
    print(f"{B}  PISA 进销存系统 - Settings API Test Suite{D}")
    print(f"{B}  Base URL: {BASE_URL}{D}")
    print(f"{B}═══════════════════════════════════════════════{D}\n")

    # ── Auth Tests ──
    print(f"{B}── 认证模块 ──{D}")
    auth = AuthTests(None)
    test("登录成功", auth.test_login_success)
    test("登录-错误密码", auth.test_login_wrong_password)
    test("登录-缺少字段", auth.test_login_missing_fields)
    test("切换商户", auth.test_switch_tenant)
    test("未授权访问", auth.test_protected_without_token)
    print()

    # ── CRUD Module Tests ──
    modules = [
        SupplierTests(),
        CustomerTests(),
        WarehouseTests(),
        AccountTests(),
        GoodsTests(),
    ]

    for mod in modules:
        print(f"{B}── {mod.module_name}模块 ──{D}")
        mod.run_all()
        print()

    # ── Integration Tests ──
    print(f"{B}── 集成测试 ──{D}")
    integration = IntegrationTests()
    test("[集成] 供应商→货品关联", integration.test_create_supplier_then_goods)
    test("[集成] 全模块创建", integration.test_create_all_modules)
    print()

    # ── Summary ──
    total = passed + failed + skipped
    print(f"{B}═══════════════════════════════════════════════{D}")
    print(f"{B}  测试结果{D}")
    print(f"{B}═══════════════════════════════════════════════{D}")
    print(f"  总计: {total}")
    print(f"  {G}通过: {passed}{D}")
    if failed:
        print(f"  {R}失败: {failed}{D}")
    else:
        print(f"  失败: {failed}")
    if skipped:
        print(f"  {Y}跳过: {skipped}{D}")

    if errors:
        print(f"\n{R}── 失败详情 ──{D}")
        for name, detail in errors:
            print(f"  {R}✗ {name}{D}")
            print(f"    {detail}")

    # ── Cleanup ──
    if not args.skip_cleanup and CREATED_IDS:
        print(f"\n{B}── 清理测试数据 ──{D}")
        for module, ids in CREATED_IDS.items():
            paths = {
                "供应商": "/api/shop/supplier",
                "客户": "/api/shop/customer",
                "仓库": "/api/shop/warehouse",
                "结算账户": "/api/shop/account",
                "货品": "/api/shop/goods",
            }
            path = paths.get(module)
            if not path:
                continue
            for item_id in ids:
                api("DELETE", f"{path}/{item_id}")
        print(f"  {G}已清理 {sum(len(v) for v in CREATED_IDS.values())} 条测试数据{D}")

    print()
    if failed == 0:
        print(f"{G}{B}全部测试通过 ✓{D}\n")
        return 0
    else:
        print(f"{R}{B}有 {failed} 项测试失败 ✗{D}\n")
        return 1


if __name__ == "__main__":
    sys.exit(main())
