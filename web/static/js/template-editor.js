(function () {
  'use strict';
  var form = document.getElementById('template-form'), box = document.getElementById('template-sections');
  if (!form || !box) return;
  var addSecBtn = document.getElementById('template-add-section'),
    enableSecBtn = document.getElementById('template-enable-sections-btn'),
    secTools = document.getElementById('template-section-tools'),
    enableWrap = document.getElementById('template-enable-sections'),
    maxRow = 0, maxSec = 0, forceSec = false;

  function n(v, d) { var x = parseInt(v, 10); return isNaN(x) ? d : x; }
  function secs() { return box.querySelectorAll('.template-editor-section'); }
  function sectioned() {
    if (secs().length > 1 || forceSec) return true;
    var t = box.querySelector('.template-editor-section-title');
    return !!(t && t.value.trim());
  }
  function syncMode() {
    var on = sectioned();
    box.classList.toggle('is-sectioned', on);
    if (secTools) secTools.hidden = !on;
    if (enableWrap) enableWrap.hidden = on;
  }
  function scan() {
    maxRow = maxSec = -1;
    box.querySelectorAll('input[name="item_row_idx"]').forEach(function (el) {
      maxRow = Math.max(maxRow, n(el.value, -1));
    });
    box.querySelectorAll('input[name="section_idx"]').forEach(function (el) {
      maxSec = Math.max(maxSec, n(el.value, -1));
    });
  }
  function syncSec(sec, si) {
    sec.setAttribute('data-section-idx', si);
    sec.querySelector('input[name="section_idx"]').value = si;
    sec.querySelectorAll('input[name="item_section_idx"]').forEach(function (el) { el.value = si; });
  }
  function syncRow(row, ri, si) {
    row.querySelector('input[name="item_row_idx"]').value = ri;
    row.querySelector('input[name="item_section_idx"]').value = si;
    row.querySelector('input[name="item_required"]').value = ri;
    var lab = row.querySelector('input[name="item_label"]'), hlp = row.querySelector('input[name="item_help"]');
    lab.id = 'item_label_' + si + '_' + ri;
    hlp.id = 'item_help_' + si + '_' + ri;
    row.querySelector('label[for^="item_label_"]').htmlFor = lab.id;
    row.querySelector('label[for^="item_help_"]').htmlFor = hlp.id;
  }
  function clearRow(row) {
    row.querySelector('input[name="item_label"]').value = '';
    row.querySelector('input[name="item_help"]').value = '';
    row.querySelector('input[name="item_required"]').checked = false;
  }
  function rowBtns(tb) {
    var rows = tb.querySelectorAll('.template-editor-row');
    rows.forEach(function (row, i) {
      row.querySelector('[data-action="move-up"]').disabled = i === 0;
      row.querySelector('[data-action="move-down"]').disabled = i === rows.length - 1;
      row.querySelector('[data-action="remove"]').disabled = rows.length <= 1;
    });
  }
  function secBtns() {
    secs().forEach(function (sec, i, all) {
      sec.querySelector('[data-action="section-up"]').disabled = i === 0;
      sec.querySelector('[data-action="section-down"]').disabled = i === all.length - 1;
      sec.querySelector('[data-action="section-remove"]').disabled = all.length <= 1;
    });
  }
  function addPoint(sec) {
    var tb = sec.querySelector('.template-section-rows'), tpl = tb.querySelector('.template-editor-row'),
      si = sec.getAttribute('data-section-idx'), row = tpl.cloneNode(true);
    clearRow(row);
    syncRow(row, ++maxRow, si);
    tb.appendChild(row);
    rowBtns(tb);
    row.querySelector('input[name="item_label"]').focus();
  }
  function addSec() {
    var tpl = box.querySelector('.template-editor-section'), sec = tpl.cloneNode(true), si = String(++maxSec);
    sec.querySelector('.template-editor-section-title').value = '';
    var tb = sec.querySelector('.template-section-rows');
    tb.innerHTML = '';
    var row = tpl.querySelector('.template-editor-row').cloneNode(true);
    clearRow(row);
    syncSec(sec, si);
    syncRow(row, ++maxRow, si);
    tb.appendChild(row);
    box.appendChild(sec);
    secBtns();
    rowBtns(tb);
    syncMode();
    sec.querySelector('.template-editor-section-title').focus();
  }
  scan();
  box.addEventListener('click', function (e) {
    var b = e.target.closest('[data-action]');
    if (!b || b.type !== 'button') return;
    var a = b.getAttribute('data-action'), sec = b.closest('.template-editor-section'), row = b.closest('.template-editor-row');
    if (a === 'add-point') return addPoint(sec);
    if (a === 'section-remove') {
      if (secs().length > 1) {
        sec.remove();
        secBtns();
        box.querySelectorAll('.template-section-rows').forEach(rowBtns);
        if (secs().length === 1) forceSec = false;
        syncMode();
      }
      return;
    }
    if (a === 'section-up' && sec.previousElementSibling) {
      box.insertBefore(sec, sec.previousElementSibling); secBtns(); return;
    }
    if (a === 'section-down' && sec.nextElementSibling) {
      box.insertBefore(sec.nextElementSibling, sec); secBtns(); return;
    }
    if (!row) return;
    var tb = row.closest('.template-section-rows');
    if (a === 'remove' && tb.querySelectorAll('.template-editor-row').length > 1) {
      row.remove(); rowBtns(tb);
    } else if (a === 'move-up' && row.previousElementSibling) {
      tb.insertBefore(row, row.previousElementSibling); rowBtns(tb);
    } else if (a === 'move-down' && row.nextElementSibling) {
      tb.insertBefore(row.nextElementSibling, row); rowBtns(tb);
    }
  });
  box.addEventListener('input', function (e) {
    if (e.target.classList.contains('template-editor-section-title')) syncMode();
  });
  if (addSecBtn) addSecBtn.addEventListener('click', addSec);
  if (enableSecBtn) enableSecBtn.addEventListener('click', function () {
    forceSec = true;
    syncMode();
    box.querySelector('.template-editor-section-title').focus();
  });
  secBtns();
  box.querySelectorAll('.template-section-rows').forEach(rowBtns);
  syncMode();
})();
