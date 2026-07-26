import { LitElement as p, css as m, html as d } from "lit";
import { property as l } from "lit/decorators.js";
import { safeDefine as c } from "../lib/safe-define.js";
import { sharedStyles as b } from "../lib/styles.js";
var f = Object.defineProperty, a = (r, s, o, h) => {
  for (var e = void 0, i = r.length - 1, t; i >= 0; i--)
    (t = r[i]) && (e = t(s, o, e) || e);
  return e && f(s, o, e), e;
};
class n extends p {
  constructor() {
    super(...arguments), this.size = "md", this.label = "Loading";
  }
  static {
    this.styles = [
      b,
      m`
      :host {
        display: inline-flex;
        vertical-align: middle;
      }

      .spinner {
        border: 2px solid var(--mb-color-border);
        border-inline-end-color: var(--mb-color-accent);
        border-radius: 50%;
        animation: spin 0.7s linear infinite;
      }

      :host([size='sm']) .spinner {
        inline-size: 0.9rem;
        block-size: 0.9rem;
      }

      :host([size='md']) .spinner,
      :host(:not([size])) .spinner {
        inline-size: 1.15rem;
        block-size: 1.15rem;
      }

      @keyframes spin {
        to {
          transform: rotate(360deg);
        }
      }

      @media (prefers-reduced-motion: reduce) {
        .spinner {
          animation: none;
          border-inline-end-color: var(--mb-color-border);
          border-block-start-color: var(--mb-color-accent);
        }
      }
    `
    ];
  }
  render() {
    return d`
      <span
        part="spinner"
        class="spinner"
        role="status"
        aria-label=${this.label}
      ></span>
    `;
  }
}
a([
  l({ reflect: !0 })
], n.prototype, "size");
a([
  l()
], n.prototype, "label");
c("mb-spinner", n);
export {
  n as MbSpinner
};
//# sourceMappingURL=spinner.js.map
