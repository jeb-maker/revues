import { LitElement as i, css as l, html as c } from "lit";
import { property as d } from "lit/decorators.js";
import { safeDefine as v } from "../lib/safe-define.js";
import { sharedStyles as m } from "../lib/styles.js";
var b = Object.defineProperty, f = (a, t, e, p) => {
  for (var r = void 0, o = a.length - 1, s; o >= 0; o--)
    (s = a[o]) && (r = s(t, e, r) || r);
  return r && b(t, e, r), r;
};
class n extends i {
  constructor() {
    super(...arguments), this.variant = "info";
  }
  static {
    this.styles = [
      m,
      l`
      :host {
        display: block;
        inline-size: 100%;
      }

      .alert {
        padding-block: var(--mb-space-3);
        padding-inline: var(--mb-space-4);
        border-radius: var(--mb-radius-md);
        border-inline-start: 4px solid currentColor;
        background: var(--mb-color-info-soft);
        color: var(--mb-color-info);
        overflow-wrap: anywhere;
        max-inline-size: 100%;
      }

      :host([variant='success']) .alert {
        background: var(--mb-color-success-soft);
        color: var(--mb-color-success);
      }

      :host([variant='warning']) .alert {
        background: var(--mb-color-warning-soft);
        color: var(--mb-color-warning);
      }

      :host([variant='danger']) .alert {
        background: var(--mb-color-danger-soft);
        color: var(--mb-color-danger);
      }
    `
    ];
  }
  get #r() {
    return this.variant === "warning" || this.variant === "danger" ? "alert" : "status";
  }
  render() {
    return c`
      <div part="base" class="alert" role=${this.#r}>
        <slot></slot>
      </div>
    `;
  }
}
f([
  d({ reflect: !0 })
], n.prototype, "variant");
v("mb-alert", n);
export {
  n as MbAlert
};
//# sourceMappingURL=alert.js.map
