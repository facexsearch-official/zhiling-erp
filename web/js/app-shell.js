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
        {key:'goods:unit',  label:'单位管理', href:'/unit.html'},
        {key:'goods:attr',  label:'商品属性', href:'/attr.html'},
        {key:'goods:price', label:'价格管理', href:'/price.html'}
      ]},
      {title:'\u00A0', items:[
        {key:'goods:combo', label:'套餐', tabLabel:'套餐列表', href:'/#goods:combo'}
      ]}
    ]},

    {key:'customer', label:'客户', icon:IC.customer, cols:[
      {title:'客户', items:[
        {key:'customer:list',  label:'客户',     tabLabel:'客户列表', href:'/customer.html'},
        {key:'customer:level', label:'价格等级', tabLabel:'价格等级', href:'/#customer:level'},
        {key:'customer:quote', label:'报价管理', tabLabel:'报价管理', href:'/#customer:quote'}
      ]}
    ]},

    {key:'purchase', label:'进货', icon:IC.purchase, cols:[
      {title:'进货', guide:true, items:[
        {key:'purchase:supplier', label:'供应商',   href:'/supplier.html'},
        {key:'purchase:new',      label:'进货',     tabLabel:'进货单列表', href:'/#purchase:new',      plus:true},
        {key:'purchase:order',    label:'进货预订', tabLabel:'进货预订列表', href:'/#purchase:order',    plus:true},
        {key:'purchase:return',   label:'进货退货', tabLabel:'进货退货列表', href:'/#purchase:return',   plus:true}
      ]}
    ]},

    {key:'sales', label:'销售', icon:IC.sales, badge:'12', cols:[
      {title:'销售', guide:true, items:[
        {key:'sales:new',    label:'销售',     tabLabel:'销售单列表',     plus:true},
        {key:'sales:order',  label:'销售预订', tabLabel:'销售预订列表',   plus:true},
        {key:'sales:return', label:'销售退货', tabLabel:'销售退货单列表', plus:true},
        {key:'sales:quote',  label:'报价',     tabLabel:'报价',           plus:true}
      ]},
      {title:'其他', items:[
        {key:'sales:commission', label:'业绩提成'}
      ]}
    ]},

    {key:'stock', label:'库存', icon:IC.stock, cols:[
      {title:'库存', guide:true, items:[
        {key:'stock:take',      label:'盘点',     tabLabel:'盘点单列表', href:'/#stock:take', plus:true},
        {key:'stock:assembly',  label:'组装',     tabLabel:'组装单列表', href:'/#stock:assembly', plus:true},
        {key:'stock:split',     label:'拆分',     tabLabel:'拆分单列表', href:'/#stock:split', plus:true}
      ]},
      {title:'查询', items:[
        {key:'stock:query',  label:'库存查询', tabLabel:'库存查询', href:'/#stock:query'},
        {key:'stock:alert',  label:'库存预警', tabLabel:'库存预警', href:'/#stock:alert'},
        {key:'stock:batch',  label:'批次查询', tabLabel:'批次查询', href:'/#stock:batch'},
        {key:'stock:expiry', label:'保质期查询', tabLabel:'保质期查询', href:'/#stock:expiry'}
      ]}
    ]},

    {key:'funds', label:'资金', icon:IC.funds, cols:[
      {title:'账户', items:[
        {key:'funds:account',  label:'账户概览', tabLabel:'账户概览', href:'/#funds:account'},
        {key:'funds:transfer', label:'转账',     tabLabel:'转账',     href:'/#funds:transfer'}
      ]},
      {title:'收支', guide:true, items:[
        {key:'funds:receipt', label:'收款',     tabLabel:'收款单列表', href:'/#funds:receipt', plus:true},
        {key:'funds:payment', label:'付款',     tabLabel:'付款单列表', href:'/#funds:payment', plus:true},
        {key:'funds:income',  label:'其他收入', tabLabel:'其他收入列表', href:'/#funds:income', plus:true},
        {key:'funds:expense', label:'其他支出', tabLabel:'其他支出列表', href:'/#funds:expense', plus:true}
      ]},
      {title:'对账', items:[
        {key:'funds:recon-customer', label:'客户对账',   tabLabel:'客户对账',   href:'/#funds:recon-customer'},
        {key:'funds:recon-supplier', label:'供应商对账', tabLabel:'供应商对账', href:'/#funds:recon-supplier'},
        {key:'funds:cashflow',       label:'资金流水',   tabLabel:'资金流水',   href:'/#funds:cashflow'}
      ]}
    ]},

    {key:'analysis', label:'分析', icon:IC.analysis, cols:[
      {title:'销售分析', items:[
        {key:'analysis:sales', label:'销售统计', tabLabel:'销售统计', href:'/#analysis:sales'},
        {key:'analysis:hot',   label:'热销分析', tabLabel:'热销分析', href:'/#analysis:hot'},
        {key:'analysis:staff', label:'员工业绩统计', tabLabel:'员工业绩统计', href:'/#analysis:staff'}
      ]},
      {title:'库存分析', items:[
        {key:'analysis:purchase', label:'进货统计', tabLabel:'进货统计', href:'/#analysis:purchase'},
        {key:'analysis:stock',    label:'库存统计', tabLabel:'库存统计', href:'/#analysis:stock'}
      ]},
      {title:'经营分析', items:[
        {key:'analysis:profit', label:'经营利润', tabLabel:'经营利润', href:'/#analysis:profit'}
      ]}
    ]},

    {sep:true},

    {key:'settings', label:'设置', icon:IC.settings, cols:[
      {title:'店铺管理', items:[
        {key:'settings:shop',  label:'商户信息', tabLabel:'商户信息', href:'/#settings:shop'},
        {key:'settings:shops', label:'门店管理', tabLabel:'门店管理', href:'/#settings:shops'},
        {key:'settings:staff', label:'员工管理'},
        {key:'settings:role',  label:'角色权限', tabLabel:'角色权限', href:'/#settings:role'},
        {key:'settings:pos',   label:'POS设备'}
      ]},
      {title:'基础设置', items:[
        {key:'settings:system', label:'系统设置', tabLabel:'系统设置', href:'/#settings:system'},
        {key:'settings:pref',   label:'用户偏好设置', tabLabel:'用户偏好设置', href:'/#settings:pref'},
        {key:'settings:print',  label:'打印设置', tabLabel:'打印设置', href:'/#settings:print'},
        {key:'settings:points', label:'积分设置', tabLabel:'积分设置', href:'/#settings:points'},
        {key:'settings:init',   label:'系统初始化'}
      ]}
    ]}
  ];

  /* ── helpers ───────────────────────────────────────────── */
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
      +   '<span class="um-view" onclick="pisaHideUserMenu();if(window.PisaShell)PisaShell.navGo(\'settings:shop\')">查看</span>'
      + '</div>'
      + '<div class="um-item" onclick="pisaHideUserMenu();pisaToast(\'操作记录开发中\')">操作记录</div>'
      + '<div class="um-item" onclick="pisaHideUserMenu();pisaToast(\'网络诊断开发中\')">网络诊断</div>'
      + '<div class="um-item um-danger" onclick="pisaLogout()">退出</div>';
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
    html += '<div class="tenant-menu-foot" onclick="pisaNavGo(\'settings:shop\')">商户信息 / 套餐订阅 ›</div>';
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
  window.pisaSwitchTenant = function(name, role){
    var list = tenants();
    var cur = null;
    list.forEach(function(t){ if(t.name === name) cur = t; });
    if(!cur) cur = {name:name, role:role||'操作员'};
    localStorage.setItem('pisa_currentTenant', JSON.stringify(cur));
    var s = byId('tenantSwitch');
    if(s) s.classList.remove('open');
    toast('已切换到「' + name + '」' + (role ? '（' + role + '）' : ''));
    setTimeout(function(){ location.href = '/'; }, 400);
  };
  window.pisaHideTenantModal = function(){
    var m = byId('tenantModal');
    if(m) m.classList.remove('show');
  };
  function addTenant(name){
    var list = tenants();
    list.push({id:Date.now(), name:name, role:'主账号'});
    localStorage.setItem('pisa_tenants', JSON.stringify(list));
    window.pisaSwitchTenant(name, '主账号');
  }
  window.pisaCreateTenant = function(){
    var m = byId('tenantModal');
    if(m){
      var s = byId('tenantSwitch'); if(s) s.classList.remove('open');
      var inp = byId('newTenantName'); if(inp) inp.value = '';
      m.classList.add('show');
      if(inp) setTimeout(function(){ inp.focus(); }, 50);
      return;
    }
    var name = (prompt('请输入商户名称') || '').trim();
    if(name) addTenant(name);
  };
  window.pisaCreateTenantFromModal = function(){
    var inp = byId('newTenantName');
    var name = ((inp && inp.value) || '').trim();
    if(!name){ toast('请输入商户名称'); return; }
    window.pisaHideTenantModal();
    addTenant(name);
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
