function a(t, e, s) {
  t.setFormValue(e, e);
}
function l(t, e, s = "", i) {
  t.setValidity(e, s, i);
}
function n(t) {
  t.setValidity({});
}
function r(t, e, s = "Please fill out this field.") {
  return t ? { flags: { customError: !0 }, message: t } : e ? { flags: { valueMissing: !0 }, message: s } : { flags: {}, message: "" };
}
export {
  n as clearValidity,
  r as constraintFlags,
  a as setFormValue,
  l as setValidity
};
//# sourceMappingURL=form.js.map
