import { LitElement as p, css as c, nothing as d, html as h } from "lit";
import { property as s } from "lit/decorators.js";
import { safeDefine as u } from "../lib/safe-define.js";
import { sharedStyles as m } from "../lib/styles.js";
var b = Object.defineProperty, r = (o, a, i, f) => {
  for (var e = void 0, l = o.length - 1, n; l >= 0; l--)
    (n = o[l]) && (e = n(a, i, e) || e);
  return e && b(a, i, e), e;
};
class t extends p {
  constructor() {
    super(...arguments), this.value = "", this.label = "", this.disabled = !1, this.checked = !1, this.name = "";
  }
  static {
    this.styles = [
      m,
      c`
      :host {
        display: block;
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
      }

      :host([disabled]) label {
        opacity: 0.55;
        cursor: not-allowed;
      }
    `
    ];
  }
  #e;
  firstUpdated() {
    this.#e = this.renderRoot.querySelector("input") ?? void 0;
  }
  focus(a) {
    this.#e?.focus(a);
  }
  #t() {
    this.checked = !0, this.dispatchEvent(
      new CustomEvent("mb-radio-select", {
        detail: { value: this.value },
        bubbles: !0,
        composed: !0
      })
    );
  }
  render() {
    return h`
      <label part="label">
        <input
          part="control"
          type="radio"
          name=${this.name || d}
          .value=${this.value}
          .checked=${this.checked}
          ?disabled=${this.disabled}
          @change=${this.#t}
        />
        <span>${this.label}<slot></slot></span>
      </label>
    `;
  }
}
r([
  s()
], t.prototype, "value");
r([
  s()
], t.prototype, "label");
r([
  s({ type: Boolean, reflect: !0 })
], t.prototype, "disabled");
r([
  s({ type: Boolean, reflect: !0 })
], t.prototype, "checked");
r([
  s()
], t.prototype, "name");
u("mb-radio", t);
export {
  t as MbRadio
};
//# sourceMappingURL=radio.js.map
