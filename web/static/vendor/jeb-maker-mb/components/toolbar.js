import { LitElement as t, css as e, html as s } from "lit";
import { safeDefine as a } from "../lib/safe-define.js";
import { sharedStyles as r } from "../lib/styles.js";
class l extends t {
  static {
    this.styles = [
      r,
      e`
      :host {
        display: block;
        inline-size: 100%;
      }

      .toolbar {
        display: flex;
        flex-wrap: wrap;
        align-items: center;
        justify-content: space-between;
        gap: var(--mb-space-3);
      }

      .start,
      .end {
        display: flex;
        flex-wrap: wrap;
        align-items: center;
        gap: var(--mb-space-2);
        min-inline-size: 0;
      }

      .end {
        margin-inline-start: auto;
      }
    `
    ];
  }
  render() {
    return s`
      <div part="toolbar" class="toolbar">
        <div part="start" class="start">
          <slot name="start"></slot>
          <slot></slot>
        </div>
        <div part="end" class="end">
          <slot name="end"></slot>
        </div>
      </div>
    `;
  }
}
a("mb-toolbar", l);
export {
  l as MbToolbar
};
//# sourceMappingURL=toolbar.js.map
