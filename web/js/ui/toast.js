// Toast notification
const Toast = {
  show(msg, type) {
    const el = document.getElementById('toast');
    if (!el) return;
    el.textContent = msg;
    el.className = 'toast show';
    if (type === 'error') el.style.background = 'var(--danger-500)';
    else el.style.background = 'var(--gray-800)';
    clearTimeout(el._tm);
    el._tm = setTimeout(() => el.classList.remove('show'), 2500);
  },
};
