import { LitElement as d, css as p, html as c } from "lit";
import { property as l } from "lit/decorators.js";
import { safeDefine as m } from "../lib/safe-define.js";
import { sharedStyles as h } from "../lib/styles.js";
var f = Object.defineProperty, n = (o, e, r, b) => {
  for (var i = void 0, t = o.length - 1, a; t >= 0; t--)
    (a = o[t]) && (i = a(e, r, i) || i);
  return i && f(e, r, i), i;
};
class s extends d {
  constructor() {
    super(...arguments), this.open = !1, this.heading = "", this.#i = !1;
  }
  static {
    this.styles = [
      h,
      p`
      :host {
        display: contents;
      }

      dialog {
        border: 1px solid var(--mb-color-border);
        border-radius: var(--mb-radius-lg);
        padding: 0;
        background: var(--mb-color-surface);
        color: var(--mb-color-fg);
        box-shadow: var(--mb-shadow);
        /* Avoid 100vw — it includes scrollbar gutters and overflows on mobile */
        inline-size: min(32rem, calc(100% - 2rem));
        max-inline-size: calc(100% - 2rem);
        margin: auto;
      }

      dialog::backdrop {
        background: rgb(20 32 27 / 45%);
      }

      .panel {
        display: flex;
        flex-direction: column;
        gap: var(--mb-space-4);
        padding: var(--mb-space-5);
        min-inline-size: 0;
        max-inline-size: 100%;
      }

      .header {
        display: flex;
        align-items: flex-start;
        justify-content: space-between;
        gap: var(--mb-space-3);
        min-inline-size: 0;
      }

      .title {
        font-family: var(--mb-font-display);
        font-size: var(--mb-font-size-xl);
        font-weight: 650;
        margin: 0;
        min-inline-size: 0;
        flex: 1;
        overflow-wrap: anywhere;
      }

      .close {
        border: 0;
        background: transparent;
        color: var(--mb-color-muted);
        font-size: 1.25rem;
        line-height: 1;
        cursor: pointer;
        padding: var(--mb-space-1);
        border-radius: var(--mb-radius-sm);
        flex-shrink: 0;
      }
    `
    ];
  }
  #e;
  #i;
  firstUpdated() {
    this.#e = this.renderRoot.querySelector("dialog") ?? void 0, this.#e?.addEventListener("close", () => {
      this.#i || (this.open && (this.open = !1), this.#t());
    }), this.#o();
  }
  updated(e) {
    e.has("open") && this.#o();
  }
  #o() {
    const e = this.#e;
    e && (this.open && !e.open ? e.showModal() : !this.open && e.open && (this.#i = !0, e.close(), this.#i = !1, this.#t()));
  }
  #t() {
    this.dispatchEvent(
      new CustomEvent("mb-close", {
        bubbles: !0,
        composed: !0
      })
    );
  }
  /** Close the modal (idempotent). Always emits `mb-close` when a close occurs. */
  close() {
    !this.open && !this.#e?.open || (this.open = !1);
  }
  #s() {
    this.close();
  }
  render() {
    return c`
      <dialog part="dialog" aria-labelledby="title" aria-modal="true">
        <div class="panel">
          <div class="header">
            <h2 class="title" id="title">${this.heading}<slot name="heading"></slot></h2>
            <button class="close" type="button" aria-label="Close" @click=${this.#s}>
              ×
            </button>
          </div>
          <div part="body">
            <slot></slot>
          </div>
          <div part="footer">
            <slot name="footer"></slot>
          </div>
        </div>
      </dialog>
    `;
  }
}
n([
  l({ type: Boolean, reflect: !0 })
], s.prototype, "open");
n([
  l()
], s.prototype, "heading");
m("mb-modal", s);
export {
  s as MbModal
};
//# sourceMappingURL=modal.js.map
