#!/usr/bin/env python3
"""
PISA 进销存系统 - 货品模块 API 测试套件.

覆盖「货品」相关的全部模块（仅 API 层）：
  - 认证 / 租户上下文
  - 单位管理       (Unit)
  - 货品分类       (GoodsCategory)
  - 规格管理       (GoodsAttribute)
  - 货品属性       (GoodsProperty)
  - 货品管理       (Goods) 含列表/搜索/筛选/分页/多单位/多规格/JSON 字段
  - 图片上传       (Upload)
  - 价格管理       (Price) 含列表展开/筛选/批量改价/内联保存
  - 跨模块集成

Usage:
    python test_settings.py [--base-url http://localhost:8080] [--skip-cleanup]
"""

import argparse
import base64
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

TOKEN = ""
TENANT_ID = 0
SUF = str(int(time.time()))[-6:]
SESSION = requests.Session()

# module -> [ids]，按模块分组，用于收尾清理
CREATED = {
    "goods": [],
    "attribute": [],
    "category": [],
    "unit": [],
    "property": [],
}


def track(module, item_id):
    if item_id:
        CREATED.setdefault(module, []).append(item_id)
    return item_id


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


def skip(name, reason=""):
    global skipped
    skipped += 1
    print(f"  {Y}○{D} {name} {Y}(跳过: {reason}){D}")


# ── HTTP helpers ────────────────────────────────────────────────────────

def api(method, path, json_data=None, params=None):
    headers = {"Content-Type": "application/json"}
    if TOKEN:
        headers["Authorization"] = f"Bearer {TOKEN}"
    url = f"{BASE_URL}{path}"
    r = SESSION.request(method, url, json=json_data, params=params, headers=headers, timeout=15)
    try:
        return r.status_code, r.json()
    except Exception:
        return r.status_code, r.text


def upload(filename, content, content_type="image/png", field="file"):
    headers = {}
    if TOKEN:
        headers["Authorization"] = f"Bearer {TOKEN}"
    files = {field: (filename, content, content_type)}
    r = SESSION.post(f"{BASE_URL}/api/shop/upload", files=files, headers=headers, timeout=30)
    try:
        return r.status_code, r.json()
    except Exception:
        return r.status_code, r.text


def dget(data):
    """Return the `data` payload whether or not the envelope is present."""
    if isinstance(data, dict) and "data" in data:
        return data["data"]
    return data


def assert_ok(status, data, msg=""):
    assert status == 200, f"{msg} 期望 HTTP 200，实际 {status}: {data}"
    assert isinstance(data, dict) and data.get("code") == 0, f"{msg} code!=0: {data}"
    return dget(data)


def assert_status(status, data, expected, msg=""):
    assert status == expected, f"{msg} 期望 HTTP {expected}，实际 {status}: {data}"
    assert isinstance(data, dict) and data.get("code") != 0, f"{msg} 期望业务错误，实际 {data}"
    return data


def assert_bad(status, data, contains=None, msg=""):
    assert_status(status, data, 400, msg)
    if contains:
        text = data.get("message", "") if isinstance(data, dict) else str(data)
        assert contains in text, f"{msg} 期望错误信息包含 '{contains}'，实际 '{text}'"


def create(module, path, payload, msg=""):
    s, d = api("POST", path, payload)
    data = assert_ok(s, d, msg or f"创建 {module}")
    item_id = data.get("id")
    assert item_id, f"{msg} 未返回 id: {data}"
    track(module, item_id)
    return data


# ── 认证 ────────────────────────────────────────────────────────────────

class AuthTests:
    def test_login_success(self):
        global TOKEN, TENANT_ID
        s, d = api("POST", "/api/auth/login", {"phone": PHONE, "password": PASSWORD})
        data = assert_ok(s, d, "登录")
        TOKEN = data.get("token")
        assert TOKEN, f"未返回 token: {data}"
        tenants = data.get("tenants") or []
        if tenants:
            TENANT_ID = tenants[0].get("tenant_id") or tenants[0].get("id") or 0

    def test_login_wrong_password(self):
        s, d = api("POST", "/api/auth/login", {"phone": PHONE, "password": "definitely-wrong"})
        assert s == 200, f"错误密码期望 HTTP 200，实际 {s}"
        assert d.get("code") != 0, f"错误密码应返回 code!=0，实际 {d}"

    def test_login_missing_fields(self):
        s, d = api("POST", "/api/auth/login", {})
        assert_bad(s, d, msg="缺少登录字段")

    def test_unauthorized_no_token(self):
        global TOKEN
        old = TOKEN
        TOKEN = ""
        s, d = api("GET", "/api/shop/goods/list")
        assert s == 401, f"无 token 期望 401，实际 {s}: {d}"
        TOKEN = old

    def test_unauthorized_bad_token(self):
        global TOKEN
        old = TOKEN
        TOKEN = "not-a-real-token"
        s, d = api("GET", "/api/shop/goods/list")
        assert s == 401, f"伪造 token 期望 401，实际 {s}: {d}"
        TOKEN = old

    def run(self):
        test("[认证] 登录成功", self.test_login_success)
        test("[认证] 登录-密码错误", self.test_login_wrong_password)
        test("[认证] 登录-缺少字段", self.test_login_missing_fields)
        test("[认证] 未授权-无 token", self.test_unauthorized_no_token)
        test("[认证] 未授权-非法 token", self.test_unauthorized_bad_token)


# ── 单位管理 ────────────────────────────────────────────────────────────

class UnitTests:
    PATH = "/api/shop/unit"

    def test_create(self):
        create("unit", self.PATH, {"name": f"单位{SUF}"}, "创建单位")

    def test_create_empty(self):
        s, d = api("POST", self.PATH, {"name": ""})
        assert_bad(s, d, "单位名称不能为空", "空名称")

    def test_create_whitespace(self):
        s, d = api("POST", self.PATH, {"name": "   "})
        assert_bad(s, d, "单位名称不能为空", "纯空格名称")

    def test_create_too_long(self):
        s, d = api("POST", self.PATH, {"name": "长" * 21})
        assert_bad(s, d, "不能超过20", "超长名称")

    def test_create_exactly_20(self):
        create("unit", self.PATH, {"name": "单" * 20}, "20 字名称")

    def test_list_all(self):
        u = create("unit", self.PATH, {"name": f"列表单位{SUF}"}, "列表单位")
        s, d = api("GET", f"{self.PATH}/all")
        data = assert_ok(s, d, "单位全部列表")
        assert any(x["id"] == u["id"] for x in data), "新建单位未出现在列表中"

    def test_update(self):
        u = create("unit", self.PATH, {"name": f"待改单位{SUF}"}, "待改单位")
        s, d = api("PUT", f"{self.PATH}/{u['id']}", {"name": f"已改单位{SUF}"})
        assert_ok(s, d, "更新单位")

    def test_update_empty(self):
        u = create("unit", self.PATH, {"name": f"更新空单位{SUF}"}, "更新空单位")
        s, d = api("PUT", f"{self.PATH}/{u['id']}", {"name": ""})
        assert_bad(s, d, "单位名称不能为空", "更新为空")

    def test_update_too_long(self):
        u = create("unit", self.PATH, {"name": f"更新长单位{SUF}"}, "更新长单位")
        s, d = api("PUT", f"{self.PATH}/{u['id']}", {"name": "长" * 21})
        assert_bad(s, d, "不能超过20", "更新超长")

    def test_update_nonexistent(self):
        s, d = api("PUT", f"{self.PATH}/999999999", {"name": "不存在"})
        assert_status(s, d, 404, "更新不存在单位")

    def test_delete(self):
        u = create("unit", self.PATH, {"name": f"删除单位{SUF}"}, "删除单位")
        s, d = api("DELETE", f"{self.PATH}/{u['id']}")
        assert_ok(s, d, "删除单位")
        if u["id"] in CREATED["unit"]:
            CREATED["unit"].remove(u["id"])

    def test_delete_nonexistent(self):
        s, d = api("DELETE", f"{self.PATH}/999999999")
        assert_status(s, d, 404, "删除不存在单位")

    def test_delete_referenced_guard(self):
        u = create("unit", self.PATH, {"name": f"被引用单位{SUF}"}, "被引用单位")
        g = create("goods", "/api/shop/goods",
                   {"name": f"引用单位货品{SUF}", "code": f"U{SUF}", "unit_id": u["id"]}, "引用单位货品")
        s, d = api("DELETE", f"{self.PATH}/{u['id']}")
        assert_bad(s, d, "已被货品使用", "删除被引用单位")
        api("DELETE", f"/api/shop/goods/{g['id']}")
        CREATED["goods"].remove(g["id"])
        s2, d2 = api("DELETE", f"{self.PATH}/{u['id']}")
        assert_ok(s2, d2, "货品删除后应可删除单位")
        CREATED["unit"].remove(u["id"])

    def run(self):
        test("[单位] 新增", self.test_create)
        test("[单位] 新增-空名称", self.test_create_empty)
        test("[单位] 新增-纯空格", self.test_create_whitespace)
        test("[单位] 新增-超长(21)", self.test_create_too_long)
        test("[单位] 新增-恰好20字", self.test_create_exactly_20)
        test("[单位] 列表", self.test_list_all)
        test("[单位] 更新", self.test_update)
        test("[单位] 更新-空名称", self.test_update_empty)
        test("[单位] 更新-超长", self.test_update_too_long)
        test("[单位] 更新-不存在", self.test_update_nonexistent)
        test("[单位] 删除", self.test_delete)
        test("[单位] 删除-不存在", self.test_delete_nonexistent)
        test("[单位] 删除-被货品引用拦截", self.test_delete_referenced_guard)


# ── 货品分类 ────────────────────────────────────────────────────────────

class CategoryTests:
    PATH = "/api/shop/category"

    def test_create_top(self):
        create("category", self.PATH, {"name": f"顶级分类{SUF}"}, "创建顶级分类")

    def test_create_child(self):
        parent = create("category", self.PATH, {"name": f"父分类{SUF}"}, "父分类")
        create("category", self.PATH, {"name": f"子分类{SUF}", "parent_id": parent["id"]}, "子分类")

    def test_create_empty(self):
        s, d = api("POST", self.PATH, {"name": ""})
        assert_bad(s, d, "分类名称不能为空", "空分类名")

    def test_create_too_long(self):
        s, d = api("POST", self.PATH, {"name": "类" * 31})
        assert_bad(s, d, "不能超过30", "超长分类名")

    def test_create_bad_parent(self):
        s, d = api("POST", self.PATH, {"name": f"坏父分类{SUF}", "parent_id": 999999999})
        assert_bad(s, d, "上级类别不存在", "父分类不存在")

    def test_list(self):
        c = create("category", self.PATH, {"name": f"列表分类{SUF}"}, "列表分类")
        s, d = api("GET", f"{self.PATH}/all")
        data = assert_ok(s, d, "分类列表")
        assert any(x["id"] == c["id"] for x in data), "新建分类未出现在列表"

    def test_update(self):
        c = create("category", self.PATH, {"name": f"待改分类{SUF}"}, "待改分类")
        s, d = api("PUT", f"{self.PATH}/{c['id']}", {"name": f"已改分类{SUF}"})
        assert_ok(s, d, "更新分类")

    def test_update_self_parent(self):
        c = create("category", self.PATH, {"name": f"自身父分类{SUF}"}, "自身父分类")
        s, d = api("PUT", f"{self.PATH}/{c['id']}", {"name": c["name"], "parent_id": c["id"]})
        assert_bad(s, d, "不能是自身", "父分类为自身")

    def test_update_bad_parent(self):
        c = create("category", self.PATH, {"name": f"坏父更新分类{SUF}"}, "坏父更新分类")
        s, d = api("PUT", f"{self.PATH}/{c['id']}", {"name": c["name"], "parent_id": 999999999})
        assert_bad(s, d, "上级类别不存在", "更新父分类不存在")

    def test_delete_with_children(self):
        parent = create("category", self.PATH, {"name": f"删父分类{SUF}"}, "删父分类")
        create("category", self.PATH, {"name": f"删子分类{SUF}", "parent_id": parent["id"]}, "删子分类")
        s, d = api("DELETE", f"{self.PATH}/{parent['id']}")
        assert_bad(s, d, "子分类", "删除有子分类的父分类")

    def test_delete_with_goods(self):
        c = create("category", self.PATH, {"name": f"占用分类{SUF}"}, "占用分类")
        g = create("goods", "/api/shop/goods",
                   {"name": f"占用分类货品{SUF}", "code": f"C{SUF}", "category_id": c["id"]}, "占用分类货品")
        s, d = api("DELETE", f"{self.PATH}/{c['id']}")
        assert_bad(s, d, "商品", "删除被货品占用的分类")
        api("DELETE", f"/api/shop/goods/{g['id']}")
        CREATED["goods"].remove(g["id"])

    def test_delete(self):
        c = create("category", self.PATH, {"name": f"删除分类{SUF}"}, "删除分类")
        s, d = api("DELETE", f"{self.PATH}/{c['id']}")
        assert_ok(s, d, "删除分类")
        CREATED["category"].remove(c["id"])

    def test_delete_nonexistent(self):
        s, d = api("DELETE", f"{self.PATH}/999999999")
        assert_status(s, d, 404, "删除不存在分类")

    def run(self):
        test("[分类] 新增-顶级", self.test_create_top)
        test("[分类] 新增-子级", self.test_create_child)
        test("[分类] 新增-空名称", self.test_create_empty)
        test("[分类] 新增-超长(31)", self.test_create_too_long)
        test("[分类] 新增-父级不存在", self.test_create_bad_parent)
        test("[分类] 列表", self.test_list)
        test("[分类] 更新", self.test_update)
        test("[分类] 更新-父级为自身", self.test_update_self_parent)
        test("[分类] 更新-父级不存在", self.test_update_bad_parent)
        test("[分类] 删除-有子分类拦截", self.test_delete_with_children)
        test("[分类] 删除-被货品占用拦截", self.test_delete_with_goods)
        test("[分类] 删除", self.test_delete)
        test("[分类] 删除-不存在", self.test_delete_nonexistent)


# ── 规格管理（辅助属性） ────────────────────────────────────────────────

class AttributeTests:
    PATH = "/api/shop/attribute"

    def test_create_with_values(self):
        create("attribute", self.PATH, {"name": f"规格{SUF}", "values": ["红", "蓝"]}, "创建规格")

    def test_create_no_values(self):
        a = create("attribute", self.PATH, {"name": f"空值规格{SUF}"}, "空值规格")
        assert json.loads(a["values"]) == [], f"空值规格 values 应为 []，实际 {a['values']}"

    def test_create_empty_name(self):
        s, d = api("POST", self.PATH, {"name": "", "values": []})
        assert_bad(s, d, "规格名称不能为空", "空规格名")

    def test_create_too_long(self):
        s, d = api("POST", self.PATH, {"name": "规" * 65, "values": []})
        assert_bad(s, d, "不能超过64", "超长规格名")

    def test_list_all(self):
        a = create("attribute", self.PATH, {"name": f"列表规格{SUF}", "values": ["v1"]}, "列表规格")
        s, d = api("GET", f"{self.PATH}/all")
        data = assert_ok(s, d, "规格全部列表")
        assert any(x["id"] == a["id"] for x in data), "新建规格未出现在列表"

    def test_list_filter_name(self):
        create("attribute", self.PATH, {"name": f"过滤名{SUF}", "values": []}, "过滤名规格")
        s, d = api("GET", f"{self.PATH}/list", params={"name": f"过滤名{SUF}"})
        data = assert_ok(s, d, "按名称过滤规格")
        assert len(data) >= 1, "按名称过滤无结果"

    def test_list_filter_content(self):
        token = f"内容{SUF}"
        create("attribute", self.PATH, {"name": f"过滤内容规格{SUF}", "values": [token]}, "过滤内容规格")
        s, d = api("GET", f"{self.PATH}/list", params={"content": token})
        data = assert_ok(s, d, "按内容过滤规格")
        assert len(data) >= 1, "按内容过滤无结果"

    def test_update(self):
        a = create("attribute", self.PATH, {"name": f"待改规格{SUF}", "values": ["a"]}, "待改规格")
        s, d = api("PUT", f"{self.PATH}/{a['id']}", {"name": f"已改规格{SUF}", "values": ["a", "b"]})
        data = assert_ok(s, d, "更新规格")
        assert json.loads(data["values"]) == ["a", "b"], "规格值未全量替换"

    def test_update_empty_name(self):
        a = create("attribute", self.PATH, {"name": f"更新空规格{SUF}", "values": []}, "更新空规格")
        s, d = api("PUT", f"{self.PATH}/{a['id']}", {"name": "", "values": []})
        assert_bad(s, d, "规格名称不能为空", "更新规格空名")

    def test_add_value(self):
        a = create("attribute", self.PATH, {"name": f"加值规格{SUF}", "values": ["x"]}, "加值规格")
        s, d = api("POST", f"{self.PATH}/{a['id']}/value", {"value": "y"})
        data = assert_ok(s, d, "追加规格值")
        assert "y" in json.loads(data["values"]), "追加的值未保存"

    def test_add_duplicate_value(self):
        a = create("attribute", self.PATH, {"name": f"重复值规格{SUF}", "values": ["dup"]}, "重复值规格")
        s, d = api("POST", f"{self.PATH}/{a['id']}/value", {"value": "dup"})
        data = assert_ok(s, d, "追加重复值")
        assert json.loads(data["values"]).count("dup") == 1, "重复值被写入"

    def test_add_empty_value(self):
        a = create("attribute", self.PATH, {"name": f"空值追加规格{SUF}", "values": []}, "空值追加规格")
        s, d = api("POST", f"{self.PATH}/{a['id']}/value", {"value": "  "})
        assert_bad(s, d, "规格值不能为空", "追加空值")

    def test_add_value_nonexistent(self):
        s, d = api("POST", f"{self.PATH}/999999999/value", {"value": "x"})
        assert_status(s, d, 404, "向不存在规格追加值")

    def test_delete_referenced_guard(self):
        a = create("attribute", self.PATH, {"name": f"被引用规格{SUF}", "values": ["s"]}, "被引用规格")
        spec_groups = json.dumps([{"name": a["name"], "has_image": False,
                                   "values": [{"name": "s", "image": ""}]}], ensure_ascii=False)
        g = create("goods", "/api/shop/goods",
                   {"name": f"引用规格货品{SUF}", "code": f"A{SUF}", "spec_groups": spec_groups}, "引用规格货品")
        s, d = api("DELETE", f"{self.PATH}/{a['id']}")
        assert_bad(s, d, "已被货品使用", "删除被引用规格")
        api("DELETE", f"/api/shop/goods/{g['id']}")
        CREATED["goods"].remove(g["id"])
        s2, d2 = api("DELETE", f"{self.PATH}/{a['id']}")
        assert_ok(s2, d2, "货品删除后应可删除规格")
        CREATED["attribute"].remove(a["id"])

    def test_delete(self):
        a = create("attribute", self.PATH, {"name": f"删除规格{SUF}", "values": []}, "删除规格")
        s, d = api("DELETE", f"{self.PATH}/{a['id']}")
        assert_ok(s, d, "删除规格")
        CREATED["attribute"].remove(a["id"])

    def test_delete_nonexistent(self):
        s, d = api("DELETE", f"{self.PATH}/999999999")
        assert_status(s, d, 404, "删除不存在规格")

    def run(self):
        test("[规格] 新增-带值", self.test_create_with_values)
        test("[规格] 新增-无值", self.test_create_no_values)
        test("[规格] 新增-空名称", self.test_create_empty_name)
        test("[规格] 新增-超长(65)", self.test_create_too_long)
        test("[规格] 列表-全部", self.test_list_all)
        test("[规格] 列表-按名称过滤", self.test_list_filter_name)
        test("[规格] 列表-按内容过滤", self.test_list_filter_content)
        test("[规格] 更新-全量替换", self.test_update)
        test("[规格] 更新-空名称", self.test_update_empty_name)
        test("[规格] 追加规格值", self.test_add_value)
        test("[规格] 追加-重复值去重", self.test_add_duplicate_value)
        test("[规格] 追加-空值", self.test_add_empty_value)
        test("[规格] 追加-规格不存在", self.test_add_value_nonexistent)
        test("[规格] 删除-被货品引用拦截", self.test_delete_referenced_guard)
        test("[规格] 删除", self.test_delete)
        test("[规格] 删除-不存在", self.test_delete_nonexistent)


# ── 货品属性（品牌/产地/材质等） ────────────────────────────────────────

class PropertyTests:
    PATH = "/api/shop/property"

    def _existing_count(self):
        s, d = api("GET", f"{self.PATH}/all")
        return len(assert_ok(s, d, "货品属性列表"))

    def _new(self, name, ptype=1, values=None):
        payload = {"name": name, "type": ptype}
        if values is not None:
            payload["values"] = values
        return create("property", self.PATH, payload, "货品属性")

    def _drop(self, pid):
        api("DELETE", f"{self.PATH}/{pid}")
        if pid in CREATED.get("property", []):
            CREATED["property"].remove(pid)

    def test_create_select(self):
        p = self._new(f"属性{SUF}", 1, ["a", "b"])
        try:
            assert p["type"] == 1, f"类型应为 1，实际 {p['type']}"
        finally:
            self._drop(p["id"])

    def test_create_input(self):
        p = self._new(f"输入属性{SUF}", 2)
        try:
            assert p["type"] == 2, f"类型应为 2，实际 {p['type']}"
        finally:
            self._drop(p["id"])

    def test_create_empty_name(self):
        s, d = api("POST", self.PATH, {"name": "", "type": 1})
        assert_bad(s, d, "属性名称不能为空", "空属性名")

    def test_create_too_long(self):
        s, d = api("POST", self.PATH, {"name": "属" * 21, "type": 1})
        assert_bad(s, d, "不能超过20", "超长属性名")

    def test_create_invalid_type(self):
        p = self._new(f"非法类型属性{SUF}", 99)
        try:
            assert p["type"] == 1, f"非法类型应归一为 1，实际 {p['type']}"
        finally:
            self._drop(p["id"])

    def test_list(self):
        p = self._new(f"列表属性{SUF}", 1)
        try:
            s, d = api("GET", f"{self.PATH}/all")
            data = assert_ok(s, d, "货品属性列表")
            assert any(x["id"] == p["id"] for x in data), "新建属性未出现在列表"
        finally:
            self._drop(p["id"])

    def test_update(self):
        p = self._new(f"待改属性{SUF}", 1, ["x"])
        try:
            s, d = api("PUT", f"{self.PATH}/{p['id']}",
                       {"name": f"已改属性{SUF}", "type": 1, "values": ["x", "y"]})
            data = assert_ok(s, d, "更新属性")
            assert json.loads(data["values"]) == ["x", "y"], "属性值未更新"
        finally:
            self._drop(p["id"])

    def test_update_too_long(self):
        p = self._new(f"更新长属性{SUF}", 1)
        try:
            s, d = api("PUT", f"{self.PATH}/{p['id']}", {"name": "属" * 21, "type": 1})
            assert_bad(s, d, "不能超过20", "更新超长属性名")
        finally:
            self._drop(p["id"])

    def test_max_six(self):
        existing = self._existing_count()
        if existing >= 6:
            skip("[属性] 最多6项限制", f"当前已有 {existing} 项，跳过")
            return
        mine = []
        try:
            for i in range(6 - existing):
                p = self._new(f"限额属性{SUF}_{i}", 1)
                mine.append(p["id"])
            s, d = api("POST", self.PATH, {"name": f"超额属性{SUF}", "type": 1})
            assert_bad(s, d, "最多可添加6项属性", "第7项属性")
        finally:
            for pid in mine:
                self._drop(pid)

    def test_delete(self):
        p = self._new(f"删除属性{SUF}", 1)
        s, d = api("DELETE", f"{self.PATH}/{p['id']}")
        assert_ok(s, d, "删除属性")
        if p["id"] in CREATED.get("property", []):
            CREATED["property"].remove(p["id"])

    def test_delete_nonexistent(self):
        s, d = api("DELETE", f"{self.PATH}/999999999")
        assert s in (200, 404), f"删除不存在属性期望 200/404，实际 {s}"

    def run(self):
        test("[属性] 新增-选择型", self.test_create_select)
        test("[属性] 新增-输入型", self.test_create_input)
        test("[属性] 新增-空名称", self.test_create_empty_name)
        test("[属性] 新增-超长(21)", self.test_create_too_long)
        test("[属性] 新增-非法类型归一", self.test_create_invalid_type)
        test("[属性] 列表", self.test_list)
        test("[属性] 更新", self.test_update)
        test("[属性] 更新-超长", self.test_update_too_long)
        test("[属性] 最多6项限制", self.test_max_six)
        test("[属性] 删除", self.test_delete)
        test("[属性] 删除-不存在", self.test_delete_nonexistent)


# ── 货品管理 ────────────────────────────────────────────────────────────

class GoodsTests:
    PATH = "/api/shop/goods"

    def test_create_minimal(self):
        g = create("goods", self.PATH, {"name": f"最简货品{SUF}"}, "最简货品")
        assert g["status"] == 1, f"新建货品 status 应为 1，实际 {g['status']}"

    def test_create_empty_name(self):
        s, d = api("POST", self.PATH, {"name": "   "})
        assert_bad(s, d, "货品名称不能为空", "空货品名")

    def test_create_brand_too_long(self):
        s, d = api("POST", self.PATH, {"name": f"长品牌货品{SUF}", "brand": "牌" * 101})
        assert_bad(s, d, "不能超过100", "超长品牌")

    def test_create_full(self):
        create("goods", self.PATH, {
            "name": f"完整货品{SUF}",
            "code": f"FULL{SUF}",
            "barcode": f"69{SUF}0000001",
            "brand": f"品牌{SUF}",
            "origin": f"产地{SUF}",
            "remark": "备注内容",
            "spec": "规格描述",
            "purchase_price": 12.5,
            "retail_price": 20.0,
            "wholesale_price": 16.0,
            "has_multi_unit": 0,
            "enable_stock_alert": 1,
            "min_stock": 5,
            "max_stock": 50,
            "init_cost": 12.5,
        }, "完整货品")

    def test_get_by_id(self):
        g = create("goods", self.PATH, {"name": f"查询货品{SUF}", "code": f"GET{SUF}"}, "查询货品")
        s, d = api("GET", f"{self.PATH}/{g['id']}")
        data = assert_ok(s, d, "按 ID 查询货品")
        for field in ("id", "name", "code", "status", "units", "prices", "stocks"):
            assert field in data, f"货品详情缺少字段 '{field}': {data}"
        assert isinstance(data["units"], list), "units 应为数组"
        assert isinstance(data["prices"], list), "prices 应为数组"
        assert isinstance(data["stocks"], list), "stocks 应为数组"

    def test_get_nonexistent(self):
        s, d = api("GET", f"{self.PATH}/999999999")
        assert_status(s, d, 404, "查询不存在货品")

    def test_list_structure(self):
        s, d = api("GET", f"{self.PATH}/list", params={"page": 1, "page_size": 5})
        data = assert_ok(s, d, "货品列表")
        for field in ("list", "total", "page", "page_size"):
            assert field in data, f"列表缺少字段 '{field}': {data}"

    def test_search_by_name(self):
        g = create("goods", self.PATH, {"name": f"搜索名称货品{SUF}", "code": f"SN{SUF}"}, "搜索名称货品")
        s, d = api("GET", f"{self.PATH}/list", params={"keyword": g["name"]})
        data = assert_ok(s, d, "按名称搜索")
        assert any(x["id"] == g["id"] for x in data["list"]), "按名称未搜到"

    def test_search_by_code(self):
        g = create("goods", self.PATH, {"name": f"搜索编码货品{SUF}", "code": f"SC{SUF}"}, "搜索编码货品")
        s, d = api("GET", f"{self.PATH}/list", params={"keyword": f"SC{SUF}"})
        data = assert_ok(s, d, "按编码搜索")
        assert any(x["id"] == g["id"] for x in data["list"]), "按编码未搜到"

    def test_search_by_barcode(self):
        bc = f"99{SUF}12345"
        g = create("goods", self.PATH, {"name": f"搜索条码货品{SUF}", "code": f"SB{SUF}", "barcode": bc}, "搜索条码货品")
        s, d = api("GET", f"{self.PATH}/list", params={"keyword": bc})
        data = assert_ok(s, d, "按条码搜索")
        assert any(x["id"] == g["id"] for x in data["list"]), "按条码未搜到"

    def test_category_filter_includes_children(self):
        parent = create("category", self.PATH.replace("/goods", "/category"), {"name": f"筛选父类{SUF}"}, "筛选父类")
        child = create("category", self.PATH.replace("/goods", "/category"),
                       {"name": f"筛选子类{SUF}", "parent_id": parent["id"]}, "筛选子类")
        g = create("goods", self.PATH, {"name": f"子类货品{SUF}", "code": f"CF{SUF}", "category_id": child["id"]}, "子类货品")
        s, d = api("GET", f"{self.PATH}/list", params={"category_id": parent["id"], "page_size": 100})
        data = assert_ok(s, d, "按父分类筛选")
        assert any(x["id"] == g["id"] for x in data["list"]), "父分类筛选未包含子类商品"

    def test_uncategorized_filter(self):
        g = create("goods", self.PATH, {"name": f"未分类货品{SUF}", "code": f"UN{SUF}"}, "未分类货品")
        s, d = api("GET", f"{self.PATH}/list", params={"category_id": -1, "page_size": 100})
        data = assert_ok(s, d, "未分类筛选")
        assert any(x["id"] == g["id"] for x in data["list"]), "未分类筛选未包含无分类商品"

    def test_pagination(self):
        for i in range(3):
            create("goods", self.PATH, {"name": f"分页货品{SUF}_{i}", "code": f"PG{SUF}_{i}"}, "分页货品")
        s, d = api("GET", f"{self.PATH}/list", params={"page": 1, "page_size": 2})
        data = assert_ok(s, d, "分页查询")
        assert len(data["list"]) <= 2, f"page_size=2 返回超过 2 条: {len(data['list'])}"
        assert data["total"] >= 3, "分页 total 异常"

    def test_update(self):
        g = create("goods", self.PATH, {"name": f"待改货品{SUF}", "code": f"UP{SUF}"}, "待改货品")
        s, d = api("PUT", f"{self.PATH}/{g['id']}", {"name": f"已改货品{SUF}"})
        data = assert_ok(s, d, "更新货品")
        assert data["name"] == f"已改货品{SUF}", "货品名称未更新"

    def test_update_preserves_status(self):
        g = create("goods", self.PATH, {"name": f"状态货品{SUF}", "code": f"ST{SUF}"}, "状态货品")
        s, d = api("PUT", f"{self.PATH}/{g['id']}", {"name": f"状态货品改{SUF}"})
        data = assert_ok(s, d, "更新后状态")
        assert data["status"] == 1, f"更新后 status 应保持 1，实际 {data['status']}"

    def test_update_empty_name(self):
        g = create("goods", self.PATH, {"name": f"更新空货品{SUF}", "code": f"UE{SUF}"}, "更新空货品")
        s, d = api("PUT", f"{self.PATH}/{g['id']}", {"name": ""})
        assert_bad(s, d, "货品名称不能为空", "更新货品空名")

    def test_update_nonexistent(self):
        s, d = api("PUT", f"{self.PATH}/999999999", {"name": "不存在"})
        assert_status(s, d, 404, "更新不存在货品")

    def test_delete(self):
        g = create("goods", self.PATH, {"name": f"删除货品{SUF}", "code": f"DL{SUF}"}, "删除货品")
        s, d = api("DELETE", f"{self.PATH}/{g['id']}")
        assert_ok(s, d, "删除货品")
        CREATED["goods"].remove(g["id"])
        s2, d2 = api("GET", f"{self.PATH}/{g['id']}")
        assert_status(s2, d2, 404, "删除后查询")

    def test_delete_nonexistent(self):
        s, d = api("DELETE", f"{self.PATH}/999999999")
        assert s in (200, 404), f"删除不存在货品期望 200/404，实际 {s}"

    def test_next_code(self):
        s, d = api("GET", f"{self.PATH}/next-code")
        data = assert_ok(s, d, "生成编号")
        code = data.get("code", "")
        assert code.startswith("SP"), f"编号应以 SP 开头: {code}"
        assert len(code) >= 14, f"编号长度异常: {code}"

    def test_brands(self):
        brand = f"品牌列表{SUF}"
        create("goods", self.PATH, {"name": f"品牌货品{SUF}", "code": f"BR{SUF}", "brand": brand}, "品牌货品")
        s, d = api("GET", f"{self.PATH}/brands")
        data = assert_ok(s, d, "品牌列表")
        assert brand in data, f"品牌 {brand} 未出现在品牌列表"

    def test_origins(self):
        origin = f"产地列表{SUF}"
        create("goods", self.PATH, {"name": f"产地货品{SUF}", "code": f"OR{SUF}", "origin": origin}, "产地货品")
        s, d = api("GET", f"{self.PATH}/origins")
        data = assert_ok(s, d, "产地列表")
        assert origin in data, f"产地 {origin} 未出现在产地列表"

    def test_multi_unit(self):
        u1 = create("unit", "/api/shop/unit", {"name": f"主单位{SUF}"}, "主单位")
        u2 = create("unit", "/api/shop/unit", {"name": f"辅单位{SUF}"}, "辅单位")
        g = create("goods", self.PATH, {
            "name": f"多单位货品{SUF}", "code": f"MU{SUF}",
            "unit_id": u1["id"], "has_multi_unit": 1,
            "units": [
                {"unit_id": u2["id"], "unit_name": u2["name"], "factor": 12, "is_main": 0},
            ],
        }, "多单位货品")
        s, d = api("GET", f"{self.PATH}/{g['id']}")
        data = assert_ok(s, d, "多单位货品详情")
        assert len(data["units"]) == 2, f"应含主单位+辅单位共 2 个，实际 {len(data['units'])}"
        aux = [u for u in data["units"] if not u["is_main"]]
        assert len(aux) == 1, f"应有 1 个辅单位，实际 {len(aux)}"
        assert float(aux[0]["factor"]) == 12, "辅单位换算系数错误"
        assert data["main_unit"] == u1["name"], f"主单位聚合错误: {data['main_unit']}"

    def test_multi_spec(self):
        u1 = create("unit", "/api/shop/unit", {"name": f"规格主单位{SUF}"}, "规格主单位")
        spec_groups = [{"name": "颜色", "has_image": False,
                        "values": [{"name": "米色", "image": ""}, {"name": "黑色", "image": ""}]}]
        price_rows = {u1["name"]: {
            "米色": {"purchase_price": "10", "retail_price": "20", "wholesale_price": "15"},
            "黑色": {"purchase_price": "10", "retail_price": "25", "wholesale_price": "15"},
        }}
        stock_rows = {"米色": {"stock": "5", "init_cost": "10"},
                      "黑色": {"stock": "3", "init_cost": "10"}}
        g = create("goods", self.PATH, {
            "name": f"多规格货品{SUF}", "code": f"MS{SUF}", "has_multi_spec": 1,
            "unit_id": u1["id"],
            "spec_groups": json.dumps(spec_groups, ensure_ascii=False),
            "price_rows": json.dumps(price_rows, ensure_ascii=False),
            "stock_rows": json.dumps(stock_rows, ensure_ascii=False),
        }, "多规格货品")
        s, d = api("GET", f"{self.PATH}/{g['id']}")
        data = assert_ok(s, d, "多规格货品详情")
        assert len(data["prices"]) == 2, f"应有 2 条价格，实际 {len(data['prices'])}"
        assert len(data["stocks"]) == 2, f"应有 2 条库存，实际 {len(data['stocks'])}"
        assert data["retail_min"] == 20 and data["retail_max"] == 25, \
            f"零售价范围错误: {data['retail_min']}~{data['retail_max']}"
        assert data["total_stock"] == 8, f"总库存应为 8，实际 {data['total_stock']}"

    def test_json_fields_roundtrip(self):
        u1 = create("unit", "/api/shop/unit", {"name": f"JSON主单位{SUF}"}, "JSON主单位")
        col = f"等级{SUF}"
        price_rows = {u1["name"]: {"": {"purchase_price": "10", "retail_price": "20",
                                        "wholesale_price": "15", "custom": {col: "5"}}}}
        price_columns = [col, "等级B"]
        stock_rows = {"": {"stock": "3", "init_cost": "10"}}
        spec_groups = [{"name": "颜色", "has_image": False, "values": [{"name": "红", "image": ""}]}]
        images = ["/uploads/x1.png", "/uploads/x2.png"]
        g = create("goods", self.PATH, {
            "name": f"JSON货品{SUF}", "code": f"JS{SUF}", "unit_id": u1["id"],
            "price_rows": json.dumps(price_rows, ensure_ascii=False),
            "price_columns": json.dumps(price_columns, ensure_ascii=False),
            "stock_rows": json.dumps(stock_rows, ensure_ascii=False),
            "spec_groups": json.dumps(spec_groups, ensure_ascii=False),
            "images": json.dumps(images, ensure_ascii=False),
            "suppliers": json.dumps([], ensure_ascii=False),
        }, "JSON货品")
        s, d = api("GET", f"{self.PATH}/{g['id']}")
        data = assert_ok(s, d, "JSON货品详情")
        assert json.loads(data["price_columns"]) == price_columns, "price_columns 往返不一致"
        assert json.loads(data["spec_groups"]) == spec_groups, "spec_groups 往返不一致"
        assert json.loads(data["images"]) == images, "images 往返不一致"
        # 价格 / 库存落到子表
        assert len(data["prices"]) == 1, f"应有 1 条价格，实际 {len(data['prices'])}"
        p = data["prices"][0]
        assert p["unit_key"] == u1["name"], f"unit_key 错误: {p}"
        assert p["purchase_price"] == 10 and p["retail_price"] == 20, f"价格错误: {p}"
        assert json.loads(p["custom"]).get(col) == 5, f"自定义价错误: {p['custom']}"
        assert len(data["stocks"]) == 1, f"应有 1 条库存，实际 {len(data['stocks'])}"
        assert data["stocks"][0]["stock"] == 3, f"库存错误: {data['stocks'][0]}"

    def test_list_units_summary(self):
        u1 = create("unit", "/api/shop/unit", {"name": f"摘要主单位{SUF}"}, "摘要主单位")
        u2 = create("unit", "/api/shop/unit", {"name": f"摘要辅单位{SUF}"}, "摘要辅单位")
        price_rows = {
            u1["name"]: {"": {"code": f"SUM{SUF}-1", "purchase_price": "10",
                              "retail_price": "20", "wholesale_price": "15"}},
            u2["name"]: {"": {"code": f"SUM{SUF}-2", "purchase_price": "100",
                              "retail_price": "200", "wholesale_price": "150"}},
        }
        g = create("goods", self.PATH, {
            "name": f"摘要货品{SUF}", "code": f"SUM{SUF}",
            "unit_id": u1["id"], "has_multi_unit": 1,
            "units": [{"unit_id": u2["id"], "unit_name": u2["name"], "factor": 10, "is_main": 0}],
            "price_rows": json.dumps(price_rows, ensure_ascii=False),
        }, "摘要货品")
        s, d = api("GET", f"{self.PATH}/list", params={"keyword": f"SUM{SUF}", "page_size": 100})
        data = assert_ok(s, d, "货品列表摘要")
        item = next((x for x in data["list"] if x["id"] == g["id"]), None)
        assert item is not None, "未找到摘要货品"
        summaries = item.get("units_summary") or []
        assert len(summaries) == 2, f"应有 2 个单位摘要，实际 {len(summaries)}"
        by_name = {u["unit_name"]: u for u in summaries}
        assert by_name[u1["name"]]["retail_min"] == 20, "主单位零售价摘要错误"
        assert by_name[u2["name"]]["retail_min"] == 200, "辅单位零售价摘要错误"
        assert f"SUM{SUF}-1" in by_name[u1["name"]]["codes"], "主单位编号摘要错误"
        assert item["main_unit"] == u1["name"], "列表主单位错误"

    def test_plan_limit_route_present(self):
        # MaxGoods 限额中间件已挂载；当前套餐额度较大，仅验证创建仍可用。
        g = create("goods", self.PATH, {"name": f"限额探测货品{SUF}", "code": f"PL{SUF}"}, "限额探测货品")
        assert g.get("id"), "限额探测创建失败"

    def run(self):
        test("[货品] 新增-最简", self.test_create_minimal)
        test("[货品] 新增-空名称", self.test_create_empty_name)
        test("[货品] 新增-品牌超长", self.test_create_brand_too_long)
        test("[货品] 新增-完整字段", self.test_create_full)
        test("[货品] 按ID查询(含units/prices/stocks)", self.test_get_by_id)
        test("[货品] 查询-不存在", self.test_get_nonexistent)
        test("[货品] 列表结构", self.test_list_structure)
        test("[货品] 搜索-名称", self.test_search_by_name)
        test("[货品] 搜索-编码", self.test_search_by_code)
        test("[货品] 搜索-条码", self.test_search_by_barcode)
        test("[货品] 筛选-父分类含子类", self.test_category_filter_includes_children)
        test("[货品] 筛选-未分类", self.test_uncategorized_filter)
        test("[货品] 分页", self.test_pagination)
        test("[货品] 更新", self.test_update)
        test("[货品] 更新-保留status", self.test_update_preserves_status)
        test("[货品] 更新-空名称", self.test_update_empty_name)
        test("[货品] 更新-不存在", self.test_update_nonexistent)
        test("[货品] 删除", self.test_delete)
        test("[货品] 删除-不存在", self.test_delete_nonexistent)
        test("[货品] 生成编号", self.test_next_code)
        test("[货品] 品牌列表", self.test_brands)
        test("[货品] 产地列表", self.test_origins)
        test("[货品] 多单位", self.test_multi_unit)
        test("[货品] 多规格", self.test_multi_spec)
        test("[货品] JSON字段往返", self.test_json_fields_roundtrip)
        test("[货品] 列表单位摘要", self.test_list_units_summary)
        test("[货品] 套餐限额路由", self.test_plan_limit_route_present)


# ── 图片上传 ────────────────────────────────────────────────────────────

PNG_1PX = base64.b64decode(
    "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mNkYPhfDwAChwGA60e6kgAAAABJRU5ErkJggg=="
)


class UploadTests:
    def test_upload_png(self):
        s, d = upload(f"test{SUF}.png", PNG_1PX)
        data = assert_ok(s, d, "上传 PNG")
        assert data.get("url", "").startswith("/uploads/"), f"返回 URL 异常: {data}"

    def test_upload_bad_ext(self):
        s, d = upload(f"test{SUF}.txt", b"hello", content_type="text/plain")
        assert_bad(s, d, "仅支持", "非法扩展名")

    def test_upload_missing_file(self):
        headers = {}
        if TOKEN:
            headers["Authorization"] = f"Bearer {TOKEN}"
        r = SESSION.post(f"{BASE_URL}/api/shop/upload", files={}, headers=headers, timeout=10)
        assert_bad(r.status_code, r.json(), "未收到文件", "缺少文件")

    def test_upload_oversize(self):
        big = b"\x89PNG\r\n\x1a\n" + b"0" * (5 * 1024 * 1024 + 100)
        s, d = upload(f"big{SUF}.png", big)
        assert_bad(s, d, "5MB", "超过 5MB")

    def run(self):
        test("[上传] 合法 PNG", self.test_upload_png)
        test("[上传] 非法扩展名", self.test_upload_bad_ext)
        test("[上传] 缺少文件", self.test_upload_missing_file)
        test("[上传] 超过5MB", self.test_upload_oversize)


# ── 价格管理 ────────────────────────────────────────────────────────────

class PriceTests:
    LIST = "/api/shop/price/list"
    BATCH = "/api/shop/price/batch"
    ROWS = "/api/shop/price/rows"
    _unit = None

    def _main_unit(self):
        if PriceTests._unit is None:
            PriceTests._unit = create("unit", "/api/shop/unit",
                                      {"name": f"价格主单位{SUF}"}, "价格主单位")
        return PriceTests._unit

    def _make_simple(self, retail=100.0, wholesale=200.0, purchase=10.0, custom=50.0):
        u = self._main_unit()
        col = f"等级{SUF}"
        price_rows = {u["name"]: {"": {
            "purchase_price": str(purchase), "retail_price": str(retail),
            "wholesale_price": str(wholesale), "custom": {col: str(custom)}}}}
        g = create("goods", "/api/shop/goods", {
            "name": f"价格货品{SUF}",
            "code": f"PR{SUF}",
            "unit_id": u["id"],
            "purchase_price": purchase, "retail_price": retail, "wholesale_price": wholesale,
            "price_columns": json.dumps([col], ensure_ascii=False),
            "price_rows": json.dumps(price_rows, ensure_ascii=False),
        }, "价格货品")
        return g, col

    def _find_row(self, goods_id, unit_key=None, spec_key=None):
        s, d = api("GET", self.LIST, params={"page_size": 500})
        data = assert_ok(s, d, "价格列表")
        for r in data["list"]:
            if r["goods_id"] != goods_id:
                continue
            if unit_key is not None and r["unit_key"] != unit_key:
                continue
            if spec_key is not None and r["spec_key"] != spec_key:
                continue
            return r
        raise AssertionError(f"未找到价格行 goods={goods_id} unit={unit_key} spec={spec_key}")

    def test_list_structure(self):
        s, d = api("GET", self.LIST)
        data = assert_ok(s, d, "价格列表")
        for field in ("list", "total", "columns", "page", "page_size"):
            assert field in data, f"价格列表缺少字段 '{field}'"

    def test_simple_row_key(self):
        g, col = self._make_simple()
        row = self._find_row(g["id"])
        assert row["purchase_price"] == 10, f"进货价错误: {row}"
        assert row["retail_price"] == 100, f"零售价错误: {row}"
        assert row["custom"].get(col) == 50, f"自定义价格错误: {row}"

    def test_columns_union(self):
        g, col = self._make_simple()
        s, d = api("GET", self.LIST, params={"keyword": f"PR{SUF}", "page_size": 100})
        data = assert_ok(s, d, "价格列表列并集")
        assert col in data["columns"], f"自定义列 {col} 未出现在 columns: {data['columns']}"

    def test_multi_unit_expansion(self):
        u1 = create("unit", "/api/shop/unit", {"name": f"价主单位2{SUF}"}, "价主单位2")
        u2 = create("unit", "/api/shop/unit", {"name": f"价辅单位2{SUF}"}, "价辅单位2")
        col = f"价等级{SUF}"
        price_rows = {
            u1["name"]: {"规格A": {"code": f"PU{SUF}-1", "purchase_price": "10", "retail_price": "20",
                                   "wholesale_price": "15", "custom": {col: "1"}}},
            u2["name"]: {"规格A": {"code": f"PU{SUF}-2", "purchase_price": "100", "retail_price": "200",
                                   "wholesale_price": "150", "custom": {col: "2"}}},
        }
        g = create("goods", "/api/shop/goods", {
            "name": f"多单位价格货品{SUF}", "code": f"PU{SUF}",
            "unit_id": u1["id"], "has_multi_unit": 1,
            "units": [{"unit_id": u2["id"], "unit_name": u2["name"], "factor": 10, "is_main": 0}],
            "price_columns": json.dumps([col], ensure_ascii=False),
            "price_rows": json.dumps(price_rows, ensure_ascii=False),
        }, "多单位价格货品")
        s, d = api("GET", self.LIST, params={"keyword": f"PU{SUF}", "page_size": 100})
        data = assert_ok(s, d, "多单位价格列表")
        keys = {r["key"] for r in data["list"]}
        assert f"{g['id']}|{u1['name']}|规格A" in keys, f"缺少主单位行: {keys}"
        assert f"{g['id']}|{u2['name']}|规格A" in keys, f"缺少辅单位行: {keys}"

    def test_filter_keyword(self):
        g, col = self._make_simple()
        s, d = api("GET", self.LIST, params={"keyword": f"PR{SUF}", "page_size": 100})
        data = assert_ok(s, d, "价格关键字筛选")
        assert any(r["goods_id"] == g["id"] for r in data["list"]), "关键字筛选无结果"

    def test_filter_spec(self):
        g, col = self._make_simple()
        s, d = api("GET", self.LIST, params={"spec": "不存在的规格xyz", "page_size": 100})
        data = assert_ok(s, d, "价格规格筛选")
        assert all("不存在的规格xyz" in (r["spec"] or "") for r in data["list"]), "规格筛选未生效"

    def test_filter_brand(self):
        brand = f"价格品牌{SUF}"
        u = self._main_unit()
        col = f"等级{SUF}"
        price_rows = {u["name"]: {"": {"purchase_price": "1", "retail_price": "2",
                                       "wholesale_price": "3", "custom": {col: "0"}}}}
        g = create("goods", "/api/shop/goods", {
            "name": f"品牌价格货品{SUF}", "code": f"PB{SUF}", "brand": brand,
            "unit_id": u["id"],
            "price_columns": json.dumps([col], ensure_ascii=False),
            "price_rows": json.dumps(price_rows, ensure_ascii=False),
        }, "品牌价格货品")
        s, d = api("GET", self.LIST, params={"brand": brand, "page_size": 100})
        data = assert_ok(s, d, "价格品牌筛选")
        assert data["list"] and all(r["goods_id"] == g["id"] for r in data["list"]), "品牌筛选未生效"

    def test_filter_hide_disabled(self):
        s, d = api("GET", self.LIST, params={"hide_disabled": "1", "page_size": 100})
        assert_ok(s, d, "隐藏停用筛选")

    def test_page_size_clamp(self):
        s, d = api("GET", self.LIST, params={"page_size": 1000})
        data = assert_ok(s, d, "page_size 上限")
        assert data["page_size"] == 50, f"page_size 应被限制为 50，实际 {data['page_size']}"

    def test_batch_selected_up(self):
        g, col = self._make_simple(retail=100.0)
        s, d = api("PUT", self.BATCH, {
            "scope": "selected", "keys": [f"{g['id']}|{self._main_unit()['name']}|"],
            "changes": [{"field": "retail_price", "mode": "up", "percent": 10}],
        })
        data = assert_ok(s, d, "批量改价-选中-上调")
        assert data.get("count") == 1, f"应修改 1 个货品，实际 {data}"
        row = self._find_row(g["id"])
        assert row["retail_price"] == 110, f"零售价应为 110，实际 {row['retail_price']}"

    def test_batch_custom_down(self):
        g, col = self._make_simple(custom=50.0)
        s, d = api("PUT", self.BATCH, {
            "scope": "selected", "keys": [f"{g['id']}|{self._main_unit()['name']}|"],
            "changes": [{"field": f"custom:{col}", "mode": "down", "percent": 50}],
        })
        assert_ok(s, d, "批量改价-自定义-下调")
        row = self._find_row(g["id"])
        assert row["custom"].get(col) == 25, f"自定义价格应为 25，实际 {row['custom']}"

    def test_batch_all_with_filter(self):
        g, col = self._make_simple(retail=200.0)
        s, d = api("PUT", self.BATCH, {
            "scope": "all", "keys": [],
            "filters": {"keyword": f"PR{SUF}"},
            "changes": [{"field": "retail_price", "mode": "up", "percent": 50}],
        })
        data = assert_ok(s, d, "批量改价-全部")
        assert data.get("count") >= 1, f"应至少修改 1 个货品，实际 {data}"
        row = self._find_row(g["id"])
        assert row["retail_price"] == 300, f"零售价应为 300，实际 {row['retail_price']}"

    def test_batch_empty_changes(self):
        s, d = api("PUT", self.BATCH, {"scope": "all", "changes": []})
        assert_bad(s, d, "请填写改价内容", "空改价内容")

    def test_batch_nonexistent_keys(self):
        s, d = api("PUT", self.BATCH, {
            "scope": "selected", "keys": ["999999999|__simple__|"],
            "changes": [{"field": "retail_price", "mode": "up", "percent": 10}],
        })
        data = assert_ok(s, d, "批量改价-不存在")
        assert data.get("count") == 0, f"不存在货品应修改 0 个，实际 {data}"

    def test_rows_save_simple(self):
        g, col = self._make_simple(retail=100.0)
        s, d = api("PUT", self.ROWS, {"rows": [{
            "goods_id": g["id"], "unit_key": self._main_unit()["name"], "spec_key": "",
            "purchase_price": 11, "retail_price": 250, "wholesale_price": 220,
            "custom": {col: 7},
        }]})
        assert_ok(s, d, "内联保存价格")
        row = self._find_row(g["id"])
        assert row["retail_price"] == 250, f"零售价应为 250，实际 {row['retail_price']}"
        assert row["custom"].get(col) == 7, f"自定义价格应为 7，实际 {row['custom']}"
        # 主表价格同步
        sg, sd = api("GET", f"/api/shop/goods/{g['id']}")
        goods = assert_ok(sg, sd, "货品主表价格")
        assert goods["retail_price"] == 250, f"主表零售价应同步为 250，实际 {goods['retail_price']}"

    def test_rows_save_multi_unit(self):
        u1 = create("unit", "/api/shop/unit", {"name": f"行保存主单位{SUF}"}, "行保存主单位")
        u2 = create("unit", "/api/shop/unit", {"name": f"行保存辅单位{SUF}"}, "行保存辅单位")
        col = f"行等级{SUF}"
        price_rows = {u1["name"]: {"规格B": {"code": f"RS{SUF}-1", "purchase_price": "10",
                                            "retail_price": "20", "wholesale_price": "15",
                                            "custom": {col: "1"}}}}
        g = create("goods", "/api/shop/goods", {
            "name": f"行保存货品{SUF}", "code": f"RS{SUF}",
            "unit_id": u1["id"], "has_multi_unit": 1,
            "units": [{"unit_id": u2["id"], "unit_name": u2["name"], "factor": 5, "is_main": 0}],
            "price_columns": json.dumps([col], ensure_ascii=False),
            "price_rows": json.dumps(price_rows, ensure_ascii=False),
        }, "行保存货品")
        s, d = api("PUT", self.ROWS, {"rows": [{
            "goods_id": g["id"], "unit_key": u1["name"], "spec_key": "规格B",
            "purchase_price": 12, "retail_price": 33, "wholesale_price": 22, "custom": {col: 9},
        }]})
        assert_ok(s, d, "内联保存多单位价格")
        row = self._find_row(g["id"], u1["name"], "规格B")
        assert row["retail_price"] == 33, f"多单位零售价应为 33，实际 {row['retail_price']}"

    def test_rows_empty(self):
        s, d = api("PUT", self.ROWS, {"rows": []})
        assert_bad(s, d, "没有需要保存的数据", "空行保存")

    def run(self):
        test("[价格] 列表结构", self.test_list_structure)
        test("[价格] 简单货品行/自定义价", self.test_simple_row_key)
        test("[价格] 自定义列并集", self.test_columns_union)
        test("[价格] 多单位展开", self.test_multi_unit_expansion)
        test("[价格] 筛选-关键字", self.test_filter_keyword)
        test("[价格] 筛选-规格", self.test_filter_spec)
        test("[价格] 筛选-品牌", self.test_filter_brand)
        test("[价格] 筛选-隐藏停用", self.test_filter_hide_disabled)
        test("[价格] page_size 上限", self.test_page_size_clamp)
        test("[价格] 批量改价-选中上调", self.test_batch_selected_up)
        test("[价格] 批量改价-自定义下调", self.test_batch_custom_down)
        test("[价格] 批量改价-全部+过滤", self.test_batch_all_with_filter)
        test("[价格] 批量改价-空内容", self.test_batch_empty_changes)
        test("[价格] 批量改价-不存在", self.test_batch_nonexistent_keys)
        test("[价格] 内联保存-简单", self.test_rows_save_simple)
        test("[价格] 内联保存-多单位", self.test_rows_save_multi_unit)
        test("[价格] 内联保存-空行", self.test_rows_empty)


# ── 跨模块集成 ──────────────────────────────────────────────────────────

class IntegrationTests:
    def test_category_unit_goods_flow(self):
        cat = create("category", "/api/shop/category", {"name": f"集成分类{SUF}"}, "集成分类")
        unit = create("unit", "/api/shop/unit", {"name": f"集成单位{SUF}"}, "集成单位")
        g = create("goods", "/api/shop/goods", {
            "name": f"集成货品{SUF}", "code": f"INT{SUF}",
            "category_id": cat["id"], "unit_id": unit["id"],
            "purchase_price": 10, "retail_price": 20, "wholesale_price": 15,
        }, "集成货品")
        s, d = api("GET", f"/api/shop/goods/{g['id']}")
        data = assert_ok(s, d, "集成货品详情")
        assert data["category_id"] == cat["id"], "分类未关联"
        assert data["unit_id"] == unit["id"], "单位未关联"
        # 删除拦截
        sc, dc = api("DELETE", f"/api/shop/category/{cat['id']}")
        assert_bad(sc, dc, "商品", "集成-分类删除拦截")
        su, du = api("DELETE", f"/api/shop/unit/{unit['id']}")
        assert_bad(su, du, "已被货品使用", "集成-单位删除拦截")

    def test_attribute_to_spec_groups(self):
        attr = create("attribute", "/api/shop/attribute",
                      {"name": f"集成规格{SUF}", "values": ["A", "B"]}, "集成规格")
        spec_groups = json.dumps([{"name": attr["name"], "has_image": False,
                                   "values": [{"name": "A", "image": ""}, {"name": "B", "image": ""}]}],
                                 ensure_ascii=False)
        g = create("goods", "/api/shop/goods",
                   {"name": f"集成规格货品{SUF}", "code": f"IA{SUF}", "spec_groups": spec_groups}, "集成规格货品")
        sa, da = api("DELETE", f"/api/shop/attribute/{attr['id']}")
        assert_bad(sa, da, "已被货品使用", "集成-规格删除拦截")
        api("DELETE", f"/api/shop/goods/{g['id']}")
        CREATED["goods"].remove(g["id"])

    def test_goods_to_price_flow(self):
        col = f"集成等级{SUF}"
        price_rows = {"__simple__": {"": {"purchase_price": "5", "retail_price": "50",
                                          "wholesale_price": "40", "custom": {col: "8"}}}}
        g = create("goods", "/api/shop/goods", {
            "name": f"集成价格货品{SUF}", "code": f"IP{SUF}",
            "retail_price": 50, "purchase_price": 5, "wholesale_price": 40,
            "price_columns": json.dumps([col], ensure_ascii=False),
            "price_rows": json.dumps(price_rows, ensure_ascii=False),
        }, "集成价格货品")
        # 出现在价格列表
        s, d = api("GET", "/api/shop/price/list", params={"keyword": f"IP{SUF}", "page_size": 100})
        data = assert_ok(s, d, "集成-价格列表")
        assert any(r["goods_id"] == g["id"] for r in data["list"]), "货品未出现在价格列表"
        # 批量改价后可见
        api("PUT", "/api/shop/price/batch", {
            "scope": "selected", "keys": [f"{g['id']}||"],
            "changes": [{"field": "retail_price", "mode": "up", "percent": 100}],
        })
        s2, d2 = api("GET", "/api/shop/price/list", params={"keyword": f"IP{SUF}", "page_size": 100})
        data2 = assert_ok(s2, d2, "集成-改价后价格列表")
        row = [r for r in data2["list"] if r["goods_id"] == g["id"]][0]
        assert row["retail_price"] == 100, f"改价后零售价应为 100，实际 {row['retail_price']}"

    def run(self):
        test("[集成] 分类+单位+货品 关联与删除拦截", self.test_category_unit_goods_flow)
        test("[集成] 规格→货品 spec_groups 删除拦截", self.test_attribute_to_spec_groups)
        test("[集成] 货品→价格列表→批量改价", self.test_goods_to_price_flow)


# ── 清理 ────────────────────────────────────────────────────────────────

CLEANUP_PATHS = {
    "goods": "/api/shop/goods",
    "attribute": "/api/shop/attribute",
    "category": "/api/shop/category",
    "unit": "/api/shop/unit",
    "property": "/api/shop/property",
}


def cleanup():
    print(f"\n{B}── 清理测试数据 ──{D}")
    total = 0
    for module in ("goods", "attribute", "category", "unit", "property"):
        path = CLEANUP_PATHS[module]
        ids = list(CREATED.get(module, []))
        # 分类需先删子级：按创建逆序
        if module == "category":
            ids = list(reversed(ids))
        for item_id in ids:
            api("DELETE", f"{path}/{item_id}")
            total += 1
    print(f"  {G}已清理 {total} 条测试数据{D}")


# ── Main ────────────────────────────────────────────────────────────────

def main():
    global BASE_URL

    parser = argparse.ArgumentParser(description="PISA 货品模块 API 测试套件")
    parser.add_argument("--base-url", default=BASE_URL, help="API base URL")
    parser.add_argument("--skip-cleanup", action="store_true", help="跳过收尾清理")
    args = parser.parse_args()
    BASE_URL = args.base_url.rstrip("/")

    print(f"\n{B}═══════════════════════════════════════════════{D}")
    print(f"{B}  PISA 进销存系统 - 货品模块 API 测试套件{D}")
    print(f"{B}  Base URL: {BASE_URL}   后缀: {SUF}{D}")
    print(f"{B}═══════════════════════════════════════════════{D}\n")

    print(f"{B}── 认证模块 ──{D}")
    AuthTests().run()
    print()

    modules = [
        ("单位管理", UnitTests()),
        ("货品分类", CategoryTests()),
        ("规格管理", AttributeTests()),
        ("货品属性", PropertyTests()),
        ("货品管理", GoodsTests()),
        ("图片上传", UploadTests()),
        ("价格管理", PriceTests()),
    ]
    for title, mod in modules:
        print(f"{B}── {title} ──{D}")
        mod.run()
        print()

    print(f"{B}── 集成测试 ──{D}")
    IntegrationTests().run()
    print()

    total = passed + failed + skipped
    print(f"{B}═══════════════════════════════════════════════{D}")
    print(f"{B}  测试结果{D}")
    print(f"{B}═══════════════════════════════════════════════{D}")
    print(f"  总计: {total}")
    print(f"  {G}通过: {passed}{D}")
    print(f"  {R}失败: {failed}{D}" if failed else f"  失败: {failed}")
    if skipped:
        print(f"  {Y}跳过: {skipped}{D}")

    if errors:
        print(f"\n{R}── 失败详情 ──{D}")
        for name, detail in errors:
            print(f"  {R}✗ {name}{D}")
            print(f"    {detail}")

    if not args.skip_cleanup:
        cleanup()

    print()
    if failed == 0:
        print(f"{G}{B}全部测试通过 ✓{D}\n")
        return 0
    print(f"{R}{B}有 {failed} 项测试失败 ✗{D}\n")
    return 1


if __name__ == "__main__":
    sys.exit(main())
