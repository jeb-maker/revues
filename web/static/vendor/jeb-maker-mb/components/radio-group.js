import { LitElement as p, css as f, nothing as u, html as h } from "lit";
import { property as a } from "lit/decorators.js";
import { setFormValue as c, constraintFlags as b, setValidity as m, clearValidity as v } from "../lib/form.js";
import { safeDefine as y } from "../lib/safe-define.js";
import { sharedStyles as g } from "../lib/styles.js";
import "./radio.js";
var k = Object.defineProperty, o = (s, e, t, d) => {
  for (var r = void 0, l = s.length - 1, n; l >= 0; l--)
    (n = s[l]) && (r = n(e, t, r) || r);
  return r && k(e, t, r), r;
};
function w(s) {
  if (!s) return [];
  try {
    const e = JSON.parse(s);
    return Array.isArray(e) ? e.filter(
      (t) => !!t && typeof t == "object" && typeof t.value == "string" && typeof t.label == "string"
    ).map((t) => ({
      value: t.value,
      label: t.label,
      disabled: !!t.disabled
    })) : [];
  } catch {
    return [];
  }
}
class i extends p {
  constructor() {
    super(...arguments), this.label = "", this.error = "", this.value = "", this.name = "", this.disabled = !1, this.required = !1, this.invalid = !1, this.options = [], this.#t = this.attachInternals(), this.#r = !1, this.#a = "", this.#o = !1, this.#e = !1, this.#s = /* @__PURE__ */ new WeakMap(), this.#n = (e) => {
      const t = e.detail?.value;
      t != null && (this.#e = !0, this.value = t, this.dispatchEvent(
        new CustomEvent("mb-change", {
          detail: { value: this.value },
          bubbles: !0,
          composed: !0
        })
      ));
    }, this.#d = (e) => {
      if (!["ArrowDown", "ArrowUp", "ArrowRight", "ArrowLeft"].includes(e.key)) return;
      const t = this.#u().filter((n) => !n.disabled);
      if (!t.length) return;
      e.preventDefault();
      const d = t.findIndex((n) => n.value === this.value), r = e.key === "ArrowDown" || e.key === "ArrowRight" ? 1 : -1, l = t[(d + r + t.length) % t.length];
      this.#e = !0, this.value = l.value, l.focus(), this.dispatchEvent(
        new CustomEvent("mb-change", {
          detail: { value: this.value },
          bubbles: !0,
          composed: !0
        })
      );
    };
  }
  static {
    this.formAssociated = !0;
  }
  static {
    this.styles = [
      g,
      f`
      :host {
        display: block;
      }

      fieldset {
        margin: 0;
        padding: 0;
        border: 0;
        min-inline-size: 0;
      }

      legend {
        font-size: var(--mb-font-size-sm);
        font-weight: 600;
        margin-block-end: var(--mb-space-2);
      }

      .options {
        display: flex;
        flex-direction: column;
        gap: var(--mb-space-2);
      }

      .error {
        margin: var(--mb-space-2) 0 0;
        color: var(--mb-color-danger);
        font-size: var(--mb-font-size-sm);
      }
    `
    ];
  }
  #t;
  #r;
  #a;
  #o;
  #e;
  #s;
  get #l() {
    return this.disabled || this.#r;
  }
  connectedCallback() {
    super.connectedCallback(), this.#o || (this.#a = this.value, this.#o = !0), this.addEventListener("mb-radio-select", this.#n), this.addEventListener("keydown", this.#d);
  }
  disconnectedCallback() {
    super.disconnectedCallback(), this.removeEventListener("mb-radio-select", this.#n), this.removeEventListener("keydown", this.#d);
  }
  firstUpdated() {
    this.#i(), this.#p();
  }
  updated(e) {
    (e.has("value") || e.has("name") || e.has("disabled") || e.has("options")) && this.#i(), (e.has("value") || e.has("required") || e.has("error") || e.has("name") || e.has("disabled")) && this.#p();
  }
  formDisabledCallback(e) {
    this.#r = e, this.requestUpdate(), this.#i();
  }
  formResetCallback() {
    this.#e = !1, this.value = this.#a, this.error = "", this.invalid = !1;
  }
  #h() {
    return this.renderRoot.querySelector("slot")?.assignedElements({ flatten: !0 }).filter((e) => e.localName === "mb-radio") ?? [];
  }
  #u() {
    const e = [
      ...this.renderRoot.querySelectorAll(".options > mb-radio")
    ];
    return [...this.#h(), ...e];
  }
  #i() {
    const e = this.#u();
    for (const t of e)
      t.name = this.name || "mb-radio-group", t.checked = t.value === this.value;
    for (const t of this.#h())
      this.#s.has(t) || this.#s.set(t, t.disabled), t.disabled = this.#l || !!this.#s.get(t);
  }
  #p() {
    c(this.#t, this.name ? this.value : null);
    const e = this.required && !this.value, { flags: t, message: d } = b(
      this.error,
      e,
      "Please select an option."
    );
    d ? (m(this.#t, t, d), this.invalid = !!this.error || this.#e) : (v(this.#t), this.invalid = !1);
  }
  #n;
  #d;
  #f() {
    this.#i();
  }
  render() {
    return h`
      <fieldset part="fieldset" ?disabled=${this.#l}>
        ${this.label ? h`<legend part="legend">${this.label}</legend>` : u}
        <div class="options" part="options" role="radiogroup" aria-invalid=${this.invalid ? "true" : "false"}>
          <slot @slotchange=${this.#f}></slot>
          ${this.options.map(
      (e) => h`
              <mb-radio
                .value=${e.value}
                .label=${e.label}
                ?disabled=${!!e.disabled || this.#l}
                ?checked=${e.value === this.value}
                .name=${this.name || "mb-radio-group"}
              ></mb-radio>
            `
    )}
        </div>
        ${this.error ? h`<p class="error" role="alert">${this.error}</p>` : u}
      </fieldset>
    `;
  }
}
o([
  a()
], i.prototype, "label");
o([
  a()
], i.prototype, "error");
o([
  a()
], i.prototype, "value");
o([
  a({ reflect: !0 })
], i.prototype, "name");
o([
  a({ type: Boolean, reflect: !0 })
], i.prototype, "disabled");
o([
  a({ type: Boolean, reflect: !0 })
], i.prototype, "required");
o([
  a({ type: Boolean, reflect: !0 })
], i.prototype, "invalid");
o([
  a({
    attribute: "options",
    converter: {
      fromAttribute: w,
      toAttribute(s) {
        return s?.length ? JSON.stringify(s) : null;
      }
    }
  })
], i.prototype, "options");
y("mb-radio-group", i);
export {
  i as MbRadioGroup
};
//# sourceMappingURL=radio-group.js.map
