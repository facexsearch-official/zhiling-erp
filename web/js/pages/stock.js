// Inventory module pages
const StockPage = {
  list: [
    { id: 1, goods: '不锈钢螺丝 M6', code: 'SS001', warehouse: '主仓库', qty: 5000, cost: 0.15, min: 1000, max: 20000 },
    { id: 2, goods: '十字螺丝刀', code: 'TG002', warehouse: '主仓库', qty: 120, cost: 8.5, min: 50, max: 500 },
    { id: 3, goods: '电工胶带', code: 'DJ003', warehouse: '主仓库', qty: 30, cost: 3.2, min: 50, max: 1000 },
  ],

  renderStock() {
    const cols = [
      { key: 'goods', label: '货品名称' }, { key: 'code', label: '编码' },
      { key: 'warehouse', label: '仓库' }, { key: 'qty', label: '库存数量', align: 'right',
        render: (v, r) => `<span class="${v<=r.min?'text-danger':v>=r.max?'text-warning':''}">${v.toLocaleString()}</span>` },
      { key: 'cost', label: '成本价', align: 'right', render: v => '¥'+v.toFixed(2) },
      { key: 'min', label: '最低库存', align: 'right' },
    ];
    return `<div class="page-head"><div class="page-title">库存查询</div><div class="toolbar-right"><button class="btn btn-secondary">导出</button></div></div>
      <div class="filter-bar">
        <input class="search-input" placeholder="搜索货品名称/编码">
        <select class="form-input form-select" style="width:160px"><option value="">全部仓库</option><option>主仓库</option></select>
      </div>
      ${Table.render(cols, this.list)}`;
  },

  renderTransfer() {
    return `<div class="page-head"><div class="page-title">仓库调拨</div><button class="btn btn-primary">新增调拨</button></div>
      <div class="card"><div class="table-empty">暂无调拨记录</div></div>`;
  },

  renderOtherInbound() {
    return `<div class="page-head"><div class="page-title">其他入库</div><button class="btn btn-primary">新增入库</button></div>
      <div class="card"><div class="table-empty">暂无入库记录</div></div>`;
  },

  renderOtherOutbound() {
    return `<div class="page-head"><div class="page-title">其他出库</div><button class="btn btn-primary">新增出库</button></div>
      <div class="card"><div class="table-empty">暂无出库记录</div></div>`;
  },

  renderStockCheck() {
    return `<div class="page-head"><div class="page-title">盘点管理</div><button class="btn btn-primary">新增盘点</button></div>
      <div class="card"><div class="table-empty">暂无盘点记录</div></div>`;
  },
};
