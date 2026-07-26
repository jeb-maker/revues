import { LitElement as c, css as f, nothing as b, html as l } from "lit";
import { property as i } from "lit/decorators.js";
import { repeat as m } from "lit/directives/repeat.js";
import { safeDefine as u } from "../lib/safe-define.js";
import { sharedStyles as h } from "../lib/styles.js";
var d = Object.defineProperty, p = (t, r, e, y) => {
  for (var a = void 0, n = t.length - 1, s; n >= 0; n--)
    (s = t[n]) && (a = s(r, e, a) || a);
  return a && d(r, e, a), a;
};
function v(t) {
  if (!t) return [];
  try {
    const r = JSON.parse(t);
    return Array.isArray(r) ? r.filter(
      (e) => !!e && typeof e == "object" && typeof e.label == "string"
    ).map((e) => ({
      label: e.label,
      href: e.href,
      current: !!e.current
    })) : [];
  } catch {
    return [];
  }
}
class o extends c {
  constructor() {
    super(...arguments), this.label = "Breadcrumb", this.items = [];
  }
  static {
    this.styles = [
      h,
      f`
      :host {
        display: block;
      }

      nav ol {
        display: flex;
        flex-wrap: wrap;
        align-items: center;
        gap: var(--mb-space-1) var(--mb-space-2);
        margin: 0;
        padding: 0;
        list-style: none;
        font-size: var(--mb-font-size-sm);
      }

      li,
      ::slotted(li) {
        display: inline-flex;
        align-items: center;
        gap: var(--mb-space-2);
        min-inline-size: 0;
        list-style: none;
      }

      li:not(:last-child)::after {
        content: '/';
        color: var(--mb-color-muted);
      }

      a {
        color: var(--mb-color-accent);
        text-decoration: none;
        overflow-wrap: anywhere;
      }

      a:hover {
        text-decoration: underline;
      }

      [aria-current='page'] {
        color: var(--mb-color-muted);
        font-weight: 600;
      }
    `
    ];
  }
  render() {
    return l`
      <nav part="nav" aria-label=${this.label}>
        <ol part="list">
          ${this.items.length ? m(
      this.items,
      (r) => `${r.href ?? ""}:${r.label}`,
      (r) => l`
                  <li part="item">
                    ${r.current || !r.href ? l`<span aria-current=${r.current ? "page" : b}
                          >${r.label}</span
                        >` : l`<a href=${r.href}>${r.label}</a>`}
                  </li>
                `
    ) : l`<slot></slot>`}
        </ol>
      </nav>
    `;
  }
}
p([
  i()
], o.prototype, "label");
p([
  i({
    attribute: "items",
    converter: {
      fromAttribute: v,
      toAttribute(t) {
        return t?.length ? JSON.stringify(t) : null;
      }
    }
  })
], o.prototype, "items");
u("mb-breadcrumbs", o);
export {
  o as MbBreadcrumbs
};
//# sourceMappingURL=breadcrumbs.js.map
