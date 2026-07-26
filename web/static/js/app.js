(function () {
  'use strict';

  if ('serviceWorker' in navigator) {
    navigator.serviceWorker.getRegistrations().then(function (regs) {
      regs.forEach(function (reg) { reg.unregister(); });
    });
  }
  if ('caches' in window) {
    caches.keys().then(function (keys) {
      keys.forEach(function (key) { caches.delete(key); });
    });
  }

  document.body.addEventListener('mb-change', function (e) {
    var t = e.target;
    if (!t || !t.matches || !t.matches('mb-select.admin-nav-jump__select')) return;
    var v = e.detail && e.detail.value;
    if (v) location.href = v;
  });

  function showToast(msg, isError) {
    document.dispatchEvent(new CustomEvent('mb-toast', {
      detail: { message: msg, variant: isError ? 'danger' : 'success' }
    }));
  }

  document.body.addEventListener('toast:success', function (e) {
    showToast((e.detail && e.detail.message) || 'Action effectuée', false);
  });
  document.body.addEventListener('toast:error', function (e) {
    showToast((e.detail && e.detail.message) || 'Erreur', true);
  });
})();
