(() => {
  if (window.__uvyThemeInitialized) return;
  window.__uvyThemeInitialized = true;
  const root = document.documentElement;
  const selector = '[data-uvy-theme-toggle], #theme-toggle';
  const system = matchMedia('(prefers-color-scheme: dark)');
  let preference = null;
  try {
    const cookie = document.cookie.split('; ').find(value => value.startsWith('uvy-theme='));
    const saved = cookie?.slice('uvy-theme='.length);
    if (saved === 'light' || saved === 'dark') preference = saved;
  } catch {}
  const sync = () => {
    const dark = preference ? preference === 'dark' : system.matches;
    root.dataset.theme = dark ? 'dark' : 'light';
    root.style.colorScheme = dark ? 'dark' : 'light';
    document.querySelectorAll(selector).forEach(button => {
      button.hidden = false;
      button.setAttribute('aria-label', dark ? 'Switch to light mode' : 'Switch to dark mode');
      button.title = dark ? 'Switch to light mode' : 'Switch to dark mode';
    });
  };
  sync();
  system.addEventListener('change', () => { if (!preference) sync(); });
  document.addEventListener('click', event => {
    if (!(event.target instanceof Element) || !event.target.closest(selector)) return;
    preference = root.dataset.theme === 'dark' ? 'light' : 'dark';
    try {
      document.cookie = `uvy-theme=${preference}; Path=/; Max-Age=31536000; SameSite=Lax${location.protocol === 'https:' ? '; Secure' : ''}`;
    } catch {}
    sync();
  });
  if (document.readyState === 'loading') document.addEventListener('DOMContentLoaded', sync, { once: true });
  else sync();
  document.addEventListener('htmx:after:swap', sync);
  document.addEventListener('htmx:afterSwap', sync);
})();

// JWKS navigation remains ordinary full-page links enhanced by HTMX.
(() => {
  const nav = document.querySelector('.left-panel');
  const toggle = document.querySelector('#navToggle');
  const content = document.querySelector('#content');
  if (!nav || !toggle || !content) return;
  let search, provider;
  const initializeFilters = () => {
    search = nav.querySelector('#searchInput');
    provider = nav.querySelector('#providerFilter');
    if (!search || !provider) return;
    search.addEventListener('input', applyFilters);
    provider.addEventListener('change', applyFilters);
    applyFilters();
    syncPage();
  };
  const closeNavigation = () => {
    nav.classList.remove('nav-open');
    toggle.setAttribute('aria-expanded', 'false');
    toggle.setAttribute('aria-label', 'Open navigation');
  };
  const applyFilters = () => {
    const query = search.value.trim().toLowerCase();
    nav.querySelectorAll('.service-link').forEach(link => {
      const providers = link.dataset.providers.trim().split(/\s+/).filter(Boolean);
      link.hidden = !(link.querySelector('.service-name').textContent.toLowerCase().includes(query) && (provider.value === 'all' || (provider.value === 'other' ? !providers.length : providers.includes(provider.value))));
    });
  };
  const syncPage = () => {
    const heading = content.querySelector('[data-page-title]');
    if (heading) document.title = heading.dataset.pageTitle;
    nav.querySelectorAll('.service-link').forEach(link => {
      if (new URL(link.href).pathname === location.pathname) link.setAttribute('aria-current', 'page');
      else link.removeAttribute('aria-current');
    });
  };
  toggle.addEventListener('click', () => {
    const open = nav.classList.toggle('nav-open');
    toggle.setAttribute('aria-expanded', String(open));
    toggle.setAttribute('aria-label', open ? 'Close navigation' : 'Open navigation');
    if (open && search) search.focus();
  });
  nav.addEventListener('click', event => { if (event.target.closest('a')) closeNavigation(); });
  document.addEventListener('keydown', event => {
    if (event.key === 'Escape' && nav.classList.contains('nav-open')) { closeNavigation(); toggle.focus(); }
  });
  // HTMX 4 dispatches after:swap on the requesting link, not the swap target.
  document.body.addEventListener('htmx:after:swap', event => {
    if (event.target === nav) initializeFilters();
    else { syncPage(); content.focus(); }
  });
  window.addEventListener('popstate', () => { requestAnimationFrame(syncPage); });
  document.body.addEventListener('htmx:history:restore', syncPage);
  initializeFilters();
})();
