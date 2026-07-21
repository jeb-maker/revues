import { LitElement as c, css as d, html as m } from "lit";
import { property as i } from "lit/decorators.js";
import { safeDefine as h } from "../lib/safe-define.js";
import { sharedStyles as b } from "../lib/styles.js";
var p = Object.defineProperty, r = (a, e, t, u) => {
  for (var s = void 0, n = a.length - 1, l; n >= 0; n--)
    (l = a[n]) && (s = l(e, t, s) || s);
  return s && p(e, t, s), s;
};
class o extends c {
  constructor() {
    super(...arguments), this.open = !1, this.variant = "info", this.autoDismiss = 4e3, this.message = "", this.#t = 0, this.#e = (e) => {
      const t = e.detail;
      t && (t.variant && (this.variant = t.variant), t.message != null && (this.message = t.message), t.autoDismiss != null && (this.autoDismiss = t.autoDismiss), this.show());
    };
  }
  static {
    this.styles = [
      b,
      d`
      :host {
        display: block;
        position: fixed;
        inset-block-end: var(--mb-space-5);
        inset-inline: var(--mb-space-4);
        z-index: 1000;
        pointer-events: none;
      }

      :host(:not([open])) {
        visibility: hidden;
      }

      .toast {
        pointer-events: auto;
        display: flex;
        align-items: flex-start;
        justify-content: space-between;
        gap: var(--mb-space-3);
        max-inline-size: 28rem;
        margin-inline: auto;
        padding-block: var(--mb-space-3);
        padding-inline: var(--mb-space-4);
        border-radius: var(--mb-radius-md);
        border: 1px solid var(--mb-color-border);
        background: var(--mb-color-surface);
        box-shadow: var(--mb-shadow);
        color: var(--mb-color-fg);
      }

      :host([variant='success']) .toast {
        border-color: var(--mb-color-success);
        background: var(--mb-color-success-soft);
        color: var(--mb-color-success);
      }

      :host([variant='danger']) .toast {
        border-color: var(--mb-color-danger);
        background: var(--mb-color-danger-soft);
        color: var(--mb-color-danger);
      }

      :host([variant='info']) .toast {
        border-color: var(--mb-color-info);
        background: var(--mb-color-info-soft);
        color: var(--mb-color-info);
      }

      .message {
        flex: 1;
        font-size: var(--mb-font-size-sm);
        font-weight: 600;
      }

      button {
        appearance: none;
        border: 0;
        background: transparent;
        color: inherit;
        cursor: pointer;
        font: inherit;
        font-weight: 700;
        line-height: 1;
        padding: 0;
      }
    `
    ];
  }
  #t;
  #e;
  connectedCallback() {
    super.connectedCallback(), document.addEventListener("mb-toast", this.#e);
  }
  disconnectedCallback() {
    super.disconnectedCallback(), document.removeEventListener("mb-toast", this.#e), this.#s();
  }
  updated(e) {
    e.has("open") && (this.open ? this.#o() : this.#s());
  }
  show(e, t) {
    e != null && (this.message = e), t && (this.variant = t), this.open = !0;
  }
  hide() {
    this.open = !1;
  }
  #o() {
    this.#s(), this.autoDismiss > 0 && (this.#t = window.setTimeout(() => this.hide(), this.autoDismiss));
  }
  #s() {
    this.#t && (window.clearTimeout(this.#t), this.#t = 0);
  }
  #i() {
    this.hide(), this.dispatchEvent(
      new CustomEvent("mb-close", { bubbles: !0, composed: !0 })
    );
  }
  render() {
    const e = this.variant === "danger" ? "alert" : "status";
    return m`
      <div
        part="toast"
        class="toast"
        role=${e}
        aria-live=${this.variant === "danger" ? "assertive" : "polite"}
        ?hidden=${!this.open}
      >
        <div part="message" class="message">${this.message}<slot></slot></div>
        <button type="button" part="close" aria-label="Dismiss" @click=${this.#i}>
          ×
        </button>
      </div>
    `;
  }
}
r([
  i({ type: Boolean, reflect: !0 })
], o.prototype, "open");
r([
  i({ reflect: !0 })
], o.prototype, "variant");
r([
  i({ type: Number, attribute: "auto-dismiss" })
], o.prototype, "autoDismiss");
r([
  i()
], o.prototype, "message");
h("mb-toast", o);
export {
  o as MbToast
};
//# sourceMappingURL=toast.js.map
