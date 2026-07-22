import { LitElement as p, css as c, nothing as n, html as m } from "lit";
import { property as a } from "lit/decorators.js";
import { safeDefine as h } from "../lib/safe-define.js";
import { sharedStyles as u } from "../lib/styles.js";
var d = Object.defineProperty, i = (s, e, o, v) => {
  for (var r = void 0, l = s.length - 1, b; l >= 0; l--)
    (b = s[l]) && (r = b(e, o, r) || r);
  return r && d(e, o, r), r;
};
class t extends p {
  constructor() {
    super(...arguments), this.value = 0, this.max = 100, this.percent = null, this.label = "";
  }
  static {
    this.styles = [
      u,
      c`
      :host {
        display: block;
        inline-size: 100%;
      }

      .wrap {
        display: flex;
        flex-direction: column;
        gap: var(--mb-space-1);
      }

      .label {
        font-size: var(--mb-font-size-sm);
        color: var(--mb-color-muted);
      }

      .track {
        inline-size: 100%;
        block-size: 0.5rem;
        border-radius: var(--mb-radius-sm);
        background: var(--mb-color-border);
        overflow: clip;
      }

      .bar {
        block-size: 100%;
        background: var(--mb-color-accent);
        border-radius: inherit;
        transition: inline-size var(--mb-transition);
      }
    `
    ];
  }
  get #e() {
    if (this.percent != null && !Number.isNaN(this.percent))
      return Math.min(100, Math.max(0, this.percent));
    const e = this.max > 0 ? this.max : 100;
    return Math.min(100, Math.max(0, this.value / e * 100));
  }
  get #r() {
    return this.percent != null && !Number.isNaN(this.percent) ? this.#e : this.value;
  }
  get #t() {
    return this.percent != null && !Number.isNaN(this.percent) ? 100 : this.max > 0 ? this.max : 100;
  }
  render() {
    const e = this.#e;
    return m`
      <div class="wrap">
        ${this.label ? m`<div part="label" class="label" id="label">${this.label}</div>` : n}
        <div
          part="track"
          class="track"
          role="progressbar"
          aria-valuemin="0"
          aria-valuenow=${this.#r}
          aria-valuemax=${this.#t}
          aria-labelledby=${this.label ? "label" : n}
          aria-label=${this.label ? n : this.getAttribute("aria-label") || "Progress"}
        >
          <div part="bar" class="bar" style="inline-size: ${e}%"></div>
        </div>
        <slot></slot>
      </div>
    `;
  }
}
i([
  a({ type: Number })
], t.prototype, "value");
i([
  a({ type: Number })
], t.prototype, "max");
i([
  a({ type: Number })
], t.prototype, "percent");
i([
  a()
], t.prototype, "label");
h("mb-progress", t);
export {
  t as MbProgress
};
//# sourceMappingURL=progress.js.map
