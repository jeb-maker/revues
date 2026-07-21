import { LitElement as f, css as c, nothing as l, html as o } from "lit";
import { property as s } from "lit/decorators.js";
import { setFormValue as n, constraintFlags as m, setValidity as b, clearValidity as y } from "../lib/form.js";
import { safeDefine as v } from "../lib/safe-define.js";
import { sharedStyles as $, fieldStyles as g, fieldLabelState as q } from "../lib/styles.js";
var x = Object.defineProperty, r = (p, t, e, h) => {
  for (var a = void 0, u = p.length - 1, d; u >= 0; u--)
    (d = p[u]) && (a = d(t, e, a) || a);
  return a && x(t, e, a), a;
};
class i extends f {
  constructor() {
    super(...arguments), this.label = "", this.hint = "", this.error = "", this.value = "", this.name = "", this.placeholder = "", this.type = "text", this.disabled = !1, this.required = !1, this.invalid = !1, this.density = "default", this.hideLabel = !1, this.min = "", this.max = "", this.step = "", this.accept = "", this.multiple = !1, this.#e = this.attachInternals(), this.#l = !1, this.#a = "", this.#h = !1, this.#s = !1;
  }
  static {
    this.formAssociated = !0;
  }
  static {
    this.styles = [
      $,
      g,
      c`
      :host {
        display: block;
      }

      input[type='file'].control {
        padding-block: var(--mb-space-2);
      }
    `
    ];
  }
  #e;
  #l;
  #i;
  #a;
  #h;
  #s;
  get #o() {
    return this.disabled || this.#l;
  }
  get #t() {
    return this.type === "file";
  }
  get #n() {
    return this.getAttribute("aria-label") ?? "";
  }
  connectedCallback() {
    super.connectedCallback(), this.#h || (this.#a = this.value, this.#h = !0);
  }
  firstUpdated() {
    this.#i = this.renderRoot.querySelector("input") ?? void 0, this.#r();
  }
  updated(t) {
    (t.has("value") || t.has("required") || t.has("error") || t.has("disabled") || t.has("name") || t.has("type")) && this.#r();
  }
  formDisabledCallback(t) {
    this.#l = t, this.requestUpdate();
  }
  formResetCallback() {
    this.#s = !1, this.value = this.#a, this.error = "", this.invalid = !1, this.#t && this.#i && (this.#i.value = "");
  }
  #p() {
    const t = this.#i?.files;
    if (!this.name || !t?.length) {
      n(this.#e, null);
      return;
    }
    if (t.length === 1) {
      n(this.#e, t[0]);
      return;
    }
    const e = new FormData();
    for (const h of t)
      e.append(this.name, h);
    n(this.#e, e);
  }
  #r() {
    this.#t ? this.#p() : n(this.#e, this.name ? this.value : null);
    const t = this.required && (this.#t ? !this.#i?.files?.length : !this.value), { flags: e, message: h } = m(this.error, t);
    h ? (b(this.#e, e, h, this.#i), this.invalid = !!this.error || this.#s) : (y(this.#e), this.invalid = !1);
  }
  #u(t) {
    const e = t.target;
    this.#s = !0, this.#t || (this.value = e.value), this.#r(), this.dispatchEvent(
      new CustomEvent("mb-input", {
        detail: { value: this.value, files: e.files },
        bubbles: !0,
        composed: !0
      })
    );
  }
  #d(t) {
    const e = t.target;
    this.#s = !0, this.#t || (this.value = e.value), this.#r(), this.dispatchEvent(
      new CustomEvent("mb-change", {
        detail: { value: this.value, files: e.files },
        bubbles: !0,
        composed: !0
      })
    );
  }
  #f(t) {
    if (t.key !== "Enter" || t.defaultPrevented || this.#t) return;
    const e = this.#e.form;
    e && (t.preventDefault(), e.requestSubmit());
  }
  render() {
    const t = [this.hint && !this.error ? "hint" : "", this.error ? "error" : ""].filter(Boolean).join(" "), { labelText: e, hideVisually: h, controlAriaLabel: a } = q(
      this.label,
      this.hideLabel,
      this.#n
    );
    return o`
      <div class="field">
        ${e ? o`<label
              part="label"
              class="label${h ? " visually-hidden" : ""}"
              for="control"
              >${e}</label
            >` : l}
        <input
          id="control"
          part="control"
          class="control"
          .type=${this.type}
          .value=${this.#t ? "" : this.value}
          name=${this.name || l}
          placeholder=${this.placeholder || l}
          min=${this.type === "number" && this.min !== "" ? this.min : l}
          max=${this.type === "number" && this.max !== "" ? this.max : l}
          step=${this.type === "number" && this.step !== "" ? this.step : l}
          accept=${this.#t && this.accept ? this.accept : l}
          ?multiple=${this.#t && this.multiple}
          ?disabled=${this.#o}
          ?required=${this.required}
          aria-invalid=${this.invalid ? "true" : "false"}
          aria-label=${a || l}
          aria-describedby=${t || l}
          @input=${this.#u}
          @change=${this.#d}
          @keydown=${this.#f}
        />
        ${this.hint && !this.error ? o`<p id="hint" class="hint">${this.hint}</p>` : l}
        ${this.error ? o`<p id="error" class="error" role="alert">${this.error}</p>` : l}
      </div>
    `;
  }
}
r([
  s()
], i.prototype, "label");
r([
  s()
], i.prototype, "hint");
r([
  s()
], i.prototype, "error");
r([
  s()
], i.prototype, "value");
r([
  s({ reflect: !0 })
], i.prototype, "name");
r([
  s()
], i.prototype, "placeholder");
r([
  s({ reflect: !0 })
], i.prototype, "type");
r([
  s({ type: Boolean, reflect: !0 })
], i.prototype, "disabled");
r([
  s({ type: Boolean, reflect: !0 })
], i.prototype, "required");
r([
  s({ type: Boolean, reflect: !0 })
], i.prototype, "invalid");
r([
  s({ reflect: !0 })
], i.prototype, "density");
r([
  s({ type: Boolean, reflect: !0, attribute: "hide-label" })
], i.prototype, "hideLabel");
r([
  s()
], i.prototype, "min");
r([
  s()
], i.prototype, "max");
r([
  s()
], i.prototype, "step");
r([
  s()
], i.prototype, "accept");
r([
  s({ type: Boolean })
], i.prototype, "multiple");
v("mb-input", i);
export {
  i as MbInput
};
//# sourceMappingURL=input.js.map
