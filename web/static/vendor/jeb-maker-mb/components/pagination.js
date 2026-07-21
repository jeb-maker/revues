import { LitElement as b, css as d, html as r } from "lit";
import { property as t } from "lit/decorators.js";
import { safeDefine as v } from "../lib/safe-define.js";
import { sharedStyles as c } from "../lib/styles.js";
var m = Object.defineProperty, a = (o, i, l, f) => {
  for (var s = void 0, n = o.length - 1, p; n >= 0; n--)
    (p = o[n]) && (s = p(i, l, s) || s);
  return s && m(i, l, s), s;
};
class e extends b {
  constructor() {
    super(...arguments), this.prevUrl = "", this.nextUrl = "", this.prevDisabled = !1, this.nextDisabled = !1, this.status = "", this.prevLabel = "Previous", this.nextLabel = "Next", this.label = "Pagination";
  }
  static {
    this.styles = [
      c,
      d`
      :host {
        display: block;
      }

      nav {
        display: flex;
        flex-wrap: wrap;
        align-items: center;
        justify-content: space-between;
        gap: var(--mb-space-3);
      }

      .status {
        color: var(--mb-color-muted);
        font-size: var(--mb-font-size-sm);
      }

      .actions {
        display: inline-flex;
        gap: var(--mb-space-2);
      }

      a,
      span.disabled {
        display: inline-flex;
        align-items: center;
        min-block-size: 2.25rem;
        padding-inline: var(--mb-space-3);
        border: 1px solid var(--mb-color-border);
        border-radius: var(--mb-radius-md);
        background: var(--mb-color-surface);
        color: var(--mb-color-fg);
        font-size: var(--mb-font-size-sm);
        font-weight: 600;
        text-decoration: none;
      }

      span.disabled {
        opacity: 0.45;
        cursor: not-allowed;
      }
    `
    ];
  }
  render() {
    const i = this.prevDisabled || !this.prevUrl, l = this.nextDisabled || !this.nextUrl;
    return r`
      <nav part="nav" aria-label=${this.label}>
        <div part="status" class="status">${this.status}<slot name="status"></slot></div>
        <div part="actions" class="actions">
          <slot name="prev">
            ${i ? r`<span class="disabled" aria-disabled="true">${this.prevLabel}</span>` : r`<a part="prev" href=${this.prevUrl}>${this.prevLabel}</a>`}
          </slot>
          <slot name="next">
            ${l ? r`<span class="disabled" aria-disabled="true">${this.nextLabel}</span>` : r`<a part="next" href=${this.nextUrl}>${this.nextLabel}</a>`}
          </slot>
        </div>
      </nav>
    `;
  }
}
a([
  t({ attribute: "prev-url" })
], e.prototype, "prevUrl");
a([
  t({ attribute: "next-url" })
], e.prototype, "nextUrl");
a([
  t({ type: Boolean, attribute: "prev-disabled" })
], e.prototype, "prevDisabled");
a([
  t({ type: Boolean, attribute: "next-disabled" })
], e.prototype, "nextDisabled");
a([
  t()
], e.prototype, "status");
a([
  t({ attribute: "prev-label" })
], e.prototype, "prevLabel");
a([
  t({ attribute: "next-label" })
], e.prototype, "nextLabel");
a([
  t()
], e.prototype, "label");
v("mb-pagination", e);
export {
  e as MbPagination
};
//# sourceMappingURL=pagination.js.map
