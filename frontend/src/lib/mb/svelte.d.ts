/**
 * Vendored from `@jeb-maker/mb` v0.5.1 (`svelte.d.ts`).
 * Svelte intrinsic element typings for `@jeb-maker/mb` custom elements.
 * Augments `svelte/elements` (`SvelteHTMLElements`) — no hard Svelte dependency.
 *
 * Reference via:
 *   import '@jeb-maker/mb/svelte'
 * or:
 *   /// <reference types="@jeb-maker/mb/svelte" />
 *
 * Keep attribute shapes in sync with `jsx.d.ts`. Custom `mb-*` events are typed
 * as `onmb-*` handlers (Svelte listens with `onmb-change={…}` / `on:mb-change`).
 */

type MbBaseAttrs = {
  children?: unknown;
  class?: string;
  style?: unknown;
  slot?: string;
  id?: string;
  title?: string;
  role?: string;
  'aria-label'?: string;
  'aria-current'?: string;
  // Revues host: native composed events (shadow DOM) — not in upstream svelte.d.ts yet.
  oninput?: (e: Event) => void;
  onchange?: (e: Event) => void;
  onclick?: (e: MouseEvent) => void;
  [key: string]: unknown;
};

type MbChangeDetail = {
  value: string;
  checked?: boolean;
  files?: FileList | null;
};

type MbSelectDetail = {
  value: string;
  label: string;
  href?: string;
};

type MbSortDetail = {
  key: string;
  direction: 'asc' | 'desc';
};

type MbToggleDetail = {
  expanded: boolean;
};

type MbSectionToggleDetail = {
  id: string;
  collapsed: boolean;
};

type MbReorderDetail = {
  rowId: string;
  fromSection: string;
  toSection: string;
  beforeId: string | null;
  afterId: string | null;
  order: Array<{ id: string; section: string }>;
};

type MbChangeHandler = { 'onmb-change'?: (e: CustomEvent<MbChangeDetail>) => void };
type MbInputHandler = { 'onmb-input'?: (e: CustomEvent<MbChangeDetail>) => void };
type MbSelectHandler = { 'onmb-select'?: (e: CustomEvent<MbSelectDetail>) => void };
type MbCloseHandler = { 'onmb-close'?: (e: CustomEvent<void>) => void };
type MbToggleHandler = { 'onmb-toggle'?: (e: CustomEvent<MbToggleDetail>) => void };
type MbSortHandler = { 'onmb-sort'?: (e: CustomEvent<MbSortDetail>) => void };
type MbSectionToggleHandler = {
  'onmb-section-toggle'?: (e: CustomEvent<MbSectionToggleDetail>) => void;
};
type MbReorderHandler = { 'onmb-reorder'?: (e: CustomEvent<MbReorderDetail>) => void };

declare module 'svelte/elements' {
  export interface SvelteHTMLElements {
    'mb-button': MbBaseAttrs & {
      variant?: 'primary' | 'secondary' | 'ghost' | 'danger';
      size?: 'sm' | 'md' | 'lg';
      type?: 'button' | 'submit' | 'reset';
      disabled?: boolean;
      loading?: boolean;
      name?: string;
      value?: string;
      href?: string;
      target?: string;
      rel?: string;
      'icon-only'?: boolean;
    };
    'mb-input': MbBaseAttrs &
      MbChangeHandler &
      MbInputHandler & {
        label?: string;
        hint?: string;
        error?: string;
        type?: 'text' | 'email' | 'password' | 'search' | 'url' | 'tel' | 'number' | 'file' | 'date';
        value?: string;
        name?: string;
        placeholder?: string;
        disabled?: boolean;
        required?: boolean;
        invalid?: boolean;
        density?: 'default' | 'compact';
        'hide-label'?: boolean;
        min?: string | number;
        max?: string | number;
        step?: string | number;
        accept?: string;
        multiple?: boolean;
        pattern?: string;
        maxlength?: number;
        minlength?: number;
        autocomplete?: string;
        readonly?: boolean;
        'missing-message'?: string;
        'invalid-message'?: string;
      };
    'mb-textarea': MbBaseAttrs &
      MbChangeHandler &
      MbInputHandler & {
        label?: string;
        hint?: string;
        error?: string;
        value?: string;
        name?: string;
        placeholder?: string;
        disabled?: boolean;
        required?: boolean;
        invalid?: boolean;
        rows?: number;
        density?: 'default' | 'compact';
        'hide-label'?: boolean;
        maxlength?: number;
        minlength?: number;
        autocomplete?: string;
        readonly?: boolean;
        'missing-message'?: string;
        'invalid-message'?: string;
      };
    'mb-select': MbBaseAttrs &
      MbChangeHandler & {
        label?: string;
        hint?: string;
        error?: string;
        value?: string;
        name?: string;
        disabled?: boolean;
        required?: boolean;
        invalid?: boolean;
        density?: 'default' | 'compact';
        'hide-label'?: boolean;
        placeholder?: string;
        options?: Array<{ value: string; label: string; disabled?: boolean }> | string;
        'missing-message'?: string;
        'invalid-message'?: string;
      };
    'mb-combobox': MbBaseAttrs &
      MbChangeHandler &
      MbInputHandler &
      MbSelectHandler & {
        label?: string;
        hint?: string;
        error?: string;
        value?: string;
        name?: string;
        placeholder?: string;
        type?: 'text' | 'search';
        disabled?: boolean;
        required?: boolean;
        invalid?: boolean;
        density?: 'default' | 'compact';
        'hide-label'?: boolean;
        open?: boolean;
        loading?: boolean;
        options?:
          | Array<{
              value: string;
              label: string;
              disabled?: boolean;
              group?: string;
              href?: string;
            }>
          | string;
        'empty-message'?: string;
        'loading-message'?: string;
        'close-on-select'?: boolean;
        'close-on-blur'?: boolean;
        'missing-message'?: string;
        'invalid-message'?: string;
      };
    'mb-checkbox': MbBaseAttrs &
      MbChangeHandler & {
        label?: string;
        error?: string;
        name?: string;
        value?: string;
        checked?: boolean;
        indeterminate?: boolean;
        disabled?: boolean;
        required?: boolean;
        invalid?: boolean;
        'missing-message'?: string;
      };
    'mb-radio': MbBaseAttrs & {
      label?: string;
      value?: string;
      name?: string;
      checked?: boolean;
      disabled?: boolean;
    };
    'mb-radio-group': MbBaseAttrs &
      MbChangeHandler & {
        label?: string;
        error?: string;
        value?: string;
        name?: string;
        disabled?: boolean;
        required?: boolean;
        invalid?: boolean;
        options?: Array<{ value: string; label: string; disabled?: boolean }> | string;
        'missing-message'?: string;
        'invalid-message'?: string;
      };
    'mb-badge': MbBaseAttrs & {
      variant?: 'neutral' | 'success' | 'warning' | 'danger' | 'info';
      href?: string;
      size?: 'sm' | 'md';
    };
    'mb-alert': MbBaseAttrs & {
      variant?: 'info' | 'success' | 'warning' | 'danger';
    };
    'mb-card': MbBaseAttrs;
    'mb-modal': MbBaseAttrs &
      MbCloseHandler & {
        open?: boolean;
        heading?: string;
        'close-label'?: string;
      };
    'mb-progress': MbBaseAttrs & {
      value?: number;
      max?: number;
      percent?: number;
      label?: string;
      'fallback-label'?: string;
    };
    'mb-segmented-control': MbBaseAttrs & {
      label?: string;
    };
    'mb-empty-state': MbBaseAttrs & {
      heading?: string;
    };
    'mb-pagination': MbBaseAttrs & {
      'prev-url'?: string;
      'next-url'?: string;
      'prev-disabled'?: boolean;
      'next-disabled'?: boolean;
      status?: string;
      'prev-label'?: string;
      'next-label'?: string;
      label?: string;
    };
    'mb-toast': MbBaseAttrs &
      MbCloseHandler & {
        open?: boolean;
        variant?: 'success' | 'danger' | 'info';
        'auto-dismiss'?: number;
        message?: string;
        'dismiss-label'?: string;
      };
    /** @deprecated Prefer `mb-badge` (same chip; tag defaults to size md). */
    'mb-tag': MbBaseAttrs & {
      variant?: 'neutral' | 'success' | 'warning' | 'danger' | 'info';
      href?: string;
      size?: 'sm' | 'md';
    };
    'mb-breadcrumbs': MbBaseAttrs & {
      label?: string;
      items?: Array<{ href?: string; label: string; current?: boolean }> | string;
    };
    'mb-nav': MbBaseAttrs & {
      label?: string;
      open?: boolean;
    };
    'mb-nav-toggle': MbBaseAttrs &
      MbToggleHandler & {
        expanded?: boolean;
        for?: string;
        'label-open'?: string;
        'label-close'?: string;
      };
    'mb-avatar': MbBaseAttrs & {
      src?: string;
      alt?: string;
      name?: string;
      size?: 'sm' | 'md';
      'fallback-label'?: string;
    };
    'mb-spinner': MbBaseAttrs & {
      size?: 'sm' | 'md';
      label?: string;
    };
    'mb-toolbar': MbBaseAttrs;
    'mb-table': MbBaseAttrs &
      MbSortHandler &
      MbSectionToggleHandler &
      MbReorderHandler & {
        label?: string;
        columns?: string;
        density?: 'default' | 'compact';
        layout?: 'auto' | 'table' | 'cards';
        sections?:
          | Array<{
              id: string;
              label: string;
              collapsed?: boolean;
              meta?: string;
              count?: boolean;
            }>
          | string;
        'sort-key'?: string;
        'sort-direction'?: 'asc' | 'desc';
        reorderable?: boolean;
        'reorder-label'?: string;
        'sort-label'?: string;
        'hide-count'?: boolean;
        'sticky-header'?: boolean;
      };
    'mb-table-row': MbBaseAttrs & {
      head?: boolean;
      section?: string;
      'sort-value'?: string;
    };
    'mb-table-cell': MbBaseAttrs & {
      label?: string;
      align?: 'start' | 'center' | 'end';
      primary?: boolean;
      'hide-label'?: boolean;
      actions?: boolean;
      'sort-key'?: string;
      sortable?: boolean;
      'sort-value'?: string;
    };
  }
}

export {};
