// Purchase module pages
const PurchasePage = {
  orders: [
    { id: 1, order_no: 'CGDD202609130001', supplier: '东莞五金厂', date: '2026-09-13', total: 12500, status: 3, remark: '' },
    { id: 2, order_no: 'CGDD202609130002', supplier: '广州包装材料', date: '2026-09-13', total: 3200, status: 1, remark: '' },
  ],
  list: [
    { id: 1, order_no: 'GH202609130001', supplier: '东莞五金厂', date: '2026-09-13', total: 12500, paid: 12500, status: 3 },
    { id: 2, order_no: 'GH202609130002', supplier: '广州包装材料', date: '2026-09-13', total: 3200, paid: 0, status: 2 },
  ],
  returns: [
    { id: 1, order_no: 'TH202609130001', supplier: '东莞五金厂', date: '2026-09-13', total: 800, status: 3 },
  ],

  renderOrder() {
    const cols = [
      { key: 'order_no', label: '订单编号' },
      { key: 'supplier', label: '供应商' },
      { key: 'date', label: '下单日期' },
      { key: 'total', label: '订单金额', align: 'right', render: v => '¥' + v.toLocaleString() },
      { key: 'status', label: '状态', render: (v) => Table.statusBadge(v) },
    ];
    const actions = (row) => `<button class="btn btn-ghost btn-sm">查看</button><button class="btn btn-ghost btn-sm">编辑</button>`;
    return `<div class="page-head"><div class="page-title">采购订单</div><button class="btn btn-primary" onclick="PurchasePage.showCreate()">新增订单</button></div>
      <div class="toolbar"><div class="toolbar-left"><input class="search-input" placeholder="搜索订单编号/供应商"></div></div>
      ${Table.render(cols, this.orders, { actions })}`;
  },

  renderList() {
    const cols = [
      { key: 'order_no', label: '进货单号' },
      { key: 'supplier', label: '供应商' },
      { key: 'date', label: '进货日期' },
      { key: 'total', label: '进货金额', align: 'right', render: v => '¥' + v.toLocaleString() },
      { key: 'paid', label: '已付金额', align: 'right', render: v => '¥' + v.toLocaleString() },
      { key: 'status', label: '状态', render: (v) => Table.statusBadge(v) },
    ];
    const actions = (row) => `<button class="btn btn-ghost btn-sm">查看</button>${row.status===2?'<button class="btn btn-primary btn-sm">审核</button>':''}`;
    return `<div class="page-head"><div class="page-title">进货管理</div><button class="btn btn-primary" onclick="PurchasePage.showCreate()">新增进货</button></div>
      <div class="toolbar"><div class="toolbar-left"><input class="search-input" placeholder="搜索单号/供应商"></div></div>
      ${Table.render(cols, this.list, { actions })}`;
  },

  renderReturn() {
    const cols = [
      { key: 'order_no', label: '退货单号' },
      { key: 'supplier', label: '供应商' },
      { key: 'date', label: '退货日期' },
      { key: 'total', label: '退货金额', align: 'right', render: v => '¥' + v.toLocaleString() },
      { key: 'status', label: '状态', render: (v) => Table.statusBadge(v) },
    ];
    const actions = (row) => `<button class="btn btn-ghost btn-sm">查看</button>`;
    return `<div class="page-head"><div class="page-title">进货退货</div><button class="btn btn-primary">新增退货</button></div>
      <div class="toolbar"><div class="toolbar-left"><input class="search-input" placeholder="搜索单号/供应商"></div></div>
      ${Table.render(cols, this.returns, { actions })}`;
  },

  showCreate() {
    Modal.show('新增进货单', `
      <div class="form-row">
        ${Form.renderSelect('供应商', 'supplier_id', [{value:'1',label:'东莞五金厂'},{value:'2',label:'广州包装材料'}], {required:true})}
        ${Form.renderInput('进货日期', 'bill_date', {type:'date', value:'2026-09-13', required:true})}
      </div>
      <div class="form-row">
        ${Form.renderSelect('仓库', 'warehouse_id', [{value:'1',label:'主仓库'}])}
        ${Form.renderInput('备注', 'remark', {placeholder:'选填'})}
      </div>
    `, () => { Toast.show('进货单已创建'); Modal.hide(); });
  },
};
