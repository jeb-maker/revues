import { LitElement as d, css as c, html as r } from "lit";
import { property as p } from "lit/decorators.js";
import { safeDefine as m } from "../lib/safe-define.js";
import { sharedStyles as h } from "../lib/styles.js";
var g = Object.defineProperty, f = (e, a, s, i) => {
  for (var t = void 0, o = e.length - 1, n; o >= 0; o--)
    (n = e[o]) && (t = n(a, s, t) || t);
  return t && g(a, s, t), t;
};
class l extends d {
  constructor() {
    super(...arguments), this.heading = "";
  }
  static {
    this.styles = [
      h,
      c`
      :host {
        display: block;
        inline-size: 100%;
      }

      .panel {
        display: flex;
        flex-direction: column;
        align-items: flex-start;
        gap: var(--mb-space-3);
        padding-block: var(--mb-space-6);
        padding-inline: var(--mb-space-5);
        border: 1px dashed var(--mb-color-border);
        border-radius: var(--mb-radius-lg);
        background: var(--mb-color-surface);
      }

      .heading {
        margin: 0;
        font-family: var(--mb-font-display);
        font-size: var(--mb-font-size-lg);
        font-weight: 650;
        line-height: var(--mb-line-height-tight);
        color: var(--mb-color-fg);
      }

      .body {
        color: var(--mb-color-muted);
        font-size: var(--mb-font-size-md);
      }

      .actions {
        display: flex;
        flex-wrap: wrap;
        gap: var(--mb-space-2);
      }

      .actions:not([data-has-content]) {
        display: none;
      }
    `
    ];
  }
  #t(a) {
    const i = a.target.assignedNodes({ flatten: !0 }).length > 0;
    this.renderRoot.querySelector(".actions")?.toggleAttribute("data-has-content", i);
  }
  render() {
    return r`
      <div part="panel" class="panel">
        ${this.heading ? r`<h2 part="heading" class="heading">${this.heading}</h2>` : r`<slot name="heading"></slot>`}
        <div part="body" class="body">
          <slot></slot>
        </div>
        <div part="actions" class="actions">
          <slot name="actions" @slotchange=${this.#t}></slot>
        </div>
      </div>
    `;
  }
}
f([
  p()
], l.prototype, "heading");
m("mb-empty-state", l);
export {
  l as MbEmptyState
};
//# sourceMappingURL=empty-state.js.map
