// Report module pages
const ReportPage = {
  renderPurchase() {
    return `<div class="page-head"><div class="page-title">采购报表</div></div>
      <div class="filter-bar">
        <label>日期：</label><input class="form-input" type="date" value="2026-09-01" style="width:160px">
        <span>~</span><input class="form-input" type="date" value="2026-09-30" style="width:160px">
        <button class="btn btn-primary">查询</button>
      </div>
      <div class="dashboard-stats">
        <div class="stat-card"><div class="stat-label">采购总额</div><div class="stat-value">¥45,800</div></div>
        <div class="stat-card"><div class="stat-label">采购单数</div><div class="stat-value">28</div></div>
        <div class="stat-card"><div class="stat-label">供应商数</div><div class="stat-value">12</div></div>
        <div class="stat-card"><div class="stat-label">已付金额</div><div class="stat-value">¥38,200</div></div>
      </div>
      <div class="card"><div class="card-title">采购趋势</div><div class="chart-container" style="height:200px;display:flex;align-items:center;justify-content:center;color:var(--gray-400)">图表区域（Chart.js）</div></div>`;
  },

  renderSales() {
    return `<div class="page-head"><div class="page-title">销售报表</div></div>
      <div class="filter-bar">
        <label>日期：</label><input class="form-input" type="date" value="2026-09-01" style="width:160px">
        <span>~</span><input class="form-input" type="date" value="2026-09-30" style="width:160px">
        <button class="btn btn-primary">查询</button>
      </div>
      <div class="dashboard-stats">
        <div class="stat-card"><div class="stat-label">销售总额</div><div class="stat-value">¥68,500</div></div>
        <div class="stat-card"><div class="stat-label">销售单数</div><div class="stat-value">45</div></div>
        <div class="stat-card"><div class="stat-label">客户数</div><div class="stat-value">23</div></div>
        <div class="stat-card"><div class="stat-label">毛利润</div><div class="stat-value text-success">¥18,200</div></div>
      </div>`;
  },

  renderStock() {
    return `<div class="page-head"><div class="page-title">库存报表</div></div>
      <div class="dashboard-stats">
        <div class="stat-card"><div class="stat-label">库存品种</div><div class="stat-value">128</div></div>
        <div class="stat-card"><div class="stat-label">库存总值</div><div class="stat-value">¥86,400</div></div>
        <div class="stat-card"><div class="stat-label">预警品种</div><div class="stat-value text-danger">5</div></div>
        <div class="stat-card"><div class="stat-label">仓库数</div><div class="stat-value">2</div></div>
      </div>`;
  },

  renderFinance() {
    return `<div class="page-head"><div class="page-title">财务报表</div></div>
      <div class="dashboard-stats">
        <div class="stat-card"><div class="stat-label">应收余额</div><div class="stat-value text-warning">¥35,600</div></div>
        <div class="stat-card"><div class="stat-label">应付余额</div><div class="stat-value text-danger">¥22,300</div></div>
        <div class="stat-card"><div class="stat-label">现金余额</div><div class="stat-value">¥25,600</div></div>
        <div class="stat-card"><div class="stat-label">银行余额</div><div class="stat-value">¥158,000</div></div>
      </div>`;
  },
};
