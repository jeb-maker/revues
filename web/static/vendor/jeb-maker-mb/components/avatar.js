import { LitElement as m, css as f, nothing as c, html as n } from "lit";
import { property as s, state as h } from "lit/decorators.js";
import { safeDefine as d } from "../lib/safe-define.js";
import { sharedStyles as v } from "../lib/styles.js";
var u = Object.defineProperty, i = (a, e, l, g) => {
  for (var t = void 0, o = a.length - 1, p; o >= 0; o--)
    (p = a[o]) && (t = p(e, l, t) || t);
  return t && u(e, l, t), t;
};
class r extends m {
  constructor() {
    super(...arguments), this.src = "", this.alt = "", this.name = "", this.size = "md", this._failed = !1;
  }
  static {
    this.styles = [
      v,
      f`
      :host {
        display: inline-flex;
        vertical-align: middle;
      }

      .avatar {
        display: inline-flex;
        align-items: center;
        justify-content: center;
        overflow: clip;
        border-radius: 50%;
        background: var(--mb-color-accent-soft);
        color: var(--mb-color-accent);
        font-weight: 700;
        line-height: 1;
        user-select: none;
      }

      :host([size='sm']) .avatar {
        inline-size: 1.75rem;
        block-size: 1.75rem;
        font-size: 0.7rem;
      }

      :host([size='md']) .avatar,
      :host(:not([size])) .avatar {
        inline-size: 2.25rem;
        block-size: 2.25rem;
        font-size: 0.8rem;
      }

      img {
        inline-size: 100%;
        block-size: 100%;
        object-fit: cover;
      }
    `
    ];
  }
  get #e() {
    const e = this.name.trim().split(/\s+/).filter(Boolean);
    return e.length ? e.length === 1 ? e[0].slice(0, 2).toUpperCase() : (e[0][0] + e[e.length - 1][0]).toUpperCase() : "?";
  }
  #t() {
    this._failed = !0;
  }
  updated(e) {
    e.has("src") && (this._failed = !1);
  }
  render() {
    const e = !!this.src && !this._failed;
    return n`
      <span part="base" class="avatar" role=${e ? c : "img"} aria-label=${e ? c : this.alt || this.name || "Avatar"}>
        ${e ? n`<img part="image" src=${this.src} alt=${this.alt} @error=${this.#t} />` : n`<span part="initials">${this.#e}</span>`}
      </span>
    `;
  }
}
i([
  s({ reflect: !0 })
], r.prototype, "src");
i([
  s()
], r.prototype, "alt");
i([
  s()
], r.prototype, "name");
i([
  s({ reflect: !0 })
], r.prototype, "size");
i([
  h()
], r.prototype, "_failed");
d("mb-avatar", r);
export {
  r as MbAvatar
};
//# sourceMappingURL=avatar.js.map
