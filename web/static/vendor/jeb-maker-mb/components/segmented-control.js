import { LitElement as l, css as n, html as c } from "lit";
import { property as d } from "lit/decorators.js";
import { safeDefine as b } from "../lib/safe-define.js";
import { sharedStyles as m } from "../lib/styles.js";
var p = Object.defineProperty, f = (r, t, s, u) => {
  for (var e = void 0, o = r.length - 1, a; o >= 0; o--)
    (a = r[o]) && (e = a(t, s, e) || e);
  return e && p(t, s, e), e;
};
class i extends l {
  constructor() {
    super(...arguments), this.label = "Filters";
  }
  static {
    this.styles = [
      m,
      n`
      :host {
        display: block;
        max-inline-size: 100%;
      }

      .scroller {
        overflow-x: auto;
        -webkit-overflow-scrolling: touch;
        max-inline-size: 100%;
      }

      .list {
        display: inline-flex;
        min-inline-size: 100%;
        gap: 0;
        padding: var(--mb-space-1);
        border: 1px solid var(--mb-color-border);
        border-radius: var(--mb-radius-md);
        background: var(--mb-color-surface);
      }

      ::slotted(a),
      ::slotted(button) {
        appearance: none;
        border: 0;
        background: transparent;
        color: var(--mb-color-muted);
        font: inherit;
        font-weight: 600;
        font-size: var(--mb-font-size-sm);
        text-decoration: none;
        padding-block: var(--mb-space-2);
        padding-inline: var(--mb-space-3);
        border-radius: var(--mb-radius-sm);
        white-space: nowrap;
        cursor: pointer;
      }

      ::slotted(a:focus-visible),
      ::slotted(button:focus-visible) {
        outline: var(--mb-focus-ring);
        outline-offset: var(--mb-focus-offset);
      }

      ::slotted([aria-current='page']),
      ::slotted([aria-selected='true']),
      ::slotted(.is-active) {
        background: var(--mb-color-accent-soft);
        color: var(--mb-color-accent);
      }
    `
    ];
  }
  render() {
    return c`
      <nav part="nav" class="scroller" aria-label=${this.label}>
        <div part="list" class="list" role="list">
          <slot></slot>
        </div>
      </nav>
    `;
  }
}
f([
  d()
], i.prototype, "label");
b("mb-segmented-control", i);
export {
  i as MbSegmentedControl
};
//# sourceMappingURL=segmented-control.js.map
