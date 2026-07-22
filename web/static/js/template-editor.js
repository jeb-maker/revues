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
lab.id = 'item_label_' + si + '_' + ri;
hlp.id = 'item_help_' + si + '_' + ri;
}
function clearRow(row) {
var lab = row.querySelector('mb-input[name="item_label"]');
lab.value = '';
lab.removeAttribute('value');
var help = row.querySelector('[name="item_help"]');
if (help) {
help.value = '';
help.removeAttribute('value');
}
var req = row.querySelector('mb-checkbox[name="item_required"]');
req.checked = false;
req.removeAttribute('checked');
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
si = sec.getAttribute('data-section-idx'), row = tpl.cloneNode(true);
clearRow(row);
syncRow(row, ++maxRow, si);
container.appendChild(row);
rowBtns(container);
row.querySelector('mb-input[name="item_label"]').focus();
}
function addSec() {
var tpl = box.querySelector('.template-editor__section'), sec = tpl.cloneNode(true), si = String(++maxSec);
var title = sec.querySelector('.template-editor__section-title');
title.value = '';
title.removeAttribute('value');
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
