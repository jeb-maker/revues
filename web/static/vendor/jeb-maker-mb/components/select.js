import { LitElement as c, css as f, nothing as h, html as p } from "lit";
import { property as l, state as b } from "lit/decorators.js";
import { repeat as v } from "lit/directives/repeat.js";
import { setFormValue as y, constraintFlags as m, setValidity as g, clearValidity as $ } from "../lib/form.js";
import { safeDefine as O } from "../lib/safe-define.js";
import { sharedStyles as S, fieldStyles as q, fieldLabelState as _ } from "../lib/styles.js";
var B = Object.defineProperty, i = (a, t, e, n) => {
  for (var r = void 0, o = a.length - 1, d; o >= 0; o--)
    (d = a[o]) && (r = d(t, e, r) || r);
  return r && B(t, e, r), r;
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
    super(...arguments), this.label = "", this.hint = "", this.error = "", this.value = "", this.name = "", this.disabled = !1, this.required = !1, this.invalid = !1, this.density = "default", this.hideLabel = !1, this.placeholder = "", this.options = [], this._slottedOptions = [], this.#e = this.attachInternals(), this.#i = !1, this.#l = "", this.#r = !1, this.#s = !1;
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
  get #p() {
    return this.disabled || this.#i;
  }
  get #o() {
    return this._slottedOptions.length ? this._slottedOptions : this.options;
  }
  /** Non-empty options rendered after the placeholder option. */
  get #d() {
    return this.#o.filter((t) => t.value !== "");
  }
  get #u() {
    const t = this.#o.find((e) => e.value === "");
    return t?.label ? t.label : this.placeholder;
  }
  get #c() {
    return this.getAttribute("aria-label") ?? "";
  }
  connectedCallback() {
    super.connectedCallback(), this.#r || (this.#l = this.value, this.#r = !0), this.#f();
  }
  firstUpdated() {
    this.#t = this.renderRoot.querySelector("select") ?? void 0, this.#h();
  }
  updated(t) {
    (t.has("value") || t.has("required") || t.has("error") || t.has("options") || t.has("_slottedOptions") || t.has("disabled") || t.has("name")) && this.#h();
  }
  formDisabledCallback(t) {
    this.#i = t, this.requestUpdate();
  }
  formResetCallback() {
    this.#s = !1, this.value = this.#l, this.error = "", this.invalid = !1;
  }
  #a(t) {
    return t instanceof HTMLOptionElement ? {
      value: t.value,
      label: t.label || t.textContent?.trim() || t.value,
      disabled: t.disabled
    } : null;
  }
  #f() {
    const t = [...this.querySelectorAll(":scope > option")].map((e) => this.#a(e)).filter((e) => e != null);
    t.length && (this._slottedOptions = t);
  }
  #b() {
    const t = this.renderRoot.querySelector('slot[name="options"]'), e = this.renderRoot.querySelector("slot:not([name])"), r = [
      ...t?.assignedElements({ flatten: !0 }) ?? [],
      ...e?.assignedElements({ flatten: !0 }) ?? []
    ].map((u) => this.#a(u)).filter((u) => u != null), o = JSON.stringify(this._slottedOptions), d = JSON.stringify(r);
    o !== d && (this._slottedOptions = r);
  }
  #n() {
    this.#b();
  }
  #h() {
    this.#t && this.#t.value !== this.value && (this.#t.value = this.value), y(this.#e, this.name ? this.value : null);
    const t = this.required && !this.value, { flags: e, message: n } = m(
      this.error,
      t,
      "Please select an option."
    );
    n ? (g(this.#e, e, n, this.#t), this.invalid = !!this.error || this.#s) : ($(this.#e), this.invalid = !1);
  }
  #v(t) {
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
    const t = [this.hint && !this.error ? "hint" : "", this.error ? "error" : ""].filter(Boolean).join(" "), { labelText: e, hideVisually: n, controlAriaLabel: r } = _(
      this.label,
      this.hideLabel,
      this.#c
    );
    return p`
      <div class="field">
        ${e ? p`<label
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
          ?disabled=${this.#p}
          ?required=${this.required}
          aria-invalid=${this.invalid ? "true" : "false"}
          aria-label=${r || h}
          aria-describedby=${t || h}
          .value=${this.value}
          @change=${this.#v}
        >
          <option value="" ?disabled=${this.required}>${this.#u}</option>
          ${v(
      this.#d,
      (o) => o.value,
      (o) => p`
              <option value=${o.value} ?disabled=${!!o.disabled}>
                ${o.label}
              </option>
            `
    )}
        </select>
        ${this.hint && !this.error ? p`<p id="hint" class="hint">${this.hint}</p>` : h}
        ${this.error ? p`<p id="error" class="error" role="alert">${this.error}</p>` : h}
      </div>
      <slot name="options" @slotchange=${this.#n}></slot>
      <slot @slotchange=${this.#n}></slot>
    `;
  }
}
i([
  l()
], s.prototype, "label");
i([
  l()
], s.prototype, "hint");
i([
  l()
], s.prototype, "error");
i([
  l()
], s.prototype, "value");
i([
  l({ reflect: !0 })
], s.prototype, "name");
i([
  l({ type: Boolean, reflect: !0 })
], s.prototype, "disabled");
i([
  l({ type: Boolean, reflect: !0 })
], s.prototype, "required");
i([
  l({ type: Boolean, reflect: !0 })
], s.prototype, "invalid");
i([
  l({ reflect: !0 })
], s.prototype, "density");
i([
  l({ type: Boolean, reflect: !0, attribute: "hide-label" })
], s.prototype, "hideLabel");
i([
  l()
], s.prototype, "placeholder");
i([
  l({
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
