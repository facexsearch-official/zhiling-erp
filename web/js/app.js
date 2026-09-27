// Main app entry
(function() {
  // Check login
  if (!State.isLoggedIn()) {
    window.location.href = '/login.html';
    return;
  }

  // Init sidebar
  Sidebar.render();

  // Set user info
  const initial = State.nickname.charAt(0);
  document.getElementById('userAvatar').textContent = initial;
  document.getElementById('sidebarUser').innerHTML =
    '<div class="sidebar-user-avatar">' + initial + '</div>' +
    '<div class="sidebar-user-info"><div class="sidebar-user-name">' + State.nickname + '</div><div class="sidebar-user-role">' + (State.role === 1 ? '主账号' : '操作员') + '</div></div>';

  // Set tenant info
  document.getElementById('currentTenant').textContent = State.tenantName;
  document.getElementById('tenantBadge').textContent = State.tenantName.charAt(0);

  // Build tenant menu
  const tenantMenu = document.getElementById('tenantMenu');
  tenantMenu.innerHTML =
    '<div class="tenant-menu-title">切换商户</div>' +
    '<div class="tenant-option active"><span class="tenant-badge">' + State.tenantName.charAt(0) + '</span>' + State.tenantName + '<span class="role">主账号</span></div>' +
    '<div style="border-top:1px solid var(--gray-100);margin-top:4px;padding-top:4px">' +
    '<div class="tenant-option" style="color:var(--primary-600);font-weight:500" onclick="createNewTenant()">＋ 创建新商户</div>' +
    '</div>';

  // Register routes
  Router.register('home', () => { Sidebar.setActive('home'); document.getElementById('pageTitle').textContent = '工作台'; HomePage.render(); });
  Router.register('purchase-order', () => { Sidebar.setActive('purchase-order'); document.getElementById('pageTitle').textContent = '采购订单'; document.getElementById('content').innerHTML = PurchasePage.renderOrder(); });
  Router.register('purchase', () => { Sidebar.setActive('purchase'); document.getElementById('pageTitle').textContent = '进货管理'; document.getElementById('content').innerHTML = PurchasePage.renderList(); });
  Router.register('purchase-return', () => { Sidebar.setActive('purchase-return'); document.getElementById('pageTitle').textContent = '进货退货'; document.getElementById('content').innerHTML = PurchasePage.renderReturn(); });
  Router.register('sale-order', () => { Sidebar.setActive('sale-order'); document.getElementById('pageTitle').textContent = '销售订单'; document.getElementById('content').innerHTML = SalePage.renderOrder(); });
  Router.register('sale', () => { Sidebar.setActive('sale'); document.getElementById('pageTitle').textContent = '销货管理'; document.getElementById('content').innerHTML = SalePage.renderList(); });
  Router.register('sale-return', () => { Sidebar.setActive('sale-return'); document.getElementById('pageTitle').textContent = '销货退货'; document.getElementById('content').innerHTML = SalePage.renderReturn(); });
  Router.register('stock', () => { Sidebar.setActive('stock'); document.getElementById('pageTitle').textContent = '库存查询'; document.getElementById('content').innerHTML = StockPage.renderStock(); });
  Router.register('transfer', () => { Sidebar.setActive('transfer'); document.getElementById('pageTitle').textContent = '仓库调拨'; document.getElementById('content').innerHTML = StockPage.renderTransfer(); });
  Router.register('other-inbound', () => { Sidebar.setActive('other-inbound'); document.getElementById('pageTitle').textContent = '其他入库'; document.getElementById('content').innerHTML = StockPage.renderOtherInbound(); });
  Router.register('other-outbound', () => { Sidebar.setActive('other-outbound'); document.getElementById('pageTitle').textContent = '其他出库'; document.getElementById('content').innerHTML = StockPage.renderOtherOutbound(); });
  Router.register('stock-check', () => { Sidebar.setActive('stock-check'); document.getElementById('pageTitle').textContent = '盘点管理'; document.getElementById('content').innerHTML = StockPage.renderStockCheck(); });
  Router.register('receipt', () => { Sidebar.setActive('receipt'); document.getElementById('pageTitle').textContent = '收款管理'; document.getElementById('content').innerHTML = FinancePage.renderReceipt(); });
  Router.register('payment', () => { Sidebar.setActive('payment'); document.getElementById('pageTitle').textContent = '付款管理'; document.getElementById('content').innerHTML = FinancePage.renderPayment(); });
  Router.register('account', () => { Sidebar.setActive('account'); document.getElementById('pageTitle').textContent = '账户管理'; document.getElementById('content').innerHTML = FinancePage.renderAccount(); });
  Router.register('report-purchase', () => { Sidebar.setActive('report-purchase'); document.getElementById('pageTitle').textContent = '采购报表'; document.getElementById('content').innerHTML = ReportPage.renderPurchase(); });
  Router.register('report-sales', () => { Sidebar.setActive('report-sales'); document.getElementById('pageTitle').textContent = '销售报表'; document.getElementById('content').innerHTML = ReportPage.renderSales(); });
  Router.register('report-stock', () => { Sidebar.setActive('report-stock'); document.getElementById('pageTitle').textContent = '库存报表'; document.getElementById('content').innerHTML = ReportPage.renderStock(); });
  Router.register('report-finance', () => { Sidebar.setActive('report-finance'); document.getElementById('pageTitle').textContent = '财务报表'; document.getElementById('content').innerHTML = ReportPage.renderFinance(); });
  Router.register('goods', () => { Sidebar.setActive('goods'); document.getElementById('pageTitle').textContent = '商品管理'; document.getElementById('content').innerHTML = SettingPage.renderGoods(); });
  Router.register('customer', () => { Sidebar.setActive('customer'); document.getElementById('pageTitle').textContent = '客户管理'; document.getElementById('content').innerHTML = SettingPage.renderCustomer(); });
  Router.register('supplier', () => { Sidebar.setActive('supplier'); document.getElementById('pageTitle').textContent = '供应商管理'; document.getElementById('content').innerHTML = SettingPage.renderSupplier(); });
  Router.register('warehouse', () => { Sidebar.setActive('warehouse'); document.getElementById('pageTitle').textContent = '仓库管理'; document.getElementById('content').innerHTML = SettingPage.renderWarehouse(); });
  Router.register('staff', () => { Sidebar.setActive('staff'); document.getElementById('pageTitle').textContent = '职员管理'; document.getElementById('content').innerHTML = SettingPage.renderStaff(); });
  Router.register('shop', () => { Sidebar.setActive('shop'); document.getElementById('pageTitle').textContent = '门店管理'; document.getElementById('content').innerHTML = SettingPage.renderShop(); });
  Router.register('tenant-info', () => { Sidebar.setActive('tenant-info'); document.getElementById('pageTitle').textContent = '商户信息'; document.getElementById('content').innerHTML = SettingPage.renderTenantInfo(); });
  Router.register('plan', () => { Sidebar.setActive('plan'); document.getElementById('pageTitle').textContent = '套餐订阅'; document.getElementById('content').innerHTML = SettingPage.renderPlan(); });

  // Init router
  Router.init();
})();

function createNewTenant() {
  const name = prompt('请输入新商户名称：');
  if (!name) return;
  Toast.show('商户「' + name + '」已创建');
}
