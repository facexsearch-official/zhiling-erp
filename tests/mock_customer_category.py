#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
PISA 进销存 — 客户分类 Mock 数据脚本（Python）

为指定商户生成「客户分类」测试数据（默认 100 条），并建立两级分类（父/子）。
默认作用于测试租户，避免污染真实数据。

用法:
    python3 tests/mock_customer_category.py                 # 生成 100 条（幂等，缺多少补多少）
    python3 tests/mock_customer_category.py --count 200
    python3 tests/mock_customer_category.py --clean         # 先清空该商户客户分类再生成
    python3 tests/mock_customer_category.py --tenant 2100945443734163456 --clean
"""
import argparse
import sys

try:
    import pymysql
except ImportError:
    sys.exit("缺少依赖 pymysql，请先执行: pip install pymysql")

import os
DB = dict(host=os.environ.get("PISA_DB_HOST", "127.0.0.1"), port=int(os.environ.get("PISA_DB_PORT", "3306")),
          user=os.environ.get("PISA_DB_USER", "root"), password=os.environ.get("PISA_DB_PASS", ""),
          database=os.environ.get("PISA_DB_NAME", "pisa"), charset="utf8mb4", autocommit=True)

TEST_TENANT_ID = 7000000000000000001


def id_base(tenant_id):
    """按商户派生独立 ID 段，避免不同商户之间主键冲突。"""
    return 7200000000000000000 + (tenant_id % 1000000) * 100000


TOP_NAMES = ["普通客户", "VIP客户", "批发客户", "零售客户", "企业客户",
             "电商客户", "团购客户", "会员客户", "经销商", "其他"]


def build_rows(tenant_id, count):
    """生成 count 条分类：每 10 条一组，组内第 1 条为父分类，其余为子分类。"""
    rows = []
    parent_ids = []
    base = id_base(tenant_id)
    for i in range(count):
        rid = base + i + 1
        if i % 10 == 0:
            name = "%s%02d" % (TOP_NAMES[(i // 10) % len(TOP_NAMES)], i // (10 * len(TOP_NAMES)) + 1)
            rows.append((rid, tenant_id, name, 0, i))
            parent_ids.append(rid)
        else:
            pid = parent_ids[-1]
            name = "子分类%03d" % i
            rows.append((rid, tenant_id, name, pid, i))
    return rows


def main():
    ap = argparse.ArgumentParser(description="客户分类 Mock 数据")
    ap.add_argument("--count", type=int, default=100, help="生成数量（默认 100）")
    ap.add_argument("--tenant", type=int, default=TEST_TENANT_ID, help="商户 ID（默认测试租户）")
    ap.add_argument("--clean", action="store_true", help="先清空该商户的客户分类")
    args = ap.parse_args()

    conn = pymysql.connect(**DB)
    cur = conn.cursor()

    if args.clean:
        n = cur.execute("DELETE FROM customer_categories WHERE tenant_id=%s", (args.tenant,))
        print("已清空客户分类 %d 条" % n)

    cur.execute("SELECT COUNT(*) FROM customer_categories WHERE tenant_id=%s", (args.tenant,))
    existing = cur.fetchone()[0]
    need = args.count - existing
    if need <= 0:
        print("客户分类已存在 %d 条（≥%d），无需新增" % (existing, args.count))
    else:
        rows = build_rows(args.tenant, args.count)[existing:]
        cur.executemany(
            "INSERT INTO customer_categories (id, tenant_id, name, parent_id, sort) VALUES (%s,%s,%s,%s,%s)",
            rows,
        )
        print("新增客户分类 %d 条" % len(rows))

    cur.execute("SELECT COUNT(*) FROM customer_categories WHERE tenant_id=%s", (args.tenant,))
    total = cur.fetchone()[0]
    cur.execute("SELECT COUNT(*) FROM customer_categories WHERE tenant_id=%s AND parent_id>0", (args.tenant,))
    children = cur.fetchone()[0]
    print("租户 %d 客户分类合计: %d 条（其中子分类 %d 条）" % (args.tenant, total, children))

    cur.close()
    conn.close()


if __name__ == "__main__":
    main()
