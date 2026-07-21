import { LitElement as p, css as c, nothing as u, html as h } from "lit";
import { property as a } from "lit/decorators.js";
import { setFormValue as f, constraintFlags as b, setValidity as m, clearValidity as v } from "../lib/form.js";
import { safeDefine as y } from "../lib/safe-define.js";
import { sharedStyles as g } from "../lib/styles.js";
import "./radio.js";
var k = Object.defineProperty, o = (s, e, t, l) => {
  for (var r = void 0, n = s.length - 1, d; n >= 0; n--)
    (d = s[n]) && (r = d(e, t, r) || r);
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
    super(...arguments), this.label = "", this.error = "", this.value = "", this.name = "", this.disabled = !1, this.required = !1, this.invalid = !1, this.options = [], this.#t = this.attachInternals(), this.#i = !1, this.#r = "", this.#a = !1, this.#e = !1, this.#l = (e) => {
      const t = e.detail?.value;
      t != null && (this.#e = !0, this.value = t, this.dispatchEvent(
        new CustomEvent("mb-change", {
          detail: { value: this.value },
          bubbles: !0,
          composed: !0
        })
      ));
    }, this.#n = (e) => {
      if (!["ArrowDown", "ArrowUp", "ArrowRight", "ArrowLeft"].includes(e.key)) return;
      const t = this.#d().filter((d) => !d.disabled);
      if (!t.length) return;
      e.preventDefault();
      const l = t.findIndex((d) => d.value === this.value), r = e.key === "ArrowDown" || e.key === "ArrowRight" ? 1 : -1, n = t[(l + r + t.length) % t.length];
      this.#e = !0, this.value = n.value, n.focus(), this.dispatchEvent(
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
      c`
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
  #i;
  #r;
  #a;
  #e;
  get #o() {
    return this.disabled || this.#i;
  }
  connectedCallback() {
    super.connectedCallback(), this.#a || (this.#r = this.value, this.#a = !0), this.addEventListener("mb-radio-select", this.#l), this.addEventListener("keydown", this.#n);
  }
  disconnectedCallback() {
    super.disconnectedCallback(), this.removeEventListener("mb-radio-select", this.#l), this.removeEventListener("keydown", this.#n);
  }
  firstUpdated() {
    this.#s(), this.#h();
  }
  updated(e) {
    (e.has("value") || e.has("name") || e.has("disabled") || e.has("options")) && this.#s(), (e.has("value") || e.has("required") || e.has("error") || e.has("name") || e.has("disabled")) && this.#h();
  }
  formDisabledCallback(e) {
    this.#i = e, this.requestUpdate(), this.#s();
  }
  formResetCallback() {
    this.#e = !1, this.value = this.#r, this.error = "", this.invalid = !1;
  }
  #d() {
    const e = this.renderRoot.querySelector("slot")?.assignedElements({ flatten: !0 }).filter((l) => l.localName === "mb-radio") ?? [], t = [
      ...this.renderRoot.querySelectorAll(".options > mb-radio")
    ];
    return [...e, ...t];
  }
  #s() {
    const e = this.#d();
    for (const t of e)
      t.name = this.name || "mb-radio-group", t.checked = t.value === this.value, this.#o && (t.disabled = !0);
  }
  #h() {
    f(this.#t, this.name ? this.value : null);
    const e = this.required && !this.value, { flags: t, message: l } = b(
      this.error,
      e,
      "Please select an option."
    );
    l ? (m(this.#t, t, l), this.invalid = !!this.error || this.#e) : (v(this.#t), this.invalid = !1);
  }
  #l;
  #n;
  #u() {
    this.#s();
  }
  render() {
    return h`
      <fieldset part="fieldset" ?disabled=${this.#o}>
        ${this.label ? h`<legend part="legend">${this.label}</legend>` : u}
        <div class="options" part="options" role="radiogroup" aria-invalid=${this.invalid ? "true" : "false"}>
          <slot @slotchange=${this.#u}></slot>
          ${this.options.map(
      (e) => h`
              <mb-radio
                .value=${e.value}
                .label=${e.label}
                ?disabled=${!!e.disabled || this.#o}
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
