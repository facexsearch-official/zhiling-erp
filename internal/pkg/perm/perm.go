package perm

import (
	"encoding/json"
	"strings"

	"pisa_server/internal/model"
	"pisa_server/internal/pkg/response"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// 权限 key 常量（与前端 PERM_TREE 保持一致）
const (
	GoodsView = "goods.goods.view"
	GoodsAdd  = "goods.goods.add"
	GoodsEdit = "goods.goods.edit"
	GoodsDel  = "goods.goods.del"

	UnitView = "goods.unit.view"
	UnitAdd  = "goods.unit.add"
	UnitEdit = "goods.unit.edit"
	UnitDel  = "goods.unit.del"

	AttrView = "goods.attr.view"
	AttrAdd  = "goods.attr.add"
	AttrEdit = "goods.attr.edit"
	AttrDel  = "goods.attr.del"

	GoodsPriceView = "goods.price.view"
	GoodsPriceEdit = "goods.price.edit"

	ComboView = "goods.combo.view"
	ComboAdd  = "goods.combo.add"
	ComboDel  = "goods.combo.del"

	CustomerView = "customer.customer.view"
	CustomerAdd  = "customer.customer.add"
	CustomerEdit = "customer.customer.edit"
	CustomerDel  = "customer.customer.del"

	CustomerLevelView = "customer.level.view"
	CustomerLevelAdd  = "customer.level.add"
	CustomerLevelEdit = "customer.level.edit"
	CustomerLevelDel  = "customer.level.del"

	CustomerQuoteView = "customer.quote.view"
	CustomerQuoteEdit = "customer.quote.edit"

	SaleView = "sale.sale.view"
	SaleAdd  = "sale.sale.add"
	SaleVoid = "sale.sale.void"

	SaleOrderView = "sale.order.view"
	SaleOrderAdd  = "sale.order.add"
	SaleOrderVoid = "sale.order.void"

	SaleReturnView = "sale.return.view"
	SaleReturnAdd  = "sale.return.add"
	SaleReturnVoid = "sale.return.void"

	SaleQuoteView = "sale.quote.view"
	SaleQuoteAdd  = "sale.quote.add"
	SaleQuoteVoid = "sale.quote.void"

	PurchaseView = "purchase.purchase.view"
	PurchaseAdd  = "purchase.purchase.add"
	PurchaseEdit = "purchase.purchase.edit"
	PurchaseVoid = "purchase.purchase.void"

	PurchaseOrderView = "purchase.order.view"
	PurchaseOrderAdd  = "purchase.order.add"
	PurchaseOrderVoid = "purchase.order.void"

	PurchaseReturnView = "purchase.return.view"
	PurchaseReturnAdd  = "purchase.return.add"
	PurchaseReturnVoid = "purchase.return.void"

	SupplierView = "purchase.supplier.view"
	SupplierAdd  = "purchase.supplier.add"
	SupplierEdit = "purchase.supplier.edit"
	SupplierDel  = "purchase.supplier.del"

	StockCountView = "stock.count.view"
	StockCountAdd  = "stock.count.add"
	StockCountVoid = "stock.count.void"

	AssemblyView = "stock.assembly.view"
	AssemblyAdd  = "stock.assembly.add"
	AssemblyEdit = "stock.assembly.edit"
	AssemblyVoid = "stock.assembly.void"

	SplitView = "stock.split.view"
	SplitAdd  = "stock.split.add"
	SplitEdit = "stock.split.edit"
	SplitVoid = "stock.split.void"

	StockQueryView = "stock.query.view"
	StockQueryCost = "stock.query.cost"

	StockAlertView  = "stock.alert.view"
	StockBatchView  = "stock.batch.view"
	StockExpiryView = "stock.expiry.view"

	AccountView = "funds.account.normal.view"
	AccountAdd  = "funds.account.normal.add"
	AccountEdit = "funds.account.normal.edit"
	AccountDel  = "funds.account.normal.del"

	ReceiptView = "funds.receipt.view"
	ReceiptAdd  = "funds.receipt.add"
	ReceiptEdit = "funds.receipt.edit"
	ReceiptVoid = "funds.receipt.void"

	PaymentView = "funds.payment.view"
	PaymentAdd  = "funds.payment.add"
	PaymentEdit = "funds.payment.edit"
	PaymentVoid = "funds.payment.void"

	IncomeView = "funds.income.view"
	IncomeAdd  = "funds.income.add"
	IncomeEdit = "funds.income.edit"
	IncomeVoid = "funds.income.void"

	ExpenseView = "funds.expense.view"
	ExpenseAdd  = "funds.expense.add"
	ExpenseEdit = "funds.expense.edit"
	ExpenseVoid = "funds.expense.void"

	IncomeTypeView = "funds.incometype.view"
	IncomeTypeAdd  = "funds.incometype.add"
	IncomeTypeEdit = "funds.incometype.edit"
	IncomeTypeDel  = "funds.incometype.del"

	TransferView = "funds.transfer.view"
	TransferAdd  = "funds.transfer.add"

	CustomerReconView = "funds.customerrecon.view"
	SupplierReconView = "funds.supplierrecon.view"
	CashflowView      = "funds.cashflow.view"

	AnalysisSalesView    = "analysis.sales.view"
	AnalysisHotView      = "analysis.hot.view"
	AnalysisStaffView    = "analysis.staff.view"
	AnalysisPurchaseView = "analysis.purchase.view"
	AnalysisStockView    = "analysis.stock.view"
	AnalysisProfitView   = "analysis.profit.view"

	TenantView = "settings.tenant.view"
	TenantEdit = "settings.tenant.edit"

	ShopView = "settings.shop.view"
	ShopAdd  = "settings.shop.add"
	ShopEdit = "settings.shop.edit"
	ShopDel  = "settings.shop.del"

	WarehouseView = "settings.warehouse.view"
	WarehouseAdd  = "settings.warehouse.add"
	WarehouseEdit = "settings.warehouse.edit"
	WarehouseDel  = "settings.warehouse.del"

	StaffView = "settings.staff.view"
	StaffAdd  = "settings.staff.add"
	StaffEdit = "settings.staff.edit"
	StaffDel  = "settings.staff.del"

	RoleView = "settings.role.view"
	RoleAdd  = "settings.role.add"
	RoleEdit = "settings.role.edit"
	RoleDel  = "settings.role.del"

	SystemView = "settings.system.view"
	SystemEdit = "settings.system.edit"

	PrintView = "settings.print.view"
	PrintEdit = "settings.print.edit"
)

// Set 权限集合（支持 * 与 a.b.* 通配）
type Set struct {
	all bool
	m   map[string]bool
}

func newSet() *Set { return &Set{m: map[string]bool{}} }

// AllSet 返回拥有全部权限的集合
func AllSet() *Set { return &Set{all: true, m: map[string]bool{"*": true}} }

// Parse 解析权限 JSON
func Parse(raw string) *Set {
	s := newSet()
	if strings.TrimSpace(raw) == "" {
		return s
	}
	var obj map[string]interface{}
	if err := json.Unmarshal([]byte(raw), &obj); err != nil {
		return s
	}
	for k, v := range obj {
		if b, ok := v.(bool); ok {
			if !b {
				continue
			}
		}
		s.m[k] = true
		if k == "*" {
			s.all = true
		}
	}
	return s
}

// Allowed 判断是否拥有某权限
func (s *Set) Allowed(key string) bool {
	if s == nil || key == "" {
		return false
	}
	if s.all || s.m["*"] || s.m[key] {
		return true
	}
	parts := strings.Split(key, ".")
	for i := len(parts) - 1; i >= 1; i-- {
		if s.m[strings.Join(parts[:i], ".")+".*"] {
			return true
		}
	}
	return false
}

// Raw 返回原始 map（含通配），用于下发前端
func (s *Set) Raw() map[string]bool {
	if s == nil {
		return map[string]bool{}
	}
	out := make(map[string]bool, len(s.m))
	for k, v := range s.m {
		out[k] = v
	}
	return out
}

func loadMember(db *gorm.DB, userID, tenantID int64) *model.UserTenant {
	var ut model.UserTenant
	if err := db.Where("user_id = ? AND tenant_id = ? AND status = 1", userID, tenantID).First(&ut).Error; err != nil {
		return nil
	}
	return &ut
}

// Load 解析用户在商户下的功能权限
// 规则：主账号/超级管理员/未绑定自定义角色（RoleID=0）→ 全部权限
func Load(db *gorm.DB, userID, tenantID int64) *Set {
	if db == nil {
		return AllSet()
	}
	ut := loadMember(db, userID, tenantID)
	if ut == nil {
		return newSet()
	}
	if ut.IsOwner == 1 || ut.Role == 1 || ut.RoleID == 0 {
		return AllSet()
	}
	var role model.Role
	if err := db.Where("id = ? AND tenant_id = ?", ut.RoleID, tenantID).First(&role).Error; err != nil {
		return newSet()
	}
	return Parse(role.Permissions)
}

// LoadSensitive 解析用户在商户下的敏感数据权限
func LoadSensitive(db *gorm.DB, userID, tenantID int64) *Set {
	if db == nil {
		return AllSet()
	}
	ut := loadMember(db, userID, tenantID)
	if ut == nil {
		return newSet()
	}
	if ut.IsOwner == 1 || ut.Role == 1 || ut.RoleID == 0 {
		return AllSet()
	}
	var role model.Role
	if err := db.Where("id = ? AND tenant_id = ?", ut.RoleID, tenantID).First(&role).Error; err != nil {
		return newSet()
	}
	return Parse(role.SensitiveData)
}

const ctxKey = "perm_set"
const sensKey = "perm_sensitive"

// FromContext 从请求上下文获取（带缓存）功能权限集合
func FromContext(c *gin.Context, db *gorm.DB) *Set {
	if v, ok := c.Get(ctxKey); ok {
		if s, ok := v.(*Set); ok {
			return s
		}
	}
	s := Load(db, c.GetInt64("user_id"), c.GetInt64("tenant_id"))
	c.Set(ctxKey, s)
	return s
}

// SensitiveFromContext 从请求上下文获取（带缓存）敏感数据权限集合
func SensitiveFromContext(c *gin.Context, db *gorm.DB) *Set {
	if v, ok := c.Get(sensKey); ok {
		if s, ok := v.(*Set); ok {
			return s
		}
	}
	s := LoadSensitive(db, c.GetInt64("user_id"), c.GetInt64("tenant_id"))
	c.Set(sensKey, s)
	return s
}

// Routes 路由 → 权限 key 映射，key 为 "METHOD /api/shop/xxx"
type Routes map[string]string

// Enforce 按路由表校验权限；未登记的路由放行
func Enforce(db *gorm.DB, routes Routes) gin.HandlerFunc {
	return func(c *gin.Context) {
		key, ok := routes[c.Request.Method+" "+c.FullPath()]
		if !ok {
			c.Next()
			return
		}
		if !FromContext(c, db).Allowed(key) {
			response.Forbidden(c, "无操作权限")
			c.Abort()
			return
		}
		c.Next()
	}
}
