(function () {
'use strict';
var form = document.getElementById('template-form'), box = document.getElementById('template-sections');
if (!form || !box) return;
var addSecBtn = document.getElementById('template-add-section'),
enableSecBtn = document.getElementById('template-enable-sections-btn'),
secTools = document.getElementById('template-section-tools'),
enableWrap = document.getElementById('template-enable-sections'),
maxRow = 0, maxSec = 0, forceSec = false, dragRow = null;
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
var cat = row.querySelector('[name="item_section"]');
lab.id = 'item_label_' + si + '_' + ri;
hlp.id = 'item_help_' + si + '_' + ri;
if (cat) cat.id = 'item_section_' + si + '_' + ri;
}
function clearFieldValue(el) {
if (!el) return;
el.value = '';
el.setAttribute('value', '');
}
function clearRow(row) {
clearFieldValue(row.querySelector('mb-input[name="item_label"]'));
clearFieldValue(row.querySelector('[name="item_help"]'));
clearFieldValue(row.querySelector('[name="item_section"]'));
var req = row.querySelector('mb-checkbox[name="item_required"]');
req.checked = false;
req.removeAttribute('checked');
}
function setFieldValue(el, v) {
if (!el) return;
el.value = v;
el.setAttribute('value', v);
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
var rows = container.querySelectorAll('.template-editor__point'), multi = rows.length > 1;
rows.forEach(function (row, i) {
setDisabled(row, 'move-up', i === 0);
setDisabled(row, 'move-down', i === rows.length - 1);
setDisabled(row, 'remove', !multi);
/* Keep drag/actions cells visible so desktop columns stay aligned. */
row.querySelectorAll('.template-editor__drag-handle').forEach(function (el) {
el.draggable = multi; el.disabled = !multi;
});
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
var container = sec.querySelector('.template-editor__points'),
rows = container.querySelectorAll('.template-editor__point'),
last = rows[rows.length - 1],
tpl = rows[0],
si = sec.getAttribute('data-section-idx'),
prevCat = '',
row;
if (last) {
var prev = last.querySelector('[name="item_section"]');
if (prev) prevCat = String(prev.value || '');
}
row = tpl.cloneNode(true);
clearRow(row);
if (prevCat) setFieldValue(row.querySelector('[name="item_section"]'), prevCat);
syncRow(row, ++maxRow, si);
container.appendChild(row);
rowBtns(container);
resyncFields(row);
row.querySelector('mb-input[name="item_label"]').focus();
}
function addSec() {
var tpl = box.querySelector('.template-editor__section'), sec = tpl.cloneNode(true), si = String(++maxSec);
var title = sec.querySelector('.template-editor__section-title');
clearFieldValue(title);
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
function clearDrag() {
if (dragRow) dragRow.classList.remove('is-dragging');
box.querySelectorAll('.is-drag-over').forEach(function (el) { el.classList.remove('is-drag-over'); });
dragRow = null;
}
scan();
box.addEventListener('click', function (e) {
var b = e.target.closest('[data-action]');
if (!b || b.type !== 'button') return;
var a = b.getAttribute('data-action'), sec = b.closest('.template-editor__section'), row = b.closest('.template-editor__point');
if (a === 'add-point') return addPoint(sec);
if (a === 'section-remove') {
if (secs().length > 1) {
sec.remove(); secBtns();
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
box.addEventListener('dragstart', function (e) {
var handle = e.target.closest('.template-editor__drag-handle');
if (!handle || handle.disabled) return;
dragRow = handle.closest('.template-editor__point');
if (!dragRow) return;
dragRow.classList.add('is-dragging');
e.dataTransfer.effectAllowed = 'move';
e.dataTransfer.setData('text/plain', 'row');
});
box.addEventListener('dragend', clearDrag);
box.addEventListener('dragover', function (e) {
var row = e.target.closest('.template-editor__point');
if (!dragRow || !row || row === dragRow || row.parentNode !== dragRow.parentNode) return;
e.preventDefault();
e.dataTransfer.dropEffect = 'move';
box.querySelectorAll('.is-drag-over').forEach(function (el) {
if (el !== row) el.classList.remove('is-drag-over');
});
row.classList.add('is-drag-over');
var rect = row.getBoundingClientRect();
if (e.clientY < rect.top + rect.height / 2) row.parentNode.insertBefore(dragRow, row);
else row.parentNode.insertBefore(dragRow, row.nextElementSibling);
});
box.addEventListener('drop', function (e) {
if (!dragRow) return;
e.preventDefault();
rowBtns(dragRow.closest('.template-editor__points'));
resyncFields(dragRow);
clearDrag();
});
box.addEventListener('mb-input', function (e) {
if (e.target.classList.contains('template-editor__section-title')) syncMode();
});
if (addSecBtn) addSecBtn.addEventListener('click', addSec);
if (enableSecBtn) enableSecBtn.addEventListener('click', function () {
forceSec = true; syncMode();
box.querySelector('.template-editor__section-title').focus();
});
secBtns();
box.querySelectorAll('.template-editor__points').forEach(rowBtns);
syncMode();
})();
