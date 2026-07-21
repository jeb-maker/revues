import { LitElement as p, css as m, nothing as o, html as n } from "lit";
import { property as t } from "lit/decorators.js";
import { setFormValue as d } from "../lib/form.js";
import { safeDefine as h } from "../lib/safe-define.js";
import { sharedStyles as f } from "../lib/styles.js";
var u = Object.defineProperty, r = (l, i, s, g) => {
  for (var a = void 0, c = l.length - 1, b; c >= 0; c--)
    (b = l[c]) && (a = b(i, s, a) || a);
  return a && u(i, s, a), a;
};
class e extends p {
  constructor() {
    super(...arguments), this.variant = "primary", this.size = "md", this.type = "button", this.disabled = !1, this.loading = !1, this.name = "", this.value = "", this.href = "", this.target = "", this.rel = "", this.iconOnly = !1, this.#e = this.attachInternals(), this.#r = !1;
  }
  static {
    this.formAssociated = !0;
  }
  static {
    this.styles = [
      f,
      m`
      :host {
        display: inline-block;
      }

      .base {
        display: inline-flex;
        align-items: center;
        justify-content: center;
        gap: var(--mb-space-2);
        max-inline-size: 100%;
        border: 1px solid transparent;
        border-radius: var(--mb-radius-md);
        font: inherit;
        font-weight: 600;
        cursor: pointer;
        white-space: normal;
        text-align: center;
        text-decoration: none;
        overflow-wrap: anywhere;
        transition:
          background-color var(--mb-transition),
          color var(--mb-transition),
          border-color var(--mb-transition),
          opacity var(--mb-transition);
      }

      .base:disabled,
      .base[aria-disabled='true'] {
        cursor: not-allowed;
        opacity: 0.55;
        pointer-events: none;
      }

      :host([size='sm']) .base {
        min-block-size: 2rem;
        padding-inline: var(--mb-space-3);
        font-size: var(--mb-font-size-sm);
      }

      :host([size='md']) .base {
        min-block-size: 2.5rem;
        padding-inline: var(--mb-space-4);
        font-size: var(--mb-font-size-md);
      }

      :host([size='lg']) .base {
        min-block-size: 3rem;
        padding-inline: var(--mb-space-5);
        font-size: var(--mb-font-size-lg);
      }

      :host([icon-only][size='sm']) .base {
        min-inline-size: 2rem;
        padding-inline: 0;
      }

      :host([icon-only][size='md']) .base,
      :host([icon-only]:not([size])) .base {
        min-inline-size: 2.5rem;
        padding-inline: 0;
      }

      :host([icon-only][size='lg']) .base {
        min-inline-size: 3rem;
        padding-inline: 0;
      }

      :host([variant='primary']) .base {
        background: var(--mb-color-accent);
        color: var(--mb-color-on-accent);
      }

      :host([variant='secondary']) .base {
        background: var(--mb-color-surface);
        color: var(--mb-color-fg);
        border-color: var(--mb-color-border);
      }

      :host([variant='ghost']) .base {
        background: transparent;
        color: var(--mb-color-accent);
      }

      :host([variant='danger']) .base {
        background: var(--mb-color-danger);
        color: var(--mb-color-on-danger);
      }

      .spinner {
        inline-size: 1em;
        block-size: 1em;
        border: 2px solid currentColor;
        border-inline-end-color: transparent;
        border-radius: 50%;
        animation: spin 0.7s linear infinite;
      }

      @keyframes spin {
        to {
          transform: rotate(360deg);
        }
      }
    `
    ];
  }
  #e;
  #r;
  get #t() {
    return this.disabled || this.loading || this.#r;
  }
  get #i() {
    return !!this.href;
  }
  get #a() {
    return this.getAttribute("aria-label") ?? "";
  }
  formDisabledCallback(i) {
    this.#r = i, this.requestUpdate();
  }
  #s(i) {
    if (this.#t) {
      i.preventDefault(), i.stopImmediatePropagation();
      return;
    }
    if (this.#i) return;
    const s = this.#e.form;
    s && (this.type === "submit" ? (this.name && d(this.#e, this.value), s.requestSubmit(), queueMicrotask(() => d(this.#e, null))) : this.type === "reset" && s.reset());
  }
  render() {
    const i = n`
      ${this.loading ? n`<span class="spinner" aria-hidden="true"></span>` : o}
      <slot></slot>
    `, s = this.#a || o;
    return this.#i ? n`
        <a
          part="base"
          class="base"
          href=${this.#t ? o : this.href}
          target=${this.target || o}
          rel=${this.rel || (this.target === "_blank" ? "noopener noreferrer" : o)}
          aria-disabled=${this.#t ? "true" : "false"}
          aria-busy=${this.loading ? "true" : "false"}
          aria-label=${s}
          @click=${this.#s}
        >
          ${i}
        </a>
      ` : n`
      <button
        part="base"
        class="base"
        type="button"
        ?disabled=${this.#t}
        aria-busy=${this.loading ? "true" : "false"}
        aria-label=${s}
        @click=${this.#s}
      >
        ${i}
      </button>
    `;
  }
}
r([
  t({ reflect: !0 })
], e.prototype, "variant");
r([
  t({ reflect: !0 })
], e.prototype, "size");
r([
  t({ reflect: !0 })
], e.prototype, "type");
r([
  t({ type: Boolean, reflect: !0 })
], e.prototype, "disabled");
r([
  t({ type: Boolean, reflect: !0 })
], e.prototype, "loading");
r([
  t({ reflect: !0 })
], e.prototype, "name");
r([
  t()
], e.prototype, "value");
r([
  t({ reflect: !0 })
], e.prototype, "href");
r([
  t({ reflect: !0 })
], e.prototype, "target");
r([
  t({ reflect: !0 })
], e.prototype, "rel");
r([
  t({ type: Boolean, reflect: !0, attribute: "icon-only" })
], e.prototype, "iconOnly");
h("mb-button", e);
export {
  e as MbButton
};
//# sourceMappingURL=button.js.map
