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

  var h = document.querySelector('.hamburger');
  var n = document.querySelector('.site-nav');
  if (h && n) {
    n.classList.remove('site-nav--nojs');
    h.addEventListener('click', function () {
      var e = h.getAttribute('aria-expanded') === 'true';
      h.setAttribute('aria-expanded', !e);
      n.classList.toggle('site-nav--open');
    });
  }

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
