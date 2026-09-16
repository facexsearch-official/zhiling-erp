/* ============================================================
   PISA 多标签页 (Tab) 管理
   ============================================================ */
window.PisaTabs = (function(){
  var STORAGE_KEY = 'pisa_tabs';
  var MAX_TABS = 12;
  var HOME = {key:'home', label:'首页', href:'/', closable:false};
  var _active = 'home';

  var IC_HOME = '<svg class="ptab-home-icon" viewBox="0 0 24 24"><path d="M3 9.5 12 3l9 6.5V20a1 1 0 0 1-1 1H4a1 1 0 0 1-1-1z"/><path d="M9 21v-7h6v7"/></svg>';
  var IC_FULL = '<svg viewBox="0 0 24 24"><path d="M8 3H5a2 2 0 0 0-2 2v3"/><path d="M21 8V5a2 2 0 0 0-2-2h-3"/><path d="M3 16v3a2 2 0 0 0 2 2h3"/><path d="M16 21h3a2 2 0 0 0 2-2v-3"/></svg>';

  function load(){
    try{
      var raw = JSON.parse(localStorage.getItem(STORAGE_KEY) || '[]');
      return Array.isArray(raw) ? raw : [];
    }catch(e){ return []; }
  }
  function save(tabs){ localStorage.setItem(STORAGE_KEY, JSON.stringify(tabs)); }

  function ensureHome(tabs){
    if(!tabs.some(function(t){ return t.key === HOME.key; })){
      tabs.unshift({key:HOME.key, label:HOME.label, href:HOME.href, closable:false});
    }
    tabs.forEach(function(t){
      if(t.key === HOME.key){ t.closable = false; t.href = HOME.href; t.label = HOME.label; }
    });
    return tabs;
  }

  function hrefFor(key){ return key === 'home' ? '/' : '/#' + key; }

  function pathKey(){
    var h = location.hash.slice(1);
    if(h){
      try{ h = decodeURIComponent(h); }catch(e){}
      if(h) return h;
    }
    var p = location.pathname.replace(/\/$/,'') || '/';
    var map = {
      '/':'home','/index.html':'home',
      '/goods.html':'goods:list',
      '/customer.html':'customer:list',
      '/supplier.html':'purchase:supplier',
      '/warehouse.html':'stock:warehouse',
      '/account.html':'funds:account'
    };
    if(map[p]) return map[p];
    return 'home';
  }

  /* 挂载点：.main 顶部（在 .content 之前） */
  function ensureBar(){
    var bar = document.getElementById('tabbar');
    if(bar) return bar;
    var main = document.querySelector('.main');
    if(!main) return null;
    bar = document.createElement('div');
    bar.className = 'tabbar';
    bar.id = 'tabbar';
    bar.innerHTML = '<div class="tabbar-inner" id="tabbarInner"></div>'
      + '<div class="tabbar-tools"><button class="tabbar-tool" title="全屏" onclick="PisaTabs.fullscreen()">' + IC_FULL + '</button></div>';
    var content = main.querySelector('.content');
    if(content) main.insertBefore(bar, content); else main.appendChild(bar);
    return bar;
  }

  function render(){
    var bar = ensureBar();
    if(!bar) return;
    var inner = document.getElementById('tabbarInner');
    if(!inner) return;
    var tabs = ensureHome(load());
    var html = '';
    tabs.forEach(function(t){
      var cls = 'ptab' + (t.key === _active ? ' active' : '');
      var icon = t.key === 'home' ? IC_HOME : '';
      var close = t.closable
        ? '<span class="ptab-close" data-close="' + esc(t.key) + '" title="关闭">'
          + '<svg viewBox="0 0 24 24"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg></span>'
        : '';
      html += '<div class="' + cls + '" data-tab="' + esc(t.key) + '" title="' + esc(t.label) + '">'
            + icon + '<span class="ptab-label">' + esc(t.label) + '</span>' + close + '</div>';
    });
    inner.innerHTML = html;
    scrollActiveIntoView();
  }

  function scrollActiveIntoView(){
    var inner = document.getElementById('tabbarInner');
    if(!inner) return;
    var el = inner.querySelector('.ptab.active');
    if(el && el.scrollIntoView){ try{ el.scrollIntoView({block:'nearest', inline:'nearest'}); }catch(e){} }
  }

  function esc(s){
    return String(s == null ? '' : s)
      .replace(/&/g,'&amp;').replace(/</g,'&lt;').replace(/>/g,'&gt;')
      .replace(/"/g,'&quot;').replace(/'/g,'&#39;');
  }

  function touch(key, label, href){
    if(!key) return;
    var tabs = ensureHome(load());
    var t = tabs.filter(function(x){ return x.key === key; })[0];
    if(t){
      if(label) t.label = label;
      if(href) t.href = href;
    }else{
      tabs.push({key:key, label:label || key, href:href || hrefFor(key), closable:key !== 'home'});
    }
    var closable = tabs.filter(function(x){ return x.closable; });
    while(closable.length > MAX_TABS){
      var idx = tabs.findIndex(function(x){ return x.closable && x.key !== key; });
      if(idx === -1) break;
      tabs.splice(idx, 1);
      closable = tabs.filter(function(x){ return x.closable; });
    }
    save(tabs);
    _active = key;
    render();
  }

  function go(key){
    var tabs = ensureHome(load());
    var t = tabs.filter(function(x){ return x.key === key; })[0];
    if(!t) return;
    _active = key;
    render();
    location.href = t.href || hrefFor(key);
  }

  function close(key){
    var tabs = ensureHome(load());
    var t = tabs.filter(function(x){ return x.key === key; })[0];
    if(!t || !t.closable) return;
    var wasActive = (_active === key);
    var pos = tabs.indexOf(t);
    tabs.splice(pos, 1);
    save(tabs);
    if(wasActive){
      var next = tabs[Math.min(pos, tabs.length - 1)] || tabs[0];
      _active = next.key;
      render();
      location.href = next.href || hrefFor(next.key);
    }else{
      render();
    }
  }

  function closeOthers(key){
    var tabs = ensureHome(load()).filter(function(t){ return !t.closable || t.key === key; });
    save(ensureHome(tabs));
    _active = key;
    render();
  }

  function closeAll(){
    var tabs = ensureHome(load()).filter(function(t){ return !t.closable; });
    save(ensureHome(tabs));
    _active = 'home';
    render();
    location.href = '/';
  }

  function fullscreen(){
    var el = document.documentElement;
    if(!document.fullscreenElement){
      if(el.requestFullscreen) el.requestFullscreen();
    }else{
      if(document.exitFullscreen) document.exitFullscreen();
    }
  }

  function init(page){
    page = page || {key:pathKey()};
    var tabs = ensureHome(load());
    var t = tabs.filter(function(x){ return x.key === page.key; })[0];
    if(t){
      if(page.label) t.label = page.label;
      if(page.href) t.href = page.href;
    }else if(page.key){
      tabs.push({key:page.key, label:page.label || page.key, href:page.href || hrefFor(page.key), closable:page.key !== 'home'});
    }
    save(tabs);
    _active = page.key || 'home';
    ensureBar();
    render();
  }

  /* ─── 事件委托 ─── */
  document.addEventListener('click', function(e){
    var closeEl = e.target.closest ? e.target.closest('.ptab-close') : null;
    if(closeEl){
      e.stopPropagation();
      close(closeEl.getAttribute('data-close'));
      return;
    }
    var tabEl = e.target.closest ? e.target.closest('.ptab') : null;
    if(tabEl){
      var key = tabEl.getAttribute('data-tab');
      if(key !== _active) go(key);
    }
  });

  /* 右键菜单 */
  var menuEl = null;
  function hideMenu(){ if(menuEl){ menuEl.remove(); menuEl = null; } }
  document.addEventListener('contextmenu', function(e){
    var tabEl = e.target.closest ? e.target.closest('.ptab') : null;
    if(!tabEl) return;
    e.preventDefault();
    var key = tabEl.getAttribute('data-tab');
    hideMenu();
    menuEl = document.createElement('div');
    menuEl.className = 'ptab-ctxmenu';
    menuEl.innerHTML =
      '<div class="ptab-ctxmenu-item" data-act="close">关闭当前</div>' +
      '<div class="ptab-ctxmenu-item" data-act="others">关闭其他</div>' +
      '<div class="ptab-ctxmenu-item" data-act="all">关闭全部</div>';
    menuEl.style.left = e.clientX + 'px';
    menuEl.style.top = e.clientY + 'px';
    document.body.appendChild(menuEl);
    menuEl.addEventListener('click', function(ev){
      var item = ev.target.closest('.ptab-ctxmenu-item');
      if(!item) return;
      var act = item.getAttribute('data-act');
      hideMenu();
      if(act === 'close') close(key);
      else if(act === 'others') closeOthers(key);
      else if(act === 'all') closeAll();
    });
  });
  document.addEventListener('click', function(e){
    if(menuEl && !e.target.closest('.ptab-ctxmenu')) hideMenu();
  });
  document.addEventListener('keydown', function(e){ if(e.key === 'Escape') hideMenu(); });
  window.addEventListener('scroll', hideMenu, true);

  return {
    init: init, touch: touch, go: go, close: close,
    closeOthers: closeOthers, closeAll: closeAll, render: render,
    ensureBar: ensureBar, hrefFor: hrefFor, fullscreen: fullscreen,
    activeKey: function(){ return _active; }
  };
})();
