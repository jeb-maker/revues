import { LitElement as p, css as m, html as c } from "lit";
import { property as n } from "lit/decorators.js";
import { safeDefine as d } from "../lib/safe-define.js";
import { sharedStyles as f } from "../lib/styles.js";
var v = Object.defineProperty, i = (a, r, s, b) => {
  for (var e = void 0, t = a.length - 1, l; t >= 0; t--)
    (l = a[t]) && (e = l(r, s, e) || e);
  return e && v(r, s, e), e;
};
class o extends p {
  constructor() {
    super(...arguments), this.label = "Primary", this.open = !1;
  }
  static {
    this.styles = [
      f,
      m`
      :host {
        display: block;
      }

      :host([hidden]) {
        display: none !important;
      }

      nav {
        display: flex;
        flex-wrap: wrap;
        align-items: center;
        gap: var(--mb-space-1) var(--mb-space-3);
      }

      ::slotted(a) {
        color: var(--mb-color-muted);
        font-weight: 600;
        font-size: var(--mb-font-size-sm);
        text-decoration: none;
        padding-block: var(--mb-space-2);
        padding-inline: var(--mb-space-2);
        border-radius: var(--mb-radius-sm);
      }

      ::slotted(a:hover) {
        color: var(--mb-color-fg);
      }

      ::slotted(a[aria-current='page']),
      ::slotted(a.is-active) {
        color: var(--mb-color-accent);
        background: var(--mb-color-accent-soft);
      }

      ::slotted(a:focus-visible) {
        outline: var(--mb-focus-ring);
        outline-offset: var(--mb-focus-offset);
      }

      @media (max-width: 36rem) {
        :host(:not([open]):not([data-always-visible])) {
          display: none;
        }
      }
    `
    ];
  }
  render() {
    return c`
      <nav part="nav" aria-label=${this.label}>
        <slot></slot>
      </nav>
    `;
  }
}
i([
  n()
], o.prototype, "label");
i([
  n({ type: Boolean, reflect: !0 })
], o.prototype, "open");
d("mb-nav", o);
export {
  o as MbNav
};
//# sourceMappingURL=nav.js.map
