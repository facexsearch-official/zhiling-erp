// Simple state management
const State = {
  get(key) { return localStorage.getItem(key); },
  set(key, val) { localStorage.setItem(key, val); },
  remove(key) { localStorage.removeItem(key); },
  clear() { localStorage.clear(); },

  get token() { return this.get('token'); },
  get tenantId() { return parseInt(this.get('tenant_id')) || 0; },
  get tenantName() { return this.get('tenant_name') || '-'; },
  get role() { return parseInt(this.get('role')) || 0; },
  get userId() { return parseInt(this.get('user_id')) || 0; },
  get nickname() { return this.get('nickname') || '用户'; },

  isLoggedIn() { return !!this.token && !!this.tenantId; },
};
