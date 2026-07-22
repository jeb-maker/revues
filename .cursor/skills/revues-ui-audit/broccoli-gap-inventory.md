# Revues × @jeb-maker/mb — gap inventory

Date: 2026-07-20  
Scope: read-only scan of `web/templates/**`, `web/static/css/**` vs `@jeb-maker/mb` v0.1.0 (README + `src/components/*`, Storybook).  
Do **not** open GitHub issues from this doc yet — drafts below are for parent triage.

**mb today:** `mb-button`, `mb-input`, `mb-textarea`, `mb-select`, `mb-checkbox`, `mb-badge`, `mb-alert`, `mb-card`, `mb-modal` + `tokens.css` + typography classes `.mb-title` / `.mb-body` / `.mb-body-sm`. Peer: `lit`.

**Legend — Gap?**
- **No** — usable as-is (maybe thin host CSS / attribute mapping)
- **Partial** — component exists but missing API/variant/SSR mode required by Revues
- **Yes** — no mb equivalent; need new component or explicit “host CSS forever” decision

---

## Mapping table

| Pattern Revues | Usage (pages / partials) | mb equivalent | Gap? | Proposed component/API for issue |
|----------------|--------------------------|---------------|------|----------------------------------|
| **Button primary** (`.button`) | Almost all pages; CTAs in toolbars, forms, empty states, page-actions | `mb-button` `variant="primary"` | Partial | Loading exists; need **href / render-as-`<a>`** (Revues styles many CTAs as `<a class="button">`). Document `type="submit"` + Form Association for Go forms. |
| **Button secondary** (`.button-secondary`) | Lists, exports, filters, admin | `mb-button` `variant="secondary"` | Partial | Same link-button gap; density `size="sm"` ≈ `.button-sm`. |
| **Button ghost** (`.button-ghost`) | Header, cancel, clear filters, inline table actions | `mb-button` `variant="ghost"` | Partial | mb ghost is accent-colored; Revues ghost is bordered text. Decide token remap or `ghost`/`ghost-muted` variants. |
| **Button danger** (`.button-danger`) | Archive template (`checklist_template_form`), destructive admin | — | **Yes** | `mb-button` **`variant="danger"`** (uses `--mb-color-danger` / on-danger). |
| **Icon-only / compact button** (`.button--icon`, ↑↓× in editor) | `base` page-actions; `checklist_template_form` reorder | — | Partial | `size="sm"` + **`icon-only`** prop (`aria-label` required) or slotted leading icon API. |
| **Link styled as text** (`.link-button`) | Rare | — | Yes (P1) | Optional `mb-button variant="link"` or keep host CSS. |
| **Text input** (`stack-form` `<input type="text">`) | Forms: subjects, templates, bug report, Jira, admin labels… | `mb-input` `type="text"` | Partial | Map `label`/`hint`/`error` props ← `.field-hint` / `.field-error`. Need **`maxlength`**, **`autocomplete`**, SSR-friendly attrs. |
| **Search input** (toolbar `type="search"`) | `runs_list`, `my_tasks`, `templates_index`, `subjects_list`, `checklist_templates_list`, `run_wizard_subjects` | `mb-input` `type="search"` | Partial | Need **compact / unlabeled** mode (`aria-label` or `label` + `hide-label`) + toolbar width. |
| **Email / password / url** | `admin_smtp`, `admin_jira`, `admin_notion`, `admin_webhooks`, `admin_users`, `subject_show` invite | `mb-input` `email`/`password`/`url` | Partial | Same attr gaps (`autocomplete`, placeholders for “leave blank to keep”). |
| **Number input** (`type="number"` port) | `admin_smtp` | — | **Yes** | Extend `mb-input` with `type="number"` + `min`/`max`/`step`. |
| **File upload** (`type="file"` + accept) | `run_item_show` attachments | — | **Yes** | **`mb-file-input`** (or `mb-input type="file"`): label, hint, accept, Form Associated, multipart-safe. |
| **Textarea** (stack forms + bug report) | Many forms; Jira create | `mb-textarea` | Partial | `rows`, label/hint/error OK; need `maxlength`. |
| **Inline compact textarea** (`.comment-inline`) | `run_item_row_fragment` (HTMX blur/Enter) | `mb-textarea` | **Yes** (API) | **`density="compact"`** / `rows="1"` auto-grow; **no block label**; honor HTML **`form="…"`** association to external form id. |
| **Select (labeled stack)** | Filters on `run_show`; org switcher; admin jump; many admin forms | `mb-select` | **Partial → blocker** | Options today are **`options` JS property only** — unusable from Go `html/template` without a boot script. Need **slotted `<option>`** and/or **`options` JSON attribute** parsed on connect. |
| **Compact status select** (`.status-select`) | Run grid HTMX | `mb-select` | **Yes** (API) | Compact + unlabeled + `form=` + native `change` (or composed `mb-change` documented for HTMX). |
| **Compact assign select** (`.assign-select`) | Run grid HTMX | `mb-select` | **Yes** (API) | Same; empty option “Non assigné”. |
| **Checkbox** | Template required flags; admin policies/webhooks/SMTP TLS | `mb-checkbox` | Partial | Works for labeled rows; ensure `name`/`value`/`checked` SSR attrs; group layout stays host. |
| **Radio group** | `org_select`, `admin_jira` instance type | — | **Yes** | **`mb-radio-group`** + `mb-radio` (form-associated, keyboard). |
| **Status / role badge** (`.status-badge.status-*`, `.role-badge`) | Runs, tasks, items, header role, privacy “Privé”, integrations enabled/disabled | `mb-badge` variants `neutral|success|warning|danger|info` | Partial | Document mapping: `ok/done/enabled`→success, `nok`→danger, `na`→warning, `in_progress/admin`→info, `pending/draft/archived`→neutral. Optional **`pill`** shape + status tokens. |
| **Tag / chip** (`.tag` domains/labels) | `templates_index`, `subject_show`, lists | — | **Yes** (P1) | **`mb-tag`** (neutral chip, not status). |
| **Flash text success/error** (`.success` / `.error` `role=status|alert`) | Nearly every mutating page | `mb-alert` | Partial | Revues often uses **inline text**, not panels. Either add `mb-alert density="inline"` / `variant` + no border, or keep host classes and use `mb-alert` for panels only. |
| **Warning / danger panels** (`.warning-panel`, `.login-alert`, `.form-danger-zone`) | `login`, `run_show` NOK list, archive zone | `mb-alert` + `mb-card` | Partial | Alert covers warning/danger; **danger-zone card** may need `mb-card` tone or host CSS. |
| **Invite banner** (`.invite-banner`) | `base` layout | `mb-alert` | Partial | Horizontal banner + inline form actions — compose alert + buttons; or host. |
| **Card / section** (`.card`) | Ubiquitous content chrome | `mb-card` (slots header/body/footer) | Partial | Revues puts `<h2>` in default slot; migration = move titles to `slot="header"`. No API gap if slots accepted. |
| **Empty state** (`.empty-state` + actions) | `runs_list`, `my_tasks`, `templates_index`, `run_wizard_subjects`, subjects lists… | — | **Yes** | **`mb-empty-state`**: title, description slot, actions slot. |
| **List toolbar** (`.list-toolbar` search + actions) | All major list pages | — | **Yes** (P1) | **`mb-toolbar`** layout primitive (CSS + slots: start/end) — or host forever. |
| **Segmented filter tabs** (`.segmented-tabs`) | `runs_list` (Tous/En cours/Terminées), `my_tasks` (statuts) | — | **Yes** | **`mb-segmented-control`** (link or button items, `aria-current` / `selected`). |
| **Data table** (`.data-table`, sticky, `--cards` mobile) | Lists + run items + admin membership | — | **Yes** (host OK) | Prefer **host CSS pattern** (tables + HTMX rows). Optional Storybook “table recipe”; full `mb-table` is heavy — **not P0** unless DS wants ownership. |
| **Progress bar** (`.run-progress*`, `.data-table__progress`, `.run-card__bar`) | `run_show` fragment, runs list / subject show cards | — | **Yes** | **`mb-progress`**: `value`/`max` or `percent`, optional label, `aria-*`. |
| **Pagination** (`.pagination`) | `runs_list` | — | **Yes** | **`mb-pagination`**: prev/next links, status text, `aria-disabled`. |
| **Breadcrumbs** (`.breadcrumb`) | `base` when ≥2 crumbs | — | **Yes** | **`mb-breadcrumbs`**: list of `{href,label}` or slotted links + separators. |
| **Page header / title / actions** (`.page-title`, `.page-header`, `.page-actions`) | Layout shell | Typography `.mb-title` + host | Partial | Keep layout in host; map H1 to `.mb-title` or tokens. |
| **Site nav** (`.site-nav__link.is-active`) | `site_nav` | — | **Yes** (P1) | **`mb-nav`** / nav-link recipe; active state + separators. |
| **Hamburger + responsive nav** | `base` ≤36rem | — | **Yes** (P1) | **`mb-nav-toggle`** or document pattern with `mb-button` icon + host CSS. |
| **Admin nav** (`.admin-nav` + jump `<select>`) | Admin pages via `admin_nav` | — | Yes (P1) | Subnav / jump-select pattern; can stay host + `mb-select` once SSR options land. |
| **Org / account switcher** | `base` header | compose `mb-select` + `mb-button` | Partial | Blocked on select SSR; compact header density. |
| **Avatar + account chip** | `base` header | — | Yes (P1) | **`mb-avatar`** (image + fallback initials). |
| **Toast** (`.toast`, JS-filled on `run_show`) | Run check HTMX feedback | — | **Yes** | **`mb-toast`** / toaster: show/hide API, success/error, `aria-live`. |
| **Spinner / HTMX indicator** (`.spinner`, `.htmx-indicator`) | Run item rows | `mb-button` loading spinner only | Partial | Standalone **`mb-spinner`** for HTMX indicators. |
| **Modal / confirm** | Clôture uses `hx-confirm` (native); no `<dialog>` in templates today | `mb-modal` | Partial | Exists for future richer confirms; **document HTMX + `mb-close`**. Not blocking migration if `hx-confirm` kept. |
| **Help callout** (`.help-text`, `.closing-note`) | Run items, closing note | `mb-alert variant="info"` | Partial | Closing-note accent bar ≈ alert; may keep host. |
| **Event timeline** (`.event-timeline`) | `run_item_show` | — | Yes (P1) | **`mb-timeline`** or host-only recipe. |
| **Attachment thumb / download chip** | `attachment_display_fragment`, run row PJ | — | Yes (P1) | Host media pattern; optional `mb-file-chip`. |
| **Collapsible details/summary** | Template editor, subject meta, wizard advanced | — | Yes (P1) | Native `<details>` OK; optional `mb-disclosure`. |
| **Skip link / visually-hidden** | `base` | — | No (host a11y) | Stay in host CSS. |
| **Icons (SVG sprite via `icon`)** | Widespread | — | Yes (P1) | DS icon set or keep Revues inline SVG helper. |
| **Template editor layout** | `checklist_template_form` + `editor.css` | compose primitives | Yes (app) | App-specific; not a DS primitive. |
| **Tokens / theme** (`app.css` `:root` light/dark) | Global | `tokens.css` (`--mb-*`, Fraunces/Source Sans) | **Partial** | Coexistence plan (see Integration). Revues system fonts + blue accent vs mb moss/paper — deliberate remapping or dual token layers. |

---

## Coverage notes (Revues-heavy patterns)

Highest-frequency migration surfaces:

1. **Shell** — header, hamburger, site nav, breadcrumbs, page title/actions, footer (`base.html`, `site_nav.html`).
2. **Lists** — toolbar + search + (segmented tabs) + data-table--cards + empty states (`runs_list`, `my_tasks`, `templates_index`, `subjects_list`, …).
3. **Run execution** — progress, filter bar, status/assign selects, comment textarea, HTMX row swap, toast, spinner (`run_show`, `run_item_row_fragment`, `run.css`).
4. **Forms** — stack-form cards, field hint/error, checkboxes, danger zone (`checklist_template_form`, admin_*, `subject_form`, `bug_report`).
5. **Detail chrome** — badges, tags, warning panels, attachments, timeline (`run_item_show`, `subject_show`).

---

## Prioritized issues to open on `jeb-maker/miniature-broccoli`

### P0 — must-have before Revues can migrate core UI

#### 1. SSR-friendly `mb-select` options (blocker)

**Title:** `mb-select`: support slotted `<option>` and/or JSON `options` attribute for non-JS hosts  

**Body draft:**
```
## Context
Revues is Go + html/template + HTMX (no SPA). Selects are rendered server-side with
selected options (status, assignee, org switcher, filters, admin jump).

## Problem
`mb-select` only accepts `options` as a JS property (`attribute: false`).
Templates cannot set Lit array properties. Forcing a client hydrator defeats MPA goals.

## Proposal
1. Prefer light-DOM / slotted native `<option>` (and `<optgroup>`) mirrored into the internal `<select>`, OR
2. Accept `options='[{"value":"ok","label":"OK"}]'` as a reflected attribute (JSON parse on connect/update).
3. Document SSR: `value` + `name` + `required` as attributes before upgrade.
4. Preserve form-associated behavior and `mb-change`.

## Acceptance
- Storybook example with zero JS option wiring beyond importing the CE.
- Vitest: slotted options + selected value submit via FormData.
```

#### 2. Compact unlabeled form controls (run grid)

**Title:** Form controls: compact density + aria-label-only mode for table/HTMX cells  

**Body draft:**
```
## Context
Revues run checklist grid uses compact `.status-select`, `.assign-select`, and
`.comment-inline` textarea inside table cells, often with `form="item-form-N"`
pointing at a visually-hidden form, driven by HTMX on change/blur.

## Proposal
For `mb-select`, `mb-input`, `mb-textarea` (and checkbox if needed):
- `density="compact" | "default"`
- Allow empty `label` when `aria-label` (or `label` + `hide-label`) is set
- Honor the HTML `form` attribute (associate with form id outside the cell)
- Ensure composed events (`mb-change` / `mb-input`) bubble through shadow for HTMX
  (document whether HTMX should listen to `mb-change` or native events)

## Acceptance
- Compact controls fit ~2.1rem row height; no forced block label margin.
- FormData includes values when control uses `form="..."`.
- Storybook “table cell” example.
```

#### 3. `mb-button`: danger + link appearance

**Title:** `mb-button`: add `danger` variant and optional `href` (anchor rendering)  

**Body draft:**
```
## Context
Revues uses `.button-danger` for archive/destructive actions and extensively styles
`<a class="button|button-secondary|button-ghost">` as CTAs (page-actions, toolbars,
empty states) to keep progressive enhancement without JS.

## Proposal
- `variant="danger"` using danger tokens.
- If `href` is set, render a styled `<a>` (or wrap) with the same visual variants;
  omit form association when acting as a link.
- Optional `icon-only` boolean requiring accessible name.

## Acceptance
- Visual variants documented in Storybook.
- Link buttons are keyboard-focusable and do not submit forms accidentally.
```

#### 4. `mb-progress`

**Title:** Add `mb-progress` for determinate progress bars  

**Body draft:**
```
## Context
Revues shows run completion as a bar + “done/total” on run detail and as a
CSS progress cue on run list cards (`--run-progress`).

## Proposal
`mb-progress` with `value`/`max` or `percent`, optional slotted/label text,
correct `role="progressbar"` + ARIA valuemin/valuenow/valuemax.
Parts: track, bar, label.

## Acceptance
- Storybook; unit test for ARIA; works without JS beyond CE definition once attributes set.
```

#### 5. `mb-segmented-control` (filter tabs)

**Title:** Add `mb-segmented-control` for URL filter tabs  

**Body draft:**
```
## Context
Revues `/revues` and `/mes-taches` use `.segmented-tabs` as navigational filters
(plain links with `.is-active` / aria semantics), not JS tabs.

## Proposal
Component or documented recipe:
- Slot or items with `href` + `selected`
- Horizontal scroll on narrow viewports
- Active styles via tokens

## Acceptance
- Can be fully SSR’d as links; no client router required.
```

#### 6. `mb-empty-state`

**Title:** Add `mb-empty-state` layout primitive  

**Body draft:**
```
## Context
Most Revues list pages have a dedicated empty-state block (title, muted copy, CTA buttons).

## Proposal
`mb-empty-state` with default slot (body), optional `heading` prop or header slot,
and `actions` slot. Use surface/border tokens; no illustration required for v1.

## Acceptance
- Storybook; usable inside/outside `mb-card`.
```

#### 7. `mb-pagination`

**Title:** Add `mb-pagination` (prev/next + status)  

**Body draft:**
```
## Context
Revues paginates `/revues` (25/page) with prev/next anchors and “Page X sur Y (total)”.

## Proposal
Light component: slots or props for prevUrl/nextUrl/disabled, status text.
No full page-number matrix required for v1.

## Acceptance
- SSR attributes; aria-label on nav; disabled prev/next not focusable or aria-disabled.
```

#### 8. `mb-toast` (or toast helper)

**Title:** Add `mb-toast` for transient status messages  

**Body draft:**
```
## Context
Run item HTMX updates surface feedback via a fixed `.toast` region on run show.

## Proposal
`mb-toast` with variants success/danger, `open`/`show()` API, auto-dismiss option,
`role="status"|"alert"` + aria-live. Prefer single instance / event-based API.

## Acceptance
- Works when toast host is outside swapped HTMX targets.
- Reduced-motion: no transform requirement.
```

#### 9. File input + number input

**Title:** Extend inputs: `type="file"` (or `mb-file-input`) and `type="number"`  

**Body draft:**
```
## Context
Revues uploads evidence (JPEG/PNG/WebP/PDF ≤5MB) on run items and uses number
inputs (SMTP port). Current `mb-input` types: text|email|password|search|url|tel.

## Proposal
- number: min/max/step + Form Association
- file: accept, multiple(false), label/hint/error; document multipart + FormData behavior
  with FACE (file inputs have known constraints — document clearly)

## Acceptance
- Storybook; FormData includes file when using native form submit.
```

#### 10. Radio group

**Title:** Add `mb-radio-group` / `mb-radio`  

**Body draft:**
```
## Context
Revues uses radios for org selection and Jira Cloud vs Server.

## Proposal
Form-associated radio group with name, value, required, options SSR-friendly
(slotted radios or JSON). Keyboard arrow navigation.

## Acceptance
- Vitest FormData + fieldset disabled.
```

---

### P1 — nice-to-have / second wave

#### 11. Breadcrumbs

**Title:** Add `mb-breadcrumbs`  
**Body:** SSR list of links + current page (`aria-current="page"`); separator part; collapse-on-mobile optional (Revues hides ancestors on small screens via CSS).

#### 12. Tags / chips

**Title:** Add `mb-tag` for non-status labels  
**Body:** Domains and subject tags in Revues are pill chips distinct from status badges. Neutral surface; optional removable later.

#### 13. Spinner + toolbar layout

**Title:** Add `mb-spinner` and document `mb-toolbar` layout  
**Body:** Standalone spinner for HTMX indicators; toolbar = flex slots start/end matching Revues list toolbars (can be CSS-only export).

#### 14. App shell nav patterns

**Title:** Document or add `mb-nav-link` + mobile nav toggle pattern  
**Body:** Active state, short/long labels, hamburger `aria-expanded`. Revues may keep shell CSS; DS should provide tokenized recipe for consistency across apps.

#### 15. Avatar

**Title:** Add `mb-avatar`  
**Body:** Image URL + alt + size sm/md; fallback initials. Used in Revues header account.

#### 16. Timeline / disclosure / file chip

**Title:** Recipes: timeline, disclosure, attachment chip  
**Body:** Lower priority. Revues can keep host CSS (`event-timeline`, `<details>`, attachment thumb). Publish Storybook recipes before full components.

#### 17. Alert inline density + badge status tokens

**Title:** `mb-alert` inline density; optional status token aliases on `mb-badge`  
**Body:** Align with Revues `.success`/`.error` text flashes and `status-ok|nok|na|pending` vocabulary (aliases → existing variants OK if documented).

#### 18. Select optgroup + header compact select

**Title:** `mb-select`: `<optgroup>` + compact header style  
**Body:** Admin jump select uses optgroups; org switcher is dense in the header.

#### 19. Host integration guide (MPA / Go / HTMX)

**Title:** Docs: consuming `@jeb-maker/mb` from Go html/template + HTMX  
**Body:** Cover custom element tags in templates, FOUC (`:not(:defined)`), CSRF-compatible forms, FACE + multipart, event names for `hx-trigger`, token coexistence, JS budget (Lit peer), no root barrel imports. Link from README.

#### 20. Tokens coexistence / dark scheme

**Title:** Tokens: dark `color-scheme` story + non-global body reset option  
**Body:** `tokens.css` currently sets `html, body` font/background. Revues `app.css` already owns shell + dark media query. Provide `tokens-core.css` without body reset, or document override order. mb lacks dark semantic flip today; Revues has `prefers-color-scheme: dark`.

---

## Integration constraints (Go + html/template host)

### Custom elements in templates

- Emit tags directly: `<mb-button type="submit" variant="primary">Enregistrer</mb-button>`.
- Load **atomic ESM** entries from a small bundled/static module (no SPA). Respect Revues JS budget (**≤15 Ko** for HTMX+app — Lit peer may force budget renegotiation or selective adoption).
- Use mb’s `:not(:defined) { visibility: hidden }` FOUC guard; ensure CSS + module order in `base.html`.
- Progressive enhancement: critical flows (login, check items) must still work if CE fail to upgrade — prefer retaining native controls for P0 run grid until compact FACE APIs land.

### Form-associated custom elements

- Prefer FACE controls so classic `<form method="post">` + CSRF hidden fields keep working.
- **Multipart uploads** need an explicit file-control story (FACE + `FormData` file bits).
- Revues relies on HTML **`form="id"`** association for split controls/actions — mb must honor this (not only being a descendant of `<form>`).
- `fieldset disabled` is already in mb contract — good for admin forms.

### Shadow DOM labels

- Do **not** use outer `<label for="x">` targeting internals — set `label="…"`, or unlabeled + `aria-label` for compact cells.
- Migrate `.field-hint` / `.field-error` → `hint` / `error` properties (string attrs).
- Server-rendered validation: set `error="…"` and `invalid` on first paint after POST.

### `mb-select` SSR (critical)

- Until slotted/JSON options ship, **do not migrate selects** (status/assign/filters/org).
- Avoid a one-off “options JSON in data-attribute + inline script” hydrator in Revues if the DS can own the contract.

### CSS tokens coexistence with `app.css`

| Concern | Guidance |
|---------|----------|
| Variable namespaces | Keep `--mb-*` vs Revues `--accent/--surface`; map via a thin `mb-bridge.css` (`--mb-color-accent: var(--accent)` **or** the reverse once design commits to mb palette). |
| Global resets | `tokens.css` styles `html, body` and ships display fonts — **conflicts** with Revues system stack and shell layout. Prefer importing semantic tokens without body rules, or scope mb fonts to `.mb-theme` wrapper. |
| Dark mode | Revues implements dark via `@media (prefers-color-scheme: dark)`. mb semantic tokens are light-paper oriented — need DS dark tokens or bridge overrides. |
| Component CSS | Shadow styles won’t see `app.css` classes; use `::part` / CSS variables only. Host layout (tables, shell, run grid) stays in `app.css` / `run.css` / `editor.css`. |
| Budgets | Revues CI: CSS core/total gzip caps. mb tokens + fonts (woff2) may blow budget — subset fonts or don’t ship Fraunces into app chrome. |

### HTMX

- Swaps replace DOM nodes: upgraded CEs inside swap targets are destroyed/recreated — OK if attributes carry state.
- Prefer listening to **composed** events (`mb-change`, `mb-input`, `mb-close`) in `hx-trigger`, or keep native internals if DS re-exports them.
- Toast/progress OOB fragments (`hx-swap-oob`) should target stable host ids outside CE shadows.
- `hx-confirm` can remain; `mb-modal` is optional upgrade path.

### What should stay host-owned (recommend)

- App shell grid (`site-header`, `site-shell`, footer)
- Data tables + responsive card-rows
- Template editor
- Run-section collapse
- Attachment image pipeline UI
- Skip link / visually-hidden / icon helper

These are product layout, not atomic DS — recipes in Storybook beat new components.

---

## Suggested adoption order (Revues side, later)

1. Tokens bridge + typography only (no CE) — validate CSS budget.  
2. `mb-button` / `mb-badge` / `mb-alert` / `mb-card` on static forms (after danger + link API).  
3. `mb-input` / `mb-textarea` / `mb-checkbox` on admin + template forms.  
4. **Blocked** until P0 #1–#2: selects + run grid.  
5. Progress, segmented tabs, empty state, pagination, toast.  
6. Shell nav/breadcrumbs when P1 ships or keep host.

---

## Sources

- Revues: `web/templates/**`, `web/static/css/app.css`, `run.css`, `editor.css`, `.cursor/skills/revues-ui-audit/decisions.md`
- mb: https://github.com/jeb-maker/miniature-broccoli (README, `src/components/*`, `src/tokens/*`, Storybook https://jeb-maker.github.io/miniature-broccoli/)

## Issues ouvertes

Repo: [jeb-maker/miniature-broccoli](https://github.com/jeb-maker/miniature-broccoli) — opened 2026-07-20.

**Adoption Revues :** vague 1 démarrée contre **`@jeb-maker/mb@0.2.0`** (tag `v0.2.0` — pas de `v0.3.0` publié). Voir `web/static/vendor/jeb-maker-mb/README.md` et `decisions.md`.

| # inventaire | Issue | Titre | Statut côté Revues |
|--------------|-------|-------|--------------------|
| P0-1 | https://github.com/jeb-maker/miniature-broccoli/issues/7 | mb-select SSR options | **Shipped in 0.2.0** — grille points pas encore migrée (HTMX `mb-change`) |
| P0-2 | https://github.com/jeb-maker/miniature-broccoli/issues/8 | Compact form controls | **Shipped in 0.2.0** — idem grille |
| P0-3 | https://github.com/jeb-maker/miniature-broccoli/issues/9 | mb-button danger + href | **Shipped** — pilot `/login`, empty-state `/revues` |
| P0-4 | https://github.com/jeb-maker/miniature-broccoli/issues/10 | mb-progress | **Shipped** — fragments progress fiche revue |
| P0-5 | https://github.com/jeb-maker/miniature-broccoli/issues/11 | mb-segmented-control | **Shipped** — pilot `/revues` |
| P0-6 | https://github.com/jeb-maker/miniature-broccoli/issues/12 | mb-empty-state | **Shipped** — pilot `/revues` |
| P0-7 | https://github.com/jeb-maker/miniature-broccoli/issues/13 | mb-pagination | **Shipped** — pilot `/revues` |
| P0-8 | https://github.com/jeb-maker/miniature-broccoli/issues/14 | mb-toast | Vendored in boot — UI toast run pas encore branchée |
| P0-9 | https://github.com/jeb-maker/miniature-broccoli/issues/15 | File + number inputs | Vendored `mb-input` — pas de pilot upload |
| P0-10 | https://github.com/jeb-maker/miniature-broccoli/issues/16 | mb-radio-group | Vendored — pas de pilot |
| P1-19 | https://github.com/jeb-maker/miniature-broccoli/issues/17 | Docs Go/HTMX | Suivi (`docs/go-htmx.md` upstream) |
| P1-20 | https://github.com/jeb-maker/miniature-broccoli/issues/18 | Tokens coexistence | `tokens-core` + `mb-bridge.css` en place |
| Tracking | https://github.com/jeb-maker/miniature-broccoli/issues/19 | Tracking Revues consumer gaps | — |

Deferred P1 (#11–#18 inventaire): not opened yet — listed in the tracking issue.
