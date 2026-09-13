// Dashboard / Home page
const HomePage = {
  render() {
    const content = document.getElementById('content');
    content.innerHTML = `
      <div class="page-head">
        <div><div class="page-title">工作台</div><div class="page-desc">欢迎回来，${State.nickname}</div></div>
      </div>
      <div class="dashboard-stats">
        <div class="stat-card">
          <div class="stat-icon" style="background:var(--primary-50);color:var(--primary-500)"><svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="9" cy="21" r="1"/><circle cx="20" cy="21" r="1"/><path d="M1 1h4l2.68 13.39a2 2 0 002 1.61h9.72a2 2 0 002-1.61L23 6H6"/></svg></div>
          <div class="stat-value">¥12,580</div>
          <div class="stat-label">今日销售额</div>
          <div class="stat-change text-success">↑ 12.5%</div>
        </div>
        <div class="stat-card">
          <div class="stat-icon" style="background:var(--success-50);color:var(--success-500)"><svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M6 2L3 6v14a2 2 0 002 2h14a2 2 0 002-2V6l-3-4z"/><line x1="3" y1="6" x2="21" y2="6"/></svg></div>
          <div class="stat-value">¥8,320</div>
          <div class="stat-label">今日采购额</div>
          <div class="stat-change text-danger">↓ 3.2%</div>
        </div>
        <div class="stat-card">
          <div class="stat-icon" style="background:var(--warning-50);color:var(--warning-500)"><svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M10.29 3.86L1.82 18a2 2 0 001.71 3h16.94a2 2 0 001.71-3L13.71 3.86a2 2 0 00-3.42 0z"/><line x1="12" y1="9" x2="12" y2="13"/><line x1="12" y1="17" x2="12.01" y2="17"/></svg></div>
          <div class="stat-value">5</div>
          <div class="stat-label">库存预警</div>
          <div class="stat-change text-warning">需及时处理</div>
        </div>
        <div class="stat-card">
          <div class="stat-icon" style="background:var(--info-50);color:var(--info-500)"><svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><line x1="12" y1="1" x2="12" y2="23"/><path d="M17 5H9.5a3.5 3.5 0 000 7h5a3.5 3.5 0 010 7H6"/></svg></div>
          <div class="stat-value">¥35,600</div>
          <div class="stat-label">应收余额</div>
          <div class="stat-change">12 笔待收</div>
        </div>
      </div>
      <div class="dashboard-grid">
        <div class="card">
          <div class="card-title">销售趋势</div>
          <div class="chart-container"><canvas id="salesChart"></canvas></div>
        </div>
        <div class="card">
          <div class="card-title">快捷操作</div>
          <div style="display:flex;flex-direction:column;gap:8px">
            <button class="btn btn-primary" style="width:100%;justify-content:flex-start" onclick="openNav('sale')"><svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M12 5v14M5 12h14"/></svg> 新增销货单</button>
            <button class="btn btn-secondary" style="width:100%;justify-content:flex-start" onclick="openNav('purchase')"><svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M12 5v14M5 12h14"/></svg> 新增进货单</button>
            <button class="btn btn-secondary" style="width:100%;justify-content:flex-start" onclick="openNav('stock')"><svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M21 16V8a2 2 0 00-1-1.73l-7-4a2 2 0 00-2 0l-7 4A2 2 0 003 8v8a2 2 0 001 1.73l7 4a2 2 0 002 0l7-4A2 2 0 0021 16z"/></svg> 查看库存</button>
            <button class="btn btn-secondary" style="width:100%;justify-content:flex-start" onclick="openNav('receipt')"><svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><line x1="12" y1="1" x2="12" y2="23"/><path d="M17 5H9.5a3.5 3.5 0 000 7h5a3.5 3.5 0 010 7H6"/></svg> 新增收款</button>
          </div>
        </div>
      </div>
    `;
    // Placeholder chart
    setTimeout(() => {
      const canvas = document.getElementById('salesChart');
      if (canvas) {
        const ctx = canvas.getContext('2d');
        ctx.fillStyle = '#f3f4f6';
        ctx.fillRect(0, 0, canvas.width, canvas.height);
        ctx.fillStyle = '#9ca3af';
        ctx.font = '14px sans-serif';
        ctx.textAlign = 'center';
        ctx.fillText('销售趋势图表（Chart.js）', canvas.width/2, canvas.height/2);
      }
    }, 100);
  },
};
