/* ============================================================
   PISA App Shell — shared chrome
   Renders: topbar (brand/search/tenant/actions) · sidebar + flyout
            · tabbar mount · unified navigation
   ============================================================ */
(function(){
  'use strict';

  /* ── Icons ─────────────────────────────────────────────── */
  var S = 'class="nav-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.7" stroke-linecap="round" stroke-linejoin="round"';
  var IC = {
    home:'<svg '+S+'><path d="M3 9.5 12 3l9 6.5V20a1 1 0 0 1-1 1H4a1 1 0 0 1-1-1z"/><path d="M9 21v-7h6v7"/></svg>',
    goods:'<svg '+S+'><path d="M21 8a2 2 0 0 0-1-1.73l-7-4a2 2 0 0 0-2 0l-7 4A2 2 0 0 0 3 8v8a2 2 0 0 0 1 1.73l7 4a2 2 0 0 0 2 0l7-4A2 2 0 0 0 21 16Z"/><path d="m3.3 7 8.7 5 8.7-5"/><path d="M12 22V12"/></svg>',
    sales:'<svg '+S+'><path d="M4 2v20l2-1 2 1 2-1 2 1 2-1 2 1 2-1 2 1V2l-2 1-2-1-2 1-2-1-2 1-2-1-2 1Z"/><path d="M14 8H8"/><path d="M16 12H8"/><path d="M13 16H8"/></svg>',
    customer:'<svg '+S+'><path d="M16 21v-2a4 4 0 0 0-4-4H6a4 4 0 0 0-4 4v2"/><circle cx="9" cy="7" r="4"/><path d="M22 21v-2a4 4 0 0 0-3-3.87"/><path d="M16 3.13a4 4 0 0 1 0 7.75"/></svg>',
    purchase:'<svg '+S+'><path d="M14 18V6a2 2 0 0 0-2-2H4a2 2 0 0 0-2 2v11a1 1 0 0 0 1 1h2"/><path d="M15 18H9"/><path d="M19 18h2a1 1 0 0 0 1-1v-3.65a1 1 0 0 0-.22-.62l-3.48-4.35A1 1 0 0 0 17.52 8H14"/><circle cx="17" cy="18" r="2"/><circle cx="7" cy="18" r="2"/></svg>',
    stock:'<svg '+S+'><path d="M22 8.35V20a2 2 0 0 1-2 2H4a2 2 0 0 1-2-2V8.35A2 2 0 0 1 3.26 6.5l8-3.2a2 2 0 0 1 1.48 0l8 3.2A2 2 0 0 1 22 8.35Z"/><path d="M6 18h12"/><path d="M6 14h12"/></svg>',
    funds:'<svg '+S+'><path d="M19 7V4a1 1 0 0 0-1-1H5a2 2 0 0 0 0 4h15a1 1 0 0 1 1 1v4h-3a2 2 0 0 0 0 4h3a1 1 0 0 0 1-1v-2a1 1 0 0 0-1-1"/><path d="M3 5v14a2 2 0 0 0 2 2h15a1 1 0 0 0 1-1v-4"/></svg>',
    analysis:'<svg '+S+'><path d="M3 3v16a2 2 0 0 0 2 2h16"/><path d="M18 17V9"/><path d="M13 17V5"/><path d="M8 17v-3"/></svg>',
    settings:'<svg '+S+'><circle cx="12" cy="12" r="3"/><path d="M19.4 15a1.7 1.7 0 0 0 .3 1.8l.1.1a2 2 0 1 1-2.8 2.8l-.1-.1a1.7 1.7 0 0 0-2.9 1.2V21a2 2 0 1 1-4 0v-.1A1.7 1.7 0 0 0 7 19.4l-.1.1a2 2 0 1 1-2.8-2.8l.1-.1A1.7 1.7 0 0 0 3 13.7a2 2 0 1 1 0-4h.1A1.7 1.7 0 0 0 4.6 7l-.1-.1a2 2 0 1 1 2.8-2.8l.1.1A1.7 1.7 0 0 0 10 3a2 2 0 1 1 4 0v.1a1.7 1.7 0 0 0 2.9 1.2l.1-.1a2 2 0 1 1 2.8 2.8l-.1.1a1.7 1.7 0 0 0 1.2 2.9 2 2 0 1 1 0 4h-.1a1.7 1.7 0 0 0-1.5 1z"/></svg>',
    search:'<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="11" cy="11" r="8"/><line x1="21" y1="21" x2="16.65" y2="16.65"/></svg>',
    bell:'<svg viewBox="0 0 24 24"><path d="M18 8A6 6 0 0 0 6 8c0 7-3 9-3 9h18s-3-2-3-9"/><path d="M13.7 21a2 2 0 0 1-3.4 0"/></svg>',
    chevron:'<svg class="ts-arrow" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="6 9 12 15 18 9"/></svg>',
    guide:'<svg viewBox="0 0 24 24"><path d="M6 3v12"/><circle cx="18" cy="6" r="3"/><circle cx="6" cy="18" r="3"/><path d="M18 9a9 9 0 0 1-9 9"/></svg>',
    box:'<svg viewBox="0 0 24 24"><path d="m7.5 4.27 9 5.15"/><path d="M21 8a2 2 0 0 0-1-1.73l-7-4a2 2 0 0 0-2 0l-7 4A2 2 0 0 0 3 8v8a2 2 0 0 0 1 1.73l7 4a2 2 0 0 0 2 0l7-4A2 2 0 0 0 21 16Z"/><path d="m3.3 7 8.7 5 8.7-5"/><path d="M12 22V12"/></svg>',
    panel:'<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8"><rect x="3" y="4" width="18" height="16" rx="2"/><line x1="9" y1="4" x2="9" y2="20"/></svg>',
    panelR:'<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8"><rect x="3" y="4" width="18" height="16" rx="2"/><line x1="15" y1="4" x2="15" y2="20"/></svg>',
    x:'<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg>'
  };

  /* ── Navigation tree ───────────────────────────────────── */
  var NAV = [
    {key:'home', label:'首页', icon:IC.home, href:'/'},

    {key:'goods', label:'商品', icon:IC.goods, cols:[
      {title:'商品', items:[
        {key:'goods:list',  label:'商品列表', tabLabel:'商品列表', href:'/goods.html'},
        {key:'goods:spec',  label:'规格管理', href:'/spec.html'},
        {key:'goods:unit',  label:'单位管理', href:'/unit.html'}
      ]}
    ]},

    {key:'customer', label:'客户', icon:IC.customer, cols:[
      {title:'客户', items:[
        {key:'customer:list',  label:'客户',     tabLabel:'客户列表', href:'/customer.html'},
        {key:'customer:level', label:'价格等级', tabLabel:'价格等级', href:'/#customer:level'}
      ]}
    ]},

    {key:'purchase', label:'进货', icon:IC.purchase, cols:[
      {title:'进货', items:[
        {key:'purchase:supplier', label:'供应商',   href:'/supplier.html'},
        {key:'purchase:new',      label:'进货',     tabLabel:'进货单列表', href:'/#purchase:new'},
        {key:'purchase:return',   label:'进货退货', tabLabel:'进货退货列表', href:'/#purchase:return'}
      ]}
    ]},

    {key:'sales', label:'销售', icon:IC.sales, cols:[
      {title:'销售', items:[
        {key:'sales:new',    label:'销售',     tabLabel:'销售单列表',     href:'/#sales:new'},
        {key:'sales:return', label:'销售退货', tabLabel:'销售退货单列表', href:'/#sales:return'}
      ]}
    ]},

    {key:'stock', label:'库存', icon:IC.stock, cols:[
      {title:'库存', items:[
        {key:'stock:take',      label:'盘点',     tabLabel:'盘点单列表', href:'/#stock:take'}
      ]},
      {title:'查询', items:[
        {key:'stock:query',  label:'库存查询', tabLabel:'库存查询', href:'/#stock:query'},
        {key:'stock:alert',  label:'库存预警', tabLabel:'库存预警', href:'/#stock:alert'}
      ]}
    ]},

    {key:'funds', label:'资金', icon:IC.funds, cols:[
      {title:'账户', items:[
        {key:'funds:account',  label:'账户概览', tabLabel:'账户概览', href:'/#funds:account'}
      ]},
      {title:'收支', items:[
        {key:'funds:receipt', label:'收款',     tabLabel:'收款单列表', href:'/#funds:receipt'},
        {key:'funds:payment', label:'付款',     tabLabel:'付款单列表', href:'/#funds:payment'}
      ]},
      {title:'对账', items:[
        {key:'funds:cashflow',       label:'资金流水',   tabLabel:'资金流水',   href:'/#funds:cashflow'}
      ]}
    ]},

    {key:'analysis', label:'分析', icon:IC.analysis, cols:[
      {title:'销售分析', items:[
        {key:'analysis:sales', label:'销售统计', tabLabel:'销售统计', href:'/#analysis:sales'}
      ]},
      {title:'库存分析', items:[
        {key:'analysis:stock',    label:'库存统计', tabLabel:'库存统计', href:'/#analysis:stock'}
      ]}
    ]}
  ];

  /* ── helpers ───────────────────────────────────────────── */
  // 顶栏「设置」下拉项（设置不再作为侧边栏模块）
  var TOPBAR_SETTINGS = [
    {key:'settings:staff', label:'员工管理'},
    {key:'settings:role',  label:'角色权限'}
  ];
  // 导航项 → 所需「查看」权限 key
  var NAV_PERM = {
    'goods:list':'goods.goods.view','goods:spec':'goods.spec.view','goods:unit':'goods.unit.view',
    'goods:attr':'goods.attr.view','goods:price':'goods.price.view','goods:combo':'goods.combo.view',
    'customer:list':'customer.customer.view','customer:level':'customer.level.view','customer:quote':'customer.quote.view',
    'purchase:supplier':'purchase.supplier.view','purchase:new':'purchase.purchase.view',
    'purchase:order':'purchase.order.view','purchase:return':'purchase.return.view',
    'sales:new':'sale.sale.view','sales:order':'sale.order.view','sales:return':'sale.return.view',
    'sales:quote':'sale.quote.view','sales:commission':'analysis.staff.view',
    'stock:take':'stock.count.view','stock:assembly':'stock.assembly.view','stock:split':'stock.split.view',
    'stock:query':'stock.query.view','stock:alert':'stock.alert.view','stock:batch':'stock.batch.view','stock:expiry':'stock.expiry.view',
    'funds:account':'funds.account.normal.view','funds:transfer':'funds.transfer.view',
    'funds:receipt':'funds.receipt.view','funds:payment':'funds.payment.view',
    'funds:income':'funds.income.view','funds:expense':'funds.expense.view',
    'funds:recon-customer':'funds.customerrecon.view','funds:recon-supplier':'funds.supplierrecon.view','funds:cashflow':'funds.cashflow.view',
    'analysis:sales':'analysis.sales.view','analysis:hot':'analysis.hot.view','analysis:staff':'analysis.staff.view',
    'analysis:purchase':'analysis.purchase.view','analysis:stock':'analysis.stock.view','analysis:profit':'analysis.profit.view',
    'settings:shop':'settings.tenant.view','settings:shops':'settings.shop.view',
    'settings:staff':'settings.staff.view','settings:role':'settings.role.view',
    'settings:system':'settings.system.view','settings:points':'settings.points.view',
    'settings:print':'settings.print.view'
  };
  function navAllowed(key){
    var p = NAV_PERM[key];
    if(!p) return true;
    return (typeof window.can !== 'function') ? true : window.can(p);
  }

  function esc(s){
    return String(s == null ? '' : s)
      .replace(/&/g,'&amp;').replace(/</g,'&lt;').replace(/>/g,'&gt;')
      .replace(/"/g,'&quot;').replace(/'/g,'&#39;');
  }
  function byId(id){ return document.getElementById(id); }
  function ls(k, d){ try{ return JSON.parse(localStorage.getItem(k) || d); }catch(e){ return JSON.parse(d); } }

  function eachItem(fn){
    NAV.forEach(function(m){
      if(!m.cols) return;
      m.cols.forEach(function(c){ c.items.forEach(function(it){ fn(it, m, c); }); });
    });
  }
  function findItem(key){
    var found = null;
    eachItem(function(it, m){ if(it.key === key) found = {item:it, mod:m}; });
    if(!found){
      NAV.forEach(function(m){ if(m.key === key) found = {item:m, mod:m}; });
    }
    return found;
  }
  function moduleOf(key){
    if(key === 'home') return null;
    var f = findItem(key);
    return f ? f.mod : null;
  }

  var PATH_KEY = {
    '/':'home', '/index.html':'home',
    '/goods.html':'goods:list',
    '/spec.html':'goods:spec',
    '/unit.html':'goods:unit',
    '/attr.html':'goods:attr',
    '/price.html':'goods:price',
    '/customer.html':'customer:list',
    '/supplier.html':'purchase:supplier',
    '/warehouse.html':'stock:warehouse',
    '/account.html':'funds:account'
  };
  // 无对应导航项的独立页面标签
  var PAGE_LABEL = { 'stock:warehouse':'仓库管理' };
  function currentKey(){
    var h = location.hash.slice(1);
    if(h){
      try{ h = decodeURIComponent(h); }catch(e){}
      if(h) return h;
    }
    var p = location.pathname.replace(/\/$/,'') || '/';
    if(PATH_KEY[p]) return PATH_KEY[p];
    return 'home';
  }

  /* ── Toast ─────────────────────────────────────────────── */
  function toast(msg){
    if(typeof window.showToast === 'function'){ window.showToast(msg); return; }
    var t = byId('pisaToast');
    if(!t){
      t = document.createElement('div');
      t.id = 'pisaToast';
      t.style.cssText = 'position:fixed;bottom:28px;left:50%;transform:translateX(-50%) translateY(10px);'
        + 'background:#1E293B;color:#fff;font-size:14px;padding:10px 20px;border-radius:8px;'
        + 'box-shadow:0 12px 32px rgba(15,23,42,.24);opacity:0;transition:opacity .2s,transform .2s;z-index:3000';
      document.body.appendChild(t);
    }
    t.textContent = msg;
    requestAnimationFrame(function(){ t.style.opacity = '1'; t.style.transform = 'translateX(-50%) translateY(0)'; });
    clearTimeout(t._tm);
    t._tm = setTimeout(function(){ t.style.opacity = '0'; t.style.transform = 'translateX(-50%) translateY(10px)'; }, 2200);
  }
  window.pisaToast = toast;

  /* ── API helper ────────────────────────────────────────── */
  function authPost(path, body){
    var t = localStorage.getItem('pisa_token');
    return fetch(path, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json', 'Authorization': t ? ('Bearer ' + t) : '' },
      body: JSON.stringify(body || {})
    }).then(function(r){ return r.json(); });
  }

  /* ── Topbar ────────────────────────────────────────────── */
  function renderTopbar(){
    var app = document.querySelector('.app');
    if(!app || byId('topbar')) return;
    var el = document.createElement('header');
    el.className = 'topbar';
    el.id = 'topbar';
    el.innerHTML =
      '<div class="topbar-brand">'
      +   '<div class="brand-logo">P</div>'
      +   '<span class="brand-name">PISA 进销存</span>'
      + '</div>'
      + '<div class="topbar-search">' + IC.search
      +   '<input type="text" placeholder="快速查价 / 搜索商品、客户、单据...">'
      + '</div>'
      + '<div class="topbar-actions">'
      +   '<div class="tenant-switch" id="tenantSwitch" onclick="pisaToggleTenantMenu(event)">'
      +     '<span class="tenant-badge" id="tenantBadge">A</span>'
      +     '<span class="ts-name" id="currentTenant">-</span>'
      +     IC.chevron
      +     '<div class="tenant-menu" id="tenantMenu"></div>'
      +   '</div>'
      +   '<button class="topbar-icon-btn" title="消息">' + IC.bell + '<span class="dot">8</span></button>'
      +   '<div class="topbar-user" id="topbarUser" onclick="pisaUserMenu(event)">'
      +     '<div class="avatar" id="userAvatar">U</div>'
      +     '<div class="u-meta"><span class="u-name" id="userName">用户</span>'
      +     '<span class="u-sub" id="userSub">我的店铺</span></div>'
      +   '</div>'
      +   '<button class="topbar-icon-btn" id="topbarSetBtn" title="设置" onclick="pisaTopbarSettings(event)">' + IC.settings + '</button>'
      +   '<div class="user-menu" id="topbarSetMenu" style="display:none"></div>'
      + '</div>';
    app.insertBefore(el, app.firstChild);
  }

  function loadUser(){
    var u = ls('pisa_user','{}');
    var t = ls('pisa_currentTenant','null');
    var name = u.name || u.nickname || u.phone || '用户';
    var av = byId('userAvatar'), un = byId('userName'), us = byId('userSub');
    if(av) av.textContent = name.charAt(0);
    if(un) un.textContent = name;
    if(us) us.textContent = (t && t.name) ? t.name : '我的店铺';
  }
  window.pisaUserMenu = function(e){
    if(e) e.stopPropagation();
    if(byId('userMenu')){ window.pisaHideUserMenu(); return; }
    var host = byId('topbarUser'); if(!host) return;
    var u = ls('pisa_user','{}');
    var t = ls('pisa_currentTenant','null');
    var name = u.name || u.nickname || u.phone || '用户';
    var tname = (t && t.name) ? t.name : '我的店铺';
    var m = document.createElement('div');
    m.className = 'user-menu'; m.id = 'userMenu';
    m.innerHTML = ''
      + '<div class="um-head">'
      +   '<div class="um-info"><div class="um-name">' + esc(name) + '</div><div class="um-sub">' + esc(tname) + '</div></div>'
      + '</div>';
    document.body.appendChild(m);
    var r = host.getBoundingClientRect();
    m.style.top = (r.bottom + 8) + 'px';
    m.style.right = Math.max(8, window.innerWidth - r.right) + 'px';
    setTimeout(function(){ document.addEventListener('click', userMenuDocHandler); }, 0);
  };
  function userMenuDocHandler(ev){
    var m = byId('userMenu');
    var host = byId('topbarUser');
    if(m && !m.contains(ev.target) && !(host && host.contains(ev.target))) window.pisaHideUserMenu();
  }
  window.pisaHideUserMenu = function(){
    var m = byId('userMenu'); if(m && m.parentNode) m.parentNode.removeChild(m);
    document.removeEventListener('click', userMenuDocHandler);
  };
  window.pisaLogout = function(){
    try{ ['pisa_token','pisa_user','pisa_tenants','pisa_currentTenant','pisa_perms','pisa_sens'].forEach(function(k){ localStorage.removeItem(k); }); }catch(e){}
    location.href = '/login.html';
  };

  // 顶栏「设置」入口（位于用户菜单右侧）
  function topbarSetDocHandler(ev){
    var m = byId('topbarSetMenu');
    var btn = byId('topbarSetBtn');
    if(m && !m.contains(ev.target) && !(btn && btn.contains(ev.target))) window.pisaHideTopbarSettings();
  }
  window.pisaHideTopbarSettings = function(){
    var m = byId('topbarSetMenu'); if(m) m.style.display = 'none';
    document.removeEventListener('click', topbarSetDocHandler);
  };
  window.pisaTopbarSettings = function(e){
    if(e) e.stopPropagation();
    var m = byId('topbarSetMenu'); if(!m) return;
    if(m.style.display !== 'none'){ window.pisaHideTopbarSettings(); return; }
    var items = TOPBAR_SETTINGS.filter(function(it){ return navAllowed(it.key); });
    m.innerHTML = items.map(function(it){
      return '<div class="um-item" onclick="pisaHideTopbarSettings();if(window.PisaShell)PisaShell.navGo(\'' + it.key + '\')">' + esc(it.label) + '</div>';
    }).join('')
      + '<div class="um-item" onclick="pisaHideTopbarSettings();pisaToast(\'操作记录开发中\')">操作记录</div>'
      + '<div class="um-item" onclick="pisaHideTopbarSettings();pisaToast(\'网络诊断开发中\')">网络诊断</div>'
      + '<div class="um-item um-danger" onclick="pisaHideTopbarSettings();pisaLogout()">退出</div>';
    m.style.display = 'block';
    var r = e.currentTarget.getBoundingClientRect();
    m.style.top = (r.bottom + 8) + 'px';
    m.style.right = Math.max(8, window.innerWidth - r.right) + 'px';
    setTimeout(function(){ document.addEventListener('click', topbarSetDocHandler); }, 0);
  };

  /* ── Sidebar ───────────────────────────────────────────── */
  function renderSidebar(){
    var nav = byId('sidebarNav');
    if(!nav) return;
    var active = currentKey();
    var activeMod = moduleOf(active);
    var html = '';
    NAV.forEach(function(m){
      if(m.sep){ html += '<div class="snav-spacer"></div><div class="snav-sep"></div>'; return; }
      if(m.cols && !moduleHasItems(m)) return;
      var isActive = (m.key === active) || (activeMod && activeMod.key === m.key);
      html += '<div class="snav-item' + (isActive ? ' active' : '') + '"'
        + ' data-nav="' + esc(m.key) + '" title="' + esc(m.label) + '">'
        + m.icon + '<span class="snav-label">' + esc(m.label) + '</span>'
        + (m.badge ? '<span class="snav-badge">' + esc(m.badge) + '</span>' : '')
        + '</div>';
    });
    nav.innerHTML = html;
    bindSidebar();
  }

  function bindSidebar(){
    var nav = byId('sidebarNav');
    if(!nav) return;
    nav.querySelectorAll('.snav-item').forEach(function(el){
      var key = el.getAttribute('data-nav');
      var mod = NAV.filter(function(m){ return m.key === key; })[0];
      el.addEventListener('click', function(){
        if(mod && mod.cols){
          // 含子菜单的模块：优先跳到其第一个可访问页面，否则展开菜单（不再提示“开发中”）
          var target = firstNavItem(mod);
          if(target && target.href) navGo(target.key);
          else showFlyout(mod, el);
          return;
        }
        navGo(key);
      });
      if(mod && mod.cols){
        el.addEventListener('mouseenter', function(){ showFlyout(mod, el); });
        el.addEventListener('mouseleave', scheduleHide);
      }else{
        el.addEventListener('mouseenter', hideFlyoutNow);
      }
    });
  }

  /* 返回模块下第一个带 href 的菜单项 */
  function firstNavItem(mod){
    if(!mod || !mod.cols) return null;
    for(var i = 0; i < mod.cols.length; i++){
      var items = mod.cols[i].items || [];
      for(var j = 0; j < items.length; j++){
        if(items[j].href && navAllowed(items[j].key)) return items[j];
      }
    }
    return null;
  }

  /* 模块下是否有当前用户可访问的菜单项 */
  function moduleHasItems(mod){
    if(!mod || !mod.cols) return true;
    for(var i = 0; i < mod.cols.length; i++){
      var items = mod.cols[i].items || [];
      for(var j = 0; j < items.length; j++){
        if(navAllowed(items[j].key)) return true;
      }
    }
    return false;
  }

  /* ── Flyout ────────────────────────────────────────────── */
  var flyEl = null, hideTimer = null;
  function getFly(){
    if(!flyEl){
      flyEl = document.createElement('div');
      flyEl.className = 'flyout';
      flyEl.addEventListener('mouseenter', function(){ clearTimeout(hideTimer); });
      flyEl.addEventListener('mouseleave', scheduleHide);
      document.body.appendChild(flyEl);
    }
    return flyEl;
  }
  function scheduleHide(){ clearTimeout(hideTimer); hideTimer = setTimeout(hideFlyoutNow, 140); }
  function hideFlyoutNow(){ if(flyEl) flyEl.classList.remove('show'); }

  function showFlyout(mod, anchor){
    clearTimeout(hideTimer);
    var f = getFly();
    var cols = mod.cols.filter(function(c){
      return (c.items || []).some(function(it){ return navAllowed(it.key); });
    }).map(function(c){
      var head = c.title
        ? '<div class="flyout-col-title">' + esc(c.title)
          + (c.guide ? '<span class="flyout-guide">' + IC.guide + '流程引导</span>' : '')
          + '</div>'
        : '';
      var items = c.items.filter(function(it){ return navAllowed(it.key); }).map(function(it){
        return '<div class="flyout-item" data-key="' + esc(it.key) + '">'
          + '<span>' + esc(it.label) + '</span>'
          + (it.plus ? '<span class="flyout-plus">+</span>' : '')
          + '</div>';
      }).join('');
      return '<div class="flyout-col">' + head + items + '</div>';
    }).join('');
    f.innerHTML = '<div class="flyout-cols">' + cols + '</div>';
    f.classList.add('show');

    f.querySelectorAll('.flyout-item').forEach(function(el){
      el.addEventListener('click', function(){
        var k = el.getAttribute('data-key');
        hideFlyoutNow();
        navGo(k);
      });
    });

    var sb = document.querySelector('.sidebar');
    var left = sb ? sb.getBoundingClientRect().right : 112;
    var top = anchor.getBoundingClientRect().top;
    var maxTop = window.innerHeight - f.offsetHeight - 12;
    if(top > maxTop) top = Math.max(8, maxTop);
    if(top < 8) top = 8;
    f.style.left = (left + 4) + 'px';
    f.style.top = top + 'px';
  }

  /* ── Navigation ────────────────────────────────────────── */
  function navGo(key){
    var f = findItem(key);
    var href = f && f.item ? f.item.href : null;
    var label = f && f.item ? f.item.label : key;

    /* SPA (index.html) 内导航：不走整页刷新 */
    if(window.showPage){
      if(key === 'home'){ window.showPage('home'); renderSidebar(); return; }
      if(href && href.indexOf('#') >= 0){
        window.showPage(href.slice(href.indexOf('#') + 1));
        renderSidebar();
        return;
      }
      if(!href){                       // 无 href 的站内页面（含占位页）
        window.showPage(key);
        renderSidebar();
        return;
      }
    }

    if(!href){
      /* 独立页面（goods/supplier 等）点击站内页面：跳转到首页 SPA 并定位 */
      if(key === 'home'){ location.href = '/'; return; }
      var spaHref = '/#' + key;
      if(window.PisaTabs) window.PisaTabs.touch(key, label, spaHref);
      location.href = spaHref;
      return;
    }
    if(window.PisaTabs) window.PisaTabs.touch(key, label, href);
    location.href = href;
  }
  window.pisaNavGo = navGo;

  /* ── Sidebar collapse ──────────────────────────────────── */
  function renderCollapseBtn(){
    var nav = byId('sidebarNav');
    if(!nav || byId('snavToggle')) return;
    var b = document.createElement('button');
    b.className = 'snav-toggle';
    b.id = 'snavToggle';
    b.title = '收起/展开导航';
    b.innerHTML = IC.panel;
    b.setAttribute('onclick','pisaToggleSidebar()');
    document.querySelector('.sidebar').appendChild(b);
  }
  window.pisaToggleSidebar = function(){
    var sb = document.querySelector('.sidebar');
    if(!sb) return;
    var c = sb.classList.toggle('collapsed');
    localStorage.setItem('pisa_sidebar_collapsed', c ? '1' : '0');
  };

  /* ── Tenant switch ─────────────────────────────────────── */
  function tenants(){
    var t = ls('pisa_tenants','[]');
    return Array.isArray(t) ? t : [];
  }
  function renderTenantMenu(){
    var menu = byId('tenantMenu');
    if(!menu) return;
    var cur = ls('pisa_currentTenant','null');
    var list = tenants();
    var html = '<div class="tenant-menu-title">切换商户</div>';
    list.forEach(function(t, i){
      var active = (cur && cur.name === t.name) ? ' active' : '';
      html += '<div class="tenant-option' + active + '" data-tenant="' + esc(t.name) + '" data-role="' + esc(t.role||'') + '">'
        + '<span class="tenant-badge">' + esc(String(t.name||'?').charAt(0)) + '</span>'
        + esc(t.name) + '<span class="role">' + esc(t.role||'') + '</span></div>';
    });
    html += '<div class="tenant-option" data-create="1" style="color:var(--accent-700);font-weight:500">'
      + '<span style="font-size:17px;line-height:1;width:18px;text-align:center">+</span>创建新商户</div>';
    menu.innerHTML = html;

    menu.querySelectorAll('.tenant-option').forEach(function(el){
      el.addEventListener('click', function(e){
        e.stopPropagation();
        if(el.getAttribute('data-create')){ pisaCreateTenant(); return; }
        pisaSwitchTenant(el.getAttribute('data-tenant'), el.getAttribute('data-role'));
      });
    });
  }
  window.pisaToggleTenantMenu = function(e){
    if(e) e.stopPropagation();
    var s = byId('tenantSwitch');
    if(s) s.classList.toggle('open');
  };
  document.addEventListener('click', function(){
    var s = byId('tenantSwitch');
    if(s) s.classList.remove('open');
  });
  function applyTenantSession(data, tenant){
    if(data.token) localStorage.setItem('pisa_token', data.token);
    if(data.permissions) localStorage.setItem('pisa_perms', JSON.stringify(data.permissions));
    if(data.sensitive_data) localStorage.setItem('pisa_sens', JSON.stringify(data.sensitive_data));
    var list = tenants();
    var exists = false;
    list.forEach(function(t){ if(t.name === tenant.name) exists = true; });
    if(!exists) list.push(tenant);
    localStorage.setItem('pisa_tenants', JSON.stringify(list));
    localStorage.setItem('pisa_currentTenant', JSON.stringify(tenant));
  }
  window.pisaSwitchTenant = function(name, role){
    var list = tenants();
    var cur = null;
    list.forEach(function(t){ if(t.name === name) cur = t; });
    if(!cur) cur = {name:name, role:role||'操作员'};
    var tid = cur.tenant_id_str || cur.tenant_id || cur.id;
    var s = byId('tenantSwitch');
    if(s) s.classList.remove('open');
    if(!tid){ toast('该商户无法切换'); return; }
    authPost('/api/auth/switch-tenant', { tenant_id_str: String(tid) }).then(function(res){
      if(!res || res.code !== 0){ toast((res && res.message) || '切换失败'); return; }
      applyTenantSession(res.data, cur);
      toast('已切换到「' + name + '」');
      setTimeout(function(){ location.href = '/'; }, 400);
    }).catch(function(){ toast('网络错误'); });
  };
  function ensureTenantModal(){
    if(byId('tenantModal')) return;
    var d = document.createElement('div');
    d.id = 'tenantModal';
    d.className = 'pisa-modal-mask';
    d.innerHTML = '<div class="pisa-modal">'
      + '<div class="pisa-modal-head"><span class="pisa-modal-title">创建新商户</span>'
      +   '<button class="pisa-modal-x" type="button" onclick="pisaHideTenantModal()">&times;</button></div>'
      + '<div class="pisa-modal-body">'
      +   '<div class="pisa-field"><label>商户名称 <i>*</i></label><input id="newTenantName" placeholder="如：张三五金店"></div>'
      +   '<div class="pisa-field"><label>商户类型</label><select id="newTenantType"><option value="1">个体户</option><option value="2">有限公司</option><option value="3">合伙企业</option></select></div>'
      +   '<div class="pisa-field"><label>联系人</label><input id="newTenantContact" placeholder="请输入联系人"></div>'
      +   '<div class="pisa-field"><label>联系电话</label><input id="newTenantPhone" placeholder="请输入联系电话"></div>'
      +   '<div class="pisa-modal-hint">创建后该账号将成为新商户的主账号，并自动切换到新商户。</div>'
      + '</div>'
      + '<div class="pisa-modal-foot">'
      +   '<button class="pisa-btn-ghost" type="button" onclick="pisaHideTenantModal()">取消</button>'
      +   '<button class="pisa-btn-primary" id="tenantCreateBtn" type="button" onclick="pisaCreateTenantFromModal()">创建并切换</button>'
      + '</div></div>';
    document.body.appendChild(d);
    d.addEventListener('click', function(e){ if(e.target === d) window.pisaHideTenantModal(); });
    var inp = byId('newTenantName');
    if(inp) inp.addEventListener('keydown', function(e){ if(e.key === 'Enter') window.pisaCreateTenantFromModal(); });
  }
  window.pisaHideTenantModal = function(){
    var m = byId('tenantModal');
    if(m) m.classList.remove('show');
  };
  window.pisaCreateTenant = function(){
    ensureTenantModal();
    var m = byId('tenantModal');
    if(!m) return;
    var s = byId('tenantSwitch'); if(s) s.classList.remove('open');
    var inp = byId('newTenantName'); if(inp) inp.value = '';
    var c = byId('newTenantContact'); if(c) c.value = '';
    var p = byId('newTenantPhone'); if(p) p.value = '';
    m.classList.add('show');
    if(inp) setTimeout(function(){ inp.focus(); }, 50);
  };
  window.pisaCreateTenantFromModal = function(fallbackName){
    var name = (typeof fallbackName === 'string') ? fallbackName
      : (((byId('newTenantName') || {}).value) || '').trim();
    if(!name){ toast('请输入商户名称'); return; }
    var btn = byId('tenantCreateBtn');
    var old = btn ? btn.textContent : '';
    if(btn){ btn.disabled = true; btn.textContent = '创建中...'; }
    var payload = {
      name: name,
      contact_name: ((byId('newTenantContact') || {}).value || '').trim(),
      contact_phone: ((byId('newTenantPhone') || {}).value || '').trim()
    };
    var typeSel = byId('newTenantType');
    if(typeSel) payload.type = parseInt(typeSel.value, 10) || 1;
    authPost('/api/auth/create-tenant', payload).then(function(res){
      if(btn){ btn.disabled = false; btn.textContent = old; }
      if(!res || res.code !== 0){ toast((res && res.message) || '创建失败'); return; }
      applyTenantSession(res.data, res.data.tenant);
      window.pisaHideTenantModal();
      toast('已创建并切换到「' + name + '」');
      setTimeout(function(){ location.href = '/'; }, 500);
    }).catch(function(){
      if(btn){ btn.disabled = false; btn.textContent = old; }
      toast('网络错误');
    });
  };

  function loadTenant(){
    var cur = ls('pisa_currentTenant','null');
    if(!cur){
      var list = tenants();
      if(list.length){ cur = list[0]; localStorage.setItem('pisa_currentTenant', JSON.stringify(cur)); }
    }
    if(cur){
      var b = byId('tenantBadge'), n = byId('currentTenant');
      if(b) b.textContent = String(cur.name||'A').charAt(0);
      if(n) n.textContent = cur.name;
    }
    renderTenantMenu();
  }

  /* ── Boot ──────────────────────────────────────────────── */
  /* 清理旧版导航遗留的标签（key 已不在新 NAV 中） */
  var DYNAMIC_TAB_KEYS = [
    'goods:add', 'goods:view',
    'goods:combo:create',
    'stock:take:create', 'stock:take:view',
    'stock:assembly:create', 'stock:assembly:view', 'stock:recipe', 'stock:recipe:create',
    'stock:split:create', 'stock:split:view',
    'stock:flow', 'stock:cost',
    'purchase:create', 'purchase:view',
    'purchase-order:create', 'purchase-order:view',
    'purchase-return:create', 'purchase-return:view',
    'sale:create', 'sale:view',
    'sale-order:create', 'sale-order:view',
    'sale-return:create', 'sale-return:view',
    'quote:create', 'quote:view'
  ];
  function pruneTabs(){
    try{
      var raw = JSON.parse(localStorage.getItem('pisa_tabs') || '[]');
      if(!Array.isArray(raw)) return;
      var valid = {home:1};
      DYNAMIC_TAB_KEYS.forEach(function(k){ valid[k] = 1; });
      NAV.forEach(function(m){
        if(m.key) valid[m.key] = 1;
        if(m.cols) m.cols.forEach(function(c){
          c.items.forEach(function(it){ valid[it.key] = 1; });
        });
      });
      var kept = raw.filter(function(t){ return t && valid[t.key]; });
      if(kept.length !== raw.length){
        localStorage.setItem('pisa_tabs', JSON.stringify(kept));
      }
    }catch(e){}
  }

  function boot(){
    pruneTabs();
    if(localStorage.getItem('pisa_sidebar_collapsed') === '1'){
      var sb = document.querySelector('.sidebar');
      if(sb) sb.classList.add('collapsed');
    }
    renderTopbar();
    renderSidebar();
    renderCollapseBtn();
    loadUser();
    loadTenant();
    if(window.PisaTabs){
      var ck = currentKey();
      var cf = findItem(ck);
      var clabel = (cf && cf.item && (cf.item.tabLabel || cf.item.label)) || (window.TAB_LABELS && window.TAB_LABELS[ck]) || (window.EXTRA_LABELS && window.EXTRA_LABELS[ck]) || PAGE_LABEL[ck] || '';
      var chref = (cf && cf.item && cf.item.href) || (location.pathname + location.hash);
      window.PisaTabs.init({key:ck, label:clabel, href:chref});
    }
    document.addEventListener('click', hideFlyoutNow);
    window.addEventListener('scroll', hideFlyoutNow, true);
    window.addEventListener('hashchange', renderSidebar);
  }

  if(document.readyState === 'loading'){
    document.addEventListener('DOMContentLoaded', boot);
  }else{
    boot();
  }

  /* public API */
  window.PisaShell = {
    NAV: NAV,
    navGo: navGo,
    currentKey: currentKey,
    refreshSidebar: renderSidebar,
    toast: toast
  };
})();
