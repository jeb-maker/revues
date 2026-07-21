(function () {
  'use strict';

  if (typeof Reports === 'undefined' || typeof Reports.init !== 'function') {
    return;
  }

  function csrfHeaders() {
    var meta = document.querySelector('meta[name="csrf-token"]');
    return {
      'X-CSRF-Token': meta ? meta.getAttribute('content') || '' : '',
    };
  }

  function metadata() {
    var el = document.getElementById('revues-reports-meta');
    if (!el) {
      return { app: 'revues' };
    }
    try {
      var parsed = JSON.parse(el.textContent || '{}');
      if (parsed && typeof parsed === 'object') {
        return parsed;
      }
    } catch (e) {
      /* fall through */
    }
    return { app: 'revues', error: 'metadata_parse_failed' };
  }

  Reports.init({
    locale: 'fr',
    adapter: 'webhook',
    webhook: {
      auth: 'url',
      url: '/signaler/api',
      credentials: 'same-origin',
      headers: csrfHeaders,
    },
    metadata: metadata,
    // Opt-in auto error reporting (v0.3.0): uncaught error / unhandledrejection.
    // Defaults: cooldown 30s, max 5 per page session; no screenshot; deduped by signature.
    autoReport: {
      errors: true,
      maxPerSession: 5,
      cooldownMs: 30000,
    },
  });

  document.addEventListener('click', function (e) {
    var target = e.target;
    if (!target || !target.closest) {
      return;
    }
    var link = target.closest('a[data-reports-open]');
    if (!link) {
      return;
    }
    e.preventDefault();
    try {
      Reports.open();
    } catch (err) {
      window.location.href = link.getAttribute('href') || '/signaler';
    }
  });

  if (document.body && document.body.getAttribute('data-reports-auto-open') === '1') {
    try {
      Reports.open();
    } catch (err) {
      /* fallback form remains available */
    }
  }
})();
