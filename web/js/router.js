// Simple hash router
const Router = {
  routes: {},
  current: null,

  register(name, handler) {
    this.routes[name] = handler;
  },

  init() {
    window.addEventListener('hashchange', () => this.resolve());
    this.resolve();
  },

  resolve() {
    const hash = location.hash.slice(1) || 'home';
    const handler = this.routes[hash];
    if (handler) {
      this.current = hash;
      handler();
    } else {
      this.navigate('home');
    }
  },

  navigate(name) {
    location.hash = '#/' + name;
  },
};
