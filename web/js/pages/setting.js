// Settings module pages
const SettingPage = {
  goodsList: [
    { id: 1, name: '不锈钢螺丝 M6', code: 'SS001', category: '紧固件', unit: '个', price: 0.25, stock: 5000 },
    { id: 2, name: '十字螺丝刀', code: 'TG002', category: '工具', unit: '把', price: 15, stock: 120 },
    { id: 3, name: '电工胶带', code: 'DJ003', category: '电工材料', unit: '卷', price: 5.8, stock: 30 },
  ],
  customers: [
    { id: 1, name: '华润万家', code: 'C001', contact: '王经理', phone: '138-0000-1111', balance: 8500 },
    { id: 2, name: '沃尔玛', code: 'C002', contact: '李经理', phone: '138-0000-2222', balance: 15200 },
  ],
  suppliers: [
    { id: 1, name: '东莞五金厂', code: 'S001', contact: '张总', phone: '138-0000-3333', payable: 12500 },
    { id: 2, name: '广州包装材料', code: 'S002', contact: '刘经理', phone: '138-0000-4444', payable: 3200 },
  ],
  warehouses: [
    { id: 1, name: '主仓库', type: '普通仓库', address: '深圳市南山区', keeper: '张三' },
  ],
  staff: [
    { id: 1, name: '张三', role: '主账号', phone: '138-0000-0000', permissions: '全部权限' },
    { id: 2, name: '李四', role: '操作员', phone: '138-0000-5555', permissions: '销售、库存' },
  ],
  shops: [
    { id: 1, name: '张三五金店（总店）', address: '深圳市南山区科技园', is_main: true },
    { id: 2, name: '南山分店', address: '深圳市南山区蛇口', is_main: false },
  ],

  renderGoods() {
    const cols = [
      { key: 'code', label: '编码' }, { key: 'name', label: '商品名称' },
      { key: 'category', label: '分类' }, { key: 'unit', label: '单位' },
      { key: 'price', label: '零售价', align: 'right', render: v => '¥'+v.toFixed(2) },
      { key: 'stock', label: '库存', align: 'right' },
    ];
    return `<div class="page-head"><div class="page-title">商品管理</div><div><button class="btn btn-secondary">导入</button> <button class="btn btn-secondary">导出</button> <button class="btn btn-primary">新增商品</button></div></div>
      <div class="toolbar"><div class="toolbar-left"><input class="search-input" placeholder="搜索商品名称/编码"></div></div>
      ${Table.render(cols, this.goodsList, { actions: r => '<button class="btn btn-ghost btn-sm">编辑</button><button class="btn btn-ghost btn-sm text-danger">删除</button>' })}`;
  },

  renderCustomer() {
    const cols = [
      { key: 'code', label: '编码' }, { key: 'name', label: '客户名称' },
      { key: 'contact', label: '联系人' }, { key: 'phone', label: '电话' },
      { key: 'balance', label: '应收余额', align: 'right', render: v => '¥'+v.toLocaleString() },
    ];
    return `<div class="page-head"><div class="page-title">客户管理</div><div><button class="btn btn-secondary">导入</button> <button class="btn btn-primary">新增客户</button></div></div>
      <div class="toolbar"><div class="toolbar-left"><input class="search-input" placeholder="搜索客户名称/编码"></div></div>
      ${Table.render(cols, this.customers, { actions: r => '<button class="btn btn-ghost btn-sm">编辑</button>' })}`;
  },

  renderSupplier() {
    const cols = [
      { key: 'code', label: '编码' }, { key: 'name', label: '供应商名称' },
      { key: 'contact', label: '联系人' }, { key: 'phone', label: '电话' },
      { key: 'payable', label: '应付余额', align: 'right', render: v => '¥'+v.toLocaleString() },
    ];
    return `<div class="page-head"><div class="page-title">供应商管理</div><button class="btn btn-primary">新增供应商</button></div>
      <div class="toolbar"><div class="toolbar-left"><input class="search-input" placeholder="搜索供应商名称/编码"></div></div>
      ${Table.render(cols, this.suppliers, { actions: r => '<button class="btn btn-ghost btn-sm">编辑</button>' })}`;
  },

  renderWarehouse() {
    const cols = [
      { key: 'name', label: '仓库名称' }, { key: 'type', label: '类型' },
      { key: 'address', label: '地址' }, { key: 'keeper', label: '负责人' },
    ];
    return `<div class="page-head"><div class="page-title">仓库管理</div><button class="btn btn-primary">新增仓库</button></div>
      ${Table.render(cols, this.warehouses, { actions: r => '<button class="btn btn-ghost btn-sm">编辑</button>' })}`;
  },

  renderStaff() {
    const cols = [
      { key: 'name', label: '姓名' }, { key: 'role', label: '角色' },
      { key: 'phone', label: '手机号' }, { key: 'permissions', label: '权限' },
    ];
    return `<div class="page-head"><div class="page-title">职员管理</div><button class="btn btn-primary">邀请职员</button></div>
      ${Table.render(cols, this.staff, { actions: r => '<button class="btn btn-ghost btn-sm">编辑</button><button class="btn btn-ghost btn-sm text-danger">移除</button>' })}`;
  },

  renderShop() {
    const cols = [
      { key: 'name', label: '门店名称', render: (v, r) => v + (r.is_main ? ' <span class="badge badge-primary">总店</span>' : '') },
      { key: 'address', label: '地址' },
    ];
    return `<div class="page-head"><div class="page-title">门店管理</div><button class="btn btn-primary">新增门店</button></div>
      ${Table.render(cols, this.shops, { actions: r => '<button class="btn btn-ghost btn-sm">编辑</button>' })}`;
  },

  renderTenantInfo() {
    return `<div class="page-head"><div class="page-title">商户信息</div><button class="btn btn-secondary">编辑</button></div>
      <div class="card" style="max-width:640px">
        <div style="display:grid;grid-template-columns:130px 1fr">
          ${[['商户名称', State.tenantName],['商户类型','个体户'],['联系人','张三'],['联系电话','138-0000-0000'],['主账号（老板）','张三'],['当前套餐','单店版'],['套餐到期','2026-12-31'],['商户状态','正常']].map(r =>
            `<div style="padding:12px 0;border-bottom:1px solid var(--gray-100);font-size:14px;color:var(--gray-500)">${r[0]}</div><div style="padding:12px 0;border-bottom:1px solid var(--gray-100);font-size:14px;color:var(--gray-800)">${r[1]}</div>`
          ).join('')}
        </div>
      </div>`;
  },

  renderPlan() {
    return `<div class="page-head"><div class="page-title">套餐订阅</div></div>
      <div class="card"><div class="card-title">套餐用量</div>
        ${[['商品数','2,000','128',6],['员工账号','10','3',30],['门店数','1','1',100],['本月单据','不限','1,240',8]].map(u =>
          `<div style="display:flex;align-items:center;gap:12px;margin-bottom:12px">
            <span style="width:80px;font-size:14px;color:var(--gray-500)">${u[0]}</span>
            <div style="flex:1;height:6px;background:var(--gray-100);border-radius:3px;overflow:hidden"><div style="height:100%;width:${u[3]}%;background:${u[3]>=100?'var(--danger-500)':'var(--primary-500)'};border-radius:3px"></div></div>
            <span style="font-size:14px;color:var(--gray-600);white-space:nowrap">${u[2]} / ${u[1]}</span>
          </div>`
        ).join('')}
      </div>
      <div class="card mt-4"><div class="card-title">套餐选择</div>
        <div style="display:grid;grid-template-columns:repeat(4,1fr);gap:16px">
          ${[['免费版','¥0','100商品·1员工·1门店',false],['开单版','¥398','500商品·3员工·1门店',false],['单店版','¥998','2000商品·10员工·1门店',true],['多店版','¥1998','不限',false]].map(p =>
            `<div style="border:1px solid ${p[3]?'var(--primary-500)':'var(--gray-200)'};border-radius:var(--radius-lg);padding:20px;text-align:center;position:relative${p[3]?';background:var(--primary-50)':''}">
              ${p[3]?'<span class="badge badge-primary" style="position:absolute;top:8px;right:8px">当前</span>':''}
              <div style="font-size:17px;font-weight:600;margin-bottom:8px">${p[0]}</div>
              <div style="font-size:25px;font-weight:700;color:var(--primary-500)">${p[1]}<small style="font-size:13px;color:var(--gray-400)">/年</small></div>
              <div style="font-size:13px;color:var(--gray-500);margin:8px 0">${p[2]}</div>
              <button class="btn ${p[3]?'btn-secondary':'btn-primary'} btn-sm" style="width:100%">${p[3]?'续费':'升级'}</button>
            </div>`
          ).join('')}
        </div>
      </div>`;
  },
};
