#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
通过 POST 接口给运行中的服务造演示数据（不直接操作数据库）。

顺序：门店(20) → 角色(20) → 业务员(20) → 用户(200) → 客户价格等级(100)
      → 客户分类(20) → 客户(100) → 供应商(200) → 商品规格(50) → 商品单位(50)
      → 商品分类(200,深层) → 商品(500) → 进货单(100,审核) → 进货退货单(100,审核)
      → 销售单(200) → 销售退货单(200) → 盘点单(200) → 账号(20)
      → 收款单(100) / 付款单(100)

用法:
    python3 scripts/mock_api.py --base https://erp.facexsearch.com
    python3 scripts/mock_api.py --base ... --conc 2 --tag t1 --insecure
"""
import argparse
import json
import random
import string
import sys
import time
from concurrent.futures import ThreadPoolExecutor, as_completed
from datetime import datetime, timedelta

try:
    import requests
except ImportError:
    sys.exit("缺少依赖 requests，请先执行: pip install requests")

rng = random.Random(20260101)
COUNTS = {}

# ── 运行参数 ──
ARGS = None
BASE = ""
SESSION = None
TAG = ""


def ri(a, b):
    return rng.randint(a, b)


def rf(a, b):
    return round(a + rng.random() * (b - a), 2)


def pick(a):
    return a[rng.randrange(len(a))]


def rand_date(days=365):
    return (datetime.now() - timedelta(days=ri(0, days))).strftime("%Y-%m-%d")


def phone(seq):
    return "1%d%09d" % (3 + seq % 6, 100000000 + seq * 131)


# ── HTTP ──
def post_one(path, body):
    last = ""
    for _ in range(3):
        try:
            r = SESSION.post(BASE + path, json=body, timeout=30)
            j = r.json()
            if j.get("code") == 0:
                return True, j.get("data"), None
            if r.status_code >= 500:
                time.sleep(0.4)
                last = "%s %s" % (r.status_code, j.get("message"))
                continue
            return False, None, "%s %s" % (r.status_code, j.get("message"))
        except Exception as e:
            last = str(e)
            time.sleep(0.4)
    return False, None, last


def eid(data):
    """返回整数 id（Go int64 可直接解析 JSON 整数；Python int 精确，无雪花的精度问题）。"""
    if not isinstance(data, dict):
        return 0
    v = data.get("id_str") or data.get("id")
    try:
        return int(v)
    except (TypeError, ValueError):
        return 0


def bulk(path, bodies, label):
    ok = fail = 0
    ids = [None] * len(bodies)
    first_err = None
    with ThreadPoolExecutor(max_workers=ARGS.conc) as ex:
        futs = {ex.submit(post_one, path, b): i for i, b in enumerate(bodies)}
        for f in as_completed(futs):
            i = futs[f]
            good, data, msg = f.result()
            if good:
                ok += 1
                ids[i] = eid(data)
            else:
                fail += 1
                if first_err is None:
                    first_err = msg
    print("  %-14s ok=%d fail=%d" % (label, ok, fail))
    if first_err:
        print("    ! 首个错误: %s" % first_err)
    return ids


def audit(path_tpl, ids, label):
    if not ids:
        return
    bodies = [None] * len(ids)
    paths = [path_tpl % i for i in ids if i]
    ok = fail = 0
    first_err = None
    with ThreadPoolExecutor(max_workers=ARGS.conc) as ex:
        futs = [ex.submit(post_one, p, {}) for p in paths]
        for f in as_completed(futs):
            good, data, msg = f.result()
            if good:
                ok += 1
            else:
                fail += 1
                if first_err is None:
                    first_err = msg
    print("  %-14s ok=%d fail=%d" % (label, ok, fail))
    if first_err:
        print("    ! 首个错误: %s" % first_err)


# ── 数据生成 ──
def gen_shops(n):
    cities = ["北京", "上海", "广州", "深圳", "杭州", "成都", "武汉", "西安", "南京", "苏州"]
    out = []
    for i in range(1, n + 1):
        c = cities[(i - 1) % len(cities)]
        out.append({
            "name": ("总店" if i == 1 else "%s门店%02d" % (c, i)),
            "address": "%s市%s区示范路%d号" % (c, c, i * 7),
            "address_detail": "%d楼%d室" % (i % 20 + 1, i),
            "phone": phone(1000 + i),
            "type": 1 + i % 2,
            "remark": "第 %d 家门店" % i,
            "is_main": 1 if i == 1 else 0,
        })
    return out


def gen_roles(n):
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
    out = []
    for i in range(n):
        out.append({"name": names[i % len(names)], "description": descs[i % len(descs)],
                    "permissions": full if i < 3 else view, "sensitive_data": "{}"})
    return out


def gen_salesmen(n, shop_ids):
    return [{"name": "业务员%02d" % i, "phone": phone(2000 + i),
             "shop_id": pick(shop_ids)} for i in range(1, n + 1)]


def gen_users(n, role_ids):
    return [{"name": "员工%03d" % i, "phone": phone(3000 + i),
             "role": 2 if i % 10 == 0 else 3, "role_id": role_ids[i % len(role_ids)],
             "password": "123456"} for i in range(1, n + 1)]


def gen_levels(n):
    return [{"name": "价格等级%03d" % i} for i in range(1, n + 1)]


def gen_customer_categories(n):
    names = ["普通客户", "VIP客户", "企业客户", "代理商", "经销商", "批发客户", "零售客户",
             "长期合作", "新客户", "潜在客户", "重点客户", "战略客户", "个人客户", "政府客户",
             "电商客户", "线下客户", "连锁客户", "加盟商", "合作单位", "其他"]
    return [{"name": names[i % len(names)]} for i in range(n)]


def gen_customers(n, level_names, cat_ids, salesman_ids):
    addrs = ["广东省深圳市南山区", "北京市朝阳区", "上海市浦东新区", "浙江省杭州市西湖区", "江苏省南京市玄武区"]
    banks = ["中国工商银行", "招商银行", "中国建设银行", "中国农业银行", "中国银行"]
    out = []
    for i in range(1, n + 1):
        out.append({
            "name": "客户%05d" % i, "code": "KH%05d" % i, "type": 1 + i % 3,
            "category_id": pick(cat_ids), "price_level": level_names[i % len(level_names)],
            "salesman_id": pick(salesman_ids), "discount": rf(80, 100), "init_debt": rf(0, 5000),
            "contact": "联系人%d" % (i % 50 + 1), "phone": phone(4000 + i),
            "address": pick(addrs), "address_detail": "科技园%d栋%d单元" % (i % 30 + 1, i % 20 + 1),
            "email": "customer%05d@example.com" % i, "tax_no": "91%016d" % (1000000000 + i),
            "fax": "0755-%08d" % (8000000 + i), "bank_name": pick(banks),
            "bank_account": "62%015d" % (600000000000 + i),
            "birthday": "19%02d-%02d-%02d" % (70 + i % 30, i % 12 + 1, i % 28 + 1),
            "wechat": "wx_%05d" % i, "qq": str(100000 + i), "remark": "客户备注 %d" % i,
            "balance": rf(0, 20000), "total_receivable": rf(0, 100000), "total_payable": rf(0, 50000),
            "addresses": [{"receiver": "联系人%d" % (i % 50 + 1), "phone": phone(4000 + i),
                           "region": pick(addrs), "detail": "收货地址%d号" % i, "is_default": 1}],
        })
    return out


def gen_suppliers(n, cat_ids):
    addrs = ["浙江省义乌市", "广东省东莞市", "江苏省常熟市", "福建省晋江市"]
    banks = ["中国银行", "交通银行", "平安银行", "兴业银行"]
    out = []
    for i in range(1, n + 1):
        name = "供应商%04d%s" % (i, TAG)
        out.append({
            "name": name, "code": "GYS%04d" % i, "category_id": pick(cat_ids),
            "contact": "供联%d" % (i % 30 + 1), "phone": phone(5000 + i), "address": pick(addrs),
            "bank_name": pick(banks), "bank_account": "62%015d" % (600000000000 + i),
            "remark": "供应商备注 %d" % i, "init_payable": rf(0, 10000), "total_payable": rf(0, 200000),
            "email": "supplier%04d@example.com" % i, "fax": "0571-%08d" % (8000000 + i),
            "wechat": "swx_%04d" % i, "qq": str(200000 + i),
            "birthday": "19%02d-%02d-%02d" % (70 + i % 30, i % 12 + 1, i % 28 + 1),
            "province": "浙江省", "city": "杭州市", "district": "西湖区",
            "address_detail": "产业园%d号" % i,
        })
    return out


def gen_attributes(n):
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
    keys = list(vmap.keys())
    out = []
    for i in range(n):
        k = keys[i % len(keys)]
        name = k if i < len(keys) else "%s%d" % (k, i // len(keys) + 1)
        out.append({"name": name, "values": vmap[k]})
    return out


def gen_units(n):
    names = ["个", "件", "箱", "包", "袋", "瓶", "盒", "套", "对", "双", "台", "只", "支", "条",
             "张", "片", "块", "卷", "桶", "罐", "斤", "公斤", "克", "吨", "米", "厘米", "升",
             "毫升", "打", "组", "打箱", "大包", "小包", "托盘", "捆", "扎", "串", "束", "车",
             "排", "提", "筐", "篓", "篮", "袋装", "瓶装", "罐装", "盒装", "散装", "整箱"]
    return [{"name": names[i % len(names)]} for i in range(n)]


def gen_goods_categories(n):
    """顺序创建（父必须先于子），共 n 个，层级最深约 8。"""
    roots = ["食品饮料", "日用百货", "数码家电", "服装鞋帽", "办公用品", "母婴玩具", "美妆护肤", "家居家纺"]
    subs = ["一级", "二级", "三级", "四级", "系列", "分类", "专区"]
    bodies = []
    nodes = []  # (index_in_bodies, depth)
    tok = TAG
    for i, r in enumerate(roots):
        bodies.append({"name": r + tok, "parent_id": "0", "sort": i + 1})
        nodes.append((i, 1))
    seq = 0
    while len(bodies) < n:
        pidx, pdepth = nodes[rng.randrange(len(nodes))]
        if pdepth >= 8:
            continue
        seq += 1
        parent_placeholder = ("PARENT_IDX", pidx)
        bodies.append({"name": "%s%d%03d%s" % (pick(subs), pdepth + 1, seq, tok),
                       "parent_id": parent_placeholder, "sort": seq % 20 + 1})
        nodes.append((len(bodies) - 1, pdepth + 1))
    # 顺序创建，用父的真实 id 替换占位
    ids = []
    ok = fail = 0
    first_err = None
    for b in bodies:
        pv = b["parent_id"]
        if isinstance(pv, tuple):
            b = dict(b)
            b["parent_id"] = ids[pv[1]] or "0"
        good, data, msg = post_one("/shop/category", b)
        if good:
            ok += 1
            ids.append(eid(data))
        else:
            fail += 1
            ids.append("")
            if first_err is None:
                first_err = msg
    print("  %-14s ok=%d fail=%d" % ("商品分类", ok, fail))
    if first_err:
        print("    ! 首个错误: %s" % first_err)
    return [i for i in ids if i]


def gen_goods(n, unit_names, unit_ids, cat_ids, supplier_ids, shop_ids):
    adjs = ["高清", "加厚", "便携", "多功能", "家用", "商用", "大容量", "迷你", "智能", "环保",
            "轻奢", "经典", "升级款", "豪华", "简约"]
    nouns = ["水杯", "收纳盒", "数据线", "充电器", "耳机", "笔记本", "T恤", "运动鞋", "台灯", "雨伞",
             "背包", "水壶", "毛巾", "拖鞋", "剃须刀", "吹风机", "电饭煲", "炒锅", "刀具", "洗衣液",
             "抽纸", "牙膏", "洗发水", "沐浴露", "面膜", "口红", "香水", "零食", "饼干", "坚果",
             "牛奶", "咖啡", "茶叶", "矿泉水", "马克笔", "文件夹", "打印纸", "订书机", "计算器", "保温杯"]
    brands = ["小米", "华为", "美的", "海尔", "格力", "苏泊尔", "九阳", "得力", "晨光", "南极人",
              "无印", "优衣库", "宝洁", "联合利华", "三只松鼠"]
    origins = ["广东深圳", "浙江义乌", "江苏苏州", "福建泉州", "山东青岛", "上海", "北京",
               "浙江杭州", "广东东莞", "四川成都"]
    stock_shops = shop_ids[:3]
    bodies, meta = [], []
    for i in range(1, n + 1):
        uidx = rng.randrange(len(unit_names))
        unit_name = unit_names[uidx]
        unit_id = unit_ids[uidx]
        pp = rf(5, 500)
        multi_spec = (i % 4 == 0)
        spec_keys = [""]
        spec_groups = "[]"
        price_rows = None
        if multi_spec:
            spec_keys = ["S", "M", "L"] if i % 8 else ["红", "蓝", "绿"]
            spec_groups = json.dumps([{"name": "规格", "has_image": False,
                                       "values": [{"name": v, "image": ""} for v in spec_keys]}], ensure_ascii=False)
            price_rows = {unit_name: {}}
            for sk in spec_keys:
                price_rows[unit_name][sk] = {"purchase_price": pp, "retail_price": rf(pp * 1.2, pp * 2),
                                             "wholesale_price": rf(pp * 1.05, pp * 1.3), "code": "", "barcode": "",
                                             "disabled": False, "custom": {}}
        stock_rows = {}
        total_stock = 0
        for sh in stock_shops:
            stock_rows[str(sh)] = {}
            for sk in spec_keys:
                st = ri(0, 200)
                total_stock += st
                stock_rows[str(sh)][sk] = {"stock": st, "min_stock": ri(5, 50), "safe_stock": ri(5, 60),
                                           "max_stock": ri(100, 1000), "init_cost": pp}
        g = {
            "name": "%s%s%d" % (pick(adjs), pick(nouns), i), "code": "SP%06d%s" % (i, TAG),
            "barcode": "69%011d" % (10000000 + i), "category_id": pick(cat_ids),
            "unit_id": unit_id, "supplier_id": pick(supplier_ids), "main_unit": unit_name,
            "sales_unit": unit_name, "purchase_unit": unit_name,
            "purchase_price": pp, "retail_price": rf(pp * 1.2, pp * 2), "wholesale_price": rf(pp * 1.05, pp * 1.3),
            "init_cost": pp, "status": 1, "remark": "商品备注%d" % i,
            "brand": pick(brands), "origin": pick(origins),
            "has_multi_unit": 0, "has_multi_spec": 1 if multi_spec else 0,
            "enable_stock_alert": i % 2, "min_stock": ri(5, 50), "safety_stock": ri(5, 60), "max_stock": ri(100, 1000),
            "has_batch": i % 2, "has_shelf_life": i % 2, "shelf_life_days": ri(30, 730),
            "expiry_alert": i % 2, "expiry_warn_days": ri(7, 60), "has_serial": 1 if i % 3 == 0 else 0,
            "images": "[]", "spec_groups": spec_groups, "price_columns": '["等级1","等级2"]',
            "current_stock": total_stock,
            "stock_rows": json.dumps(stock_rows, ensure_ascii=False),
        }
        if price_rows is not None:
            g["price_rows"] = json.dumps(price_rows, ensure_ascii=False)
        bodies.append(g)
        meta.append(spec_keys)
    ids = bulk("/shop/goods", bodies, "商品")
    goods = [(ids[i], meta[i]) for i in range(len(ids)) if ids[i]]
    return goods


def gen_accounts(n, shop_ids):
    names = ["现金", "微信", "支付宝", "中国银行", "工商银行", "建设银行", "农业银行", "招商银行",
             "交通银行", "平安银行", "兴业银行", "浦发银行", "中信银行", "光大银行", "民生银行",
             "广发银行", "华夏银行", "邮储银行", "备用金", "其他"]
    out = []
    for i, nm in enumerate(names[:n]):
        t = 1 if i == 0 else (3 if i <= 2 else 2)
        out.append({"name": nm, "type": t, "bank_name": nm, "card_no": "62%015d" % (600000000000 + i),
                    "shop_id": shop_ids[i % len(shop_ids)], "balance": rf(1000, 500000),
                    "remark": "账户-%s" % nm, "sort": i + 1})
    return out


def gen_items(goods, n):
    items = []
    for _ in range(n):
        gid, specs = pick(goods)
        qty = ri(1, 20)
        price = rf(5, 500)
        items.append({"goods_id": gid, "spec_key": pick(specs), "quantity": qty,
                      "unit_price": price, "remark": ""})
    return items


def objs_of(path):
    try:
        r = SESSION.get(BASE + path, timeout=30, verify=not ARGS.insecure)
        data = r.json().get("data")
        if isinstance(data, dict):
            lst = data.get("list") or []
        elif isinstance(data, list):
            lst = data
        else:
            lst = []
        out = []
        for x in lst:
            v = x.get("id_str") or x.get("id")
            try:
                out.append({"id": int(v), "name": x.get("name", "") or ""})
            except (TypeError, ValueError):
                pass
        return out
    except Exception as e:
        print("  ! 拉取 %s 失败: %s" % (path, e))
        return []


def main():
    global ARGS, BASE, SESSION, TAG
    p = argparse.ArgumentParser()
    p.add_argument("--base", default="https://erp.facexsearch.com")
    p.add_argument("--conc", type=int, default=2)
    p.add_argument("--tag", default="")
    p.add_argument("--phone", default="admin")
    p.add_argument("--password", default="123456")
    p.add_argument("--insecure", action="store_true")
    p.add_argument("--resume", action="store_true", help="主数据已存在，仅用现有数据继续造 商品+单据")
    ARGS = p.parse_args()
    BASE = ARGS.base.rstrip("/") + "/api"
    TAG = ARGS.tag or ("M" + "".join(rng.choice(string.ascii_uppercase + string.digits) for _ in range(4)))

    SESSION = requests.Session()
    if ARGS.insecure:
        SESSION.verify = False
        try:
            import urllib3
            urllib3.disable_warnings()
        except Exception:
            pass

    print("登录 %s ..." % (BASE + "/auth/login"))
    r = SESSION.post(BASE + "/auth/login", json={"phone": ARGS.phone, "password": ARGS.password},
                     timeout=30, verify=not ARGS.insecure)
    j = r.json()
    if j.get("code") != 0:
        sys.exit("登录失败: %s" % j)
    SESSION.headers["Authorization"] = "Bearer " + j["data"]["token"]
    print("登录成功  (tag=%s)\n" % TAG)

    t0 = time.time()

    if ARGS.resume:
        print("[resume] 从现有数据读取主数据 id ...")
        shop_ids = [o["id"] for o in objs_of("/shop/shops")]
        sm_ids = [o["id"] for o in objs_of("/shop/salesmen")]
        unl = objs_of("/shop/unit/all")
        unit_names = [o["name"] for o in unl]
        unit_ids = [o["id"] for o in unl]
        goods_cat_ids = [o["id"] for o in objs_of("/shop/category/all")]
        supplier_ids = [o["id"] for o in objs_of("/shop/supplier/all")]
        cust_ids = [o["id"] for o in objs_of("/shop/customer/all")]
        account_ids = [o["id"] for o in objs_of("/shop/account/all")]
        print("  shops=%d salesmen=%d units=%d cats=%d suppliers=%d customers=%d accounts=%d" % (
            len(shop_ids), len(sm_ids), len(unit_names), len(goods_cat_ids),
            len(supplier_ids), len(cust_ids), len(account_ids)))
    else:
        shop_ids = [i for i in bulk("/shop/shops", gen_shops(20), "门店") if i]
        role_ids = [i for i in bulk("/shop/role", gen_roles(20), "角色") if i]
        sm_ids = [i for i in bulk("/shop/salesmen", gen_salesmen(20, shop_ids), "业务员") if i]
        bulk("/shop/staff/users", gen_users(200, role_ids), "用户")
        level_names = ["价格等级%03d" % i for i in range(1, 101)]
        bulk("/shop/price-level", gen_levels(100), "价格等级")
        cat_ids = [i for i in bulk("/shop/customer-category", gen_customer_categories(20), "客户分类") if i]
        cust_ids = [i for i in bulk("/shop/customer", gen_customers(100, level_names, cat_ids, sm_ids), "客户") if i]
        supplier_ids = [i for i in bulk("/shop/supplier", gen_suppliers(200, cat_ids), "供应商") if i]
        bulk("/shop/attribute", gen_attributes(50), "商品规格")
        unit_bodies = gen_units(50)
        unit_ids = [i for i in bulk("/shop/unit", unit_bodies, "商品单位") if i]
        unit_names = [b["name"] for b in unit_bodies]
        goods_cat_ids = gen_goods_categories(200)
        account_ids = [i for i in bulk("/shop/account", gen_accounts(20, shop_ids), "账号") if i]

    if not (shop_ids and sm_ids and unit_ids and goods_cat_ids and supplier_ids and cust_ids):
        sys.exit("缺少必要主数据（shops/salesmen/units/categories/suppliers/customers），已中止")

    goods = gen_goods(500, unit_names, unit_ids, goods_cat_ids, supplier_ids, shop_ids)
    if not goods:
        sys.exit("商品创建全部失败，已中止（请检查上面的错误信息）")

    # 单据
    print("  (进货单/退货单创建后自动审核联库存)")
    purch_ids = []
    bodies = []
    purchased = []
    for i in range(1, 101):
        items = gen_items(goods, ri(1, 5))
        purchased.extend([(it["goods_id"], it["spec_key"], it["quantity"]) for it in items])
        bodies.append({"shop_id": pick(shop_ids), "warehouse_id": 0, "supplier_id": pick(supplier_ids),
                       "salesman_id": pick(sm_ids), "account_id": 0, "bill_date": rand_date(),
                       "discount": float(ri(90, 100)), "freight": float(ri(0, 100)), "deposit_offset": 0,
                       "paid_amount": rf(0, 1000), "invoice_status": i % 2, "related_no": "",
                       "attachments": "[]", "remark": "进货单备注 %d" % i, "items": items})
    purch_ids = bulk("/shop/purchase", bodies, "进货单")
    purch_ids = [i for i in purch_ids if i]
    audit("/shop/purchase/%s/audit", purch_ids, "进货审核")

    pr_bodies = []
    for i in range(1, 101):
        # 从已采购商品中取，保证有库存可退
        items = []
        for _ in range(ri(1, 3)):
            gid, sk, qty = pick(purchased)
            items.append({"goods_id": gid, "spec_key": sk, "quantity": min(qty, ri(1, 3)),
                          "unit_price": rf(5, 500), "remark": ""})
        pr_bodies.append({"shop_id": pick(shop_ids), "warehouse_id": 0, "supplier_id": pick(supplier_ids),
                          "salesman_id": pick(sm_ids), "account_id": 0, "bill_date": rand_date(),
                          "deposit_offset": 0, "paid_amount": rf(0, 500), "invoice_status": i % 2,
                          "related_no": "", "attachments": "[]", "remark": "进货退货备注 %d" % i,
                          "items": items})
    pr_ids = bulk("/shop/purchase-return", pr_bodies, "进货退货")
    pr_ids = [i for i in pr_ids if i]
    audit("/shop/purchase-return/%s/audit", pr_ids, "退货审核")

    # 销售 / 退货 需要客户 id
    sale_bodies = []
    for i in range(1, 201):
        items = gen_items(goods, ri(1, 4))
        sale_bodies.append({"shop_id": pick(shop_ids), "warehouse_id": 0, "customer_id": pick(cust_ids),
                            "salesman_id": pick(sm_ids), "account_id": 0, "bill_date": rand_date(),
                            "discount": float(ri(90, 100)), "round_off": 0, "received_amount": rf(0, 500),
                            "invoice_status": i % 2, "print_status": i % 2, "attachments": "[]",
                            "remark": "销售单备注 %d" % i, "items": items})
    bulk("/shop/sale", sale_bodies, "销售单")

    sr_bodies = []
    for i in range(1, 201):
        items = gen_items(goods, ri(1, 3))
        sr_bodies.append({"shop_id": pick(shop_ids), "warehouse_id": 0, "customer_id": pick(cust_ids),
                          "salesman_id": pick(sm_ids), "account_id": 0, "bill_date": rand_date(),
                          "round_off": 0, "received_amount": rf(0, 300), "print_status": i % 2,
                          "attachments": "[]", "remark": "销售退货备注 %d" % i, "items": items})
    bulk("/shop/sale-return", sr_bodies, "销售退货")

    spec_pairs = [(g[0], sk) for g in goods for sk in g[1]]
    sc_bodies = []
    for i in range(1, 201):
        items = []
        for _ in range(ri(1, 5)):
            gid, sk = pick(spec_pairs)
            bq = ri(0, 200)
            aq = max(0, bq + ri(-10, 10))
            items.append({"goods_id": gid, "book_qty": bq, "actual_qty": aq, "diff_qty": aq - bq, "remark": ""})
        sc_bodies.append({"shop_id": pick(shop_ids), "warehouse_id": 0, "salesman_id": pick(sm_ids),
                          "bill_date": rand_date(), "attachments": "[]", "remark": "盘点备注 %d" % i,
                          "items": items})
    bulk("/shop/stock-count", sc_bodies, "盘点单")

    rec_bodies = []
    for i in range(1, 101):
        rec_bodies.append({"shop_id": pick(shop_ids), "related_no": "",
                           "type": pick(["直接收款", "预收款", "销售收款"]), "customer_id": pick(cust_ids),
                           "salesman_id": pick(sm_ids), "bill_date": rand_date(), "amount": rf(100, 50000),
                           "discount_amount": rf(0, 200), "deposit_offset": 0, "account_id": pick(account_ids),
                           "attachments": "[]", "remark": "收款备注 %d" % i})
    bulk("/shop/receipt", rec_bodies, "收款单")

    pay_bodies = []
    for i in range(1, 101):
        pay_bodies.append({"shop_id": pick(shop_ids), "related_no": "",
                           "type": pick(["直接付款", "预付款", "采购付款"]), "supplier_id": pick(supplier_ids),
                           "salesman_id": pick(sm_ids), "bill_date": rand_date(), "amount": rf(100, 50000),
                           "discount_amount": rf(0, 200), "account_id": pick(account_ids),
                           "attachments": "[]", "remark": "付款备注 %d" % i})
    bulk("/shop/payment", pay_bodies, "付款单")

    print("\n完成，用时 %.1fs" % (time.time() - t0))


if __name__ == "__main__":
    main()
