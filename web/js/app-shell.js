/* ============================================================
   PISA App Shell — shared chrome
   Renders: topbar (brand/search/tenant/actions) · sidebar + flyout
            · help panel · tabbar mount · unified navigation
   ============================================================ */
(function(){
  'use strict';

  /* ── Icons ─────────────────────────────────────────────── */
  var S = 'class="nav-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.7" stroke-linecap="round" stroke-linejoin="round"';
  var IC = {
    home:'<svg '+S+'><path d="M3 9.5 12 3l9 6.5V20a1 1 0 0 1-1 1H4a1 1 0 0 1-1-1z"/><path d="M9 21v-7h6v7"/></svg>',
    goods:'<svg '+S+'><path d="M21 8a2 2 0 0 0-1-1.73l-7-4a2 2 0 0 0-2 0l-7 4A2 2 0 0 0 3 8v8a2 2 0 0 0 1 1.73l7 4a2 2 0 0 0 2 0l7-4A2 2 0 0 0 21 16Z"/><path d="m3.3 7 8.7 5 8.7-5"/><path d="M12 22V12"/></svg>',
    sales:'<svg '+S+'><rect width="8" height="4" x="8" y="2" rx="1"/><path d="M16 4h2a2 2 0 0 1 2 2v14a2 2 0 0 1-2 2H6a2 2 0 0 1-2-2V6a2 2 0 0 1 2-2h2"/><path d="M12 11h4"/><path d="M12 16h4"/><path d="M8 11h.01"/><path d="M8 16h.01"/></svg>',
    customer:'<svg '+S+'><path d="M19 21v-2a4 4 0 0 0-4-4H9a4 4 0 0 0-4 4v2"/><circle cx="12" cy="7" r="4"/></svg>',
    purchase:'<svg '+S+'><circle cx="8" cy="21" r="1"/><circle cx="19" cy="21" r="1"/><path d="M2 2h2l2.6 12.4a2 2 0 0 0 2 1.6h9.8a2 2 0 0 0 2-1.6L21.5 7H5.1"/></svg>',
    stock:'<svg '+S+'><path d="M22 8.35V20a2 2 0 0 1-2 2H4a2 2 0 0 1-2-2V8.35A2 2 0 0 1 3.26 6.5l8-3.2a2 2 0 0 1 1.48 0l8 3.2A2 2 0 0 1 22 8.35Z"/><path d="M6 18h12"/><path d="M6 14h12"/></svg>',
    funds:'<svg '+S+'><rect width="20" height="12" x="2" y="6" rx="2"/><circle cx="12" cy="12" r="2"/><path d="M6 12h.01M18 12h.01"/></svg>',
    analysis:'<svg '+S+'><path d="M3 3v16a2 2 0 0 0 2 2h16"/><path d="m19 9-5 5-4-4-3 3"/></svg>',
    settings:'<svg '+S+'><circle cx="12" cy="12" r="3"/><path d="M19.4 15a1.7 1.7 0 0 0 .3 1.8l.1.1a2 2 0 1 1-2.8 2.8l-.1-.1a1.7 1.7 0 0 0-2.9 1.2V21a2 2 0 1 1-4 0v-.1A1.7 1.7 0 0 0 7 19.4l-.1.1a2 2 0 1 1-2.8-2.8l.1-.1A1.7 1.7 0 0 0 3 13.7a2 2 0 1 1 0-4h.1A1.7 1.7 0 0 0 4.6 7l-.1-.1a2 2 0 1 1 2.8-2.8l.1.1A1.7 1.7 0 0 0 10 3a2 2 0 1 1 4 0v.1a1.7 1.7 0 0 0 2.9 1.2l.1-.1a2 2 0 1 1 2.8 2.8l-.1.1a1.7 1.7 0 0 0 1.2 2.9 2 2 0 1 1 0 4h-.1a1.7 1.7 0 0 0-1.5 1z"/></svg>',
    search:'<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="11" cy="11" r="8"/><line x1="21" y1="21" x2="16.65" y2="16.65"/></svg>',
    help:'<svg viewBox="0 0 24 24"><circle cx="12" cy="12" r="10"/><path d="M9.1 9a3 3 0 0 1 5.8 1c0 2-3 3-3 3"/><path d="M12 17h.01"/></svg>',
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
        {key:'goods:list',  label:'商品',     href:'/goods.html'},
        {key:'goods:spec',  label:'规格管理'},
        {key:'goods:unit',  label:'单位管理'},
        {key:'goods:attr',  label:'商品属性'},
        {key:'goods:price', label:'价格管理'}
      ]},
      {title:'套餐', items:[
        {key:'goods:combo', label:'套餐'}
      ]}
    ]},

    {key:'sales', label:'销售', icon:IC.sales, badge:'12', cols:[
      {title:'销售', guide:true, items:[
        {key:'sales:new',    label:'销售',     plus:true},
        {key:'sales:order',  label:'销售预订', plus:true},
        {key:'sales:return', label:'销售退货', plus:true},
        {key:'sales:quote',  label:'报价',     plus:true}
      ]},
      {title:'其他', items:[
        {key:'sales:commission', label:'业绩提成'}
      ]}
    ]},

    {key:'customer', label:'客户', icon:IC.customer, cols:[
      {title:'客户', items:[
        {key:'customer:list',  label:'客户',     href:'/customer.html'},
        {key:'customer:level', label:'价格等级'},
        {key:'customer:quote', label:'报价管理'}
      ]}
    ]},

    {key:'purchase', label:'进货', icon:IC.purchase, cols:[
      {title:'进货', guide:true, items:[
        {key:'purchase:supplier', label:'供应商',   href:'/supplier.html'},
        {key:'purchase:new',      label:'进货',     href:'/#purchase:new',      plus:true},
        {key:'purchase:order',    label:'进货预订', href:'/#purchase:order',    plus:true},
        {key:'purchase:return',   label:'进货退货', href:'/#purchase:return',   plus:true}
      ]}
    ]},

    {key:'stock', label:'库存', icon:IC.stock, cols:[
      {title:'库存', guide:true, items:[
        {key:'stock:take',      label:'盘点',     plus:true},
        {key:'stock:assembly',  label:'组装',     plus:true},
        {key:'stock:split',     label:'拆分',     plus:true},
        {key:'stock:warehouse', label:'仓库管理', href:'/warehouse.html'}
      ]},
      {title:'查询', items:[
        {key:'stock:query',  label:'库存查询'},
        {key:'stock:alert',  label:'库存预警'},
        {key:'stock:batch',  label:'批次查询'},
        {key:'stock:expiry', label:'保质期查询'}
      ]}
    ]},

    {key:'funds', label:'资金', icon:IC.funds, cols:[
      {title:'账户', items:[
        {key:'funds:account',  label:'账户概览', href:'/account.html'},
        {key:'funds:transfer', label:'转账'}
      ]},
      {title:'收支', guide:true, items:[
        {key:'funds:receipt', label:'收款',     plus:true},
        {key:'funds:payment', label:'付款',     plus:true},
        {key:'funds:income',  label:'其他收入', plus:true},
        {key:'funds:expense', label:'其他支出', plus:true}
      ]},
      {title:'对账', items:[
        {key:'funds:recon-customer', label:'客户对账'},
        {key:'funds:recon-supplier', label:'供应商对账'},
        {key:'funds:cashflow',       label:'资金流水'}
      ]}
    ]},

    {key:'analysis', label:'分析', icon:IC.analysis, cols:[
      {title:'销售分析', items:[
        {key:'analysis:sales', label:'销售统计'},
        {key:'analysis:hot',   label:'热销分析'},
        {key:'analysis:staff', label:'员工绩效统计'}
      ]},
      {title:'库存分析', items:[
        {key:'analysis:purchase', label:'进货统计'},
        {key:'analysis:stock',    label:'库存统计'}
      ]},
      {title:'经营分析', items:[
        {key:'analysis:profit', label:'经营利润'}
      ]}
    ]},

    {sep:true},

    {key:'settings', label:'设置', icon:IC.settings, cols:[
      {title:'店铺管理', items:[
        {key:'settings:shop',  label:'店铺信息'},
        {key:'settings:staff', label:'员工管理'},
        {key:'settings:role',  label:'角色权限'},
        {key:'settings:pos',   label:'POS设备'}
      ]},
      {title:'基础设置', items:[
        {key:'settings:system', label:'系统设置'},
        {key:'settings:pref',   label:'用户偏好设置'},
        {key:'settings:print',  label:'打印设置'},
        {key:'settings:points', label:'积分设置'},
        {key:'settings:init',   label:'系统初始化'}
      ]}
    ]}
  ];

  /* ── helpers ───────────────────────────────────────────── */
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
    '/customer.html':'customer:list',
    '/supplier.html':'purchase:supplier',
    '/warehouse.html':'stock:warehouse',
    '/account.html':'funds:account'
  };
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
        + 'background:#1E293B;color:#fff;font-size:13px;padding:10px 20px;border-radius:8px;'
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
      +   '<button class="topbar-icon-btn" title="帮助中心" onclick="pisaToggleHelp(true)">' + IC.help + '</button>'
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
    toast('账号设置开发中');
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
      el.addEventListener('click', function(){ navGo(key); });
      if(mod && mod.cols){
        el.addEventListener('mouseenter', function(){ showFlyout(mod, el); });
        el.addEventListener('mouseleave', scheduleHide);
      }else{
        el.addEventListener('mouseenter', hideFlyoutNow);
      }
    });
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
    var cols = mod.cols.map(function(c){
      var head = c.title
        ? '<div class="flyout-col-title">' + esc(c.title)
          + (c.guide ? '<span class="flyout-guide">' + IC.guide + '流程引导</span>' : '')
          + '</div>'
        : '';
      var items = c.items.map(function(it){
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
      toast('「' + label + '」功能开发中');
      return;
    }
    if(window.PisaTabs) window.PisaTabs.touch(key, label, href);
    location.href = href;
  }
  window.pisaNavGo = navGo;

  /* ── Help panel ────────────────────────────────────────── */
  var HELP_LINKS = ['库存预警','商品基础管理指南','批量修改商品基础信息','库存预警设置','商品多单位管理','商品批量导入'];
  function renderHelp(){
    var p = byId('helpPanel');
    if(!p) return;
    p.innerHTML =
      '<div class="help-head">'
      + '<span class="help-title">帮助中心 »</span>'
      + '<button class="help-collapse" title="收起" onclick="pisaToggleHelp(false)">' + IC.x + '</button>'
      + '</div>'
      + '<div class="help-section">'
      +   '<div class="help-section-title">相关问题</div>'
      +   HELP_LINKS.map(function(t){ return '<span class="help-link" onclick="pisaHelpLink(\'' + esc(t) + '\')">' + esc(t) + '</span>'; }).join('')
      + '</div>'
      + '<div class="help-section">'
      +   '<div class="help-section-title">服务推荐</div>'
      +   '<div class="help-card" onclick="pisaToast(\'订阅版本功能开发中\')">'
      +     '<div class="help-card-icon">' + IC.box + '</div>'
      +     '<div><div class="help-card-title">订阅版本</div><div class="help-card-sub">多仓库管理</div></div>'
      +   '</div>'
      + '</div>';

    if(!byId('helpReopen')){
      var btn = document.createElement('button');
      btn.className = 'help-reopen';
      btn.id = 'helpReopen';
      btn.title = '展开帮助中心';
      btn.innerHTML = IC.help;
      btn.setAttribute('onclick','pisaToggleHelp(true)');
      document.body.appendChild(btn);
    }
    var collapsed = localStorage.getItem('pisa_help_collapsed') === '1';
    applyHelp(!collapsed);
  }
  function applyHelp(open){
    var p = byId('helpPanel'), r = byId('helpReopen');
    if(p) p.classList.toggle('collapsed', !open);
    if(r) r.classList.toggle('show', !open);
    localStorage.setItem('pisa_help_collapsed', open ? '0' : '1');
  }
  window.pisaToggleHelp = function(open){ applyHelp(!!open); };
  window.pisaHelpLink = function(t){ toast('「' + t + '」帮助文档开发中'); };

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
      + '<span style="font-size:16px;line-height:1;width:18px;text-align:center">+</span>创建新商户</div>';
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
  function pruneTabs(){
    try{
      var raw = JSON.parse(localStorage.getItem('pisa_tabs') || '[]');
      if(!Array.isArray(raw)) return;
      var valid = {home:1};
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
    renderHelp();
    loadUser();
    loadTenant();
    if(window.PisaTabs){
      var ck = currentKey();
      var cf = findItem(ck);
      var clabel = (cf && cf.item && cf.item.label) || ck;
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
