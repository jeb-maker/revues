import { LitElement as c, css as f, nothing as h, html as d } from "lit";
import { property as o, state as b } from "lit/decorators.js";
import { repeat as v } from "lit/directives/repeat.js";
import { setFormValue as y, constraintFlags as m, setValidity as g, clearValidity as $ } from "../lib/form.js";
import { safeDefine as O } from "../lib/safe-define.js";
import { sharedStyles as S, fieldStyles as q, fieldLabelState as _ } from "../lib/styles.js";
var B = Object.defineProperty, i = (a, t, e, n) => {
  for (var l = void 0, r = a.length - 1, p; r >= 0; r--)
    (p = a[r]) && (l = p(t, e, l) || l);
  return l && B(t, e, l), l;
};
function C(a) {
  if (!a) return [];
  try {
    const t = JSON.parse(a);
    return Array.isArray(t) ? t.filter(
      (e) => !!e && typeof e == "object" && typeof e.value == "string" && typeof e.label == "string"
    ).map((e) => ({
      value: e.value,
      label: e.label,
      disabled: !!e.disabled
    })) : [];
  } catch {
    return [];
  }
}
class s extends c {
  constructor() {
    super(...arguments), this.label = "", this.hint = "", this.error = "", this.value = "", this.name = "", this.disabled = !1, this.required = !1, this.invalid = !1, this.density = "default", this.hideLabel = !1, this.options = [], this._slottedOptions = [], this.#e = this.attachInternals(), this.#i = !1, this.#l = "", this.#r = !1, this.#s = !1;
  }
  static {
    this.formAssociated = !0;
  }
  static {
    this.styles = [
      S,
      q,
      f`
      :host {
        display: block;
      }

      slot[name='options'] {
        display: none;
      }
    `
    ];
  }
  #e;
  #i;
  #t;
  #l;
  #r;
  #s;
  get #h() {
    return this.disabled || this.#i;
  }
  get #d() {
    return this._slottedOptions.length ? this._slottedOptions : this.options;
  }
  get #p() {
    return this.getAttribute("aria-label") ?? "";
  }
  connectedCallback() {
    super.connectedCallback(), this.#r || (this.#l = this.value, this.#r = !0), this.#u();
  }
  firstUpdated() {
    this.#t = this.renderRoot.querySelector("select") ?? void 0, this.#n();
  }
  updated(t) {
    (t.has("value") || t.has("required") || t.has("error") || t.has("options") || t.has("_slottedOptions") || t.has("disabled") || t.has("name")) && this.#n();
  }
  formDisabledCallback(t) {
    this.#i = t, this.requestUpdate();
  }
  formResetCallback() {
    this.#s = !1, this.value = this.#l, this.error = "", this.invalid = !1;
  }
  #o(t) {
    return t instanceof HTMLOptionElement ? {
      value: t.value,
      label: t.label || t.textContent?.trim() || t.value,
      disabled: t.disabled
    } : null;
  }
  #u() {
    const t = [...this.querySelectorAll(":scope > option")].map((e) => this.#o(e)).filter((e) => e != null);
    t.length && (this._slottedOptions = t);
  }
  #c() {
    const t = this.renderRoot.querySelector('slot[name="options"]'), e = this.renderRoot.querySelector("slot:not([name])"), l = [
      ...t?.assignedElements({ flatten: !0 }) ?? [],
      ...e?.assignedElements({ flatten: !0 }) ?? []
    ].map((u) => this.#o(u)).filter((u) => u != null), r = JSON.stringify(this._slottedOptions), p = JSON.stringify(l);
    r !== p && (this._slottedOptions = l);
  }
  #a() {
    this.#c();
  }
  #n() {
    this.#t && this.#t.value !== this.value && (this.#t.value = this.value), y(this.#e, this.name ? this.value : null);
    const t = this.required && !this.value, { flags: e, message: n } = m(
      this.error,
      t,
      "Please select an option."
    );
    n ? (g(this.#e, e, n, this.#t), this.invalid = !!this.error || this.#s) : ($(this.#e), this.invalid = !1);
  }
  #f(t) {
    const e = t.target;
    this.#s = !0, this.value = e.value, this.dispatchEvent(
      new CustomEvent("mb-change", {
        detail: { value: this.value },
        bubbles: !0,
        composed: !0
      })
    );
  }
  render() {
    const t = [this.hint && !this.error ? "hint" : "", this.error ? "error" : ""].filter(Boolean).join(" "), { labelText: e, hideVisually: n, controlAriaLabel: l } = _(
      this.label,
      this.hideLabel,
      this.#p
    );
    return d`
      <div class="field">
        ${e ? d`<label
              part="label"
              class="label${n ? " visually-hidden" : ""}"
              for="control"
              >${e}</label
            >` : h}
        <select
          id="control"
          part="control"
          class="control"
          name=${this.name || h}
          ?disabled=${this.#h}
          ?required=${this.required}
          aria-invalid=${this.invalid ? "true" : "false"}
          aria-label=${l || h}
          aria-describedby=${t || h}
          .value=${this.value}
          @change=${this.#f}
        >
          <option value="" ?disabled=${this.required}></option>
          ${v(
      this.#d,
      (r) => r.value,
      (r) => d`
              <option value=${r.value} ?disabled=${!!r.disabled}>
                ${r.label}
              </option>
            `
    )}
        </select>
        ${this.hint && !this.error ? d`<p id="hint" class="hint">${this.hint}</p>` : h}
        ${this.error ? d`<p id="error" class="error" role="alert">${this.error}</p>` : h}
      </div>
      <slot name="options" @slotchange=${this.#a}></slot>
      <slot @slotchange=${this.#a}></slot>
    `;
  }
}
i([
  o()
], s.prototype, "label");
i([
  o()
], s.prototype, "hint");
i([
  o()
], s.prototype, "error");
i([
  o()
], s.prototype, "value");
i([
  o({ reflect: !0 })
], s.prototype, "name");
i([
  o({ type: Boolean, reflect: !0 })
], s.prototype, "disabled");
i([
  o({ type: Boolean, reflect: !0 })
], s.prototype, "required");
i([
  o({ type: Boolean, reflect: !0 })
], s.prototype, "invalid");
i([
  o({ reflect: !0 })
], s.prototype, "density");
i([
  o({ type: Boolean, reflect: !0, attribute: "hide-label" })
], s.prototype, "hideLabel");
i([
  o({
    attribute: "options",
    converter: {
      fromAttribute: C,
      toAttribute(a) {
        return a?.length ? JSON.stringify(a) : null;
      }
    }
  })
], s.prototype, "options");
i([
  b()
], s.prototype, "_slottedOptions");
O("mb-select", s);
export {
  s as MbSelect
};
//# sourceMappingURL=select.js.map
