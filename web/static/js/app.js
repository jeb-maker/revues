(function () {
  'use strict';
  var h = document.querySelector('.hamburger');
  var n = document.querySelector('.nav-tabs');
  if (h && n) {
    n.classList.add('nav-tabs--nojs');
    h.addEventListener('click', function () {
      var e = h.getAttribute('aria-expanded') === 'true';
      h.setAttribute('aria-expanded', !e);
      n.classList.toggle('nav-tabs--open');
    });
  }
})();
