import { LitElement as h, css as n, html as c } from "lit";
import { state as l } from "lit/decorators.js";
import { safeDefine as f } from "../lib/safe-define.js";
import { sharedStyles as p } from "../lib/styles.js";
var b = Object.defineProperty, i = (t, e, o, m) => {
  for (var r = void 0, a = t.length - 1, d; a >= 0; a--)
    (d = t[a]) && (r = d(e, o, r) || r);
  return r && b(e, o, r), r;
};
class s extends h {
  constructor() {
    super(...arguments), this._hasHeader = !1, this._hasFooter = !1;
  }
  static {
    this.styles = [
      p,
      n`
      :host {
        display: block;
        inline-size: 100%;
      }

      .card {
        background: var(--mb-color-surface);
        border: 1px solid var(--mb-color-border);
        border-radius: var(--mb-radius-lg);
        overflow: clip;
        max-inline-size: 100%;
      }

      .header,
      .body,
      .footer {
        padding-block: var(--mb-space-4);
        padding-inline: var(--mb-space-5);
        min-inline-size: 0;
        overflow-wrap: anywhere;
      }

      .header {
        display: none;
        border-block-end: 1px solid var(--mb-color-border);
        font-family: var(--mb-font-display);
        font-weight: 650;
      }

      .footer {
        display: none;
        border-block-start: 1px solid var(--mb-color-border);
      }

      :host([data-has-header]) .header,
      :host([data-has-footer]) .footer {
        display: block;
      }

      ::slotted([slot='header']),
      ::slotted([slot='footer']) {
        display: block;
      }
    `
    ];
  }
  #e(e) {
    const o = e.target;
    this._hasHeader = o.assignedNodes({ flatten: !0 }).length > 0, this.toggleAttribute("data-has-header", this._hasHeader);
  }
  #o(e) {
    const o = e.target;
    this._hasFooter = o.assignedNodes({ flatten: !0 }).length > 0, this.toggleAttribute("data-has-footer", this._hasFooter);
  }
  render() {
    return c`
      <article part="card" class="card">
        <header class="header" part="header">
          <slot name="header" @slotchange=${this.#e}></slot>
        </header>
        <div class="body" part="body">
          <slot></slot>
        </div>
        <footer class="footer" part="footer">
          <slot name="footer" @slotchange=${this.#o}></slot>
        </footer>
      </article>
    `;
  }
}
i([
  l()
], s.prototype, "_hasHeader");
i([
  l()
], s.prototype, "_hasFooter");
f("mb-card", s);
export {
  s as MbCard
};
//# sourceMappingURL=card.js.map
