import { LitElement as p, css as d, nothing as b, html as m } from "lit";
import { property as r } from "lit/decorators.js";
import { safeDefine as u } from "../lib/safe-define.js";
import { sharedStyles as h } from "../lib/styles.js";
var f = Object.defineProperty, n = (i, e, a, c) => {
  for (var t = void 0, s = i.length - 1, l; s >= 0; s--)
    (l = i[s]) && (t = l(e, a, t) || t);
  return t && f(e, a, t), t;
};
class o extends p {
  constructor() {
    super(...arguments), this.expanded = !1, this.for = "", this.labelOpen = "Menu", this.labelClose = "Close menu";
  }
  static {
    this.styles = [
      h,
      d`
      :host {
        display: none;
      }

      @media (max-width: 36rem) {
        :host {
          display: inline-flex;
        }
      }

      button {
        appearance: none;
        display: inline-flex;
        align-items: center;
        justify-content: center;
        min-inline-size: 2.5rem;
        min-block-size: 2.5rem;
        border: 1px solid var(--mb-color-border);
        border-radius: var(--mb-radius-md);
        background: var(--mb-color-surface);
        color: var(--mb-color-fg);
        cursor: pointer;
        font: inherit;
        font-weight: 700;
      }
    `
    ];
  }
  #e() {
    this.expanded = !this.expanded;
    const e = this.for ? document.getElementById(this.for) : null;
    e && (e.toggleAttribute("open", this.expanded), "open" in e && (e.open = this.expanded)), this.dispatchEvent(
      new CustomEvent("mb-toggle", {
        detail: { expanded: this.expanded },
        bubbles: !0,
        composed: !0
      })
    );
  }
  render() {
    return m`
      <button
        part="button"
        type="button"
        aria-expanded=${this.expanded ? "true" : "false"}
        aria-controls=${this.for || b}
        aria-label=${this.expanded ? this.labelClose : this.labelOpen}
        @click=${this.#e}
      >
        <slot>${this.expanded ? "✕" : "☰"}</slot>
      </button>
    `;
  }
}
n([
  r({ type: Boolean, reflect: !0 })
], o.prototype, "expanded");
n([
  r({ attribute: "for" })
], o.prototype, "for");
n([
  r({ attribute: "label-open" })
], o.prototype, "labelOpen");
n([
  r({ attribute: "label-close" })
], o.prototype, "labelClose");
u("mb-nav-toggle", o);
export {
  o as MbNavToggle
};
//# sourceMappingURL=nav-toggle.js.map
