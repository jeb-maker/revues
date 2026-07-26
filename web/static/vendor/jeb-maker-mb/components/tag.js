import { LitElement as f, css as p, html as n } from "lit";
import { property as l } from "lit/decorators.js";
import { safeDefine as d } from "../lib/safe-define.js";
import { sharedStyles as c } from "../lib/styles.js";
var h = Object.defineProperty, m = (r, i, a, b) => {
  for (var e = void 0, t = r.length - 1, o; t >= 0; t--)
    (o = r[t]) && (e = o(i, a, e) || e);
  return e && h(i, a, e), e;
};
class s extends f {
  constructor() {
    super(...arguments), this.href = "", this.size = "md";
  }
  static {
    this.styles = [
      c,
      p`
      :host {
        display: inline-flex;
        max-inline-size: 100%;
      }

      .tag {
        display: inline-flex;
        align-items: center;
        gap: var(--mb-space-1);
        max-inline-size: 100%;
        padding-block: 0.15rem;
        padding-inline: var(--mb-space-2);
        border: 1px solid var(--mb-color-border);
        border-radius: var(--mb-radius-sm);
        background: var(--mb-color-surface);
        color: var(--mb-color-fg);
        font-size: var(--mb-font-size-sm);
        font-weight: 600;
        line-height: 1.3;
        text-decoration: none;
        overflow-wrap: anywhere;
      }

      :host([size='sm']) .tag {
        font-size: 0.75rem;
        padding-inline: 0.4rem;
      }
    `
    ];
  }
  render() {
    return this.href ? n`
        <a part="base" class="tag" href=${this.href}>
          <slot></slot>
        </a>
      ` : n`<span part="base" class="tag"><slot></slot></span>`;
  }
}
m([
  l({ reflect: !0 })
], s.prototype, "href");
m([
  l({ reflect: !0 })
], s.prototype, "size");
d("mb-tag", s);
export {
  s as MbTag
};
//# sourceMappingURL=tag.js.map
