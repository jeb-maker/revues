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
  var n = document.querySelector('.site-nav') || document.querySelector('.nav-tabs');
  if (h && n) {
    var openClass = n.classList.contains('site-nav') ? 'site-nav--open' : 'nav-tabs--open';
    var nojsClass = n.classList.contains('site-nav') ? 'site-nav--nojs' : 'nav-tabs--nojs';
    n.classList.add(nojsClass);
    h.addEventListener('click', function () {
      var e = h.getAttribute('aria-expanded') === 'true';
      h.setAttribute('aria-expanded', !e);
      n.classList.toggle(openClass);
    });
  }

  function showToast(msg, isError) {
    var t = document.getElementById('toast');
    if (!t) return;
    t.textContent = msg;
    t.className = 'toast' + (isError ? ' toast--error' : '') + ' toast--show';
    setTimeout(function () { t.className = 'toast'; }, 3000);
  }

  document.body.addEventListener('toast:success', function (e) {
    showToast(e.detail.message || 'Action effectuée', false);
  });
  document.body.addEventListener('toast:error', function (e) {
    showToast(e.detail.message || 'Erreur', true);
  });
})();
