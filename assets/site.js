// JWKS navigation remains ordinary full-page links enhanced by local HTMX.
(() => {
  const nav = document.querySelector('.left-panel');
  const toggle = document.querySelector('#navToggle');
  const content = document.querySelector('#content');
  const search = document.querySelector('#searchInput');
  const provider = document.querySelector('#providerFilter');
  if (!nav || !toggle || !content || !search || !provider) return;
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
  search.addEventListener('input', applyFilters);
  provider.addEventListener('change', applyFilters);
  toggle.addEventListener('click', () => {
    const open = nav.classList.toggle('nav-open');
    toggle.setAttribute('aria-expanded', String(open));
    toggle.setAttribute('aria-label', open ? 'Close navigation' : 'Open navigation');
    if (open) search.focus();
  });
  nav.addEventListener('click', event => { if (event.target.closest('a')) closeNavigation(); });
  document.addEventListener('keydown', event => {
    if (event.key === 'Escape' && nav.classList.contains('nav-open')) { closeNavigation(); toggle.focus(); }
  });
  // HTMX 4 dispatches after:swap on the requesting link, not the swap target.
  document.body.addEventListener('htmx:after:swap', () => {
    syncPage(); content.focus();
  });
  window.addEventListener('popstate', () => { requestAnimationFrame(syncPage); });
  document.body.addEventListener('htmx:history:restore', syncPage);
  applyFilters();
  syncPage();
})();
