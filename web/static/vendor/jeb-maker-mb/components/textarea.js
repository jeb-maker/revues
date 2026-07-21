import { LitElement as p, css as c, nothing as l, html as h } from "lit";
import { property as i } from "lit/decorators.js";
import { setFormValue as b, constraintFlags as f, setValidity as v, clearValidity as m } from "../lib/form.js";
import { safeDefine as y } from "../lib/safe-define.js";
import { sharedStyles as $, fieldStyles as g, fieldLabelState as x } from "../lib/styles.js";
var C = Object.defineProperty, r = (n, e, s, o) => {
  for (var a = void 0, d = n.length - 1, u; d >= 0; d--)
    (u = n[d]) && (a = u(e, s, a) || a);
  return a && C(e, s, a), a;
};
class t extends p {
  constructor() {
    super(...arguments), this.label = "", this.hint = "", this.error = "", this.value = "", this.name = "", this.placeholder = "", this.disabled = !1, this.required = !1, this.invalid = !1, this.rows = 4, this.density = "default", this.hideLabel = !1, this.#t = this.attachInternals(), this.#i = !1, this.#r = "", this.#s = !1, this.#e = !1;
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

      textarea.control {
        min-block-size: 6rem;
        resize: vertical;
      }
    `
    ];
  }
  #t;
  #i;
  #a;
  #r;
  #s;
  #e;
  get #o() {
    return this.disabled || this.#i;
  }
  get #h() {
    return this.getAttribute("aria-label") ?? "";
  }
  connectedCallback() {
    super.connectedCallback(), this.#s || (this.#r = this.value, this.#s = !0);
  }
  firstUpdated() {
    this.#a = this.renderRoot.querySelector("textarea") ?? void 0, this.#l();
  }
  updated(e) {
    (e.has("value") || e.has("required") || e.has("error") || e.has("disabled") || e.has("name")) && this.#l();
  }
  formDisabledCallback(e) {
    this.#i = e, this.requestUpdate();
  }
  formResetCallback() {
    this.#e = !1, this.value = this.#r, this.error = "", this.invalid = !1;
  }
  #l() {
    b(this.#t, this.name ? this.value : null);
    const e = this.required && !this.value, { flags: s, message: o } = f(this.error, e);
    o ? (v(this.#t, s, o, this.#a), this.invalid = !!this.error || this.#e) : (m(this.#t), this.invalid = !1);
  }
  #n(e) {
    const s = e.target;
    this.#e = !0, this.value = s.value, this.dispatchEvent(
      new CustomEvent("mb-input", {
        detail: { value: this.value },
        bubbles: !0,
        composed: !0
      })
    );
  }
  #d(e) {
    const s = e.target;
    this.#e = !0, this.value = s.value, this.dispatchEvent(
      new CustomEvent("mb-change", {
        detail: { value: this.value },
        bubbles: !0,
        composed: !0
      })
    );
  }
  render() {
    const e = [this.hint && !this.error ? "hint" : "", this.error ? "error" : ""].filter(Boolean).join(" "), { labelText: s, hideVisually: o, controlAriaLabel: a } = x(
      this.label,
      this.hideLabel,
      this.#h
    );
    return h`
      <div class="field">
        ${s ? h`<label
              part="label"
              class="label${o ? " visually-hidden" : ""}"
              for="control"
              >${s}</label
            >` : l}
        <textarea
          id="control"
          part="control"
          class="control"
          .value=${this.value}
          name=${this.name || l}
          placeholder=${this.placeholder || l}
          rows=${this.rows}
          ?disabled=${this.#o}
          ?required=${this.required}
          aria-invalid=${this.invalid ? "true" : "false"}
          aria-label=${a || l}
          aria-describedby=${e || l}
          @input=${this.#n}
          @change=${this.#d}
        ></textarea>
        ${this.hint && !this.error ? h`<p id="hint" class="hint">${this.hint}</p>` : l}
        ${this.error ? h`<p id="error" class="error" role="alert">${this.error}</p>` : l}
      </div>
    `;
  }
}
r([
  i()
], t.prototype, "label");
r([
  i()
], t.prototype, "hint");
r([
  i()
], t.prototype, "error");
r([
  i()
], t.prototype, "value");
r([
  i({ reflect: !0 })
], t.prototype, "name");
r([
  i()
], t.prototype, "placeholder");
r([
  i({ type: Boolean, reflect: !0 })
], t.prototype, "disabled");
r([
  i({ type: Boolean, reflect: !0 })
], t.prototype, "required");
r([
  i({ type: Boolean, reflect: !0 })
], t.prototype, "invalid");
r([
  i({ type: Number })
], t.prototype, "rows");
r([
  i({ reflect: !0 })
], t.prototype, "density");
r([
  i({ type: Boolean, reflect: !0, attribute: "hide-label" })
], t.prototype, "hideLabel");
y("mb-textarea", t);
export {
  t as MbTextarea
};
//# sourceMappingURL=textarea.js.map
