#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
PISA 进销存 — 分类接口测试（Python）

覆盖：商品分类 / 客户分类 / 供应商分类 的
  列表 / 新增子类(字符串 parent_id) / 新增顶级(空 parent_id) / 修改 / 列表校验 / 删除
直连运行中的服务（默认 http://127.0.0.1:8080），退出码 0 表示全部通过。

用法:
    python3 tests/test_category.py
    PISA_BASE=http://127.0.0.1:8081 python3 tests/test_category.py
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

BASE_HOST = os.environ.get("PISA_BASE", "http://127.0.0.1:8080")
BASE = BASE_HOST + "/api"
ROOT = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))

PASS = 0
FAIL = 0
FAILURES = []


def load_secret():
    with open(os.path.join(ROOT, "config.yaml"), "r", encoding="utf-8") as f:
        for line in f:
            m = re.match(r"\s*secret:\s*[\"']?([^\"'\s]+)", line)
            if m:
                return m.group(1)
    return "pisa-secret-key-change-in-production"


def load_dsn():
    with open(os.path.join(ROOT, "config.yaml"), "r", encoding="utf-8") as f:
        txt = f.read()
    m = re.search(r'dsn:\s*"([^"]+)"', txt)
    return m.group(1) if m else "root:@tcp(127.0.0.1:3306)/pisa"


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


def resolve_tenant():
    """解析 tenant_id 与主账号 user_id（用于签发有权限的 JWT）"""
    tid = int(os.environ["PISA_TENANT_ID"]) if os.environ.get("PISA_TENANT_ID") else 2104556395360686080
    uid = int(os.environ["PISA_USER_ID"]) if os.environ.get("PISA_USER_ID") else 0
    try:
        import pymysql
        m = re.match(r"([^:]+):([^@]*)@tcp\(([^:]+):(\d+)\)/([^?]+)", load_dsn())
        conn = pymysql.connect(user=m.group(1), password=m.group(2), host=m.group(3),
                               port=int(m.group(4)), database=m.group(5), charset="utf8mb4")
        cur = conn.cursor()
        if not os.environ.get("PISA_TENANT_ID"):
            cur.execute("SELECT id FROM tenants WHERE name=%s LIMIT 1", ("演示商户",))
            row = cur.fetchone()
            if row:
                tid = int(row[0])
        if not uid:
            cur.execute(
                "SELECT ut.user_id FROM user_tenants ut "
                "WHERE ut.tenant_id=%s ORDER BY ut.is_owner DESC, ut.role ASC LIMIT 1", (tid,))
            row = cur.fetchone()
            if row:
                uid = int(row[0])
        conn.close()
    except Exception as e:
        print("  ! 解析商户/用户失败: %s" % e)
    return tid, (uid or tid)


SECRET = load_secret()
TENANT_ID, USER_ID = resolve_tenant()
TOKEN = make_token(SECRET, USER_ID, TENANT_ID)
SESSION = requests.Session()


def do(method, path, body=None):
    try:
        r = SESSION.request(method, BASE + path, json=body,
                            headers={"Authorization": "Bearer " + TOKEN}, timeout=30)
        try:
            return r.status_code, r.json()
        except ValueError:
            return r.status_code, None
    except Exception:
        return 0, None


def code_of(b):
    if not b:
        return None
    try:
        return int(b.get("code"))
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


def ok(name, st, body):
    global PASS, FAIL
    if st != 200 or body is None:
        FAIL += 1
        FAILURES.append("%s — HTTP %s" % (name, st))
        print("  x %-34s HTTP %s" % (name, st))
        return False
    if code_of(body) != 0:
        FAIL += 1
        FAILURES.append("%s — %s" % (name, body.get("message")))
        print("  x %-34s %s" % (name, body.get("message")))
        return False
    PASS += 1
    print("  ok %-33s" % name)
    return True


def assert_true(name, cond, detail=""):
    global PASS, FAIL
    if cond:
        PASS += 1
        print("  ok %-33s" % name)
    else:
        FAIL += 1
        FAILURES.append("%s — %s" % (name, detail))
        print("  x %-34s %s" % (name, detail))


def list_items(path):
    _, b = do("GET", path)
    data = (b or {}).get("data")
    if isinstance(data, dict):
        return data.get("list") or []
    if isinstance(data, list):
        return data
    return []


def find_by_name(path, name):
    for it in list_items(path):
        if it.get("name") == name:
            return str(it.get("id_str") or it.get("id") or "")
    return ""


CATS = {
    "商品分类": {
        "list": "/shop/category/all",
        "create": "/shop/category",
        "item": "/shop/category/%s",
    },
    "客户分类": {
        "list": "/shop/customer-category/list",
        "create": "/shop/customer-category",
        "item": "/shop/customer-category/%s",
    },
    "供应商分类": {
        "list": "/shop/supplier-category/all",
        "create": "/shop/supplier-category",
        "item": "/shop/supplier-category/%s",
    },
}


def test_category(label, cfg):
    stamp = int(time.time() * 1000)
    print("[%s] 开始" % label)

    items = list_items(cfg["list"])
    assert_true("%s-列表非空" % label, len(items) > 0, "列表为空")
    parent = ""
    if items:
        parent = str(items[0].get("id_str") or items[0].get("id") or "")
    print("  - 共 %d 条，父级 id=%s" % (len(items), parent))

    child_name = "测试子类_%d" % stamp
    top_name = "测试顶级_%d" % stamp
    created = []

    # 新增子类（字符串 parent_id）
    st, b = do("POST", cfg["create"], {"name": child_name, "parent_id": parent})
    if not ok("%s-新增子类(字符串父ID)" % label, st, b):
        return
    cid = find_by_name(cfg["list"], child_name)
    assert_true("%s-子类已入库" % label, cid != "", "列表未找到 %s" % child_name)
    if cid:
        created.append(cid)

    # 新增顶级（空 parent_id）
    st, b = do("POST", cfg["create"], {"name": top_name, "parent_id": ""})
    if ok("%s-新增顶级(空父ID)" % label, st, b):
        tid = find_by_name(cfg["list"], top_name)
        assert_true("%s-顶级已入库" % label, tid != "", "列表未找到 %s" % top_name)
        if tid:
            created.append(tid)

    # 修改
    if cid:
        st, b = do("PUT", cfg["item"] % cid, {"name": child_name + "_改", "parent_id": parent})
        if ok("%s-修改" % label, st, b):
            assert_true("%s-修改生效" % label, find_by_name(cfg["list"], child_name + "_改") != "")

    # 删除清理
    for c in created:
        st, b = do("DELETE", cfg["item"] % c)
        ok("%s-删除" % label, st, b)
    assert_true("%s-清理完成" % label,
                find_by_name(cfg["list"], child_name) == "" and find_by_name(cfg["list"], child_name + "_改") == "")
    print("")


def main():
    print("服务: %s  商户: %d  用户: %d" % (BASE_HOST, TENANT_ID, USER_ID))
    if code_of(do("GET", "/shop/category/all")[1]) is None:
        print("无法访问服务，请确认已启动")
        return 1
    for label, cfg in CATS.items():
        test_category(label, cfg)
    print("=" * 48)
    print("通过 %d，失败 %d" % (PASS, FAIL))
    for f in FAILURES:
        print("  - " + f)
    return 0 if FAIL == 0 else 1


if __name__ == "__main__":
    sys.exit(main())
