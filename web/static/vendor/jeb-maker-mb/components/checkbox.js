import { LitElement as p, css as u, nothing as h, html as d } from "lit";
import { property as i } from "lit/decorators.js";
import { setFormValue as m, setValidity as f, clearValidity as b } from "../lib/form.js";
import { safeDefine as v } from "../lib/safe-define.js";
import { sharedStyles as y } from "../lib/styles.js";
var k = Object.defineProperty, s = (l, e, r, n) => {
  for (var a = void 0, o = l.length - 1, c; o >= 0; o--)
    (c = l[o]) && (a = c(e, r, a) || a);
  return a && k(e, r, a), a;
};
class t extends p {
  constructor() {
    super(...arguments), this.label = "", this.error = "", this.name = "", this.value = "on", this.checked = !1, this.indeterminate = !1, this.disabled = !1, this.required = !1, this.invalid = !1, this.#e = this.attachInternals(), this.#s = !1, this.#r = !1, this.#a = !1, this.#i = !1;
  }
  static {
    this.formAssociated = !0;
  }
  static {
    this.styles = [
      y,
      u`
      :host {
        display: inline-block;
      }

      label {
        display: inline-flex;
        align-items: flex-start;
        gap: var(--mb-space-2);
        cursor: pointer;
        font-size: var(--mb-font-size-md);
      }

      input {
        margin-block-start: 0.2rem;
        accent-color: var(--mb-color-accent);
        inline-size: 1.1rem;
        block-size: 1.1rem;
      }

      input:disabled {
        cursor: not-allowed;
      }

      :host([disabled]) label {
        opacity: 0.55;
        cursor: not-allowed;
      }

      .error {
        margin: var(--mb-space-1) 0 0;
        color: var(--mb-color-danger);
        font-size: var(--mb-font-size-sm);
      }
    `
    ];
  }
  #e;
  #s;
  #t;
  #r;
  #a;
  #i;
  get #h() {
    return this.disabled || this.#s;
  }
  connectedCallback() {
    super.connectedCallback(), this.#a || (this.#r = this.checked, this.#a = !0);
  }
  firstUpdated() {
    this.#t = this.renderRoot.querySelector("input") ?? void 0, this.#l(), this.#o();
  }
  updated(e) {
    e.has("indeterminate") && this.#l(), (e.has("checked") || e.has("value") || e.has("required") || e.has("error") || e.has("disabled") || e.has("name")) && this.#o();
  }
  formDisabledCallback(e) {
    this.#s = e, this.requestUpdate();
  }
  formResetCallback() {
    this.#i = !1, this.checked = this.#r, this.indeterminate = !1, this.error = "", this.invalid = !1;
  }
  #l() {
    this.#t && (this.#t.indeterminate = this.indeterminate);
  }
  #o() {
    m(this.#e, this.name && this.checked ? this.value : null);
    const e = this.required && !this.checked, r = this.error || (e ? "Please check this box." : "");
    if (r) {
      const n = this.error ? { customError: !0 } : { valueMissing: !0 };
      f(this.#e, n, r, this.#t), this.invalid = !!this.error || this.#i;
    } else
      b(this.#e), this.invalid = !1;
  }
  #n(e) {
    const r = e.target;
    this.#i = !0, this.checked = r.checked, this.indeterminate = !1, this.dispatchEvent(
      new CustomEvent("mb-change", {
        detail: { checked: this.checked, value: this.value },
        bubbles: !0,
        composed: !0
      })
    );
  }
  render() {
    const e = this.error ? "error" : "";
    return d`
      <label part="label">
        <input
          part="control"
          type="checkbox"
          .checked=${this.checked}
          name=${this.name || h}
          value=${this.value}
          ?disabled=${this.#h}
          ?required=${this.required}
          aria-invalid=${this.invalid ? "true" : "false"}
          aria-describedby=${e || h}
          @change=${this.#n}
        />
        <span>${this.label}<slot></slot></span>
      </label>
      ${this.error ? d`<p id="error" class="error" role="alert">${this.error}</p>` : h}
    `;
  }
}
s([
  i()
], t.prototype, "label");
s([
  i()
], t.prototype, "error");
s([
  i({ reflect: !0 })
], t.prototype, "name");
s([
  i()
], t.prototype, "value");
s([
  i({ type: Boolean, reflect: !0 })
], t.prototype, "checked");
s([
  i({ type: Boolean, reflect: !0 })
], t.prototype, "indeterminate");
s([
  i({ type: Boolean, reflect: !0 })
], t.prototype, "disabled");
s([
  i({ type: Boolean, reflect: !0 })
], t.prototype, "required");
s([
  i({ type: Boolean, reflect: !0 })
], t.prototype, "invalid");
v("mb-checkbox", t);
export {
  t as MbCheckbox
};
//# sourceMappingURL=checkbox.js.map
