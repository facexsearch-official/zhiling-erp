// API client with auto token injection
const API = {
  baseURL: '/api',

  async request(method, path, body) {
    const token = localStorage.getItem('token');
    const headers = { 'Content-Type': 'application/json' };
    if (token) headers['Authorization'] = 'Bearer ' + token;

    const opts = { method, headers };
    if (body && method !== 'GET') opts.body = JSON.stringify(body);

    try {
      const res = await fetch(this.baseURL + path, opts);
      if (res.status === 401) {
        localStorage.clear();
        window.location.href = '/login.html';
        return null;
      }
      const data = await res.json();
      return data;
    } catch (e) {
      console.error('API error:', e);
      Toast.show('网络请求失败', 'error');
      return null;
    }
  },

  get(path) { return this.request('GET', path); },
  post(path, body) { return this.request('POST', path, body); },
  put(path, body) { return this.request('PUT', path, body); },
  del(path) { return this.request('DELETE', path); },
};
