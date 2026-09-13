// Sales module pages
const SalePage = {
  orders: [
    { id: 1, order_no: 'KHDD202609130001', customer: '华润万家', date: '2026-09-13', total: 8500, status: 3 },
    { id: 2, order_no: 'KHDD202609130002', customer: '沃尔玛', date: '2026-09-13', total: 15200, status: 1 },
  ],
  list: [
    { id: 1, order_no: 'XH202609130001', customer: '华润万家', date: '2026-09-13', total: 8500, received: 8500, status: 3 },
    { id: 2, order_no: 'XH202609130002', customer: '沃尔玛', date: '2026-09-13', total: 15200, received: 0, status: 2 },
  ],
  returns: [
    { id: 1, order_no: 'XHT202609130001', customer: '华润万家', date: '2026-09-13', total: 1200, status: 3 },
  ],

  renderOrder() {
    const cols = [
      { key: 'order_no', label: '订单编号' }, { key: 'customer', label: '客户' },
      { key: 'date', label: '下单日期' }, { key: 'total', label: '订单金额', align: 'right', render: v => '¥'+v.toLocaleString() },
      { key: 'status', label: '状态', render: v => Table.statusBadge(v) },
    ];
    return `<div class="page-head"><div class="page-title">销售订单</div><button class="btn btn-primary">新增订单</button></div>
      <div class="toolbar"><div class="toolbar-left"><input class="search-input" placeholder="搜索订单编号/客户"></div></div>
      ${Table.render(cols, this.orders, { actions: (r) => '<button class="btn btn-ghost btn-sm">查看</button>' })}`;
  },

  renderList() {
    const cols = [
      { key: 'order_no', label: '销货单号' }, { key: 'customer', label: '客户' },
      { key: 'date', label: '销货日期' }, { key: 'total', label: '销货金额', align: 'right', render: v => '¥'+v.toLocaleString() },
      { key: 'received', label: '已收金额', align: 'right', render: v => '¥'+v.toLocaleString() },
      { key: 'status', label: '状态', render: v => Table.statusBadge(v) },
    ];
    return `<div class="page-head"><div class="page-title">销货管理</div><button class="btn btn-primary">新增销货</button></div>
      <div class="toolbar"><div class="toolbar-left"><input class="search-input" placeholder="搜索单号/客户"></div></div>
      ${Table.render(cols, this.list, { actions: (r) => `<button class="btn btn-ghost btn-sm">查看</button>${r.status===2?'<button class="btn btn-primary btn-sm">审核</button>':''}` })}`;
  },

  renderReturn() {
    const cols = [
      { key: 'order_no', label: '退货单号' }, { key: 'customer', label: '客户' },
      { key: 'date', label: '退货日期' }, { key: 'total', label: '退货金额', align: 'right', render: v => '¥'+v.toLocaleString() },
      { key: 'status', label: '状态', render: v => Table.statusBadge(v) },
    ];
    return `<div class="page-head"><div class="page-title">销货退货</div><button class="btn btn-primary">新增退货</button></div>
      <div class="toolbar"><div class="toolbar-left"><input class="search-input" placeholder="搜索单号/客户"></div></div>
      ${Table.render(cols, this.returns, { actions: (r) => '<button class="btn btn-ghost btn-sm">查看</button>' })}`;
  },
};
