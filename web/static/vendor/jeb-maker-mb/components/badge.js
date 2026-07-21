import { LitElement as i, css as l, html as c } from "lit";
import { property as m } from "lit/decorators.js";
import { safeDefine as b } from "../lib/safe-define.js";
import { sharedStyles as p } from "../lib/styles.js";
var d = Object.defineProperty, f = (o, s, n, v) => {
  for (var r = void 0, a = o.length - 1, e; a >= 0; a--)
    (e = o[a]) && (r = e(s, n, r) || r);
  return r && d(s, n, r), r;
};
class t extends i {
  constructor() {
    super(...arguments), this.variant = "neutral";
  }
  static {
    this.styles = [
      p,
      l`
      :host {
        display: inline-flex;
      }

      span {
        display: inline-flex;
        align-items: center;
        gap: var(--mb-space-1);
        padding-block: 0.15rem;
        padding-inline: var(--mb-space-2);
        border-radius: var(--mb-radius-sm);
        font-size: var(--mb-font-size-sm);
        font-weight: 600;
        line-height: 1.3;
        background: var(--mb-color-border);
        color: var(--mb-color-fg);
      }

      :host([variant='success']) span {
        background: var(--mb-color-success-soft);
        color: var(--mb-color-success);
      }

      :host([variant='warning']) span {
        background: var(--mb-color-warning-soft);
        color: var(--mb-color-warning);
      }

      :host([variant='danger']) span {
        background: var(--mb-color-danger-soft);
        color: var(--mb-color-danger);
      }

      :host([variant='info']) span {
        background: var(--mb-color-info-soft);
        color: var(--mb-color-info);
      }
    `
    ];
  }
  render() {
    return c`<span part="base"><slot></slot></span>`;
  }
}
f([
  m({ reflect: !0 })
], t.prototype, "variant");
b("mb-badge", t);
export {
  t as MbBadge
};
//# sourceMappingURL=badge.js.map
