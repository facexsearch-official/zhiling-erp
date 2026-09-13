// Finance module pages
const FinancePage = {
  accounts: [
    { id: 1, name: '现金账户', type: '现金', balance: 25600 },
    { id: 2, name: '工商银行', type: '银行', balance: 158000 },
  ],
  receipts: [
    { id: 1, order_no: 'SK202609130001', customer: '华润万家', date: '2026-09-13', amount: 8500, status: 3 },
  ],
  payments: [
    { id: 1, order_no: 'FK202609130001', supplier: '东莞五金厂', date: '2026-09-13', amount: 12500, status: 3 },
  ],

  renderReceipt() {
    const cols = [
      { key: 'order_no', label: '收款单号' }, { key: 'customer', label: '客户' },
      { key: 'date', label: '收款日期' }, { key: 'amount', label: '收款金额', align: 'right', render: v => '¥'+v.toLocaleString() },
      { key: 'status', label: '状态', render: v => Table.statusBadge(v) },
    ];
    return `<div class="page-head"><div class="page-title">收款管理</div><button class="btn btn-primary">新增收款</button></div>
      ${Table.render(cols, this.receipts, { actions: r => '<button class="btn btn-ghost btn-sm">查看</button>' })}`;
  },

  renderPayment() {
    const cols = [
      { key: 'order_no', label: '付款单号' }, { key: 'supplier', label: '供应商' },
      { key: 'date', label: '付款日期' }, { key: 'amount', label: '付款金额', align: 'right', render: v => '¥'+v.toLocaleString() },
      { key: 'status', label: '状态', render: v => Table.statusBadge(v) },
    ];
    return `<div class="page-head"><div class="page-title">付款管理</div><button class="btn btn-primary">新增付款</button></div>
      ${Table.render(cols, this.payments, { actions: r => '<button class="btn btn-ghost btn-sm">查看</button>' })}`;
  },

  renderAccount() {
    const cols = [
      { key: 'name', label: '账户名称' }, { key: 'type', label: '类型' },
      { key: 'balance', label: '余额', align: 'right', render: v => '¥'+v.toLocaleString() },
    ];
    return `<div class="page-head"><div class="page-title">账户管理</div><button class="btn btn-primary">新增账户</button></div>
      ${Table.render(cols, this.accounts, { actions: r => '<button class="btn btn-ghost btn-sm">编辑</button>' })}`;
  },
};
