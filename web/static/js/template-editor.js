(function () {
'use strict';
var form = document.getElementById('template-form'), box = document.getElementById('template-sections');
if (!form || !box) return;
var addSecBtn = document.getElementById('template-add-section'),
enableSecBtn = document.getElementById('template-enable-sections-btn'),
secTools = document.getElementById('template-section-tools'),
enableWrap = document.getElementById('template-enable-sections'),
maxRow = 0, maxSec = 0, forceSec = false;
// #region agent log
function dbg(hypothesisId, location, message, data) {
  try {
    var payload = {hypothesisId: hypothesisId, location: location, message: message, data: data || {}, timestamp: Date.now()};
    fetch('http://127.0.0.1:7399/', {method: 'POST', headers: {'Content-Type': 'application/json'}, body: JSON.stringify(payload), mode: 'cors', keepalive: true}).catch(function () {});
  } catch (e) {}
}
function snapMb(el, tag) {
  if (!el) return {tag: tag, missing: true};
  var sh = null;
  try { sh = el.shadowRoot && el.shadowRoot.querySelector('input, textarea'); } catch (e) {}
  return {
    tag: tag,
    id: el.id || '',
    name: el.getAttribute('name') || el.name || '',
    isConnected: !!el.isConnected,
    sameAs: null,
    propValue: el.value,
    attrValue: el.getAttribute('value'),
    shadowValue: sh ? sh.value : null,
    checked: typeof el.checked === 'boolean' ? el.checked : undefined
  };
}
function formFieldCounts() {
  try {
    var fd = new FormData(form);
    return {
      item_label: fd.getAll('item_label'),
      item_help: fd.getAll('item_help'),
      item_row_idx: fd.getAll('item_row_idx'),
      item_section_idx: fd.getAll('item_section_idx'),
      item_required: fd.getAll('item_required')
    };
  } catch (e) {
    return {error: String(e)};
  }
}
var _origLabels = [];
function trackOriginals() {
  _origLabels = Array.prototype.slice.call(box.querySelectorAll('mb-input[name="item_label"]'));
}
trackOriginals();
customElements.whenDefined('mb-input').then(function () {
  var Proto = customElements.get('mb-input').prototype;
  var _upd = Proto.updated;
  var _frc = Proto.formResetCallback;
  var _cc = Proto.connectedCallback;
  Proto.updated = function (changed) {
    if (this.name === 'item_label' && changed && changed.has('value')) {
      dbg('C', 'mb-input.updated', 'item_label value updated', Object.assign(snapMb(this, 'self'), {oldValue: changed.get('value'), newValue: this.value, isOrig: _origLabels.indexOf(this) === 0}));
    }
    return _upd ? _upd.call(this, changed) : undefined;
  };
  Proto.formResetCallback = function () {
    dbg('D', 'mb-input.formResetCallback', 'form reset on mb-input', Object.assign(snapMb(this, 'self'), {isOrig: _origLabels.indexOf(this) === 0}));
    return _frc ? _frc.call(this) : undefined;
  };
  Proto.connectedCallback = function () {
    dbg('A', 'mb-input.connectedCallback', 'mb-input connected', snapMb(this, 'self'));
    return _cc ? _cc.call(this) : undefined;
  };
  dbg('A', 'template-editor.js:patch', 'mb-input prototype patched', {});
});
form.addEventListener('reset', function () {
  dbg('D', 'form.reset', 'native form reset fired', formFieldCounts());
});
form.addEventListener('submit', function (e) {
  dbg('D', 'form.submit', 'form submit fired', Object.assign({defaultPrevented: e.defaultPrevented}, formFieldCounts()));
});
// #endregion
function n(v, d) { var x = parseInt(v, 10); return isNaN(x) ? d : x; }
function secs() { return box.querySelectorAll('.template-editor__section'); }
function sectioned() {
if (secs().length > 1 || forceSec) return true;
var t = box.querySelector('.template-editor__section-title');
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
var req = row.querySelector('mb-checkbox[name="item_required"]');
req.setAttribute('value', ri);
req.value = String(ri);
var lab = row.querySelector('mb-input[name="item_label"]'), hlp = row.querySelector('[name="item_help"]');
// #region agent log
dbg('B', 'syncRow:before', 'syncRow ids', {ri: ri, si: String(si), lab: snapMb(lab, 'lab'), hlp: snapMb(hlp, 'hlp'), labIsOrig0: lab === _origLabels[0]});
// #endregion
lab.id = 'item_label_' + si + '_' + ri;
hlp.id = 'item_help_' + si + '_' + ri;
}
function clearRow(row) {
var lab = row.querySelector('mb-input[name="item_label"]');
// #region agent log
var before = snapMb(lab, 'lab');
var orig0 = _origLabels[0];
dbg('B', 'clearRow:before', 'clearRow about to clear', {labIsOrig0: lab === orig0, labConnected: lab && lab.isConnected, before: before, orig0: snapMb(orig0, 'orig0')});
// #endregion
// Do not removeAttribute('value'): Lit maps missing attr → value=null,
// and form-associated setFormValue(null) omits the control from FormData
// (len(item_label)≠len(item_help) → "lignes incohérentes").
lab.value = '';
if (lab.hasAttribute('value')) lab.setAttribute('value', '');
var help = row.querySelector('[name="item_help"]');
if (help) {
help.value = '';
if (help.hasAttribute('value')) help.setAttribute('value', '');
}
var req = row.querySelector('mb-checkbox[name="item_required"]');
req.checked = false;
req.removeAttribute('checked');
// #region agent log
dbg('B', 'clearRow:after', 'clearRow done', {lab: snapMb(lab, 'lab'), orig0: snapMb(orig0, 'orig0'), labIsOrig0: lab === orig0});
// #endregion
}
function resyncFields(root) {
root.querySelectorAll('mb-input,mb-textarea,mb-select,mb-checkbox').forEach(function (el) {
if (el.requestUpdate) el.requestUpdate('value');
});
}
function setDisabled(root, action, on) {
root.querySelectorAll('[data-action="' + action + '"]').forEach(function (btn) { btn.disabled = on; });
}
function rowBtns(container) {
var rows = container.querySelectorAll('.template-editor__point');
rows.forEach(function (row, i) {
setDisabled(row, 'move-up', i === 0);
setDisabled(row, 'move-down', i === rows.length - 1);
setDisabled(row, 'remove', rows.length <= 1);
});
}
function secBtns() {
secs().forEach(function (sec, i, all) {
setDisabled(sec, 'section-up', i === 0);
setDisabled(sec, 'section-down', i === all.length - 1);
setDisabled(sec, 'section-remove', all.length <= 1);
});
}
function addPoint(sec) {
var container = sec.querySelector('.template-editor__points'), tpl = container.querySelector('.template-editor__point'),
si = sec.getAttribute('data-section-idx');
// #region agent log
trackOriginals();
var origLab = tpl.querySelector('mb-input[name="item_label"]');
dbg('A', 'addPoint:beforeClone', 'state before cloneNode', {orig: snapMb(origLab, 'orig'), form: formFieldCounts(), rowCount: container.querySelectorAll('.template-editor__point').length});
// #endregion
var row = tpl.cloneNode(true);
// #region agent log
var cloneLab = row.querySelector('mb-input[name="item_label"]');
dbg('A', 'addPoint:afterClone', 'state after cloneNode(true)', {
  orig: snapMb(origLab, 'orig'),
  clone: snapMb(cloneLab, 'clone'),
  sameRef: cloneLab === origLab,
  origShadowAfterClone: snapMb(origLab, 'orig')
});
// #endregion
clearRow(row);
// #region agent log
dbg('A', 'addPoint:afterClearRow', 'orig vs clone after clearRow(clone)', {orig: snapMb(origLab, 'orig'), clone: snapMb(cloneLab, 'clone')});
// #endregion
syncRow(row, ++maxRow, si);
container.appendChild(row);
rowBtns(container);
// #region agent log
dbg('A', 'addPoint:afterAppend', 'after appendChild', {orig: snapMb(origLab, 'orig'), clone: snapMb(cloneLab, 'clone'), form: formFieldCounts()});
queueMicrotask(function () {
  dbg('A', 'addPoint:microtask', 'microtask after add', {orig: snapMb(origLab, 'orig'), clone: snapMb(cloneLab, 'clone'), form: formFieldCounts()});
});
requestAnimationFrame(function () {
  dbg('C', 'addPoint:raf', 'raf after add (prop vs shadow)', {orig: snapMb(origLab, 'orig'), clone: snapMb(cloneLab, 'clone'), form: formFieldCounts()});
});
trackOriginals();
// #endregion
row.querySelector('mb-input[name="item_label"]').focus();
}
function addSec() {
var tpl = box.querySelector('.template-editor__section'), sec = tpl.cloneNode(true), si = String(++maxSec);
var title = sec.querySelector('.template-editor__section-title');
title.value = '';
if (title.hasAttribute('value')) title.setAttribute('value', '');
var container = sec.querySelector('.template-editor__points');
container.innerHTML = '';
var row = tpl.querySelector('.template-editor__point').cloneNode(true);
clearRow(row);
syncSec(sec, si);
syncRow(row, ++maxRow, si);
container.appendChild(row);
box.appendChild(sec);
secBtns();
rowBtns(container);
syncMode();
sec.querySelector('.template-editor__section-title').focus();
}
scan();
box.addEventListener('click', function (e) {
var b = e.target.closest('[data-action]');
// #region agent log
if (b && b.getAttribute('data-action') === 'add-point') {
  dbg('D', 'click:add-point', 'add-point click seen', {
    targetTag: e.target && e.target.tagName,
    btnTag: b.tagName,
    btnType: b.type,
    typeAttr: b.getAttribute('type'),
    defaultPrevented: e.defaultPrevented,
    eventPhase: e.eventPhase
  });
}
// #endregion
if (!b || b.type !== 'button') return;
var a = b.getAttribute('data-action'), sec = b.closest('.template-editor__section'), row = b.closest('.template-editor__point');
if (a === 'add-point') return addPoint(sec);
if (a === 'section-remove') {
if (secs().length > 1) {
sec.remove();
secBtns();
box.querySelectorAll('.template-editor__points').forEach(rowBtns);
if (secs().length === 1) forceSec = false;
syncMode();
}
return;
}
if (a === 'section-up' && sec.previousElementSibling) {
box.insertBefore(sec, sec.previousElementSibling); secBtns(); resyncFields(sec); return;
}
if (a === 'section-down' && sec.nextElementSibling) {
var moved = sec.nextElementSibling;
box.insertBefore(moved, sec); secBtns(); resyncFields(moved); return;
}
if (!row) return;
var container = row.closest('.template-editor__points');
if (a === 'remove' && container.querySelectorAll('.template-editor__point').length > 1) {
row.remove(); rowBtns(container);
} else if (a === 'move-up' && row.previousElementSibling) {
container.insertBefore(row, row.previousElementSibling); rowBtns(container); resyncFields(row);
} else if (a === 'move-down' && row.nextElementSibling) {
var movedRow = row.nextElementSibling;
container.insertBefore(movedRow, row); rowBtns(container); resyncFields(movedRow);
}
});
box.addEventListener('mb-input', function (e) {
if (e.target.classList.contains('template-editor__section-title')) syncMode();
});
if (addSecBtn) addSecBtn.addEventListener('click', addSec);
if (enableSecBtn) enableSecBtn.addEventListener('click', function () {
forceSec = true;
syncMode();
box.querySelector('.template-editor__section-title').focus();
});
secBtns();
box.querySelectorAll('.template-editor__points').forEach(rowBtns);
syncMode();
})();
