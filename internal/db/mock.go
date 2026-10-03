package db

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"math/rand"
	"os"
	"sort"
	"time"

	"pisa_server/internal/model"
	"pisa_server/internal/pkg/snowflake"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func mround(v float64) float64 { return math.Round(v*100) / 100 }

// MockIfRequested 当设置环境变量 PISA_MOCK 时执行 mock 数据导入。
//
//	PISA_MOCK=1     仅当尚未导入（客户数 < 10000）时导入
//	PISA_MOCK=force 强制清空并重新导入
func MockIfRequested(db *gorm.DB) {
	mode := os.Getenv("PISA_MOCK")
	if mode == "" {
		return
	}
	var tenant model.Tenant
	if err := db.Where("name = ?", "演示商户").First(&tenant).Error; err != nil || tenant.ID == 0 {
		log.Printf("mock: 演示商户不存在，跳过")
		return
	}
	var cnt int64
	db.Model(&model.Customer{}).Where("tenant_id = ?", tenant.ID).Count(&cnt)
	if mode != "force" && cnt >= 10000 {
		log.Printf("mock: 已存在 %d 条客户数据，跳过（如需重建设置 PISA_MOCK=force）", cnt)
		return
	}
	SeedMock(db, tenant.ID)
}

var mockRng = rand.New(rand.NewSource(20260101))

func mr(n int) int {
	if n <= 0 {
		return 0
	}
	return mockRng.Intn(n)
}
func pickStr(a []string) string { return a[mr(len(a))] }
func rf(min, max float64) float64 {
	return float64(int64((min+mockRng.Float64()*(max-min))*100+0.5)) / 100
}
func ri(min, max int) int {
	if max <= min {
		return min
	}
	return min + mockRng.Intn(max-min+1)
}
func randDayWithin(days int) string {
	return time.Now().AddDate(0, 0, -mr(days)).Format("2006-01-02")
}
func randTimeWithin(days int) time.Time {
	return time.Now().AddDate(0, 0, -mr(days)).Add(time.Duration(mr(86400)) * time.Second)
}

// SeedMock 清空并重建演示商户的测试数据
func SeedMock(db *gorm.DB, tenantID int64) {
	start := time.Now()
	log.Printf("mock: 开始清空演示商户测试数据...")
	wipeMock(db, tenantID)

	units := mockUnits(db, tenantID)
	attrs := mockGoodsAttrs(db, tenantID)

	shops := mockShops(db, tenantID)
	roles := mockRoles(db, tenantID)
	salesmen := mockSalesmen(db, tenantID, shops)
	users := mockUsers(db, tenantID, roles)

	levels := mockPriceLevels(db, tenantID)
	custCats := mockCustomerCategories(db, tenantID)
	customers := mockCustomers(db, tenantID, levels, custCats, salesmen)

	supCats := mockSupplierCategories(db, tenantID)
	suppliers := mockSuppliers(db, tenantID, supCats)

	goodsCats := mockGoodsCategories(db, tenantID)
	goods := mockGoods(db, tenantID, units, attrs, goodsCats, suppliers, shops)

	accounts := mockAccounts(db, tenantID, shops)

	mockPurchases(db, tenantID, suppliers, goods, salesmen, accounts, shops)
	mockPurchaseReturns(db, tenantID, suppliers, goods, salesmen, accounts, shops)
	mockSales(db, tenantID, customers, goods, salesmen, accounts, shops)
	mockSalesReturns(db, tenantID, customers, goods, salesmen, accounts, shops)
	mockStockCounts(db, tenantID, goods, salesmen, shops)
	mockReceipts(db, tenantID, customers, salesmen, accounts, shops)
	mockPayments(db, tenantID, suppliers, salesmen, accounts, shops)

	log.Printf("mock: 完成，用时 %s（shops=%d roles=%d users=%d levels=%d customers=%d suppliers=%d units=%d attrs=%d goodsCats=%d goods=%d accounts=%d）",
		time.Since(start), len(shops), len(roles), len(users), len(levels), len(customers),
		len(suppliers), len(units), len(attrs), len(goodsCats), len(goods), len(accounts))
}

func wipeMock(db *gorm.DB, tenantID int64) {
	tables := []string{
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
	}
	for _, t := range tables {
		// 明细表可能没有 tenant_id 时用主表关联，这里统一按 tenant_id（所有相关表均有该列）
		if err := db.Exec("DELETE FROM "+t+" WHERE tenant_id = ?", tenantID).Error; err != nil {
			log.Printf("mock wipe %s: %v", t, err)
		}
	}
	// 清理该商户下除管理员外的所有用户与关系
	var adminID int64
	db.Table("users").Where("phone = ?", "admin").Pluck("id", &adminID)
	db.Exec("DELETE FROM users WHERE id <> ? AND id IN (SELECT t.user_id FROM (SELECT user_id FROM user_tenants WHERE tenant_id = ?) t)", adminID, tenantID)
	db.Exec("DELETE FROM user_tenants WHERE tenant_id = ? AND user_id <> ?", tenantID, adminID)
}

/* ─────────── 门店 ─────────── */
func mockShops(db *gorm.DB, tenantID int64) []model.Shop {
	cities := []string{"北京", "上海", "广州", "深圳", "杭州", "成都", "武汉", "西安", "南京", "苏州"}
	list := make([]model.Shop, 0, 20)
	for i := 1; i <= 20; i++ {
		city := cities[(i-1)%len(cities)]
		s := model.Shop{
			ID: snowflake.GenID(), TenantID: tenantID,
			Name:          fmt.Sprintf("%s门店%02d", city, i),
			Address:       fmt.Sprintf("%s市%s区示范路%d号", city, city, i*7),
			AddressDetail: fmt.Sprintf("%d楼%d室", i%20+1, i),
			Phone:         fmt.Sprintf("1%d%09d", 3+i%6, 100000000+i*137),
			Logo:          "",
			Type:          int8(1 + i%2),
			Remark:        fmt.Sprintf("第 %d 家门店", i),
			IsMain:        0,
			Status:        int8(1),
			CreatedAt:     randTimeWithin(720),
		}
		if i == 1 {
			s.Name = "总店"
			s.IsMain = 1
		}
		list = append(list, s)
	}
	insertBatches(db, list)
	return list
}

/* ─────────── 角色 ─────────── */
func mockRoles(db *gorm.DB, tenantID int64) []model.Role {
	names := []string{"总经理", "副总经理", "店长", "收银员", "仓管员", "采购员", "销售员", "财务主管",
		"会计", "出纳", "客服", "理货员", "配送员", "运营专员", "市场专员", "督导", "审核员",
		"数据员", "店助", "临时工"}
	descs := []string{"拥有全部权限", "协助管理日常事务", "负责门店整体运营", "负责收银开单",
		"负责仓库进出库", "负责采购进货", "负责销售跟单", "负责财务核算", "负责账务处理",
		"负责现金收付", "负责客户接待", "负责货品整理", "负责配送", "负责线上运营", "负责市场推广",
		"负责门店巡查", "负责单据审核", "负责数据统计", "协助店长", "临时用工"}
	permsFull := `{"*":true}`
	permsView := `{"goods.goods.view":true,"sale.sale.view":true,"purchase.purchase.view":true,"stock.query.view":true,"funds.receipt.view":true,"funds.payment.view":true,"customer.customer.view":true,"purchase.supplier.view":true}`
	list := make([]model.Role, 0, 20)
	for i := 0; i < 20; i++ {
		perm := permsView
		if i < 3 {
			perm = permsFull
		}
		list = append(list, model.Role{
			ID: snowflake.GenID(), TenantID: tenantID,
			Name: names[i], Description: descs[i],
			Permissions: perm, SensitiveData: `{}`,
			IsSystem: 0, Status: 1, CreatedAt: randTimeWithin(720),
		})
	}
	insertBatches(db, list)
	return list
}

func mockSalesmen(db *gorm.DB, tenantID int64, shops []model.Shop) []model.Salesman {
	list := make([]model.Salesman, 0, 20)
	for i := 1; i <= 20; i++ {
		list = append(list, model.Salesman{
			ID: snowflake.GenID(), TenantID: tenantID,
			Name:   fmt.Sprintf("业务员%02d", i),
			Phone:  fmt.Sprintf("151%08d", 10000000+i*311),
			ShopID: shops[mr(len(shops))].ID,
			Status: 1, CreatedAt: randTimeWithin(720), UpdatedAt: time.Now(),
		})
	}
	insertBatches(db, list)
	return list
}

/* ─────────── 用户 ─────────── */
func mockUsers(db *gorm.DB, tenantID int64, roles []model.Role) []model.User {
	hash, _ := bcrypt.GenerateFromPassword([]byte("123456"), bcrypt.DefaultCost)
	list := make([]model.User, 0, 200)
	links := make([]model.UserTenant, 0, 200)
	for i := 1; i <= 200; i++ {
		suffix := 100000000 + i*1237
		u := model.User{
			ID:           snowflake.GenID(),
			Phone:        fmt.Sprintf("13%d", suffix)[:11],
			PasswordHash: string(hash),
			Nickname:     fmt.Sprintf("员工%03d", i),
			Avatar:       "",
			Status:       1,
			CreatedAt:    randTimeWithin(720),
			UpdatedAt:    time.Now(),
		}
		last := randTimeWithin(60)
		u.LastLoginAt = &last
		dt := tenantID
		u.DefaultTenantID = &dt
		list = append(list, u)
	}
	insertBatches(db, list)
	for i, u := range list {
		role := roles[i%len(roles)]
		links = append(links, model.UserTenant{
			ID: snowflake.GenID(), UserID: u.ID, TenantID: tenantID,
			IsOwner: 0, Role: int8(1 + i%3), RoleID: role.ID,
			StaffName: u.Nickname, StaffPhone: u.Phone,
			Permissions: role.Permissions, Status: 1, JoinedAt: randTimeWithin(720),
		})
	}
	insertBatches(db, links)
	return list
}

/* ─────────── 价格等级 ─────────── */
func mockPriceLevels(db *gorm.DB, tenantID int64) []model.PriceLevel {
	list := make([]model.PriceLevel, 0, 100)
	for i := 1; i <= 100; i++ {
		list = append(list, model.PriceLevel{
			ID: snowflake.GenID(), TenantID: tenantID,
			Name: fmt.Sprintf("价格等级%03d", i), Sort: i, Status: 1, CreatedAt: randTimeWithin(720),
		})
	}
	insertBatches(db, list)
	return list
}

/* ─────────── 客户分类 / 客户 ─────────── */
func mockCustomerCategories(db *gorm.DB, tenantID int64) []model.CustomerCategory {
	names := []string{"普通客户", "VIP客户", "企业客户", "代理商", "经销商", "批发客户", "零售客户",
		"长期合作", "新客户", "潜在客户", "重点客户", "战略客户", "个人客户", "政府客户", "电商客户",
		"线下客户", "连锁客户", "加盟商", "合作单位", "其他"}
	list := make([]model.CustomerCategory, 0, len(names))
	for i, n := range names {
		list = append(list, model.CustomerCategory{ID: snowflake.GenID(), TenantID: tenantID, Name: n, ParentID: 0, Sort: i + 1})
	}
	insertBatches(db, list)
	return list
}

func mockCustomers(db *gorm.DB, tenantID int64, levels []model.PriceLevel, cats []model.CustomerCategory, salesmen []model.Salesman) []model.Customer {
	total := 10000
	list := make([]model.Customer, 0, total)
	for i := 1; i <= total; i++ {
		level := levels[i%len(levels)]
		cat := cats[i%len(cats)]
		cid := model.FlexInt64(cat.ID)
		sm := salesmen[i%len(salesmen)]
		c := model.Customer{
			ID: snowflake.GenID(), TenantID: tenantID,
			Name:            fmt.Sprintf("客户%05d", i),
			Code:            fmt.Sprintf("KH%05d", i),
			Type:            int8(1 + i%3),
			CategoryID:      &cid,
			PriceLevel:      level.Name,
			SalesmanID:      sm.ID,
			Discount:        rf(80, 100),
			InitDebt:        rf(0, 5000),
			Contact:         fmt.Sprintf("联系人%d", i%50+1),
			Phone:           fmt.Sprintf("1%d%09d", 3+i%6, 200000000+i*97),
			Address:         pickStr([]string{"广东省深圳市南山区", "北京市朝阳区", "上海市浦东新区", "浙江省杭州市西湖区", "江苏省南京市玄武区"}),
			AddressDetail:   fmt.Sprintf("科技园%d栋%d单元", i%30+1, i%20+1),
			Email:           fmt.Sprintf("customer%05d@example.com", i),
			TaxNo:           fmt.Sprintf("91%016d", 1000000000+i),
			Fax:             fmt.Sprintf("0755-%08d", 8000000+i),
			BankName:        pickStr([]string{"中国工商银行", "招商银行", "中国建设银行", "中国农业银行", "中国银行"}),
			BankAccount:     fmt.Sprintf("62%015d", 600000000000+i),
			Birthday:        fmt.Sprintf("19%02d-%02d-%02d", 70+i%30, i%12+1, i%28+1),
			Wechat:          fmt.Sprintf("wx_%05d", i),
			QQ:              fmt.Sprintf("%d", 100000+i),
			Remark:          fmt.Sprintf("客户备注 %d", i),
			Balance:         rf(0, 20000),
			TotalReceivable: rf(0, 100000),
			TotalPayable:    rf(0, 50000),
			Status:          int8(1 + i%4%2),
			CreatedAt:       randTimeWithin(720), UpdatedAt: time.Now(),
		}
		if c.Status == 0 {
			c.Status = 1
		}
		list = append(list, c)
	}
	insertBatches(db, list)

	// 收货地址（前 1000 个客户各 1~2 条）
	addr := make([]model.CustomerAddress, 0, 2000)
	for i := 0; i < 1000 && i < len(list); i++ {
		n := 1 + i%2
		for j := 0; j < n; j++ {
			addr = append(addr, model.CustomerAddress{
				ID: snowflake.GenID(), TenantID: tenantID, CustomerID: list[i].ID,
				Receiver: list[i].Contact, Phone: list[i].Phone,
				Region: list[i].Address, Detail: list[i].AddressDetail, IsDefault: int8(boolToInt(j == 0)),
			})
		}
	}
	insertBatches(db, addr)
	return list
}

/* ─────────── 供应商分类 / 供应商 ─────────── */
func mockSupplierCategories(db *gorm.DB, tenantID int64) []model.SupplierCategory {
	names := []string{"常规供应商", "一级代理", "生产厂家", "贸易商", "批发商", "进口商",
		"本地供应商", "外省供应商", "长期合作", "临时供应商", "战略供应商", "备选供应商",
		"食品类", "日化类", "数码类", "服装类", "办公类", "五金类", "包装类", "其他"}
	list := make([]model.SupplierCategory, 0, len(names))
	for i, n := range names {
		list = append(list, model.SupplierCategory{ID: snowflake.GenID(), TenantID: tenantID, Name: n, ParentID: 0, Sort: i + 1})
	}
	insertBatches(db, list)
	return list
}

func mockSuppliers(db *gorm.DB, tenantID int64, cats []model.SupplierCategory) []model.Supplier {
	total := 200
	list := make([]model.Supplier, 0, total)
	for i := 1; i <= total; i++ {
		cat := cats[i%len(cats)]
		cid := model.FlexInt64(cat.ID)
		list = append(list, model.Supplier{
			ID: snowflake.GenID(), TenantID: tenantID,
			Name:         fmt.Sprintf("供应商%04d", i),
			Code:         fmt.Sprintf("GYS%04d", i),
			CategoryID:   &cid,
			Contact:      fmt.Sprintf("供联%d", i%30+1),
			Phone:        fmt.Sprintf("1%d%09d", 3+i%6, 300000000+i*131),
			Address:      pickStr([]string{"浙江省义乌市", "广东省东莞市", "江苏省常熟市", "福建省晋江市"}),
			BankName:     pickStr([]string{"中国银行", "交通银行", "平安银行", "兴业银行"}),
			BankAccount:  fmt.Sprintf("62%015d", 600000000000+i),
			Remark:       fmt.Sprintf("供应商备注 %d", i),
			InitPayable:  rf(0, 10000),
			TotalPayable: rf(0, 200000),
			Status:       int8(1 + i%4%2),
			Email:        fmt.Sprintf("supplier%04d@example.com", i),
			Fax:          fmt.Sprintf("0571-%08d", 8000000+i),
			Wechat:       fmt.Sprintf("swx_%04d", i),
			QQ:           fmt.Sprintf("%d", 200000+i),
			Birthday:     fmt.Sprintf("19%02d-%02d-%02d", 70+i%30, i%12+1, i%28+1),
			Province:     "浙江省", City: "杭州市", District: "西湖区",
			AddressDetail: fmt.Sprintf("产业园%d号", i),
			CreatedAt:     randTimeWithin(720), UpdatedAt: time.Now(),
		})
	}
	insertBatches(db, list)
	addr := make([]model.SupplierAddress, 0, 400)
	for i := 0; i < 200; i++ {
		addr = append(addr, model.SupplierAddress{
			ID: snowflake.GenID(), TenantID: tenantID, SupplierID: list[i].ID,
			Consignee: list[i].Contact, ConsigneePhone: list[i].Phone,
			Province: list[i].Province, City: list[i].City, District: list[i].District,
			AddressDetail: list[i].AddressDetail, IsDefault: 1, Sort: 1,
		})
	}
	insertBatches(db, addr)
	return list
}

/* ─────────── 单位 / 规格属性 ─────────── */
func mockUnits(db *gorm.DB, tenantID int64) []model.Unit {
	names := []string{"个", "件", "箱", "包", "袋", "瓶", "盒", "套", "对", "双",
		"台", "只", "支", "条", "张", "片", "块", "卷", "桶", "罐",
		"斤", "公斤", "克", "吨", "米", "厘米", "升", "毫升", "打", "组",
		"打箱", "大包", "小包", "托盘", "捆", "扎", "串", "束", "车", "排",
		"提", "筐", "篓", "篮", "袋装", "瓶装", "罐装", "盒装", "散装", "整箱"}
	list := make([]model.Unit, 0, len(names))
	for i, n := range names {
		list = append(list, model.Unit{ID: snowflake.GenID(), TenantID: tenantID, Name: n, CreatedAt: randTimeWithin(720)})
		_ = i
	}
	insertBatches(db, list)
	return list
}

func mockGoodsAttrs(db *gorm.DB, tenantID int64) []model.GoodsAttribute {
	vmap := map[string][]string{
		"颜色":   {"红色", "蓝色", "绿色", "黑色", "白色", "黄色"},
		"尺码":   {"S", "M", "L", "XL", "XXL"},
		"材质":   {"棉", "涤纶", "混纺", "真皮", "帆布"},
		"口味":   {"原味", "香辣", "五香", "麻辣", "香甜"},
		"规格":   {"小", "中", "大", "特大"},
		"型号":   {"A型", "B型", "C型", "Pro"},
		"容量":   {"500ml", "1L", "1.5L", "2L"},
		"重量":   {"100g", "500g", "1kg", "2kg", "5kg"},
		"功率":   {"10W", "20W", "50W", "100W"},
		"版本":   {"标准版", "豪华版", "旗舰版"},
		"款式":   {"简约", "复古", "运动", "商务"},
		"图案":   {"纯色", "条纹", "格子", "印花"},
		"香型":   {"柠檬", "薰衣草", "玫瑰", "薄荷"},
		"度数":   {"38度", "42度", "52度"},
		"产地":   {"国产", "进口"},
		"包装":   {"袋装", "盒装", "瓶装", "罐装"},
		"净含量":  {"100g", "250g", "500g", "1kg"},
		"适用人群": {"男", "女", "儿童", "通用"},
		"季节":   {"春季", "夏季", "秋季", "冬季"},
		"风格":   {"现代", "北欧", "中式", "工业"},
	}
	var base []string
	for k := range vmap {
		base = append(base, k)
	}
	sort.Strings(base)
	list := make([]model.GoodsAttribute, 0, 50)
	for i := 0; i < 50; i++ {
		key := base[i%len(base)]
		name := key
		if i >= len(base) {
			name = fmt.Sprintf("%s%d", key, i/len(base)+1)
		}
		vb, _ := json.Marshal(vmap[key])
		list = append(list, model.GoodsAttribute{
			ID: snowflake.GenID(), TenantID: tenantID, Name: name,
			Values: string(vb), Sort: i + 1, Status: 1,
		})
	}
	insertBatches(db, list)
	return list
}

/* ─────────── 商品分类（深层级，共 200 个） ─────────── */
func mockGoodsCategories(db *gorm.DB, tenantID int64) []model.GoodsCategory {
	roots := []string{"食品饮料", "日用百货", "数码家电", "服装鞋帽", "办公用品", "母婴玩具", "美妆护肤", "家居家纺"}
	list := make([]model.GoodsCategory, 0, 200)
	type node struct {
		id    int64
		depth int
	}
	var nodes []node
	for i, r := range roots {
		n := model.GoodsCategory{ID: snowflake.GenID(), TenantID: tenantID, Name: r, ParentID: 0, Sort: i + 1}
		list = append(list, n)
		nodes = append(nodes, node{n.ID, 1})
	}
	subNames := []string{"一级", "二级", "三级", "四级", "系列", "分类", "专区", "专区"}
	seq := 0
	for len(list) < 200 {
		parent := nodes[mr(len(nodes))]
		if parent.depth >= 8 {
			continue
		}
		seq++
		n := model.GoodsCategory{
			ID: snowflake.GenID(), TenantID: tenantID,
			Name:     fmt.Sprintf("%s%d%03d", pickStr(subNames), parent.depth+1, seq),
			ParentID: parent.id, Sort: seq%20 + 1,
		}
		list = append(list, n)
		nodes = append(nodes, node{n.ID, parent.depth + 1})
	}
	insertBatches(db, list)
	return list
}

/* ─────────── 商品 ─────────── */
func mockGoods(db *gorm.DB, tenantID int64, units []model.Unit, attrs []model.GoodsAttribute, cats []model.GoodsCategory, suppliers []model.Supplier, shops []model.Shop) []model.Goods {
	total := 20000
	adjs := []string{"高清", "加厚", "便携", "多功能", "家用", "商用", "大容量", "迷你", "智能", "环保", "轻奢", "经典", "升级款", "豪华", "简约"}
	nouns := []string{"水杯", "收纳盒", "数据线", "充电器", "耳机", "笔记本", "T恤", "运动鞋", "台灯", "雨伞",
		"背包", "水壶", "毛巾", "拖鞋", "拖鞋", "剃须刀", "吹风机", "电饭煲", "炒锅", "刀具",
		"洗衣液", "抽纸", "牙膏", "洗发水", "沐浴露", "面膜", "口红", "香水", "零食", "饼干",
		"坚果", "牛奶", "咖啡", "茶叶", "矿泉水", "马克笔", "文件夹", "打印纸", "订书机", "计算器"}
	brands := []string{"小米", "华为", "美的", "海尔", "格力", "苏泊尔", "九阳", "得力", "晨光", "南极人", "无印", "优衣库", "宝洁", "联合利华", "三只松鼠"}
	origins := []string{"广东深圳", "浙江义乌", "江苏苏州", "福建泉州", "山东青岛", "上海", "北京", "浙江杭州", "广东东莞", "四川成都"}
	stockShops := shops
	if len(stockShops) > 3 {
		stockShops = stockShops[:3]
	}

	goods := make([]model.Goods, 0, total)
	gu := make([]model.GoodsUnit, 0, total*2)
	gp := make([]model.GoodsPrice, 0, total*3)
	gs := make([]model.GoodsStock, 0, total*6)

	for i := 1; i <= total; i++ {
		cat := cats[mr(len(cats))]
		unit := units[mr(len(units))]
		sup := suppliers[mr(len(suppliers))]
		supID := sup.ID
		catID := model.FlexInt64(cat.ID)
		unitID := unit.ID
		fUnitID := model.FlexInt64(unit.ID)
		pp := rf(5, 500)
		multiUnit := int8(0)
		if i%5 == 0 {
			multiUnit = 1
		}
		multiSpec := int8(0)
		specVals := []string(nil)
		specGroups := "[]"
		if i%4 == 0 {
			multiSpec = 1
			specVals = []string{"S", "M", "L"}
			if i%8 == 0 {
				specVals = []string{"红", "蓝", "绿"}
			}
			groups := []map[string]interface{}{
				{"name": "规格", "has_image": false, "values": specValues(specVals)},
			}
			if b, err := json.Marshal(groups); err == nil {
				specGroups = string(b)
			}
		}
		supList, _ := json.Marshal([]int64{sup.ID})
		specKeys := []string{""}
		if multiSpec == 1 {
			specKeys = specVals
		}
		mainUnit := unit.Name
		specStr := ""
		if len(specVals) > 0 {
			specStr = pickStr(specVals)
		}
		g := model.Goods{
			ID: snowflake.GenID(), TenantID: tenantID,
			Name:       fmt.Sprintf("%s%s%d", pickStr(adjs), pickStr(nouns), i),
			Code:       fmt.Sprintf("SP%06d", i),
			Barcode:    fmt.Sprintf("69%011d", 10000000+i),
			CategoryID: &catID, UnitID: &fUnitID, SupplierID: &supID,
			Suppliers: string(supList),
			ImageURL:  "",
			MaxStock:  ri(100, 1000), MinStock: ri(5, 50),
			PurchasePrice: pp, RetailPrice: rf(pp*1.2, pp*2), WholesalePrice: rf(pp*1.05, pp*1.3),
			CostMethod:   int8(1 + i%2),
			Status:       int8(1 + i%4%2),
			Remark:       fmt.Sprintf("商品备注%d", i),
			Spec:         specStr,
			Brand:        pickStr(brands),
			Origin:       pickStr(origins),
			HasMultiUnit: multiUnit, HasMultiSpec: multiSpec,
			SalesUnit: mainUnit, PurchaseUnit: mainUnit,
			EnableStockAlert: int8(i % 3 % 2), SafetyStock: ri(5, 60),
			HasBatch: int8(i % 10 % 2), HasShelfLife: int8(i % 7 % 2), ShelfLifeDays: ri(30, 730),
			ExpiryAlert: int8(i % 6 % 2), ExpiryWarnDays: ri(7, 60),
			HasSerial:    int8(i % 11 % 2),
			InitCost:     pp,
			Images:       "[]",
			SpecGroups:   specGroups,
			PriceColumns: `["等级1","等级2"]`,
			MainUnit:     mainUnit,
			CreatedAt:    randTimeWithin(720), UpdatedAt: time.Now(),
		}
		if g.Status == 0 {
			g.Status = 1
		}
		// 单位
		factor1 := 1.0
		gu = append(gu, model.GoodsUnit{ID: snowflake.GenID(), TenantID: tenantID, GoodsID: g.ID, UnitID: &unitID, UnitName: mainUnit, Factor: factor1, IsMain: 1, Sort: 1})
		if multiUnit == 1 {
			u2 := units[mr(len(units))]
			for u2.Name == mainUnit {
				u2 = units[mr(len(units))]
			}
			u2ID := u2.ID
			gu = append(gu, model.GoodsUnit{ID: snowflake.GenID(), TenantID: tenantID, GoodsID: g.ID, UnitID: &u2ID, UnitName: u2.Name, Factor: float64(ri(2, 24)), IsMain: 0, Sort: 2})
		}
		// 价格
		rmin, rmax := g.RetailPrice, g.RetailPrice
		for _, sk := range specKeys {
			pr := model.GoodsPrice{ID: snowflake.GenID(), TenantID: tenantID, GoodsID: g.ID, UnitKey: mainUnit, SpecKey: sk,
				Code: g.Code, Barcode: g.Barcode, PurchasePrice: pp, RetailPrice: g.RetailPrice, WholesalePrice: g.WholesalePrice,
				Custom: `{"等级1":0,"等级2":0}`, Sort: 1}
			if sk != "" {
				pr.RetailPrice = rf(pp*1.2, pp*2)
				if pr.RetailPrice < rmin {
					rmin = pr.RetailPrice
				}
				if pr.RetailPrice > rmax {
					rmax = pr.RetailPrice
				}
			}
			gp = append(gp, pr)
		}
		if multiUnit == 1 {
			gp = append(gp, model.GoodsPrice{ID: snowflake.GenID(), TenantID: tenantID, GoodsID: g.ID, UnitKey: gu[len(gu)-1].UnitName, SpecKey: "",
				Code: g.Code + "-1", Barcode: "", PurchasePrice: pp, RetailPrice: rf(pp*1.2, pp*2), WholesalePrice: g.WholesalePrice, Custom: `{}`, Sort: 2})
		}
		// 库存
		totalStock := 0
		for _, sp := range stockShops {
			for _, sk := range specKeys {
				st := ri(0, 200)
				gs = append(gs, model.GoodsStock{ID: snowflake.GenID(), TenantID: tenantID, GoodsID: g.ID, ShopID: sp.ID, SpecKey: sk,
					Stock: st, MinStock: g.MinStock, SafetyStock: g.SafetyStock, MaxStock: g.MaxStock, InitCost: pp})
				totalStock += st
			}
		}
		g.TotalStock = totalStock
		g.CurrentStock = totalStock
		g.RetailMin = rmin
		g.RetailMax = rmax
		goods = append(goods, g)
	}
	insertBatches(db, goods)
	insertBatches(db, gu)
	insertBatches(db, gp)
	insertBatches(db, gs)
	return goods
}

func specValues(vals []string) []map[string]interface{} {
	out := make([]map[string]interface{}, 0, len(vals))
	for _, v := range vals {
		out = append(out, map[string]interface{}{"name": v, "image": ""})
	}
	return out
}

/* ─────────── 账户 ─────────── */
func mockAccounts(db *gorm.DB, tenantID int64, shops []model.Shop) []model.Account {
	names := []string{"现金", "微信", "支付宝", "中国银行", "工商银行", "建设银行", "农业银行", "招商银行",
		"交通银行", "平安银行", "兴业银行", "浦发银行", "中信银行", "光大银行", "民生银行", "广发银行",
		"华夏银行", "邮储银行", "备用金", "其他"}
	list := make([]model.Account, 0, len(names))
	for i, n := range names {
		t := int8(2)
		if i == 0 {
			t = 1
		} else if i <= 2 {
			t = 3
		}
		list = append(list, model.Account{
			ID: snowflake.GenID(), TenantID: tenantID, Name: n, Type: t,
			BankName: n, CardNo: fmt.Sprintf("62%015d", 600000000000+i),
			ShopID: shops[i%len(shops)].ID, Balance: rf(1000, 500000),
			Remark: fmt.Sprintf("账户-%s", n), Sort: i + 1, Status: 1, CreatedAt: randTimeWithin(720),
		})
	}
	insertBatches(db, list)
	return list
}

/* ─────────── 明细辅助 ─────────── */
type mockItemBase struct {
	GoodsID int64
	Qty     int
	Price   float64
}

func genItems(goods []model.Goods, n int) ([]mockItemBase, float64) {
	items := make([]mockItemBase, 0, n)
	var total float64
	for k := 0; k < n; k++ {
		g := goods[mr(len(goods))]
		qty := ri(1, 20)
		price := g.PurchasePrice
		if price <= 0 {
			price = rf(5, 100)
		}
		amt := float64(qty) * price
		total += amt
		items = append(items, mockItemBase{g.ID, qty, price})
	}
	return items, total
}

func orderNo(prefix string, seq int) string {
	return fmt.Sprintf("%s%s%05d", prefix, time.Now().Format("20060102"), seq)
}

/* ─────────── 进货单 ─────────── */
func mockPurchases(db *gorm.DB, tenantID int64, suppliers []model.Supplier, goods []model.Goods, salesmen []model.Salesman, accounts []model.Account, shops []model.Shop) {
	total := 10000
	list := make([]model.Purchase, 0, total)
	items := make([]model.PurchaseItem, 0, total*3)
	for i := 1; i <= total; i++ {
		sup := suppliers[mr(len(suppliers))]
		sm := salesmen[mr(len(salesmen))]
		acc := accounts[mr(len(accounts))]
		shop := shops[mr(len(shops))]
		bi, sum := genItems(goods, ri(1, 5))
		disc := float64(ri(90, 100))
		discAmt := mround(sum * disc / 100)
		freight := float64(ri(0, 100))
		deposit := 0.0
		totalAmt := mround(discAmt + freight - deposit)
		paid := mround(totalAmt * float64(ri(0, 100)) / 100)
		status := int8(ri(1, 4))
		pid := snowflake.GenID()
		list = append(list, model.Purchase{
			ID: pid, TenantID: tenantID, ShopID: shop.ID, WarehouseID: 0,
			OrderNo: orderNo("JH", i), SupplierID: sup.ID, SalesmanID: sm.ID, AccountID: acc.ID,
			BillDate: randDayWithin(365), Subtotal: mround(sum), Discount: disc, DiscountedAmount: discAmt,
			Freight: freight, DepositOffset: deposit, TotalAmount: totalAmt, PaidAmount: paid, UnpaidAmount: mround(totalAmt - paid),
			InvoiceStatus: int8(i % 2), PrintStatus: int8(i % 2), RelatedNo: "",
			Attachments: "[]", Status: status, Remark: fmt.Sprintf("进货单备注 %d", i),
			CreatedBy: 0, CreatedAt: randTimeWithin(365), UpdatedAt: time.Now(),
		})
		for _, it := range bi {
			items = append(items, model.PurchaseItem{
				ID: snowflake.GenID(), TenantID: tenantID, PurchaseID: pid, GoodsID: it.GoodsID,
				SpecKey: "", Quantity: it.Qty, UnitPrice: it.Price, Amount: mround(float64(it.Qty) * it.Price),
				Remark: "", CreatedAt: time.Now(),
			})
		}
	}
	insertBatches(db, list)
	insertBatches(db, items)
}

func mockPurchaseReturns(db *gorm.DB, tenantID int64, suppliers []model.Supplier, goods []model.Goods, salesmen []model.Salesman, accounts []model.Account, shops []model.Shop) {
	total := 5000
	list := make([]model.PurchaseReturn, 0, total)
	items := make([]model.PurchaseReturnItem, 0, total*2)
	for i := 1; i <= total; i++ {
		sup := suppliers[mr(len(suppliers))]
		sm := salesmen[mr(len(salesmen))]
		acc := accounts[mr(len(accounts))]
		shop := shops[mr(len(shops))]
		bi, sum := genItems(goods, ri(1, 3))
		amt := mround(sum)
		status := int8(ri(1, 4))
		rid := snowflake.GenID()
		list = append(list, model.PurchaseReturn{
			ID: rid, TenantID: tenantID, ShopID: shop.ID, WarehouseID: 0,
			OrderNo: orderNo("JHT", i), SupplierID: sup.ID, SalesmanID: sm.ID, AccountID: acc.ID,
			BillDate: randDayWithin(365), DepositOffset: 0, PaidAmount: amt, UnpaidAmount: 0,
			InvoiceStatus: int8(i % 2), PrintStatus: int8(i % 2), RelatedNo: "",
			Attachments: "[]", TotalAmount: -amt, RefundAmount: amt, Status: status,
			Remark: fmt.Sprintf("进货退货备注 %d", i), CreatedAt: randTimeWithin(365),
		})
		for _, it := range bi {
			items = append(items, model.PurchaseReturnItem{
				ID: snowflake.GenID(), TenantID: tenantID, PurchaseReturnID: rid, GoodsID: it.GoodsID,
				SpecKey: "", Quantity: -it.Qty, UnitPrice: it.Price, Amount: mround(-float64(it.Qty) * it.Price),
				Remark: "", CreatedAt: time.Now(),
			})
		}
	}
	insertBatches(db, list)
	insertBatches(db, items)
}

/* ─────────── 销售单 ─────────── */
func mockSales(db *gorm.DB, tenantID int64, customers []model.Customer, goods []model.Goods, salesmen []model.Salesman, accounts []model.Account, shops []model.Shop) {
	total := 20000
	list := make([]model.Sale, 0, total)
	items := make([]model.SaleItem, 0, total*3)
	for i := 1; i <= total; i++ {
		cust := customers[mr(len(customers))]
		sm := salesmen[mr(len(salesmen))]
		acc := accounts[mr(len(accounts))]
		shop := shops[mr(len(shops))]
		bi, sum := genItemsAt(goods, ri(1, 4), true)
		disc := float64(ri(90, 100))
		subtotal := mround(sum)
		totalAmt := mround(subtotal * disc / 100)
		roundOff := mround(totalAmt - float64(int(totalAmt)))
		totalAmt = float64(int(totalAmt))
		received := mround(totalAmt * float64(ri(0, 100)) / 100)
		rs := int8(0)
		if received >= totalAmt {
			rs = 2
		} else if received > 0 {
			rs = 1
		}
		sid := snowflake.GenID()
		list = append(list, model.Sale{
			ID: sid, TenantID: tenantID, ShopID: shop.ID, WarehouseID: 0,
			OrderNo: orderNo("XS", i), RelatedOrderNo: "", CustomerID: cust.ID,
			SalesmanID: sm.ID, AccountID: acc.ID, BillDate: randDayWithin(365),
			Discount: disc, Subtotal: subtotal, RoundOff: roundOff, TotalAmount: totalAmt,
			ReceivedAmount: received, UnreceivedAmount: mround(totalAmt - received), ReceiveStatus: rs,
			InvoiceStatus: int8(i % 2), PrintStatus: int8(i % 2), Attachments: "[]",
			Status: int8(1 + i%3%2), Remark: fmt.Sprintf("销售单备注 %d", i),
			CreatedAt: randTimeWithin(365), UpdatedAt: time.Now(),
		})
		for _, it := range bi {
			items = append(items, model.SaleItem{
				ID: snowflake.GenID(), TenantID: tenantID, SaleID: sid, GoodsID: it.GoodsID,
				SpecKey: "", Quantity: it.Qty, UnitPrice: it.Price, Amount: mround(float64(it.Qty) * it.Price),
				Remark: "", CreatedAt: time.Now(),
			})
		}
	}
	insertBatches(db, list)
	insertBatches(db, items)
}

func mockSalesReturns(db *gorm.DB, tenantID int64, customers []model.Customer, goods []model.Goods, salesmen []model.Salesman, accounts []model.Account, shops []model.Shop) {
	total := 20000
	list := make([]model.SalesReturn, 0, total)
	items := make([]model.SalesReturnItem, 0, total*2)
	for i := 1; i <= total; i++ {
		cust := customers[mr(len(customers))]
		sm := salesmen[mr(len(salesmen))]
		acc := accounts[mr(len(accounts))]
		shop := shops[mr(len(shops))]
		bi, sum := genItemsAt(goods, ri(1, 3), true)
		totalAmt := mround(sum)
		received := mround(totalAmt * float64(ri(0, 100)) / 100)
		rid := snowflake.GenID()
		list = append(list, model.SalesReturn{
			ID: rid, TenantID: tenantID, ShopID: shop.ID, WarehouseID: 0,
			OrderNo: orderNo("XST", i), CustomerID: cust.ID, SalesmanID: sm.ID, AccountID: acc.ID,
			BillDate: randDayWithin(365), RoundOff: 0, TotalAmount: totalAmt,
			ReceivedAmount: received, UnreceivedAmount: mround(totalAmt - received),
			PrintStatus: int8(i % 2), Attachments: "[]", Status: int8(1 + i%3%2),
			Remark: fmt.Sprintf("销售退货备注 %d", i), CreatedAt: randTimeWithin(365),
		})
		for _, it := range bi {
			items = append(items, model.SalesReturnItem{
				ID: snowflake.GenID(), TenantID: tenantID, ReturnID: rid, GoodsID: it.GoodsID,
				SpecKey: "", Quantity: it.Qty, UnitPrice: it.Price, Amount: mround(float64(it.Qty) * it.Price),
				Remark: "", CreatedAt: time.Now(),
			})
		}
	}
	insertBatches(db, list)
	insertBatches(db, items)
}

func genItemsAt(goods []model.Goods, n int, sale bool) ([]mockItemBase, float64) {
	items := make([]mockItemBase, 0, n)
	var total float64
	for k := 0; k < n; k++ {
		g := goods[mr(len(goods))]
		qty := ri(1, 10)
		price := g.RetailPrice
		if price <= 0 {
			price = rf(5, 200)
		}
		total += float64(qty) * price
		items = append(items, mockItemBase{g.ID, qty, price})
	}
	return items, total
}

/* ─────────── 盘点单 ─────────── */
func mockStockCounts(db *gorm.DB, tenantID int64, goods []model.Goods, salesmen []model.Salesman, shops []model.Shop) {
	total := 5000
	list := make([]model.StockCount, 0, total)
	items := make([]model.StockCountItem, 0, total*3)
	for i := 1; i <= total; i++ {
		sm := salesmen[mr(len(salesmen))]
		shop := shops[mr(len(shops))]
		n := ri(1, 5)
		cid := snowflake.GenID()
		book, actual := 0, 0
		for k := 0; k < n; k++ {
			g := goods[mr(len(goods))]
			bq := ri(0, 200)
			aq := bq + ri(-10, 10)
			if aq < 0 {
				aq = 0
			}
			book += bq
			actual += aq
			items = append(items, model.StockCountItem{
				ID: snowflake.GenID(), TenantID: tenantID, CountID: cid, GoodsID: g.ID,
				BookQty: bq, ActualQty: aq, DiffQty: aq - bq, Remark: "", CreatedAt: time.Now(),
			})
		}
		list = append(list, model.StockCount{
			ID: cid, TenantID: tenantID, ShopID: shop.ID, WarehouseID: 0,
			OrderNo: orderNo("PD", i), SalesmanID: sm.ID, BillDate: randDayWithin(365),
			BookQty: book, ActualQty: actual, DiffQty: actual - book,
			Status: int8(1 + i%5%2), Attachments: "[]", Remark: fmt.Sprintf("盘点备注 %d", i),
			CreatedAt: randTimeWithin(365),
		})
	}
	insertBatches(db, list)
	insertBatches(db, items)
}

/* ─────────── 收款 / 付款 ─────────── */
func mockReceipts(db *gorm.DB, tenantID int64, customers []model.Customer, salesmen []model.Salesman, accounts []model.Account, shops []model.Shop) {
	total := 5000
	list := make([]model.Receipt, 0, total)
	for i := 1; i <= total; i++ {
		cust := customers[mr(len(customers))]
		sm := salesmen[mr(len(salesmen))]
		acc := accounts[mr(len(accounts))]
		shop := shops[mr(len(shops))]
		amt := rf(100, 50000)
		list = append(list, model.Receipt{
			ID: snowflake.GenID(), TenantID: tenantID, ShopID: shop.ID,
			OrderNo: orderNo("SK", i), RelatedNo: "", Type: pickStr([]string{"直接收款", "预收款", "销售收款"}),
			CustomerID: cust.ID, SalesmanID: sm.ID, BillDate: randDayWithin(365),
			Amount: amt, DiscountAmount: rf(0, 200), DepositOffset: 0,
			AccountID: acc.ID, Attachments: "[]", Status: int8(1 + i%7%2),
			Remark: fmt.Sprintf("收款备注 %d", i), CreatedAt: randTimeWithin(365),
		})
	}
	insertBatches(db, list)
}

func mockPayments(db *gorm.DB, tenantID int64, suppliers []model.Supplier, salesmen []model.Salesman, accounts []model.Account, shops []model.Shop) {
	total := 5000
	list := make([]model.Payment, 0, total)
	for i := 1; i <= total; i++ {
		sup := suppliers[mr(len(suppliers))]
		sm := salesmen[mr(len(salesmen))]
		acc := accounts[mr(len(accounts))]
		shop := shops[mr(len(shops))]
		amt := rf(100, 50000)
		list = append(list, model.Payment{
			ID: snowflake.GenID(), TenantID: tenantID, ShopID: shop.ID,
			OrderNo: orderNo("FK", i), RelatedNo: "", Type: pickStr([]string{"直接付款", "预付款", "采购付款"}),
			SupplierID: sup.ID, SalesmanID: sm.ID, BillDate: randDayWithin(365),
			Amount: amt, DiscountAmount: rf(0, 200),
			AccountID: acc.ID, Attachments: "[]", Status: int8(1 + i%7%2),
			Remark: fmt.Sprintf("付款备注 %d", i), CreatedAt: randTimeWithin(365),
		})
	}
	insertBatches(db, list)
}

/* ─────────── 工具 ─────────── */
func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

func insertBatches(db *gorm.DB, rows interface{}) {
	switch v := rows.(type) {
	case []model.Shop:
		chunkCreate(db, v)
	case []model.Role:
		chunkCreate(db, v)
	case []model.Salesman:
		chunkCreate(db, v)
	case []model.User:
		chunkCreate(db, v)
	case []model.UserTenant:
		chunkCreate(db, v)
	case []model.PriceLevel:
		chunkCreate(db, v)
	case []model.CustomerCategory:
		chunkCreate(db, v)
	case []model.Customer:
		chunkCreate(db, v)
	case []model.CustomerAddress:
		chunkCreate(db, v)
	case []model.SupplierCategory:
		chunkCreate(db, v)
	case []model.Supplier:
		chunkCreate(db, v)
	case []model.SupplierAddress:
		chunkCreate(db, v)
	case []model.Unit:
		chunkCreate(db, v)
	case []model.GoodsAttribute:
		chunkCreate(db, v)
	case []model.GoodsCategory:
		chunkCreate(db, v)
	case []model.Goods:
		chunkCreate(db, v)
	case []model.GoodsUnit:
		chunkCreate(db, v)
	case []model.GoodsPrice:
		chunkCreate(db, v)
	case []model.GoodsStock:
		chunkCreate(db, v)
	case []model.Account:
		chunkCreate(db, v)
	case []model.Purchase:
		chunkCreate(db, v)
	case []model.PurchaseItem:
		chunkCreate(db, v)
	case []model.PurchaseReturn:
		chunkCreate(db, v)
	case []model.PurchaseReturnItem:
		chunkCreate(db, v)
	case []model.Sale:
		chunkCreate(db, v)
	case []model.SaleItem:
		chunkCreate(db, v)
	case []model.SalesReturn:
		chunkCreate(db, v)
	case []model.SalesReturnItem:
		chunkCreate(db, v)
	case []model.StockCount:
		chunkCreate(db, v)
	case []model.StockCountItem:
		chunkCreate(db, v)
	case []model.Receipt:
		chunkCreate(db, v)
	case []model.Payment:
		chunkCreate(db, v)
	default:
		_ = v
	}
}

func chunkCreate[T any](db *gorm.DB, rows []T) {
	if len(rows) == 0 {
		return
	}
	const size = 500
	for i := 0; i < len(rows); i += size {
		end := i + size
		if end > len(rows) {
			end = len(rows)
		}
		if err := db.CreateInBatches(rows[i:end], size).Error; err != nil {
			log.Printf("mock batch insert error: %v", err)
			return
		}
	}
}
